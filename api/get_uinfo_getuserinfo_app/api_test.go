package getuserinfoapp

import (
	"bytes"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/get_uinfo_getuserinfo_app/protobuf"
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

func TestPackProtoMatchesPython(t *testing.T) {
	got := PackProto(4444444)
	if want := loadHex(t, "req.hex"); !bytes.Equal(got, want) {
		t.Errorf("PackProto = %x\n         want %x", got, want)
	}
}

func TestParseBodyMatchesPython(t *testing.T) {
	user, err := ParseBody(loadHex(t, "response.hex"))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if user.UserID != 4444444 {
		t.Errorf("user id = %d, want 4444444", user.UserID)
	}
	// Python does `portrait[:-13]` when the portrait carries a "?" suffix.
	if user.Portrait != "tb.1.abcdefghijklmnopqrst" {
		t.Errorf("portrait = %q, want tb.1.abcdefghijklmnopqrst", user.Portrait)
	}
	if user.UserName != "某个用户名" {
		t.Errorf("user name = %q", user.UserName)
	}
	if user.NickNameOld != "某个昵称" || user.NickName() != "某个昵称" {
		t.Errorf("nick name = %q", user.NickNameOld)
	}
	if user.Gender != enums.GenderFemale {
		t.Errorf("gender = %v, want GenderFemale", user.Gender)
	}
	if !user.IsVIP {
		t.Error("IsVIP = false, want true (v_status is 3)")
	}
	if !user.IsGod {
		t.Error("IsGod = false, want true (new_god_data.status is 1)")
	}
	if got := user.String(); got != "某个用户名" {
		t.Errorf("String() = %q", got)
	}
	if got := user.LogName(); got != "某个用户名" {
		t.Errorf("LogName() = %q", got)
	}
}

func TestFromProtoWithoutSuffix(t *testing.T) {
	// A portrait without "?" must be left untouched.
	user := UserInfoGuinfoAppFromProto(&commonpb.User{Id: 7, Portrait: "tb.1.short"})
	if user.Portrait != "tb.1.short" {
		t.Errorf("portrait = %q, want tb.1.short", user.Portrait)
	}
	if user.Gender != enums.GenderUnknown {
		t.Errorf("gender = %v, want GenderUnknown", user.Gender)
	}
	if user.Valid() {
		// UserID is 7, so Valid() must be true.
		return
	}
	t.Error("Valid() = false, want true")
}

func TestParseBodySurfacesServerError(t *testing.T) {
	res := &pb.GetUserInfoResIdl{Error: &commonpb.Error{Errorno: 340006, Errmsg: "boom"}}
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
	if u.Scheme != "http" || u.Host != "tiebac.baidu.com" || u.Path != "/c/u/user/getuserinfo" {
		t.Errorf("url = %s", u)
	}
	if u.RawQuery != "cmd=303024" {
		t.Errorf("query = %q, want cmd=303024", u.RawQuery)
	}
}
