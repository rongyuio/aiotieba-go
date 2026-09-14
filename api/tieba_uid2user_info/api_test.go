package tiebauid2userinfo

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

	pb "github.com/rongyuio/aiotieba/api/tieba_uid2user_info/protobuf"
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
	got := PackProto(987654321)
	if want := loadHex(t, "req.hex"); !bytes.Equal(got, want) {
		t.Errorf("PackProto = %x\n         want %x", got, want)
	}
}

func TestParseBodyMatchesPython(t *testing.T) {
	user, err := ParseBody(loadHex(t, "response.hex"))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if user.UserID != 4444444 || user.UserName != "某个用户名" || user.NickNameNew != "某个昵称" {
		t.Errorf("identity = %+v", user)
	}
	// Python does `portrait[:-13]` when the portrait carries a "?" suffix.
	if user.Portrait != "tb.1.abcdefghijklmnopqrst" {
		t.Errorf("portrait = %q", user.Portrait)
	}
	if user.TiebaUID != 987654321 {
		t.Errorf("tieba uid = %d, want 987654321", user.TiebaUID)
	}
	if user.Age != 8.5 {
		t.Errorf("age = %v, want 8.5", user.Age)
	}
	if user.Sign != "个性签名" {
		t.Errorf("sign = %q", user.Sign)
	}
	if !user.IsGod {
		t.Error("IsGod = false, want true")
	}
	if got := user.ShowName(); got != "某个昵称" {
		t.Errorf("ShowName() = %q", got)
	}
	if got := user.LogName(); got != "某个用户名" {
		t.Errorf("LogName() = %q", got)
	}
	if !user.Valid() {
		t.Error("Valid() = false, want true")
	}
}

func TestParseBodySurfacesServerError(t *testing.T) {
	res := &pb.GetUserByTiebaUidResIdl{Error: &commonpb.Error{Errorno: 340006, Errmsg: "boom"}}
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

func TestRequestURL(t *testing.T) {
	u := RequestURL()
	if u.Scheme != "http" || u.Host != "tiebac.baidu.com" || u.Path != "/c/u/user/getUserByTiebaUid" {
		t.Errorf("url = %s", u)
	}
	if u.RawQuery != "cmd=309702" {
		t.Errorf("query = %q, want cmd=309702", u.RawQuery)
	}
}
