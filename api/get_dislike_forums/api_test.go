package getdislikeforums

import (
	"bytes"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba/core"
	"github.com/rongyuio/aiotieba/exception"

	pb "github.com/rongyuio/aiotieba/api/get_dislike_forums/protobuf"
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
	account, err := core.NewAccount(strings.Repeat("b", 192), "")
	if err != nil {
		t.Fatalf("NewAccount: %v", err)
	}

	got := PackProto(account, 1, 20)
	if want := loadHex(t, "req.hex"); !bytes.Equal(got, want) {
		t.Errorf("PackProto = %x\n         want %x", got, want)
	}
}

func TestParseBodyMatchesPython(t *testing.T) {
	forums, err := ParseBody(loadHex(t, "response.hex"))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if forums.Len() != 1 {
		t.Fatalf("len(forums) = %d, want 1", forums.Len())
	}

	forum := forums.Objs[0]
	if forum.FID != 111 || forum.FName != "屏蔽吧" {
		t.Errorf("forum identity = %+v", forum)
	}
	if forum.MemberNum != 12 || forum.PostNum != 34 || forum.ThreadNum != 56 {
		t.Errorf("forum counters = %+v", forum)
	}
	// is_followed is not reported by this endpoint.
	if forum.IsFollowed {
		t.Error("IsFollowed = true, want false")
	}

	// has_prev is derived from the page number (3 > 1).
	page := forums.Page
	if page.CurrentPage != 3 {
		t.Errorf("current page = %d, want 3", page.CurrentPage)
	}
	if !page.HasMore {
		t.Error("HasMore = false, want true")
	}
	if !page.HasPrev {
		t.Error("HasPrev = false, want true (page 3 > 1)")
	}
	if !forums.HasMore() {
		t.Error("HasMore() = false, want true")
	}
}

func TestPageDislikeFFirstPage(t *testing.T) {
	// On page 1 there is never a previous page.
	page := PageDislikeFFromProto(&pb.GetDislikeListResIdl_DataRes{CurPage: 1, HasMore: 0})
	if page.HasPrev {
		t.Error("HasPrev = true, want false on page 1")
	}
	if page.HasMore {
		t.Error("HasMore = true, want false")
	}
}

func TestParseBodySurfacesServerError(t *testing.T) {
	res := &pb.GetDislikeListResIdl{Error: &commonpb.Error{Errorno: 340006, Errmsg: "boom"}}
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
	if u.Scheme != "https" || u.Host != "tiebac.baidu.com" || u.Path != "/c/u/user/getDislikeList" {
		t.Errorf("url = %s", u)
	}
	if u.RawQuery != "cmd=309692" {
		t.Errorf("query = %q, want cmd=309692", u.RawQuery)
	}
}
