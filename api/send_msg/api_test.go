package sendmsg

import (
	"bytes"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba/exception"

	pb "github.com/rongyuio/aiotieba/api/send_msg/protobuf"
	commonpb "github.com/rongyuio/aiotieba/protobuf"
)

// The golden fixtures under testdata/ were produced by the Python bindings
// generated from the same .proto files.

func loadHex(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("reading testdata/%s: %v", name, err)
	}
	b, err := hex.DecodeString(strings.TrimSpace(string(raw)))
	if err != nil {
		t.Fatalf("decoding testdata/%s: %v", name, err)
	}
	return b
}

func TestPackProtoMatchesPython(t *testing.T) {
	got := PackProto(4444444, "私信内容", 7001)
	if want := loadHex(t, "req.hex"); !bytes.Equal(got, want) {
		t.Errorf("PackProto = %x\n         want %x", got, want)
	}
}

func TestParseBodyMatchesPython(t *testing.T) {
	msgID, err := ParseBody(loadHex(t, "response.hex"))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if msgID != 88888 {
		t.Errorf("msg id = %d, want 88888", msgID)
	}
}

func TestParseBodyBlockInfo(t *testing.T) {
	// The endpoint reports a second failure envelope under data.blockInfo.
	res := &pb.CommitPersonalMsgResIdl{
		Data: &pb.CommitPersonalMsgResIdl_DataRes{
			BlockInfo: &pb.CommitPersonalMsgResIdl_DataRes_BlockInfo{
				BlockErrno:  340006,
				BlockErrmsg: "被屏蔽",
			},
		},
	}
	raw, err := proto.Marshal(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	_, err = ParseBody(raw)
	var serverErr *exception.TiebaServerError
	if !errors.As(err, &serverErr) {
		t.Fatalf("error = %v (%T), want *exception.TiebaServerError", err, err)
	}
	if serverErr.Code != 340006 || serverErr.Msg != "被屏蔽" {
		t.Errorf("server error = %+v", serverErr)
	}
}

func TestParseBodySurfacesServerError(t *testing.T) {
	res := &pb.CommitPersonalMsgResIdl{Error: &commonpb.Error{Errorno: 340006, Errmsg: "boom"}}
	raw, err := proto.Marshal(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	_, err = ParseBody(raw)
	var serverErr *exception.TiebaServerError
	if !errors.As(err, &serverErr) {
		t.Fatalf("error = %v (%T), want *exception.TiebaServerError", err, err)
	}
	if serverErr.Code != 340006 || serverErr.Msg != "boom" {
		t.Errorf("server error = %+v", serverErr)
	}
}

func TestCMD(t *testing.T) {
	if CMD != 205001 {
		t.Errorf("CMD = %d, want 205001", CMD)
	}
}
