package getreplys

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
	"github.com/rongyuio/aiotieba/enums"
	"github.com/rongyuio/aiotieba/exception"

	pb "github.com/rongyuio/aiotieba/api/get_replys/protobuf"
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

	got := PackProto(account, 2)
	if want := loadHex(t, "req.hex"); !bytes.Equal(got, want) {
		t.Errorf("PackProto = %x\n         want %x", got, want)
	}
}

func TestParseBodyMatchesPython(t *testing.T) {
	replys, err := ParseBody(loadHex(t, "response.hex"))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if replys.Len() != 1 {
		t.Fatalf("len(replys) = %d, want 1", replys.Len())
	}

	reply := replys.Objs[0]
	if reply.Text != "回复内容" || reply.FName != "天堂鸡汤" {
		t.Errorf("reply text/fname = %+v", reply)
	}
	if reply.TID != 111 || reply.PID != 222 || reply.PPID != 333 {
		t.Errorf("reply ids = %+v", reply)
	}
	if !reply.IsComment {
		t.Error("IsComment = false, want true (is_floor is 1)")
	}
	if reply.CreateTime != 1700000000 {
		t.Errorf("create time = %d", reply.CreateTime)
	}

	// The replyer portrait keeps its "?..." suffix stripping.
	replier := reply.User
	if replier.UserID != 4444444 || replier.Portrait != "tb.1.abcdefghijklmnopqrst" {
		t.Errorf("replyer = %+v", replier)
	}
	if replier.UserName != "某个用户名" || replier.NickNameNew != "某个昵称" {
		t.Errorf("replyer names = %+v", replier)
	}
	if replier.PrivLike != enums.PrivLikeFriend {
		t.Errorf("priv like = %v, want PrivLikeFriend", replier.PrivLike)
	}
	if replier.PrivReply != enums.PrivReplyFans {
		t.Errorf("priv reply = %v, want PrivReplyFans", replier.PrivReply)
	}
	if got := replier.ShowName(); got != "某个昵称" {
		t.Errorf("ShowName() = %q", got)
	}
	if got := replier.LogName(); got != "某个用户名" {
		t.Errorf("LogName() = %q", got)
	}
	if reply.AuthorID() != 4444444 {
		t.Errorf("AuthorID() = %d", reply.AuthorID())
	}

	// The quoted floor user has no portrait.
	if reply.PostUser.UserID != 5555555 || reply.PostUser.UserName != "楼层用户" {
		t.Errorf("post user = %+v", reply.PostUser)
	}
	if got := reply.PostUser.LogName(); got != "楼层用户" {
		t.Errorf("post user LogName() = %q", got)
	}

	// The thread author keeps the raw portrait.
	if reply.ThreadUser.UserID != 6666666 || reply.ThreadUser.Portrait != "tb.1.lz" {
		t.Errorf("thread user = %+v", reply.ThreadUser)
	}
	if got := reply.ThreadUser.LogName(); got != "楼主昵称/tb.1.lz" {
		t.Errorf("thread user LogName() = %q", got)
	}

	page := replys.Page
	if page.CurrentPage != 2 || !page.HasMore || !page.HasPrev {
		t.Errorf("page = %+v", page)
	}
	if !replys.HasMore() {
		t.Error("HasMore() = false, want true")
	}
}

func TestPrivDefaults(t *testing.T) {
	// A zero priv_sets value falls back to PUBLIC / ALL.
	user := UserInfoReplyFromProto(&commonpb.User{Id: 1})
	if user.PrivLike != enums.PrivLikePublic {
		t.Errorf("priv like = %v, want PrivLikePublic", user.PrivLike)
	}
	if user.PrivReply != enums.PrivReplyAll {
		t.Errorf("priv reply = %v, want PrivReplyAll", user.PrivReply)
	}
}

func TestParseBodySurfacesServerError(t *testing.T) {
	res := &pb.ReplyMeResIdl{Error: &commonpb.Error{Errorno: 340006, Errmsg: "boom"}}
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
	if u.Scheme != "https" || u.Host != "tiebac.baidu.com" || u.Path != "/c/u/feed/replyme" {
		t.Errorf("url = %s", u)
	}
	if u.RawQuery != "cmd=303007" {
		t.Errorf("query = %q, want cmd=303007", u.RawQuery)
	}
}
