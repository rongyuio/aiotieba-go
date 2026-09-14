package getthreads

import (
	"bytes"
	"encoding/hex"
	"errors"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/get_threads/protobuf"
)

// The wire bytes below were produced by the Python bindings generated from the
// same .proto files.

const wantReqPn1 = "0a270a0ce5a4a9e5a082e9b8a1e6b1a4101e18232001ba020d0802120931322e36342e312e31f80206"

const resWire = "0a0012ec01122c08b960120ce5a4a9e5a082e9b8a1e6b1a41a06e7949fe6b4bb2206e68385e6849f486450c80158ac028a0100220b081e180120ac02280a30013a62086f1a06e6a087e9a2982003280738e4e2cfaa065001c002de01e80280e2cfaa06c003e707f20704080a2001b80801f20807120568656c6c6ff2080c08041206407573657220787bf20814080112057469746c651a0968747470733a2f2f78f80a058a013610e7071a05756e616d6522046e69636b2a0874622e312e6162638a01070a0569636f6e31b8010cd00201ea020410013801ca08020807aa020c0a0a08051a06e585a8e983a8ca06022001"

const errWire = "0a0a08a6e0141204626f6f6d"

func mustDecode(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("decoding hex: %v", err)
	}
	return b
}

func TestPackProtoMatchesPython(t *testing.T) {
	got := PackProto("天堂鸡汤", 1, 30, 6, true, "12.64.1.1")
	if !bytes.Equal(got, mustDecode(t, wantReqPn1)) {
		t.Errorf("PackProto = %x\n         want %s", got, wantReqPn1)
	}
}

func TestPackProtoFields(t *testing.T) {
	// pn == 1 collapses to 0, every other value is kept.
	for _, pn := range []int32{1, 3, 10} {
		req := &pb.FrsPageReqIdl{}
		if err := proto.Unmarshal(PackProto("天堂鸡汤", pn, 30, 6, true, "12.64.1.1"), req); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		data := req.GetData()
		wantPn := pn
		if pn == 1 {
			wantPn = 0
		}
		if data.GetPn() != wantPn {
			t.Errorf("pn(%d) = %d, want %d", pn, data.GetPn(), wantPn)
		}
		if data.GetKw() != "天堂鸡汤" {
			t.Errorf("kw = %q", data.GetKw())
		}
		if data.GetRn() != 30 || data.GetRnNeed() != 35 {
			t.Errorf("rn/rn_need = %d/%d, want 30/35", data.GetRn(), data.GetRnNeed())
		}
		if data.GetIsGood() != 1 || data.GetSortType() != 6 {
			t.Errorf("is_good/sort_type = %d/%d, want 1/6", data.GetIsGood(), data.GetSortType())
		}
		if data.GetCommon().GetXClientType() != 2 || data.GetCommon().GetXClientVersion() != "12.64.1.1" {
			t.Errorf("common = %+v", data.GetCommon())
		}
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

func TestParseBodyMatchesPython(t *testing.T) {
	threads, err := ParseBody(mustDecode(t, resWire))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}

	// Forum
	if threads.Forum.FID != 12345 || threads.Forum.FName != "天堂鸡汤" {
		t.Errorf("forum = %+v", threads.Forum)
	}
	if threads.Forum.Category != "生活" || threads.Forum.Subcategory != "情感" {
		t.Errorf("forum categories = %q/%q", threads.Forum.Category, threads.Forum.Subcategory)
	}
	if threads.Forum.MemberNum != 100 || threads.Forum.PostNum != 300 || threads.Forum.ThreadNum != 200 {
		t.Errorf("forum counts = %+v", threads.Forum)
	}
	if !threads.Forum.HasBawu || !threads.Forum.HasRule {
		t.Errorf("forum flags = %+v", threads.Forum)
	}

	// Page
	if threads.Page.PageSize != 30 || threads.Page.CurrentPage != 1 || threads.Page.TotalPage != 10 ||
		threads.Page.TotalCount != 300 {
		t.Errorf("page = %+v", threads.Page)
	}
	if !threads.Page.HasMore || threads.Page.HasPrev {
		t.Errorf("page flags = %+v", threads.Page)
	}
	if !threads.HasMore() {
		t.Error("HasMore() = false, want true")
	}

	// Tab map
	if got := threads.TabMap["全部"]; got != 5 {
		t.Errorf("tab_map[全部] = %d, want 5", got)
	}

	// Threads
	if threads.Len() != 1 {
		t.Fatalf("thread count = %d, want 1", threads.Len())
	}
	th := threads.Objs[0]

	if th.TID != 111 || th.PID != 222 || th.Title != "标题" || th.AuthorID != 999 {
		t.Errorf("thread identity = %+v", th)
	}
	if th.FID != 12345 || th.FName != "天堂鸡汤" {
		t.Errorf("thread forum = %d/%q", th.FID, th.FName)
	}
	if th.Type != enums.ThreadTypeArticle {
		t.Errorf("thread type = %d, want %d", th.Type, enums.ThreadTypeArticle)
	}
	if th.TabID != 5 {
		t.Errorf("tab id = %d, want 5", th.TabID)
	}
	if !th.IsGood || th.IsTop || th.IsShare || th.IsHide || th.IsLivepost {
		t.Errorf("thread flags = %+v", th)
	}
	if th.ViewNum != 7 || th.ReplyNum != 3 || th.ShareNum != 1 {
		t.Errorf("thread counters = %+v", th)
	}
	if th.Agree != 10 || th.Disagree != 1 {
		t.Errorf("thread agree = %d/%d, want 10/1", th.Agree, th.Disagree)
	}
	if th.CreateTime != 1700000000 || th.LastTime != 1700000100 {
		t.Errorf("thread times = %d/%d", th.CreateTime, th.LastTime)
	}

	// User
	u := th.User
	if u.UserID != 999 || u.UserName != "uname" || u.NickNameNew != "nick" || u.Portrait != "tb.1.abc" {
		t.Errorf("user = %+v", u)
	}
	if u.Level != 12 || u.GLevel != 7 {
		t.Errorf("user levels = %d/%d", u.Level, u.GLevel)
	}
	if u.Gender != enums.GenderMale {
		t.Errorf("gender = %d, want %d", u.Gender, enums.GenderMale)
	}
	if len(u.Icons) != 1 || u.Icons[0] != "icon1" {
		t.Errorf("icons = %v", u.Icons)
	}
	if u.PrivLike != enums.PrivLikePublic || u.PrivReply != enums.PrivReplyAll {
		t.Errorf("priv = %d/%d", u.PrivLike, u.PrivReply)
	}
	if u.NickName() != "nick" || u.ShowName() != "nick" || u.LogName() != "uname" {
		t.Errorf("user names = %q/%q/%q", u.NickName(), u.ShowName(), u.LogName())
	}

	// Contents
	c := th.Contents
	if len(c.Texts) != 3 {
		t.Errorf("text fragment count = %d, want 3", len(c.Texts))
	}
	if len(c.Ats) != 1 || c.Ats[0].UserID != 123 {
		t.Errorf("ats = %+v", c.Ats)
	}
	if len(c.Links) != 1 || c.Links[0].Text != "https://x" || c.Links[0].Title != "title" {
		t.Errorf("links = %+v", c.Links)
	}
	if got := c.Text(); got != "hello@user https://x" {
		t.Errorf("contents text = %q, want %q", got, "hello@user https://x")
	}
	if got := th.Text(); got != "标题\nhello@user https://x" {
		t.Errorf("thread text = %q", got)
	}
}

func TestParseBodySurfacesServerError(t *testing.T) {
	_, err := ParseBody(mustDecode(t, errWire))
	if err == nil {
		t.Fatal("ParseBody with a server error: want error, got nil")
	}
	var serverErr *exception.TiebaServerError
	if !errors.As(err, &serverErr) {
		t.Fatalf("error type = %T, want *exception.TiebaServerError", err)
	}
	if serverErr.Code != 340006 || serverErr.Msg != "boom" {
		t.Errorf("server error = %+v, want code=340006 msg=boom", serverErr)
	}
}
