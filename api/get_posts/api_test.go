package getposts

import (
	"bytes"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/get_posts/protobuf"
	commonpb "github.com/rongyuio/aiotieba-go/protobuf"
)

// The wire bytes below were produced by the Python bindings generated from the
// same .proto files.

const wantReq1 = "0a1920c0c407681e900101ca010d0802120931322e36342e312e31"
const wantReq3 = "0a1b20c0c40728016802900101ca010d0802120931322e36342e312e31"

const resWire = "0a0012ba031225080c120ce5a4a9e5a082e9b8a1e6b1a43a06e7949fe6b4bb4206e68385e6849f606468ac021a0b081e180120ac02280a300132850108d708180220e4e2cfaa062a081206e9a696e8a18c2a1508011206e6a087e9a2981a0968747470733a2f2f7868027a3a123808ae1112091207e59b9ee5a48d20120f0804120840e69f90e4baba20789a05120a1208203ae58685e5aeb918c8e3cfaa062089064a0208039801f806aa010d220b1209e5b08fe5b0bee5b7b4aa020408072001320b08851af204050a03626f744279086f1a0ce5b896e5ad90e6a087e9a298200592011610e7071a09e6a5bce4b8bbe5908d2206e6a5bce4b8bbe80280e2cfaa06a003de01f2070408212003b80802ea08360a0ce58e9fe5b896e6a087e9a298220ce5a4a9e5a082e9b8a1e6b1a42a03333333380c6a0320800572081206e58e9fe69687c801bc036a3110f8061a09e794a8e688b7e5908d2206e698b5e7a7b02a0874622e312e646566b80105d00202fa07084950e5b19ee59cb06a151089061a06e5b182e4b8bb2a0874622e312e6768696a1610e7071a09e6a5bce4b8bbe5908d2206e6a5bce4b8bb6a0e109a051a09e8a2abe59b9ee5a48da802ab04"

func mustDecode(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("decoding hex: %v", err)
	}
	return b
}

func TestPackProtoMatchesPython(t *testing.T) {
	account, err := core.NewAccount("", "")
	if err != nil {
		t.Fatalf("NewAccount: %v", err)
	}

	t.Run("plain", func(t *testing.T) {
		got := PackProto(account, 123456, 1, 30, 0, false, false, false, 0)
		if !bytes.Equal(got, mustDecode(t, wantReq1)) {
			t.Errorf("PackProto = %x\n         want %s", got, wantReq1)
		}
	})

	t.Run("lz-and-rn-collapse", func(t *testing.T) {
		// rn == 1 collapses to 2 and lz is set.
		got := PackProto(account, 123456, 1, 1, 0, true, false, false, 0)
		if !bytes.Equal(got, mustDecode(t, wantReq3)) {
			t.Errorf("PackProto = %x\n         want %s", got, wantReq3)
		}
	})
}

func TestPackProtoWithComments(t *testing.T) {
	account, err := core.NewAccount(strings.Repeat("b", 192), "")
	if err != nil {
		t.Fatalf("NewAccount: %v", err)
	}

	req := &pb.PbPageReqIdl{}
	if err := proto.Unmarshal(PackProto(account, 123456, 2, 10, 2, true, true, true, 3), req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	data := req.GetData()
	if got := data.GetCommon().GetBDUSS(); got != strings.Repeat("b", 192) {
		t.Errorf("common.BDUSS = %q", got)
	}
	if data.GetWithFloor() != 1 || data.GetFloorSortType() != 1 || data.GetFloorRn() != 3 {
		t.Errorf("floor params = %d/%d/%d, want 1/1/3",
			data.GetWithFloor(), data.GetFloorSortType(), data.GetFloorRn())
	}
	if data.GetLz() != 1 || data.GetR() != 2 || data.GetPn() != 2 || data.GetRn() != 10 || data.GetKz() != 123456 {
		t.Errorf("request params = %+v", data)
	}

	// Without comments the floor parameters and the BDUSS stay unset.
	plain := &pb.PbPageReqIdl{}
	if err := proto.Unmarshal(PackProto(account, 123456, 1, 30, 0, false, false, false, 0), plain); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if plain.GetData().GetWithFloor() != 0 || plain.GetData().GetFloorRn() != 0 {
		t.Errorf("plain request must not set the floor parameters: %+v", plain.GetData())
	}
	if plain.GetData().GetCommon().GetBDUSS() != "" {
		t.Error("plain request must not carry the BDUSS")
	}
}

func TestRequestURL(t *testing.T) {
	u := RequestURL()
	if u.Scheme != "https" || u.Host != "tiebac.baidu.com" || u.Path != "/c/f/pb/page" {
		t.Errorf("url = %s", u)
	}
	if u.RawQuery != "cmd=302001" {
		t.Errorf("query = %q, want cmd=302001", u.RawQuery)
	}
}

func TestParseBodyMatchesPython(t *testing.T) {
	posts, err := ParseBody(mustDecode(t, resWire))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}

	// Forum
	if posts.Forum.FID != 12 || posts.Forum.FName != "天堂鸡汤" {
		t.Errorf("forum = %+v", posts.Forum)
	}
	if posts.Forum.Category != "生活" || posts.Forum.Subcategory != "情感" {
		t.Errorf("forum categories = %q/%q", posts.Forum.Category, posts.Forum.Subcategory)
	}
	if posts.Forum.MemberNum != 100 || posts.Forum.PostNum != 300 {
		t.Errorf("forum counts = %+v", posts.Forum)
	}

	// Page
	if !posts.HasMore() {
		t.Error("HasMore() = false, want true")
	}
	if posts.Page.PageSize != 30 || posts.Page.TotalPage != 10 || posts.Page.TotalCount != 300 {
		t.Errorf("page = %+v", posts.Page)
	}

	// Thread
	th := posts.Thread
	if th.TID != 111 || th.PID != 222 || th.Title != "帖子标题" || th.FID != 12 || th.FName != "天堂鸡汤" {
		t.Errorf("thread = %+v", th)
	}
	if th.Type != enums.ThreadTypeArticle || th.IsShare {
		t.Errorf("thread type/share = %d/%v", th.Type, th.IsShare)
	}
	if th.ViewNum != 555 || th.ReplyNum != 5 || th.ShareNum != 2 || th.Agree != 33 || th.Disagree != 3 {
		t.Errorf("thread counters = %+v", th)
	}
	if th.CreateTime != 1700000000 {
		t.Errorf("thread create time = %d", th.CreateTime)
	}
	if th.AuthorID() != 999 || th.User.NickNameNew != "楼主" {
		t.Errorf("thread author = %+v", th.User)
	}
	if len(th.Contents.Texts) != 1 || th.Contents.Text() != "原文" {
		t.Errorf("thread contents = %+v", th.Contents.Texts)
	}
	if !th.Contents.Video.Valid() || th.Contents.Video.Width != 640 {
		t.Errorf("thread video = %+v", th.Contents.Video)
	}

	// Posts: the bot floor is filtered out.
	if posts.Len() != 1 {
		t.Fatalf("post count = %d, want 1", posts.Len())
	}
	post := posts.Objs[0]
	if post.PID != 1111 || post.Floor != 2 || post.AuthorID != 888 || post.TID != 111 {
		t.Errorf("post identity = %+v", post)
	}
	if post.FID != 12 || post.FName != "天堂鸡汤" {
		t.Errorf("post forum = %d/%q", post.FID, post.FName)
	}
	if post.ReplyNum != 2 || post.Agree != 7 || post.Disagree != 1 || post.CreateTime != 1700000100 {
		t.Errorf("post counters = %+v", post)
	}
	if post.IsThreadAuthor {
		t.Error("post.IsThreadAuthor = true, want false")
	}
	if post.Sign != "小尾巴" {
		t.Errorf("post sign = %q", post.Sign)
	}
	if got := post.Text(); got != "首行https://x\n小尾巴" {
		t.Errorf("post text = %q", got)
	}
	if len(post.Contents.Links) != 1 || post.Contents.Links[0].Title != "标题" {
		t.Errorf("post links = %+v", post.Contents.Links)
	}
	if post.User.UserID != 888 || post.User.UserName != "用户名" || post.User.NickNameNew != "昵称" {
		t.Errorf("post user = %+v", post.User)
	}
	if post.User.Level != 5 || post.User.Gender != enums.GenderFemale || post.User.IP != "IP属地" {
		t.Errorf("post user extra = %+v", post.User)
	}

	// Comment: the leading "回复 @xxx " prefix is stripped.
	if len(post.Comments) != 1 {
		t.Fatalf("comment count = %d, want 1", len(post.Comments))
	}
	comment := post.Comments[0]
	if comment.PID != 2222 || comment.AuthorID != 777 || comment.ReplyToID != 666 {
		t.Errorf("comment identity = %+v", comment)
	}
	if comment.FID != 12 || comment.TID != 111 || comment.PPID != 1111 || comment.Floor != 2 {
		t.Errorf("comment context = %+v", comment)
	}
	if comment.Agree != 3 || comment.Disagree != 0 || comment.CreateTime != 1700000200 {
		t.Errorf("comment counters = %+v", comment)
	}
	if got := comment.Text(); got != "内容" {
		t.Errorf("comment text = %q, want 内容", got)
	}
	if len(comment.Contents.Ats) != 0 {
		t.Errorf("comment ats = %+v, want none (the leading @ is stripped)", comment.Contents.Ats)
	}
	if len(comment.Contents.Objs) != 1 {
		t.Errorf("comment objs = %+v, want one fragment", comment.Contents.Objs)
	}
	if comment.User.UserID != 777 || comment.User.UserName != "层主" {
		t.Errorf("comment user = %+v", comment.User)
	}
	if comment.IsThreadAuthor {
		t.Error("comment.IsThreadAuthor = true, want false")
	}
}

func TestParseBodySurfacesServerError(t *testing.T) {
	res := &pb.PbPageResIdl{Error: &commonpb.Error{Errorno: 340006, Errmsg: "boom"}}
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
