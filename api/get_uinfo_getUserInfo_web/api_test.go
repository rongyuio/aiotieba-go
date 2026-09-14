package getuserinfoweb

import (
	"errors"
	"testing"

	"github.com/rongyuio/aiotieba/exception"
)

func TestParseBody(t *testing.T) {
	body := `{"errno":0,"chatUser":{"uid":123,"portrait":"tb.1.x","uname":"u","show_nickname":"昵称"}}`
	user, err := ParseBody([]byte(body))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if user.UserID != 123 || user.Portrait != "tb.1.x" || user.UserName != "u" {
		t.Errorf("user = %+v", user)
	}
	if user.NickNameNew != "昵称" || user.NickName() != "昵称" {
		t.Errorf("nick name = %+v", user)
	}
	if got := user.ShowName(); got != "昵称" {
		t.Errorf("ShowName() = %q", got)
	}
	if got := user.String(); got != "u" {
		t.Errorf("String() = %q", got)
	}
	if got := user.LogName(); got != "u" {
		t.Errorf("LogName() = %q", got)
	}
}

func TestFromJSONEchoedUID(t *testing.T) {
	// When the server echoes the uid as the user name (both strings) the name is
	// dropped, mirroring the Python `!=` guard.
	user := UserInfoGuinfoWebFromJSON(map[string]any{
		"uid": "123", "uname": "123", "portrait": "tb.1.x",
	})
	if user.UserName != "" {
		t.Errorf("user name = %q, want empty", user.UserName)
	}

	// A numeric uid never equals the string name under Python's `!=`.
	user = UserInfoGuinfoWebFromJSON(map[string]any{
		"uid": float64(123), "uname": "123", "portrait": "tb.1.x",
	})
	if user.UserName != "123" {
		t.Errorf("user name = %q, want 123", user.UserName)
	}
}

func TestParseBodyServerError(t *testing.T) {
	_, err := ParseBody([]byte(`{"errno":340006,"errmsg":"boom"}`))
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
	if u.Scheme != "http" || u.Host != "tieba.baidu.com" || u.Path != "/im/pcmsg/query/getUserInfo" {
		t.Errorf("url = %s", u)
	}
}
