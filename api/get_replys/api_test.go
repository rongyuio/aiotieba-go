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

	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/get_replys/protobuf"
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

	// 正文取自 new_floor_info 的末项；@碎片计入 texts，表情与未知碎片不计。
	if got, want := reply.Contents.Text(), "回复内容@某人"; got != want {
		t.Errorf("contents text = %q, want %q", got, want)
	}
	if got := reply.Text(); got != reply.Contents.Text() {
		t.Errorf("Text() = %q, want %q", got, reply.Contents.Text())
	}
	if reply.FName != "天堂鸡汤" || reply.PID != 222 {
		t.Errorf("fname/pid = %q/%d", reply.FName, reply.PID)
	}
	if reply.ContentType != enums.ContentTypeComment {
		t.Errorf("content type = %v, want ContentTypeComment (is_floor is 1)", reply.ContentType)
	}
	if reply.CreateTime != 1700000000 {
		t.Errorf("create time = %d", reply.CreateTime)
	}

	contents := reply.Contents
	if contents.Len() != 4 {
		t.Errorf("len(contents) = %d, want 4 (文本 + 表情 + @ + 未知)", contents.Len())
	}
	if len(contents.Texts) != 2 || len(contents.Emojis) != 1 || len(contents.Ats) != 1 {
		t.Errorf("contents = %d texts, %d emojis, %d ats", len(contents.Texts), len(contents.Emojis), len(contents.Ats))
	}
	if got := contents.Emojis[0]; got.ID != "image_emoticon1" || got.Desc != "滑稽" {
		t.Errorf("emoji = %+v", got)
	}
	if got := contents.Ats[0]; got.Text != "@某人" || got.UserID != 7777777 || got.Portrait != "tb.1.someone" {
		t.Errorf("at = %+v", got)
	}
	if _, ok := contents.Objs[3].(classdef.FragUnknown); !ok {
		t.Errorf("objs[3] = %T, want classdef.FragUnknown", contents.Objs[3])
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

	// 父级回复取自倒数第二项，": " 引述头已被剥掉。
	post := reply.Post
	if post.PID != 333 || !post.Valid() {
		t.Errorf("post = %+v", post)
	}
	if got, want := post.Contents.Text(), "层主说什么"; got != want {
		t.Errorf("post text = %q, want %q", got, want)
	}
	if post.Contents.Len() != 1 || len(post.Contents.Ats) != 0 {
		t.Errorf("post contents = %d objs, %d ats", post.Contents.Len(), len(post.Contents.Ats))
	}
	// The quoted floor user has no portrait.
	if post.User.UserID != 5555555 || post.User.UserName != "楼层用户" {
		t.Errorf("post user = %+v", post.User)
	}
	if got := post.User.LogName(); got != "楼层用户" {
		t.Errorf("post user LogName() = %q", got)
	}

	// 主题帖取自首项。
	thread := reply.Thread
	if thread.TID != 111 || !thread.Valid() {
		t.Errorf("thread = %+v", thread)
	}
	if got, want := thread.Contents.Text(), "主题帖正文"; got != want {
		t.Errorf("thread text = %q, want %q", got, want)
	}
	// The thread author keeps the raw portrait.
	if thread.User.UserID != 6666666 || thread.User.Portrait != "tb.1.lz" {
		t.Errorf("thread user = %+v", thread.User)
	}
	if got := thread.User.LogName(); got != "楼主昵称/tb.1.lz" {
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

func TestStripQuoteHeader(t *testing.T) {
	at := &commonpb.PbContent{Type: 4, Text: "@某人", Uid: 7, Portrait: "tb.1.p"}
	text := func(s string) *commonpb.PbContent { return &commonpb.PbContent{Text: s} }

	tests := []struct {
		name      string
		content   []*commonpb.PbContent
		wantText  string
		wantObjs  int
		wantTexts int
		wantAts   int
	}{
		{
			name:      "剥掉引述头",
			content:   []*commonpb.PbContent{at, text(": 正文")},
			wantText:  "正文",
			wantObjs:  1,
			wantTexts: 1,
			wantAts:   0,
		},
		{
			name:      "剥空后连同文本碎片一起丢弃",
			content:   []*commonpb.PbContent{at, text(": ")},
			wantText:  "",
			wantObjs:  0,
			wantTexts: 0,
			wantAts:   0,
		},
		{
			name:      "首项不是@时不处理",
			content:   []*commonpb.PbContent{text("@某人: 正文")},
			wantText:  "@某人: 正文",
			wantObjs:  1,
			wantTexts: 1,
			wantAts:   0,
		},
		{
			name:      "没有引述前缀时不处理",
			content:   []*commonpb.PbContent{at, text("正文")},
			wantText:  "@某人正文",
			wantObjs:  2,
			wantTexts: 2,
			wantAts:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := ContentsReplyFromProto(&pb.ReplyMeResIdl_DataRes_ReplyList_NewFloorInfo{Content: tt.content})
			stripQuoteHeader(&c)

			if got := c.Text(); got != tt.wantText {
				t.Errorf("Text() = %q, want %q", got, tt.wantText)
			}
			if c.Len() != tt.wantObjs {
				t.Errorf("len(objs) = %d, want %d", c.Len(), tt.wantObjs)
			}
			if len(c.Texts) != tt.wantTexts {
				t.Errorf("len(texts) = %d, want %d", len(c.Texts), tt.wantTexts)
			}
			if len(c.Ats) != tt.wantAts {
				t.Errorf("len(ats) = %d, want %d", len(c.Ats), tt.wantAts)
			}
		})
	}
}

func TestReplyFromProtoWithoutNewFloorInfo(t *testing.T) {
	// new_floor_info 缺失时正文为空，父级回复与主题帖都不成立，
	// 帖子类型按 is_floor 的缺省值落回 POST。
	reply := ReplyFromProto(&pb.ReplyMeResIdl_DataRes_ReplyList{
		PostId:  222,
		IsFloor: 0,
	})

	if reply.Text() != "" || !reply.Contents.Empty() {
		t.Errorf("contents = %+v, want empty", reply.Contents)
	}
	if reply.Post.Valid() || reply.Thread.Valid() {
		t.Errorf("post/thread = %+v/%+v, want empty", reply.Post, reply.Thread)
	}
	if reply.ContentType != enums.ContentTypePost {
		t.Errorf("content type = %v, want ContentTypePost", reply.ContentType)
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
