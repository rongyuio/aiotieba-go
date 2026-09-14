package getsquareforums

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

	pb "github.com/rongyuio/aiotieba-go/api/get_square_forums/protobuf"
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
	got := PackProto(newAccount(t), "生活", 2, 20)
	if want := loadHex(t, "req.hex"); !bytes.Equal(got, want) {
		t.Errorf("PackProto = %x\n         want %x", got, want)
	}
}

func TestParseBodyMatchesPython(t *testing.T) {
	forums, err := ParseBody(loadHex(t, "response.hex"))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if forums.Len() != 2 {
		t.Fatalf("len(forums) = %d, want 2", forums.Len())
	}

	first := forums.Objs[0]
	if first.FID != 12345 || first.FName != "天堂鸡汤" {
		t.Errorf("forum identity = %+v", first)
	}
	if first.MemberNum != 100 {
		t.Errorf("member num = %d, want 100", first.MemberNum)
	}
	// post_num is derived from thread_count.
	if first.PostNum != 300 {
		t.Errorf("post num = %d, want 300 (thread_count)", first.PostNum)
	}
	if !first.IsFollowed {
		t.Error("IsFollowed = false, want true (is_like is 1)")
	}

	second := forums.Objs[1]
	if second.IsFollowed {
		t.Error("IsFollowed = true, want false (is_like is 0)")
	}

	page := forums.Page
	if page.PageSize != 20 || page.CurrentPage != 2 || page.TotalPage != 5 || page.TotalCount != 88 {
		t.Errorf("page = %+v", page)
	}
	if !page.HasMore || !page.HasPrev {
		t.Errorf("page flags = %+v", page)
	}
	if !forums.HasMore() {
		t.Error("HasMore() = false, want true")
	}
}

func TestParseBodySurfacesServerError(t *testing.T) {
	res := &pb.GetForumSquareResIdl{Error: &commonpb.Error{Errorno: 340006, Errmsg: "boom"}}
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
	if u.Scheme != "https" || u.Host != "tiebac.baidu.com" || u.Path != "/c/f/forum/getForumSquare" {
		t.Errorf("url = %s", u)
	}
	if u.RawQuery != "cmd=309653" {
		t.Errorf("query = %q, want cmd=309653", u.RawQuery)
	}
}
