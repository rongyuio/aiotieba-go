package getats

import (
	"testing"

	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/helper"
)

func TestUserInfoAtPrivSets(t *testing.T) {
	m, err := helper.ParseJSONMap([]byte(`{
		"id": 1,
		"portrait": "tb.1.a?t=1234567890",
		"name": "u",
		"name_show": "n",
		"priv_sets": {"like": 3, "reply": 5}
	}`))
	if err != nil {
		t.Fatalf("ParseJSONMap: %v", err)
	}
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
	m, _ := helper.ParseJSONMap([]byte(`{"id": 2, "portrait": "tb.1.b", "name": "u2", "name_show": "n2"}`))
	u := UserInfoAtFromJSON(m)
	if u.PrivLike != enums.PrivLikePublic {
		t.Errorf("PrivLike = %v, want Public", u.PrivLike)
	}
	if u.PrivReply != enums.PrivReplyAll {
		t.Errorf("PrivReply = %v, want All", u.PrivReply)
	}
}

func TestAtAuthorID(t *testing.T) {
	m, _ := helper.ParseJSONMap([]byte(`{
		"content": "x", "fname": "f", "thread_id": 1, "post_id": 2,
		"replyer": {"id": 9, "portrait": "tb.1.c", "name": "u", "name_show": "n"},
		"is_floor": "1", "is_first_post": "0", "time": 100
	}`))
	a := AtFromJSON(m)
	if a.AuthorID() != 9 {
		t.Errorf("AuthorID = %d, want 9", a.AuthorID())
	}
	if !a.IsComment || a.IsThread {
		t.Errorf("IsComment = %v, IsThread = %v", a.IsComment, a.IsThread)
	}
}
