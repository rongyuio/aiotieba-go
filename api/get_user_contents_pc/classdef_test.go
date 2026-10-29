package getusercontentpc

import (
	"testing"

	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/helper"
)

// contentsJSON 覆盖全部碎片类型：0/9/18/27/40 文本、2/11 表情、3/20 图像、
// 4 @、1 链接、5 视频、10 音频，外加一个未知类型。
const contentsJSON = `{
	"content": [
		{"type": 0, "text": "hello"},
		{"type": 9, "text": " 400-123"},
		{"type": 18, "text": "#话题#"},
		{"type": 27, "text": "词条"},
		{"type": 40, "text": "梗"},
		{"type": 2, "text": "image_emoticon1", "c": "微笑"},
		{"type": 11, "text": "image_emoticon2", "c": "大笑"},
		{"type": 3, "cdn_src": "https://imgsa.baidu.com/forum/abc1234567890abcdef1234567890abcd.jpg", "big_cdn_src": "https://big.jpg", "origin_src": "https://origin.jpg", "bsize": "720,960"},
		{"type": 20, "cdn_src": "https://imgsa.baidu.com/forum/0123456789abcdef0123456789abcdef.jpg", "big_cdn_src": "https://big2.jpg", "origin_src": "https://origin2.jpg", "bsize": "300,400"},
		{"type": 4, "text": "@alice ", "uid": 42, "portrait": "tb.1.abcdef?t=1234567890"},
		{"type": 1, "text": "示例", "link": "https://example.com/a"},
		{"type": 5, "link": "https://v.mp4", "src": "https://cover.jpg", "during_time": "12", "width": 1280, "height": 720},
		{"type": 10, "voice_md5": "deadbeef", "during_time": 1500},
		{"type": 7, "x": 1}
	]
}`

// threadJSON 是一份完整的 thread_info。
const threadJSON = `{
	"title": "标题",
	"fid": 12345,
	"fname": "测试吧",
	"tid": 987654,
	"first_post_id": 111,
	"author": {"id": 42, "portrait": "tb.1.abcdef?t=1234567890", "name": "alice", "name_show": "爱丽丝", "gender": 2, "priv_sets": {"like": 3, "reply": 6}},
	"forum_info": {"id": 12345, "name": "测试吧", "member_num": 999, "post_num": 123456, "is_liked": 1},
	"first_post_content": [{"type": 0, "text": "首楼正文"}],
	"thread_type": 0,
	"view_num": 100,
	"reply_num": 20,
	"share_num": 3,
	"agree": {"agree_num": 7, "disagree_num": 1},
	"create_time": 1700000000,
	"last_time_int": 1700000100,
	"is_good": 1,
	"is_top": 0,
	"is_deleted": 0
}`

func mustJSONMap(t *testing.T, raw string) map[string]any {
	t.Helper()
	m, err := helper.ParseJSONMap([]byte(raw))
	if err != nil {
		t.Fatalf("ParseJSONMap: %v", err)
	}
	return m
}

func TestContentsPcupFromJSON(t *testing.T) {
	m := mustJSONMap(t, contentsJSON)
	c := ContentsPcupFromJSON(helper.JSONSlice(m, "content"))

	// 文本碎片：0/9/18/27/40 五条，加上 @ 与链接
	if len(c.Texts) != 7 {
		t.Errorf("len(Texts) = %d, want 7", len(c.Texts))
	}
	wantText := "hello 400-123#话题#词条梗@alice https://example.com/a"
	if got := c.Text(); got != wantText {
		t.Errorf("Text() = %q, want %q", got, wantText)
	}

	if len(c.Emojis) != 2 {
		t.Fatalf("len(Emojis) = %d, want 2", len(c.Emojis))
	}
	if c.Emojis[0].ID != "image_emoticon1" || c.Emojis[0].Desc != "微笑" {
		t.Errorf("Emojis[0] = %+v", c.Emojis[0])
	}
	if c.Emojis[1].ID != "image_emoticon2" || c.Emojis[1].Desc != "大笑" {
		t.Errorf("Emojis[1] = %+v", c.Emojis[1])
	}

	if len(c.Imgs) != 2 {
		t.Fatalf("len(Imgs) = %d, want 2", len(c.Imgs))
	}
	img := c.Imgs[0]
	if img.Src != "https://imgsa.baidu.com/forum/abc1234567890abcdef1234567890abcd.jpg" {
		t.Errorf("Imgs[0].Src = %q", img.Src)
	}
	if img.BigSrc != "https://big.jpg" || img.OriginSrc != "https://origin.jpg" {
		t.Errorf("Imgs[0] big/origin = %q/%q", img.BigSrc, img.OriginSrc)
	}
	if img.ShowWidth != 720 || img.ShowHeight != 960 {
		t.Errorf("Imgs[0] size = %dx%d, want 720x960", img.ShowWidth, img.ShowHeight)
	}
	if img.Hash != "abc1234567890abcdef1234567890abcd" {
		t.Errorf("Imgs[0].Hash = %q", img.Hash)
	}
	if got := c.Imgs[1]; got.ShowWidth != 300 || got.ShowHeight != 400 {
		t.Errorf("Imgs[1] size = %dx%d, want 300x400", got.ShowWidth, got.ShowHeight)
	}

	if len(c.Ats) != 1 {
		t.Fatalf("len(Ats) = %d, want 1", len(c.Ats))
	}
	at := c.Ats[0]
	if at.Text != "@alice " || at.UserID != 42 {
		t.Errorf("Ats[0] = %+v", at)
	}
	if at.Portrait != "tb.1.abcdef" {
		t.Errorf("Ats[0].Portrait = %q, want tb.1.abcdef", at.Portrait)
	}

	if len(c.Links) != 1 {
		t.Fatalf("len(Links) = %d, want 1", len(c.Links))
	}
	if c.Links[0].Text != "https://example.com/a" || c.Links[0].Title != "示例" {
		t.Errorf("Links[0] = %+v", c.Links[0])
	}

	wantVideo := FragVideoPcup{Src: "https://v.mp4", CoverSrc: "https://cover.jpg", Duration: 12, Width: 1280, Height: 720}
	if c.Video != wantVideo {
		t.Errorf("Video = %+v, want %+v", c.Video, wantVideo)
	}
	if !c.Video.Valid() {
		t.Error("Video.Valid() = false, want true")
	}

	wantVoice := FragVoicePcup{MD5: "deadbeef", Duration: 1.5}
	if c.Voice != wantVoice {
		t.Errorf("Voice = %+v, want %+v", c.Voice, wantVoice)
	}
	if !c.Voice.Valid() {
		t.Error("Voice.Valid() = false, want true")
	}

	// objs 顺序与 content 一致，视频与音频也进 objs
	if len(c.Objs) != 14 {
		t.Fatalf("len(Objs) = %d, want 14", len(c.Objs))
	}
	if _, ok := c.Objs[0].(classdef.FragText); !ok {
		t.Errorf("Objs[0] = %T, want classdef.FragText", c.Objs[0])
	}
	if _, ok := c.Objs[5].(classdef.FragEmoji); !ok {
		t.Errorf("Objs[5] = %T, want classdef.FragEmoji", c.Objs[5])
	}
	if _, ok := c.Objs[7].(FragImagePcup); !ok {
		t.Errorf("Objs[7] = %T, want FragImagePcup", c.Objs[7])
	}
	if _, ok := c.Objs[9].(FragAtPcup); !ok {
		t.Errorf("Objs[9] = %T, want FragAtPcup", c.Objs[9])
	}
	if _, ok := c.Objs[10].(classdef.FragLink); !ok {
		t.Errorf("Objs[10] = %T, want classdef.FragLink", c.Objs[10])
	}
	if v, ok := c.Objs[11].(FragVideoPcup); !ok || v != wantVideo {
		t.Errorf("Objs[11] = %T %+v, want FragVideoPcup", c.Objs[11], c.Objs[11])
	}
	if v, ok := c.Objs[12].(FragVoicePcup); !ok || v != wantVoice {
		t.Errorf("Objs[12] = %T %+v, want FragVoicePcup", c.Objs[12], c.Objs[12])
	}
	if _, ok := c.Objs[13].(classdef.FragUnknown); !ok {
		t.Errorf("Objs[13] = %T, want classdef.FragUnknown", c.Objs[13])
	}
}

func TestContentsPcupEmpty(t *testing.T) {
	c := ContentsPcupFromJSON(nil)
	if len(c.Objs) != 0 || c.Text() != "" {
		t.Errorf("empty contents = %+v", c)
	}
	if c.Video.Valid() || c.Voice.Valid() {
		t.Error("empty contents reports a valid video or voice")
	}
}

func TestUserInfoPcuFromJSON(t *testing.T) {
	m := mustJSONMap(t, `{
		"id": 42,
		"portrait": "tb.1.abcdef?t=1234567890",
		"name": "alice",
		"name_show": "爱丽丝",
		"gender": 2,
		"priv_sets": {"like": 3, "reply": 6}
	}`)
	u := UserInfoPcuFromJSON(m)

	if u.UserID != 42 || u.Portrait != "tb.1.abcdef" || u.UserName != "alice" || u.NickNameNew != "爱丽丝" {
		t.Errorf("UserInfoPcu = %+v", u)
	}
	if u.Gender != enums.GenderFemale {
		t.Errorf("Gender = %v, want GenderFemale", u.Gender)
	}
	if u.PrivLike != enums.PrivLikeHide {
		t.Errorf("PrivLike = %v, want PrivLikeHide", u.PrivLike)
	}
	if u.PrivReply != enums.PrivReplyFollow {
		t.Errorf("PrivReply = %v, want PrivReplyFollow", u.PrivReply)
	}
	if u.NickName() != "爱丽丝" || u.ShowName() != "爱丽丝" || u.String() != "alice" || u.LogName() != "alice" {
		t.Errorf("name helpers = %q/%q/%q/%q", u.NickName(), u.ShowName(), u.String(), u.LogName())
	}

	// priv_sets 缺失时回退为 PUBLIC / ALL，gender 缺失时为 UNKNOWN
	bare := UserInfoPcuFromJSON(mustJSONMap(t, `{"id": 7}`))
	if bare.PrivLike != enums.PrivLikePublic || bare.PrivReply != enums.PrivReplyAll {
		t.Errorf("bare priv = %v/%v, want Public/All", bare.PrivLike, bare.PrivReply)
	}
	if bare.Gender != enums.GenderUnknown {
		t.Errorf("bare Gender = %v, want GenderUnknown", bare.Gender)
	}
	if bare.String() != "7" || bare.LogName() != "7" {
		t.Errorf("bare String/LogName = %q/%q", bare.String(), bare.LogName())
	}
	if (UserInfoPcu{}).Valid() {
		t.Error("zero UserInfoPcu Valid() = true, want false")
	}

	// log_name 在缺少用户名时用 昵称/portrait
	named := UserInfoPcuFromJSON(mustJSONMap(t, `{"id": 8, "portrait": "tb.1.cccc", "name_show": "无名"}`))
	if named.LogName() != "无名/tb.1.cccc" {
		t.Errorf("LogName = %q, want 无名/tb.1.cccc", named.LogName())
	}
	if named.ShowName() != "无名" {
		t.Errorf("ShowName = %q, want 无名", named.ShowName())
	}
}

func TestThreadPcupFromJSON(t *testing.T) {
	th := ThreadPcupFromJSON(mustJSONMap(t, threadJSON))

	if th.TID != 987654 || th.PID != 111 || th.FID != 12345 || th.FName != "测试吧" {
		t.Errorf("thread ids = tid %d pid %d fid %d fname %q", th.TID, th.PID, th.FID, th.FName)
	}
	if th.Title != "标题" {
		t.Errorf("Title = %q", th.Title)
	}
	if th.Text() != "标题\n首楼正文" {
		t.Errorf("Text() = %q", th.Text())
	}
	if th.Type != enums.ThreadTypeArticle {
		t.Errorf("Type = %v, want ThreadTypeArticle", th.Type)
	}
	if th.ViewNum != 100 || th.ReplyNum != 20 || th.ShareNum != 3 {
		t.Errorf("counters = %d/%d/%d", th.ViewNum, th.ReplyNum, th.ShareNum)
	}
	if th.Agree != 7 || th.Disagree != 1 {
		t.Errorf("agree/disagree = %d/%d", th.Agree, th.Disagree)
	}
	if th.CreateTime != 1700000000 || th.LastTime != 1700000100 {
		t.Errorf("times = %d/%d", th.CreateTime, th.LastTime)
	}
	if !th.IsGood || th.IsTop || th.IsDeleted {
		t.Errorf("flags = good %v top %v deleted %v", th.IsGood, th.IsTop, th.IsDeleted)
	}
	if th.AuthorID() != 42 || th.User.NickNameNew != "爱丽丝" {
		t.Errorf("author = %+v", th.User)
	}

	forum := th.Forum
	if forum.FID != 12345 || forum.FName != "测试吧" || forum.MemberNum != 999 || forum.PostNum != 123456 || !forum.IsLiked {
		t.Errorf("Forum = %+v", forum)
	}
	if !forum.Valid() {
		t.Error("Forum.Valid() = false, want true")
	}

	if !th.Valid() {
		t.Error("ThreadPcup.Valid() = false, want true")
	}

	// 未知帖子类型回退为 UNKNOWN
	unknown := ThreadPcupFromJSON(mustJSONMap(t, `{"thread_type": 999}`))
	if unknown.Type != enums.ThreadTypeUnknown {
		t.Errorf("unknown Type = %v, want ThreadTypeUnknown", unknown.Type)
	}
	if unknown.Text() != "" {
		t.Errorf("unknown Text() = %q, want empty", unknown.Text())
	}
	if unknown.Valid() {
		t.Error("unknown thread Valid() = true, want false for tid 0")
	}
}

func TestPcUserPostWithPostInfo(t *testing.T) {
	m := mustJSONMap(t, `{
		"type": 2,
		"post_info": {
			"id": 222,
			"quote_id": "111",
			"time": 1700000200,
			"author": {"id": 43, "portrait": "tb.1.bbbb", "name": "bob", "name_show": "鲍勃", "gender": 1},
			"content": [{"type": 0, "text": "回复内容"}]
		},
		"thread_info": {
			"title": "标题",
			"fid": 12345,
			"fname": "测试吧",
			"tid": 987654,
			"first_post_id": 111,
			"author": {"id": 42, "name": "alice"},
			"forum_info": {"id": 12345, "name": "测试吧"},
			"first_post_content": [{"type": 0, "text": "首楼正文"}],
			"thread_type": 0,
			"create_time": 1700000000
		}
	}`)

	p := PcUserPostFromJSON(m)

	if p.FID != 12345 {
		t.Errorf("FID = %d, want 12345 (from thread)", p.FID)
	}
	if p.PPID != 111 || p.PID != 222 {
		t.Errorf("ppid/pid = %d/%d, want 111/222", p.PPID, p.PID)
	}
	if p.User.UserID != 43 || p.User.UserName != "bob" {
		t.Errorf("User = %+v, want post_info.author", p.User)
	}
	if p.AuthorID() != 43 {
		t.Errorf("AuthorID() = %d, want 43", p.AuthorID())
	}
	if p.ContentType != enums.ContentTypePost {
		t.Errorf("ContentType = %v, want ContentTypePost", p.ContentType)
	}
	if p.CreateTime != 1700000200 {
		t.Errorf("CreateTime = %d, want 1700000200", p.CreateTime)
	}
	if p.Text() != "回复内容" {
		t.Errorf("Text() = %q, want 回复内容", p.Text())
	}
	if p.Thread.TID != 987654 {
		t.Errorf("Thread.TID = %d, want 987654", p.Thread.TID)
	}
}

func TestPcUserPostDegradesToThread(t *testing.T) {
	cases := map[string]string{
		"null":  `{"type": 1, "post_info": null, "thread_info": ` + threadJSON + `}`,
		"empty": `{"type": 1, "post_info": {}, "thread_info": ` + threadJSON + `}`,
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			p := PcUserPostFromJSON(mustJSONMap(t, raw))

			if p.Text() != "首楼正文" {
				t.Errorf("Text() = %q, want 首楼正文", p.Text())
			}
			if p.PID != 111 || p.PPID != 0 {
				t.Errorf("pid/ppid = %d/%d, want 111/0", p.PID, p.PPID)
			}
			if p.User.UserID != 42 {
				t.Errorf("User.UserID = %d, want thread author 42", p.User.UserID)
			}
			if p.CreateTime != 1700000000 {
				t.Errorf("CreateTime = %d, want thread create_time 1700000000", p.CreateTime)
			}
			if p.ContentType != enums.ContentTypeThread {
				t.Errorf("ContentType = %v, want ContentTypeThread", p.ContentType)
			}
			if p.FID != 12345 {
				t.Errorf("FID = %d, want 12345", p.FID)
			}
		})
	}
}

func TestPcUserPostsFromJSON(t *testing.T) {
	m := mustJSONMap(t, `{
		"list": [
			{
				"type": 2,
				"post_info": {"id": 1, "quote_id": 0, "time": 10, "author": {"id": 43, "name": "bob"}, "content": [{"type": 0, "text": "a"}]},
				"thread_info": {"fid": 7, "tid": 100, "first_post_id": 1, "author": {"id": 42, "name": "alice"}, "create_time": 5}
			},
			{
				"type": 2,
				"post_info": {"id": 2, "quote_id": 1, "time": 20, "author": {"id": 44, "name": "carol"}, "content": [{"type": 0, "text": "b"}]},
				"thread_info": {"fid": 7, "tid": 101, "first_post_id": 2, "author": {"id": 42, "name": "alice"}, "create_time": 6}
			}
		]
	}`)

	posts := PcUserPostsFromJSON(m)
	if posts.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", posts.Len())
	}
	// 每个条目的 author 必须来自各自的 post_info，不能被首项的 author 覆盖
	if posts.Objs[0].User.UserID != 43 || posts.Objs[1].User.UserID != 44 {
		t.Errorf("user ids = %d/%d, want 43/44", posts.Objs[0].User.UserID, posts.Objs[1].User.UserID)
	}
	if posts.Objs[0].PPID != 0 || posts.Objs[1].PPID != 1 {
		t.Errorf("ppids = %d/%d, want 0/1", posts.Objs[0].PPID, posts.Objs[1].PPID)
	}
	if posts.Objs[0].Thread.TID != 100 || posts.Objs[1].Thread.TID != 101 {
		t.Errorf("thread tids = %d/%d", posts.Objs[0].Thread.TID, posts.Objs[1].Thread.TID)
	}

	empty := PcUserPostsFromJSON(mustJSONMap(t, `{"list": []}`))
	if !empty.Empty() {
		t.Error("empty list is not Empty()")
	}
	if nilList := PcUserPostsFromJSON(mustJSONMap(t, `{}`)); !nilList.Empty() {
		t.Error("missing list is not Empty()")
	}
}
