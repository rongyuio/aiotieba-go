// The profile data models are exercised through the request packages, which
// mirrors how callers reach them. This is an external test package because
// those packages import api/profile.
package profile_test

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rongyuio/aiotieba-go/api/profile"
	"github.com/rongyuio/aiotieba-go/api/profile/get_homepage"
	"github.com/rongyuio/aiotieba-go/api/profile/get_uinfo_profile"
)

// loadResponse returns the Python generated ProfileResIdl wire bytes.
func loadResponse(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "response.hex"))
	if err != nil {
		t.Fatalf("reading testdata/response.hex: %v", err)
	}
	b, err := hex.DecodeString(strings.TrimSpace(string(raw)))
	if err != nil {
		t.Fatalf("decoding the fixture: %v", err)
	}
	return b
}

func TestParseResponseUser(t *testing.T) {
	user, err := getuinfoprofile.ParseBody(loadResponse(t))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}

	if user.UserID != 4444444 || user.UserName != "某个用户名" || user.NickNameNew != "某个昵称" {
		t.Errorf("identity = %+v", user)
	}
	// Python does `portrait[:-13]` when the portrait carries a "?" suffix.
	if user.Portrait != "tb.1.abcdefghijklmnopqrst" {
		t.Errorf("portrait = %q", user.Portrait)
	}
	if user.TiebaUID != 987654321 {
		t.Errorf("tieba uid = %d, want 987654321", user.TiebaUID)
	}
	if user.GLevel != 12 {
		t.Errorf("glevel = %d, want 12", user.GLevel)
	}
	if user.Age != 8.5 {
		t.Errorf("age = %v, want 8.5", user.Age)
	}
	if user.PostNum != 42 || user.FanNum != 7 || user.FollowNum != 8 || user.ForumNum != 9 {
		t.Errorf("counters = %+v", user)
	}
	// 3e9 does not fit in an int32.
	if user.AgreeNum != 3000000000 {
		t.Errorf("agree num = %d, want 3000000000", user.AgreeNum)
	}
	if user.Sign != "个性签名" || user.IP != "浙江" {
		t.Errorf("sign/ip = %q/%q", user.Sign, user.IP)
	}
	// The empty icon name is skipped.
	if len(user.Icons) != 2 || user.Icons[0] != "icon-a" || user.Icons[1] != "icon-b" {
		t.Errorf("icons = %v", user.Icons)
	}
	if !user.IsVIP || !user.IsGod || user.IsBlocked {
		t.Errorf("flags = vip:%v god:%v blocked:%v", user.IsVIP, user.IsGod, user.IsBlocked)
	}
	if got := user.String(); got != "某个用户名" {
		t.Errorf("String() = %q", got)
	}
	if got := user.LogName(); got != "某个用户名" {
		t.Errorf("LogName() = %q", got)
	}
}

func TestParseResponseHomepage(t *testing.T) {
	homepage, err := gethomepage.ParseBody(loadResponse(t))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}

	if len(homepage.Objs) != 1 {
		t.Fatalf("len(posts) = %d, want 1", len(homepage.Objs))
	}
	post := homepage.Objs[0]

	if post.Title != "标题" || post.FID != 111 || post.TID != 222 || post.PID != 333 {
		t.Errorf("post identity = %+v", post)
	}
	if post.FName != "天堂鸡汤" || post.CreateTime != 1700000000 {
		t.Errorf("post meta = %+v", post)
	}
	if post.ViewNum != 6 || post.ReplyNum != 4 || post.ShareNum != 5 || post.Agree != 7 || post.Disagree != 8 {
		t.Errorf("post counters = %+v", post)
	}
	// The page owner is assigned to every post.
	if post.User.UserID != 4444444 || post.AuthorID() != 4444444 {
		t.Errorf("post author = %+v", post.User)
	}
	if got := post.Text(); got != "标题\n正文" {
		t.Errorf("Text() = %q", got)
	}
	if homepage.User.UserID != 4444444 {
		t.Errorf("homepage user = %+v", homepage.User)
	}
}

func TestParseResponseContents(t *testing.T) {
	homepage, err := gethomepage.ParseBody(loadResponse(t))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	contents := homepage.Objs[0].Contents

	if got := contents.Text(); got != "正文" {
		t.Errorf("Text() = %q, want 正文", got)
	}
	if len(contents.Texts) != 1 || len(contents.Emojis) != 0 || len(contents.Ats) != 0 || len(contents.Links) != 0 {
		t.Errorf("fragments = %+v", contents)
	}

	// Images come from the media list (type != 5), not from the content.
	if len(contents.Imgs) != 1 {
		t.Fatalf("len(imgs) = %d, want 1", len(contents.Imgs))
	}
	img := contents.Imgs[0]
	if img.Src != "https://imgsrc.baidu.com/forum/pic/0123456789abcdef0123456789abcdef.jpg" {
		t.Errorf("img src = %q", img.Src)
	}
	if img.OriginSrc != "https://imgsrc.baidu.com/forum/pic/origin.jpg" || img.OriginSize != 2048 {
		t.Errorf("img origin = %+v", img)
	}
	if img.Width != 960 || img.Height != 540 {
		t.Errorf("img size = %+v", img)
	}
	if img.Hash != "0123456789abcdef0123456789abcdef" {
		t.Errorf("img hash = %q", img.Hash)
	}

	if !contents.Video.Valid() || contents.Video.Width != 1280 || contents.Video.Height != 720 {
		t.Errorf("video = %+v", contents.Video)
	}
	if !contents.Voice.Valid() || contents.Voice.MD5 != "voicemd5" || contents.Voice.Duration != 1.5 {
		t.Errorf("voice = %+v", contents.Voice)
	}

	// objs = 1 text + 1 image + 1 video + 1 voice.
	if len(contents.Objs) != 4 {
		t.Errorf("len(objs) = %d, want 4", len(contents.Objs))
	}
}

func TestRef(t *testing.T) {
	if ref := profile.ByUserID(1); ref.UserID != 1 || ref.Portrait != "" {
		t.Errorf("ByUserID = %+v", ref)
	}
	if ref := profile.ByPortrait("tb.1.x"); ref.Portrait != "tb.1.x" || ref.UserID != 0 {
		t.Errorf("ByPortrait = %+v", ref)
	}
}
