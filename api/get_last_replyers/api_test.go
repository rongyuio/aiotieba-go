package getlastreplyers

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

	pb "github.com/rongyuio/aiotieba-go/api/get_last_replyers/protobuf"
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
	// pn 1 must be sent as 0 and rn_need must be rn + 5.
	got := PackProto("天堂鸡汤", 1, 20, enums.ThreadSortReply, false)
	if want := loadHex(t, "req.hex"); !bytes.Equal(got, want) {
		t.Errorf("PackProto = %x\n         want %x", got, want)
	}
}

func TestPackProtoPageNumber(t *testing.T) {
	// Only pn 1 is remapped; any other page travels as is.
	req := &pb.FrsPageReqIdl4Lp{}
	if err := proto.Unmarshal(PackProto("f", 3, 20, enums.ThreadSortReply, false), req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req.GetData().GetPn() != 3 {
		t.Errorf("pn = %d, want 3", req.GetData().GetPn())
	}
	if req.GetData().GetRnNeed() != 25 {
		t.Errorf("rn_need = %d, want 25", req.GetData().GetRnNeed())
	}
	if req.GetData().GetCommon().GetXClientVersion() != "6.0.1" {
		t.Errorf("client version = %q, want 6.0.1", req.GetData().GetCommon().GetXClientVersion())
	}
}

func TestParseBodyMatchesPython(t *testing.T) {
	threads, err := ParseBody(loadHex(t, "response.hex"))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}

	// The endpoint reports page 0 with a page size; Python normalises it to 1.
	if threads.Page.CurrentPage != 1 {
		t.Errorf("current page = %d, want 1", threads.Page.CurrentPage)
	}
	if threads.Page.PageSize != 20 || threads.Page.TotalPage != 3 || threads.Page.TotalCount != 55 {
		t.Errorf("page = %+v", threads.Page)
	}
	if !threads.Page.HasMore || threads.Page.HasPrev {
		t.Errorf("page flags = %+v", threads.Page)
	}
	if !threads.HasMore() {
		t.Error("HasMore() = false, want true")
	}

	if threads.Forum.FID != 12345 || threads.Forum.FName != "天堂鸡汤" {
		t.Errorf("forum = %+v", threads.Forum)
	}

	if threads.Len() != 1 {
		t.Fatalf("len(threads) = %d, want 1", threads.Len())
	}
	thread := threads.Objs[0]
	if thread.Title != "标题" || thread.TID != 111 || thread.PID != 222 {
		t.Errorf("thread identity = %+v", thread)
	}
	// The forum is inherited from the list rather than carried by the thread.
	if thread.FID != 12345 || thread.FName != "天堂鸡汤" {
		t.Errorf("thread forum = %+v", thread)
	}
	if !thread.IsGood || thread.IsTop {
		t.Errorf("thread flags = %+v", thread)
	}
	if thread.CreateTime != 1700000000 || thread.LastTime != 1700000001 {
		t.Errorf("thread times = %+v", thread)
	}
	if got := thread.Text(); got != "标题" {
		t.Errorf("Text() = %q", got)
	}

	// The author portrait keeps its "?..." suffix stripping.
	author := thread.User
	if author.UserID != 4444444 || author.Portrait != "tb.1.abcdefghijklmnopqrst" {
		t.Errorf("author = %+v", author)
	}
	if author.UserName != "某个用户名" || author.NickNameOld != "某个旧昵称" {
		t.Errorf("author names = %+v", author)
	}
	if got := author.ShowName(); got != "某个旧昵称" {
		t.Errorf("author ShowName() = %q", got)
	}
	if got := author.LogName(); got != "某个用户名" {
		t.Errorf("author LogName() = %q", got)
	}
	if thread.AuthorID() != 4444444 {
		t.Errorf("AuthorID() = %d", thread.AuthorID())
	}

	// The last replier has no portrait.
	replyer := thread.LastReplyer
	if replyer.UserID != 5555555 || replyer.UserName != "最后回复者" || replyer.NickNameOld != "回复者昵称" {
		t.Errorf("last replyer = %+v", replyer)
	}
	if got := replyer.LogName(); got != "最后回复者" {
		t.Errorf("last replyer LogName() = %q", got)
	}
}

func TestPageLPZeroPage(t *testing.T) {
	// Without a page size the page number is not normalised.
	if got := PageLPFromProto(&commonpb.Page{CurrentPage: 0}); got.CurrentPage != 0 {
		t.Errorf("current page = %d, want 0", got.CurrentPage)
	}
}

func TestParseBodySurfacesServerError(t *testing.T) {
	res := &pb.FrsPageResIdl4Lp{Error: &commonpb.Error{Errorno: 340006, Errmsg: "boom"}}
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
	if u.Scheme != "http" || u.Host != "tiebac.baidu.com" || u.Path != "/c/f/frs/page" {
		t.Errorf("url = %s", u)
	}
	if u.RawQuery != "cmd=301001" {
		t.Errorf("query = %q, want cmd=301001", u.RawQuery)
	}
}
