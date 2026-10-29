package getats

import (
	"testing"

	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/helper"
)

func mustParse(t *testing.T, s string) map[string]any {
	t.Helper()
	m, err := helper.ParseJSONMap([]byte(s))
	if err != nil {
		t.Fatalf("ParseJSONMap: %v", err)
	}
	return m
}

func TestUserInfoAtPrivSets(t *testing.T) {
	m := mustParse(t, `{
		"id": 1,
		"portrait": "tb.1.a?t=1234567890",
		"name": "u",
		"name_show": "n",
		"priv_sets": {"like": 3, "reply": 5}
	}`)
	u := UserInfoAtFromJSON(m)
	if u.Portrait != "tb.1.a" {
		t.Errorf("Portrait = %q, want tb.1.a", u.Portrait)
	}
	if u.PrivLike != enums.PrivLikeHide {
		t.Errorf("PrivLike = %v, want Hide", u.PrivLike)
	}
	if u.PrivReply != enums.PrivReplyFans {
		t.Errorf("PrivReply = %v, want Fans", u.PrivReply)
	}
}

func TestUserInfoAtDefaults(t *testing.T) {
	m := mustParse(t, `{"id": 2, "portrait": "tb.1.b", "name": "u2", "name_show": "n2"}`)
	u := UserInfoAtFromJSON(m)
	if u.PrivLike != enums.PrivLikePublic {
		t.Errorf("PrivLike = %v, want Public", u.PrivLike)
	}
	if u.PrivReply != enums.PrivReplyAll {
		t.Errorf("PrivReply = %v, want All", u.PrivReply)
	}
	if u.ShowName() != "n2" || u.NickName() != "n2" {
		t.Errorf("ShowName/NickName = %q/%q", u.ShowName(), u.NickName())
	}
}

const contentsJSON = `[
	{"type": 0, "text": "前文"},
	{"type": 2, "text": "emoji01", "c": "滑稽"},
	{"src": "https://imgsrc.baidu.com/forum/pic/item/0123456789abcdef0123456789abcdef.jpg",
	 "size": 4096, "bsize": "300,200"},
	{"type": 4, "text": "@小明", "uid": 42, "portrait": "tb.1.xyz?t=1700000000"},
	{"type": 1, "link": "https://example.com/p/1", "text": "链接标题"},
	{"type": 0, "text": "后文"}
]`

func TestContentsAtSniffing(t *testing.T) {
	m := mustParse(t, `{"content": `+contentsJSON+`}`)
	c := ContentsAtFromJSON(helper.JSONSlice(m, "content"))

	if got := len(c.Objs); got != 6 {
		t.Fatalf("len(Objs) = %d, want 6", got)
	}
	if _, ok := c.Objs[0].(classdef.FragText); !ok {
		t.Errorf("Objs[0] = %T, want FragText", c.Objs[0])
	}
	if _, ok := c.Objs[1].(FragEmojiAt); !ok {
		t.Errorf("Objs[1] = %T, want FragEmojiAt", c.Objs[1])
	}
	if _, ok := c.Objs[2].(FragImageAt); !ok {
		t.Errorf("Objs[2] = %T, want FragImageAt", c.Objs[2])
	}
	if _, ok := c.Objs[3].(FragAtAt); !ok {
		t.Errorf("Objs[3] = %T, want FragAtAt", c.Objs[3])
	}
	if _, ok := c.Objs[4].(classdef.FragLink); !ok {
		t.Errorf("Objs[4] = %T, want FragLink", c.Objs[4])
	}

	// 文本碎片列表按出现顺序收纳文本、链接与 @。
	if got := len(c.Texts); got != 4 {
		t.Errorf("len(Texts) = %d, want 4", got)
	}
	if got, want := c.Text(), "前文@小明https://example.com/p/1后文"; got != want {
		t.Errorf("Text() = %q, want %q", got, want)
	}

	if got := len(c.Emojis); got != 1 {
		t.Fatalf("len(Emojis) = %d, want 1", len(c.Emojis))
	}
	if c.Emojis[0].ID != "emoji01" || c.Emojis[0].Desc != "滑稽" {
		t.Errorf("Emojis[0] = %+v", c.Emojis[0])
	}

	if got := len(c.Imgs); got != 1 {
		t.Fatalf("len(Imgs) = %d, want 1", len(c.Imgs))
	}
	img := c.Imgs[0]
	if img.Src != "https://imgsrc.baidu.com/forum/pic/item/0123456789abcdef0123456789abcdef.jpg" {
		t.Errorf("Imgs[0].Src = %q", img.Src)
	}
	if img.OriginSize != 4096 || img.ShowWidth != 300 || img.ShowHeight != 200 {
		t.Errorf("Imgs[0] size/show = %d/%d,%d", img.OriginSize, img.ShowWidth, img.ShowHeight)
	}
	if img.Hash != "0123456789abcdef0123456789abcdef" {
		t.Errorf("Imgs[0].Hash = %q", img.Hash)
	}

	if got := len(c.Ats); got != 1 {
		t.Fatalf("len(Ats) = %d, want 1", len(c.Ats))
	}
	if c.Ats[0].Text != "@小明" || c.Ats[0].UserID != 42 || c.Ats[0].Portrait != "tb.1.xyz" {
		t.Errorf("Ats[0] = %+v", c.Ats[0])
	}

	if got := len(c.Links); got != 1 {
		t.Fatalf("len(Links) = %d, want 1", len(c.Links))
	}
	if c.Links[0].Text != "https://example.com/p/1" || c.Links[0].Title != "链接标题" {
		t.Errorf("Links[0] = %+v", c.Links[0])
	}
}

const atJSON = `{
	"content": "回复了你",
	"fname": "天堂鸡汤",
	"fid": 54321,
	"thread_id": 100,
	"post_id": 200,
	"replyer": {"id": 9, "portrait": "tb.1.c?t=1700000000", "name": "u", "name_show": "n"},
	"type": 3,
	"time": 1700000000,
	"quote_pid": 300,
	"quote_user": {"id": 11, "portrait": "tb.1.d", "name": "qu", "name_show": "qn"},
	"thread_author_user": {"id": 12, "portrait": "tb.1.e", "name": "tu", "name_show": "tn"},
	"new_floor_info": [
		{"content": [{"type": 0, "text": "主题正文"}]},
		{"content": [
			{"type": 4, "text": "@楼主", "uid": 7, "portrait": "tb.1.q?t=1700000000"},
			{"type": 0, "text": ": 引用的话"}
		]},
		{"content": []}
	]
}`

func TestAtFromJSON(t *testing.T) {
	a := AtFromJSON(mustParse(t, atJSON))

	if a.AuthorID() != 9 {
		t.Errorf("AuthorID = %d, want 9", a.AuthorID())
	}
	if a.Text != "回复了你" {
		t.Errorf("Text = %q, want 回复了你", a.Text)
	}
	if a.FName != "天堂鸡汤" || a.FID != 54321 {
		t.Errorf("FName/FID = %q/%d", a.FName, a.FID)
	}
	if a.TID != 100 || a.PID != 200 {
		t.Errorf("TID/PID = %d/%d", a.TID, a.PID)
	}
	if a.ContentType != enums.ContentTypeComment {
		t.Errorf("ContentType = %v, want Comment", a.ContentType)
	}

	// Post_at 取 new_floor_info[-2] 并剥掉引述头。
	if !a.Post.Valid() {
		t.Fatal("Post.Valid() = false, want true")
	}
	if a.Post.PID != 300 {
		t.Errorf("Post.PID = %d, want 300", a.Post.PID)
	}
	if got := a.Post.Contents.Text(); got != "引用的话" {
		t.Errorf("Post.Contents.Text() = %q, want 引用的话", got)
	}
	if got := len(a.Post.Contents.Objs); got != 1 {
		t.Errorf("len(Post.Contents.Objs) = %d, want 1", got)
	}
	if got := len(a.Post.Contents.Ats); got != 0 {
		t.Errorf("len(Post.Contents.Ats) = %d, want 0", got)
	}
	if a.Post.User.UserID != 11 || a.Post.User.Portrait != "tb.1.d" {
		t.Errorf("Post.User = %+v", a.Post.User)
	}

	// Thread_at 取 new_floor_info[0]。
	if !a.Thread.Valid() {
		t.Fatal("Thread.Valid() = false, want true")
	}
	if a.Thread.TID != 100 {
		t.Errorf("Thread.TID = %d, want 100", a.Thread.TID)
	}
	if got := a.Thread.Contents.Text(); got != "主题正文" {
		t.Errorf("Thread.Contents.Text() = %q, want 主题正文", got)
	}
	if a.Thread.User.UserID != 12 || a.Thread.User.ShowName() != "tn" {
		t.Errorf("Thread.User = %+v", a.Thread.User)
	}
}

func TestPostAtRequiresMoreThanTwoFloors(t *testing.T) {
	m := mustParse(t, `{
		"quote_pid": 0,
		"new_floor_info": [
			{"content": [{"type": 0, "text": "只有两楼"}]},
			{"content": [{"type": 0, "text": "第二楼"}]}
		]
	}`)

	// 两个条件不对称：Thread_at 在非空时取首项，Post_at 需要长度大于 2。
	thread := ThreadAtFromJSON(m)
	if got := thread.Contents.Text(); got != "只有两楼" {
		t.Errorf("Thread.Contents.Text() = %q, want 只有两楼", got)
	}

	post := PostAtFromJSON(m)
	if post.Valid() {
		t.Errorf("Post.Valid() = true, want false, Objs = %d", len(post.Contents.Objs))
	}
}

func TestStripQuoteHeaderEmptyText(t *testing.T) {
	m := mustParse(t, `{
		"quote_pid": 0,
		"new_floor_info": [
			{"content": []},
			{"content": [
				{"type": 4, "text": "@某人", "uid": 1, "portrait": "tb.1.r"},
				{"type": 0, "text": ": "}
			]},
			{"content": [{"type": 0, "text": "占位"}]}
		]
	}`)

	post := PostAtFromJSON(m)
	if post.Valid() {
		t.Errorf("Post.Valid() = true, want false")
	}
	if got := len(post.Contents.Objs); got != 0 {
		t.Errorf("len(Post.Contents.Objs) = %d, want 0", got)
	}
	if got := len(post.Contents.Ats); got != 0 {
		t.Errorf("len(Post.Contents.Ats) = %d, want 0", got)
	}
}

func TestAtsFromJSON(t *testing.T) {
	m := mustParse(t, `{
		"at_list": [
			`+atJSON+`],
		"page": {"current_page": 2, "has_more": "1", "has_prev": "0"}
	}`)
	ats := AtsFromJSON(m)
	if len(ats.Objs) != 1 {
		t.Fatalf("len(Objs) = %d, want 1", len(ats.Objs))
	}
	if ats.Objs[0].AuthorID() != 9 {
		t.Errorf("AuthorID = %d, want 9", ats.Objs[0].AuthorID())
	}
	if ats.Page.CurrentPage != 2 || !ats.Page.HasMore || ats.Page.HasPrev {
		t.Errorf("Page = %+v", ats.Page)
	}
	if !ats.HasMore() {
		t.Errorf("HasMore() = false, want true")
	}
}
