package getforum

import (
	"errors"
	"testing"

	"github.com/rongyuio/aiotieba-go/exception"
)

const forumBody = `{"error_code":0,"error_msg":"","forum":{"id":12345,"name":"天堂鸡汤",` +
	`"first_class":"生活","second_class":"情感","avatar":"avatar.jpg","slogan":"吧标语",` +
	`"member_num":100,"post_num":300,"thread_num":200,"managers":[]}}`

func TestParseBody(t *testing.T) {
	forum, err := ParseBody([]byte(forumBody))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if forum.FID != 12345 || forum.FName != "天堂鸡汤" {
		t.Errorf("forum identity = %+v", forum)
	}
	if forum.Category != "生活" || forum.Subcategory != "情感" {
		t.Errorf("forum categories = %+v", forum)
	}
	if forum.SmallAvatar != "avatar.jpg" || forum.Slogan != "吧标语" {
		t.Errorf("forum avatar/slogan = %+v", forum)
	}
	if forum.MemberNum != 100 || forum.PostNum != 300 || forum.ThreadNum != 200 {
		t.Errorf("forum counts = %+v", forum)
	}
	if !forum.HasBawu {
		t.Error("HasBawu = false, want true (the managers key is present)")
	}
}

func TestParseBodyWithoutManagers(t *testing.T) {
	body := `{"error_code":0,"forum":{"id":1,"name":"n","first_class":"a","second_class":"b",` +
		`"avatar":"v","slogan":"s","member_num":0,"post_num":0,"thread_num":0}}`
	forum, err := ParseBody([]byte(body))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if forum.HasBawu {
		t.Error("HasBawu = true, want false (the managers key is missing)")
	}
}

func TestParseBodyServerError(t *testing.T) {
	_, err := ParseBody([]byte(`{"error_code":340006,"error_msg":"boom"}`))
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
	if u.Host != "tiebac.baidu.com" || u.Path != "/c/f/frs/frsBottom" {
		t.Errorf("url = %s", u)
	}
}
