package getcomments

import (
	"bytes"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba/enums"
	"github.com/rongyuio/aiotieba/exception"

	pb "github.com/rongyuio/aiotieba/api/get_comments/protobuf"
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
	t.Run("pid", func(t *testing.T) {
		got := PackProto(123456, 222, 1, false)
		if want := loadHex(t, "req_pid.hex"); !bytes.Equal(got, want) {
			t.Errorf("PackProto = %x\n         want %x", got, want)
		}
	})

	t.Run("spid", func(t *testing.T) {
		got := PackProto(123456, 333, 4, true)
		if want := loadHex(t, "req_spid.hex"); !bytes.Equal(got, want) {
			t.Errorf("PackProto = %x\n         want %x", got, want)
		}
	})
}

func TestRequestURL(t *testing.T) {
	u := RequestURL()
	if u.Scheme != "http" || u.Host != "tiebac.baidu.com" || u.Path != "/c/f/pb/floor" {
		t.Errorf("url = %s", u)
	}
	if u.RawQuery != "cmd=302002" {
		t.Errorf("query = %q, want cmd=302002", u.RawQuery)
	}
}

func TestParseBodyMatchesPython(t *testing.T) {
	comments, err := ParseBody(loadHex(t, "response.hex"))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}

	// Page: has_more/has_prev are derived from the page numbers.
	if comments.Page.CurrentPage != 2 || comments.Page.TotalPage != 5 || comments.Page.TotalCount != 150 {
		t.Errorf("page = %+v", comments.Page)
	}
	if !comments.Page.HasMore || !comments.Page.HasPrev {
		t.Errorf("page flags = %+v", comments.Page)
	}
	if !comments.HasMore() {
		t.Error("HasMore() = false, want true")
	}

	// Forum
	if comments.Forum.FID != 12 || comments.Forum.FName != "天堂鸡汤" {
		t.Errorf("forum = %+v", comments.Forum)
	}
	if comments.Forum.Category != "生活" || comments.Forum.Subcategory != "情感" {
		t.Errorf("forum categories = %+v", comments.Forum)
	}

	// Thread
	th := comments.Thread
	if th.TID != 111 || th.Title != "帖子标题" || th.ReplyNum != 9 {
		t.Errorf("thread = %+v", th)
	}
	if th.FID != 12 || th.FName != "天堂鸡汤" || th.AuthorID() != 999 {
		t.Errorf("thread context = %+v", th)
	}
	if th.Type != enums.ThreadTypeArticle {
		t.Errorf("thread type = %d", th.Type)
	}

	// Post
	post := comments.Post
	if post.PID != 222 || post.Floor != 3 || post.AuthorID() != 888 {
		t.Errorf("post = %+v", post)
	}
	if post.FID != 12 || post.FName != "天堂鸡汤" || post.TID != 111 {
		t.Errorf("post context = %+v", post)
	}
	if post.Sign != "尾巴" {
		t.Errorf("post sign = %q", post.Sign)
	}
	if got := post.Text(); got != "楼层内容\n尾巴" {
		t.Errorf("post text = %q", got)
	}
	if post.User.Level != 6 || post.User.Gender != enums.GenderMale {
		t.Errorf("post user = %+v", post.User)
	}

	// Comments
	if comments.Len() != 2 {
		t.Fatalf("comment count = %d, want 2", comments.Len())
	}

	c0 := comments.Objs[0]
	if c0.PID != 2222 || c0.AuthorID() != 777 || c0.ReplyToID != 999 {
		t.Errorf("comment 0 identity = %+v", c0)
	}
	if c0.FID != 12 || c0.FName != "天堂鸡汤" || c0.TID != 111 || c0.PPID != 222 || c0.Floor != 3 {
		t.Errorf("comment 0 context = %+v", c0)
	}
	if c0.Agree != 3 || c0.Disagree != 1 || c0.CreateTime != 1700000200 {
		t.Errorf("comment 0 counters = %+v", c0)
	}
	if c0.IsThreadAuthor {
		t.Error("comment 0 IsThreadAuthor = true, want false")
	}
	if got := c0.Text(); got != "好的" {
		t.Errorf("comment 0 text = %q, want 好的", got)
	}
	if c0.User.UserName != "回复者" {
		t.Errorf("comment 0 user = %+v", c0.User)
	}
	if len(c0.Contents.Ats) != 0 {
		t.Errorf("comment 0 ats = %+v, want none", c0.Contents.Ats)
	}

	c1 := comments.Objs[1]
	if c1.PID != 3333 || c1.AuthorID() != 999 || c1.ReplyToID != 0 {
		t.Errorf("comment 1 identity = %+v", c1)
	}
	if !c1.IsThreadAuthor {
		t.Error("comment 1 IsThreadAuthor = false, want true")
	}
	if got := c1.Text(); got != "第二条" {
		t.Errorf("comment 1 text = %q", got)
	}
	if c1.Floor != 3 || c1.PPID != 222 {
		t.Errorf("comment 1 context = %+v", c1)
	}
}

func TestParseBodySurfacesServerError(t *testing.T) {
	res := &pb.PbFloorResIdl{Error: &commonpb.Error{Errorno: 340006, Errmsg: "boom"}}
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
