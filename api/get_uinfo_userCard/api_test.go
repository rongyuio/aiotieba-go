package getuserinfousercard

import (
	"errors"
	"testing"

	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/exception"
)

const userCardBody = `{"error_code":0,"error_msg":"","data":{"user_info":{"portrait":"tb.1.abcdefghijklmnopqrstuv?1739164613",` +
	`"name_show":"昵称","tieba_uid":"987654321","sex":1,"tb_age":"8.5","total_agree_num":3000000000,` +
	`"fans_num":7,"concern_num":8,"intro":"个性签名","ip_address":"浙江"}}}`

func TestParseBody(t *testing.T) {
	user, err := ParseBody([]byte(userCardBody))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	// Python does `portrait[:-13]` when the portrait carries a "?" suffix.
	if user.Portrait != "tb.1.abcdefghijklmnopqrst" {
		t.Errorf("portrait = %q", user.Portrait)
	}
	if user.NickNameNew != "昵称" || user.NickName() != "昵称" || user.ShowName() != "昵称" {
		t.Errorf("nick name = %+v", user)
	}
	if user.TiebaUID != 987654321 {
		t.Errorf("tieba uid = %d, want 987654321 (decoded from a string)", user.TiebaUID)
	}
	if user.Gender != enums.GenderMale {
		t.Errorf("gender = %v, want GenderMale", user.Gender)
	}
	if user.Age != 8.5 {
		t.Errorf("age = %v, want 8.5", user.Age)
	}
	if user.AgreeNum != 3000000000 {
		t.Errorf("agree num = %d, want 3000000000", user.AgreeNum)
	}
	if user.FanNum != 7 || user.FollowNum != 8 {
		t.Errorf("counters = %+v", user)
	}
	if user.Sign != "个性签名" || user.IP != "浙江" {
		t.Errorf("sign/ip = %q/%q", user.Sign, user.IP)
	}
	if got := user.String(); got != "昵称" {
		t.Errorf("String() = %q", got)
	}
	if !user.Valid() {
		t.Error("Valid() = false, want true")
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
	if u.Scheme != "https" || u.Host != "tieba.baidu.com" || u.Path != "/c/u/pc/userCard" {
		t.Errorf("url = %s", u)
	}
}
