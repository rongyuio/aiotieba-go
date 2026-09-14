package getusercontents

import (
	"testing"

	"github.com/rongyuio/aiotieba/enums"
	"github.com/rongyuio/aiotieba/protobuf"

	pb "github.com/rongyuio/aiotieba/api/get_user_contents/protobuf"
)

func newPostInfoList() *protobuf.PostInfoList {
	return &protobuf.PostInfoList{
		ForumId:      999,
		ThreadId:     888,
		PostId:       777,
		CreateTime:   1600000000,
		ForumName:    "吧名",
		Title:        "标题",
		UserName:     "用户",
		UserId:       123,
		UserPortrait: "tb.1.aaa?t=1234567890",
		NameShow:     "昵称",
		ReplyNum:     5,
		ThreadType:   uint64(enums.ThreadTypeArticle),
		Agree:        &protobuf.Agree{AgreeNum: 10, DisagreeNum: 2},
		FreqNum:      100,
		ShareNum:     3,
	}
}

func TestUserThreadFromProto(t *testing.T) {
	p := newPostInfoList()
	p.FirstPostContent = []*protobuf.PbContent{
		{Type: 0, Text: "hello"},
		{Type: 1, Link: "https://x", Text: "alink"},
	}
	p.Media = []*protobuf.Media{
		{Type: 1, SmallPic: "https://x/aabbccddeeff00112233445566778899.png", Width: 640, Height: 480},
	}

	th := UserThreadFromProto(p)
	if th.FID != 999 || th.TID != 888 || th.PID != 777 || th.FName != "吧名" || th.Title != "标题" {
		t.Errorf("thread ids = %+v", th)
	}
	if th.Type != enums.ThreadTypeArticle {
		t.Errorf("Type = %v, want Article", th.Type)
	}
	if th.Agree != 10 || th.Disagree != 2 || th.ViewNum != 100 || th.ReplyNum != 5 || th.ShareNum != 3 {
		t.Errorf("counters = %+v", th)
	}
	// A link fragment contributes its URL to the text, mirroring FragLink.text.
	if th.Contents.Text() != "hellohttps://x" {
		t.Errorf("contents text = %q", th.Contents.Text())
	}
	if len(th.Contents.Imgs) != 1 || th.Contents.Imgs[0].Hash != "aabbccddeeff00112233445566778899" {
		t.Errorf("imgs = %+v", th.Contents.Imgs)
	}
}

func TestUserPostFromProto(t *testing.T) {
	p := newPostInfoList()
	p.Content = []*protobuf.PostInfoList_PostInfoContent{
		{
			PostId:     777,
			PostType:   1,
			CreateTime: 1600000000,
			PostContent: []*protobuf.PostInfoList_PostInfoContent_Abstract{
				{Type: 0, Text: "正文"},
				{Type: 1, Link: "https://x", Text: "t"},
				{Type: 10, VoiceMd5: "abc", DuringTime: "1500"},
			},
		},
	}

	posts := UserPostsFromProto(p)
	if posts.FID != 999 || posts.TID != 888 {
		t.Errorf("posts ids = %+v", posts)
	}
	if len(posts.Objs) != 1 {
		t.Fatalf("len(Objs) = %d, want 1", len(posts.Objs))
	}
	up := posts.Objs[0]
	if up.PID != 777 || !up.IsComment || up.CreateTime != 1600000000 {
		t.Errorf("post = %+v", *up)
	}
	if up.Contents.Text() != "正文https://x" {
		t.Errorf("contents text = %q", up.Contents.Text())
	}
	if up.Contents.Voice.MD5 != "abc" || up.Contents.Voice.Duration != 1.5 {
		t.Errorf("voice = %+v", up.Contents.Voice)
	}
}

func TestUserInfoUFromProto(t *testing.T) {
	p := newPostInfoList()
	u := UserInfoUFromProto(p)
	if u.UserID != 123 || u.UserName != "用户" || u.NickNameNew != "昵称" {
		t.Errorf("user = %+v", u)
	}
	if u.Portrait != "tb.1.aaa" {
		t.Errorf("portrait = %q, want tb.1.aaa", u.Portrait)
	}
}

func TestEmptyParse(t *testing.T) {
	// The parse helpers must not panic on an empty data response.
	empty := UserPostssFromProto(&pb.UserPostResIdl_DataRes{})
	if len(empty.Objs) != 0 {
		t.Errorf("empty postss = %+v", empty)
	}
	emptyThreads := UserThreadsFromProto(&pb.UserPostResIdl_DataRes{})
	if len(emptyThreads.Objs) != 0 {
		t.Errorf("empty threads = %+v", emptyThreads)
	}
}
