package getuinfopanel

import (
	"errors"
	"testing"

	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/exception"
)

const panelBody = `{"no":0,"data":{"portrait":"tb.1.x","name":"某个用户名","show_nickname":"昵称",` +
	`"name_show":"旧昵称","sex":"female","tb_age":"8.5","post_num":"1.2万",` +
	`"followed_count":123,"vipInfo":{"v_status":"3"}}}`

func TestParseBody(t *testing.T) {
	user, err := ParseBody([]byte(panelBody))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if user.Portrait != "tb.1.x" || user.UserName != "某个用户名" {
		t.Errorf("identity = %+v", user)
	}
	if user.NickNameNew != "昵称" || user.NickNameOld != "旧昵称" {
		t.Errorf("nick names = %+v", user)
	}
	if user.Gender != enums.GenderFemale {
		t.Errorf("gender = %v, want GenderFemale", user.Gender)
	}
	if user.Age != 8.5 {
		t.Errorf("age = %v, want 8.5", user.Age)
	}
	// "1.2万" is 1.2 * 1e4.
	if user.PostNum != 12000 {
		t.Errorf("post num = %d, want 12000", user.PostNum)
	}
	if user.FanNum != 123 {
		t.Errorf("fan num = %d, want 123", user.FanNum)
	}
	if !user.IsVIP {
		t.Error("IsVIP = false, want true (v_status is 3)")
	}
	if got := user.NickName(); got != "昵称" {
		t.Errorf("NickName() = %q", got)
	}
	if got := user.ShowName(); got != "昵称" {
		t.Errorf("ShowName() = %q", got)
	}
	if got := user.String(); got != "某个用户名" {
		t.Errorf("String() = %q", got)
	}
	if got := user.LogName(); got != "某个用户名" {
		t.Errorf("LogName() = %q", got)
	}
	if !user.Valid() {
		t.Error("Valid() = false, want true")
	}
}

func TestFromJSONSexMapping(t *testing.T) {
	tests := []struct {
		sex  string
		want enums.Gender
	}{
		{"male", enums.GenderMale},
		{"female", enums.GenderFemale},
		{"", enums.GenderUnknown},
		{"其它", enums.GenderUnknown},
	}
	for _, tc := range tests {
		got := UserInfoPanelFromJSON(map[string]any{"sex": tc.sex}).Gender
		if got != tc.want {
			t.Errorf("sex %q -> %v, want %v", tc.sex, got, tc.want)
		}
	}
}

func TestFromJSONMissingOptionalFields(t *testing.T) {
	// tb_age "-" means unknown and an absent vipInfo means no VIP.
	user := UserInfoPanelFromJSON(map[string]any{"tb_age": "-", "post_num": 42})
	if user.Age != 0 {
		t.Errorf("age = %v, want 0", user.Age)
	}
	if user.PostNum != 42 {
		t.Errorf("post num = %d, want 42", user.PostNum)
	}
	if user.IsVIP {
		t.Error("IsVIP = true, want false")
	}
}

func TestParseBodyServerError(t *testing.T) {
	_, err := ParseBody([]byte(`{"no":340006,"error":"boom"}`))
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
	if u.Scheme != "https" || u.Host != "tieba.baidu.com" || u.Path != "/home/get/panel" {
		t.Errorf("url = %s", u)
	}
}
