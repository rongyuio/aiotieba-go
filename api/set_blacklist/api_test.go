package setblacklist

import (
	"bytes"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/set_blacklist/protobuf"
	commonpb "github.com/rongyuio/aiotieba-go/protobuf"
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

func newAccount(t *testing.T) *core.Account {
	t.Helper()
	account, err := core.NewAccount(strings.Repeat("b", 192), "")
	if err != nil {
		t.Fatalf("NewAccount: %v", err)
	}
	return account
}

func TestPackProtoMatchesPython(t *testing.T) {
	got := PackProto(newAccount(t), 4444444, enums.BlacklistAll)
	if want := loadHex(t, "req.hex"); !bytes.Equal(got, want) {
		t.Errorf("PackProto = %x\n         want %x", got, want)
	}
}

func TestPackProtoPartialPermissions(t *testing.T) {
	// Only follow is denied; the others stay allowed (2).
	req := &pb.SetUserBlackReqIdl{}
	if err := proto.Unmarshal(PackProto(newAccount(t), 1, enums.BlacklistFollow), req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	perm := req.GetData().GetPermList()
	if perm.GetFollow() != permissionDeny || perm.GetInteract() != permissionAllow || perm.GetChat() != permissionAllow {
		t.Errorf("perm list = %+v", perm)
	}
}

func TestParseBodyMatchesPython(t *testing.T) {
	if err := ParseBody(loadHex(t, "response.hex")); err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
}

func TestParseBodySurfacesServerError(t *testing.T) {
	res := &pb.SetUserBlackResIdl{Error: &commonpb.Error{Errorno: 340006, Errmsg: "boom"}}
	raw, err := proto.Marshal(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	err = ParseBody(raw)
	var serverErr *exception.TiebaServerError
	if !errors.As(err, &serverErr) {
		t.Fatalf("error = %v (%T), want *exception.TiebaServerError", err, err)
	}
	if serverErr.Code != 340006 || serverErr.Msg != "boom" {
		t.Errorf("server error = %+v", serverErr)
	}
}

func TestRequestURL(t *testing.T) {
	u := RequestURL()
	if u.Scheme != "https" || u.Host != "tiebac.baidu.com" || u.Path != "/c/c/user/setUserBlack" {
		t.Errorf("url = %s", u)
	}
	if u.RawQuery != "cmd=309697" {
		t.Errorf("query = %q, want cmd=309697", u.RawQuery)
	}
}
