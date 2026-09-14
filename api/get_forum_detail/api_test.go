package getforumdetail

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

	pb "github.com/rongyuio/aiotieba/api/get_forum_detail/protobuf"
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
	got := PackProto(12345)
	if want := loadHex(t, "req.hex"); !bytes.Equal(got, want) {
		t.Errorf("PackProto = %x\n         want %x", got, want)
	}
}

func TestParseBodyMatchesPython(t *testing.T) {
	detail, err := ParseBody(loadHex(t, "response.hex"))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if detail.FID != 12345 || detail.FName != "天堂鸡汤" {
		t.Errorf("forum identity = %+v", detail)
	}
	if detail.Category != "生活" {
		t.Errorf("category = %q, want 生活", detail.Category)
	}
	if detail.SmallAvatar != "https://imgsrc.baidu.com/forum/avatar/abc.jpg" {
		t.Errorf("small avatar = %q", detail.SmallAvatar)
	}
	if detail.OriginAvatar != "https://imgsrc.baidu.com/forum/avatar/origin.jpg" {
		t.Errorf("origin avatar = %q", detail.OriginAvatar)
	}
	if detail.Slogan != "吧标语" {
		t.Errorf("slogan = %q", detail.Slogan)
	}
	if detail.MemberNum != 100 {
		t.Errorf("member num = %d, want 100", detail.MemberNum)
	}
	// post_num is derived from thread_count.
	if detail.PostNum != 300 {
		t.Errorf("post num = %d, want 300 (thread_count)", detail.PostNum)
	}
	if !detail.HasBawu {
		t.Error("HasBawu = false, want true (new_strategy_text is 已有吧主)")
	}
}

func TestRequestURL(t *testing.T) {
	u := RequestURL()
	if u.Scheme != "http" || u.Host != "tiebac.baidu.com" || u.Path != "/c/f/forum/getforumdetail" {
		t.Errorf("url = %s", u)
	}
	if u.RawQuery != "cmd=303021" {
		t.Errorf("query = %q, want cmd=303021", u.RawQuery)
	}
}

func TestParseBodySurfacesServerError(t *testing.T) {
	res := &pb.GetForumDetailResIdl{Error: &commonpb.Error{Errorno: 340006, Errmsg: "boom"}}
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
