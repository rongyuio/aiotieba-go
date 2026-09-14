package getblacklistold

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
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/get_blacklist_old/protobuf"
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
	got := PackProto(newAccount(t), 1, 10)
	if want := loadHex(t, "req.hex"); !bytes.Equal(got, want) {
		t.Errorf("PackProto = %x\n         want %x", got, want)
	}
}

func TestParseBodyMatchesPython(t *testing.T) {
	users, err := ParseBody(loadHex(t, "response.hex"))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if users.Len() != 1 {
		t.Fatalf("len(users) = %d, want 1", users.Len())
	}

	user := users.Objs[0]
	if user.UserID != 4444444 || user.UserName != "某个用户名" {
		t.Errorf("user = %+v", user)
	}
	// Python does `portrait[:-13]` when the portrait carries a "?" suffix.
	if user.Portrait != "tb.1.abcdefghijklmnopqrst" {
		t.Errorf("portrait = %q", user.Portrait)
	}
	if user.NickNameOld != "某个旧昵称" || user.NickName() != "某个旧昵称" {
		t.Errorf("nick name = %+v", user)
	}
	if user.UntilTime != 1700000000 {
		t.Errorf("until time = %d", user.UntilTime)
	}
	if got := user.String(); got != "某个用户名" {
		t.Errorf("String() = %q", got)
	}
	if got := user.LogName(); got != "某个用户名" {
		t.Errorf("LogName() = %q", got)
	}
	if !user.Valid() {
		t.Error("Valid() = false, want true")
	}

	page := users.Page
	if page.CurrentPage != 1 || !page.HasMore || page.HasPrev {
		t.Errorf("page = %+v", page)
	}
	if !users.HasMore() {
		t.Error("HasMore() = false, want true")
	}
}

func TestParseBodySurfacesServerError(t *testing.T) {
	res := &pb.UserMuteQueryResIdl{Error: &commonpb.Error{Errorno: 340006, Errmsg: "boom"}}
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
	if u.Scheme != "https" || u.Host != "tiebac.baidu.com" || u.Path != "/c/u/user/userMuteQuery" {
		t.Errorf("url = %s", u)
	}
	if u.RawQuery != "cmd=303028" {
		t.Errorf("query = %q, want cmd=303028", u.RawQuery)
	}
}
