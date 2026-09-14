package core

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba/protobuf"
)

func makeBLCPFrame(rpcBody, lcmBody []byte) []byte {
	out := make([]byte, 0, 8+len(rpcBody)+len(lcmBody))
	out = append(out, blcpMagic...)
	out = binary.BigEndian.AppendUint32(out, uint32(len(rpcBody)+len(lcmBody)))
	out = binary.BigEndian.AppendUint32(out, uint32(len(rpcBody)))
	out = append(out, rpcBody...)
	out = append(out, lcmBody...)
	return out
}

func TestBLCPDataBytes(t *testing.T) {
	d := &BLCPData{ServiceID: 3, MethodID: 201, RPCBody: []byte{0x01, 0x02}, LCMBody: []byte{0x03}}
	got := d.Bytes()

	want := []byte{'l', 'c', 'p', 0x01, 0, 0, 0, 3, 0, 0, 0, 2, 0x01, 0x02, 0x03}
	if !bytes.Equal(got, want) {
		t.Errorf("Bytes() = %x, want %x", got, want)
	}
}

func TestParseBLCPResponseProtobufBody(t *testing.T) {
	rpcBody := marshal(&protobuf.RpcMeta{
		Response:      &protobuf.RpcResponseMeta{ServiceId: 5, MethodId: 6, ErrorText: "success"},
		CorrelationId: 42,
	})
	lcmBody := marshal(&protobuf.RpcData{
		LcmResponse: &protobuf.LcmResponse{ErrorMsg: "success"},
	})

	rpc, body, err := ParseBLCPResponse(makeBLCPFrame(rpcBody, lcmBody))
	if err != nil {
		t.Fatalf("ParseBLCPResponse: %v", err)
	}
	if got := rpc.GetResponse().GetServiceId(); got != 5 {
		t.Errorf("service id = %d, want 5", got)
	}
	if got := rpc.GetResponse().GetMethodId(); got != 6 {
		t.Errorf("method id = %d, want 6", got)
	}
	if got := rpc.GetCorrelationId(); got != 42 {
		t.Errorf("correlation id = %d, want 42", got)
	}
	if body.RPC == nil {
		t.Fatal("body was not decoded as protobuf")
	}
	if got := body.ErrorMsg(); got != "success" {
		t.Errorf("error msg = %q, want success", got)
	}
}

func TestParseBLCPResponseJSONBody(t *testing.T) {
	rpcBody := marshal(&protobuf.RpcMeta{Response: &protobuf.RpcResponseMeta{ErrorText: "success"}})
	lcmBody := []byte(`{"err_code":0,"uk":"abc"}`)

	_, body, err := ParseBLCPResponse(makeBLCPFrame(rpcBody, lcmBody))
	if err != nil {
		t.Fatalf("ParseBLCPResponse: %v", err)
	}
	if body.RPC != nil {
		t.Fatal("body was decoded as protobuf, want JSON")
	}
	if v, ok := body.JSONValue("err_code"); !ok || v != "0" {
		t.Errorf("err_code = %q (present=%v), want 0", v, ok)
	}
	if v, ok := body.JSONValue("uk"); !ok || v != "abc" {
		t.Errorf("uk = %q (present=%v), want abc", v, ok)
	}
	if _, ok := body.JSONValue("missing"); ok {
		t.Error("JSONValue reported a missing key as present")
	}
}

func TestParseBLCPResponseGzippedJSONBody(t *testing.T) {
	rpcBody := marshal(&protobuf.RpcMeta{Response: &protobuf.RpcResponseMeta{ErrorText: "success"}})
	raw, err := gzipCompress([]byte(`{"errno":"0"}`))
	if err != nil {
		t.Fatalf("gzipCompress: %v", err)
	}

	_, body, err := ParseBLCPResponse(makeBLCPFrame(rpcBody, raw))
	if err != nil {
		t.Fatalf("ParseBLCPResponse: %v", err)
	}
	if v, ok := body.JSONValue("errno"); !ok || v != "0" {
		t.Errorf("errno = %q (present=%v), want 0", v, ok)
	}
}

func TestParseBLCPResponseRejectsBadFrames(t *testing.T) {
	if _, _, err := ParseBLCPResponse([]byte("nope")); err == nil {
		t.Error("ParseBLCPResponse on a non-BLCP frame: want error, got nil")
	}
	if _, _, err := ParseBLCPResponse(append([]byte(blcpMagic), 0, 0)); err == nil {
		t.Error("ParseBLCPResponse on a truncated frame: want error, got nil")
	}
}

func TestClientBLCPResponsesNext(t *testing.T) {
	rpcBody := marshal(&protobuf.RpcMeta{
		Response:      &protobuf.RpcResponseMeta{ServiceId: 3, MethodId: 201, ErrorText: "success"},
		CorrelationId: 77,
	})
	lcmBody := marshal(&protobuf.RpcData{LcmResponse: &protobuf.LcmResponse{ErrorMsg: "success"}})
	frame := makeBLCPFrame(rpcBody, lcmBody)

	resp := &ClientBLCPResponses{reader: bufio.NewReader(bytes.NewReader(frame))}
	got, body, err := resp.Next()
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if got.CorrelationID != 77 {
		t.Errorf("correlation id = %d, want 77", got.CorrelationID)
	}
	if got.ServiceID != 3 || got.MethodID != 201 {
		t.Errorf("service/method = %d/%d, want 3/201", got.ServiceID, got.MethodID)
	}
	if got.IfRequest {
		t.Error("IfRequest = true, want false")
	}
	if body.ErrorMsg() != "success" {
		t.Errorf("error msg = %q, want success", body.ErrorMsg())
	}
}

func TestBuildRPCBody(t *testing.T) {
	raw := BuildRPCBody(3, 201, 12345, 0, 1)

	meta := &protobuf.RpcMeta{}
	if err := proto.Unmarshal(raw, meta); err != nil {
		t.Fatalf("decoding RpcMeta: %v", err)
	}
	if meta.GetRequest().GetServiceId() != 3 || meta.GetRequest().GetMethodId() != 201 {
		t.Errorf("service/method = %d/%d, want 3/201", meta.GetRequest().GetServiceId(), meta.GetRequest().GetMethodId())
	}
	if meta.GetCorrelationId() != 12345 {
		t.Errorf("correlation id = %d, want 12345", meta.GetCorrelationId())
	}
	if meta.GetRequest().GetNeedCommon() != 1 {
		t.Errorf("need_common = %d, want 1", meta.GetRequest().GetNeedCommon())
	}
	if meta.GetAcceptCompressType() != 1 {
		t.Errorf("accept_compress_type = %d, want 1", meta.GetAcceptCompressType())
	}
	if meta.GetCompressType() != 0 {
		t.Errorf("compress_type = %d, want 0", meta.GetCompressType())
	}
	events := meta.GetRequest().GetEventList()
	if len(events) != 1 || events[0].GetEvent() != "CLCPReqBegin" {
		t.Errorf("event list = %v, want one CLCPReqBegin", events)
	}
}

func TestGetBDUKFromUserIDVector(t *testing.T) {
	// AES-CBC (key=b"AFD311832EDEEAEF", iv=b"2011121211143000") + base64url
	// without padding; the vector comes from the Python implementation.
	got, err := GetBDUKFromUserID("1234567890")
	if err != nil {
		t.Fatalf("GetBDUKFromUserID: %v", err)
	}
	const want = "J5lKUi5hb4wDg04lLRwjAg"
	if got != want {
		t.Errorf("bduk = %q, want %q", got, want)
	}
	if bytes.ContainsAny([]byte(got), "+/=") {
		t.Errorf("bduk %q is not base64url without padding", got)
	}
}

func TestGetMsgKey(t *testing.T) {
	key := GetMsgKey("bduk")
	if len(key) <= len("bduk") {
		t.Errorf("msg key = %q, want the bduk prefix plus a suffix", key)
	}
	if key[:4] != "bduk" {
		t.Errorf("msg key = %q, want it to start with the bduk", key)
	}
}
