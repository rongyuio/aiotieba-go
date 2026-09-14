package protobuf

import (
	"bytes"
	"encoding/hex"
	"testing"

	"google.golang.org/protobuf/proto"
)

// The wire bytes below were produced by the Python bindings generated from the
// same .proto files, so they validate that the Go bindings agree on field
// numbers, wire types and value encoding.

const commonWire = "0802120931322e36342e312e313a0a637569642d76616c75654080d095ffbc314a0b6d6f64656c2d76616c7565520b62647573732d76616c75655a097462732d76616c7565fa01097a69642d76616c756582022a41334544324437423943464332384538393334413346424433413935373943377c565A35464B423558539a020b4130302d5858582d595959a202083130343530355f33"

const errorWire = "08a6e014120d6572726f72206d657373616765"

func mustDecode(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("decoding hex: %v", err)
	}
	return b
}

func TestCommonReqMatchesPython(t *testing.T) {
	want := &CommonReq{
		XClientType:    2,
		XClientVersion: "12.64.1.1",
		Cuid:           "cuid-value",
		BDUSS:          "bduss-value",
		Tbs:            "tbs-value",
		CuidGalaxy2:    "A3ED2D7B9CFC28E8934A3FBD3A9579C7|VZ5FKB5XS",
		C3Aid:          "A00-XXX-YYY",
		SampleId:       "104505_3",
		ZId:            "zid-value",
		Model:          "model-value",
		XTimestamp:     1700000000000,
	}

	wire := mustDecode(t, commonWire)

	msg := &CommonReq{}
	if err := proto.Unmarshal(wire, msg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !proto.Equal(msg, want) {
		t.Errorf("decoded message mismatch:\n got %v\nwant %v", msg, want)
	}

	got, err := proto.Marshal(want)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !bytes.Equal(got, wire) {
		t.Errorf("re-encoded wire = %x, want %x", got, wire)
	}
}

func TestErrorMatchesPython(t *testing.T) {
	want := &Error{Errorno: 340006, Errmsg: "error message"}

	msg := &Error{}
	if err := proto.Unmarshal(mustDecode(t, errorWire), msg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !proto.Equal(msg, want) {
		t.Errorf("decoded message mismatch:\n got %v\nwant %v", msg, want)
	}
}
