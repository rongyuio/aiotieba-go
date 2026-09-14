package login

import (
	"errors"
	"testing"

	"github.com/rongyuio/aiotieba-go/exception"
)

func TestParseBody(t *testing.T) {
	user, tbs, err := ParseBody([]byte(`{"error_code":0,"user":{"id":123,"portrait":"tb.1.x","name":"u"},"anti":{"tbs":"abc"}}`))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if user.UserID != 123 || user.Portrait != "tb.1.x" || user.UserName != "u" {
		t.Errorf("user = %+v", user)
	}
	if tbs != "abc" {
		t.Errorf("tbs = %q, want abc", tbs)
	}
	if !user.Valid() {
		t.Error("Valid() = false, want true")
	}
	if got := user.String(); got != "u" {
		t.Errorf("String() = %q, want u", got)
	}
}

func TestParseBodyServerError(t *testing.T) {
	_, _, err := ParseBody([]byte(`{"error_code":340006,"error_msg":"boom"}`))
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
	if u.Scheme != "https" || u.Host != "tiebac.baidu.com" || u.Path != "/c/s/login" {
		t.Errorf("url = %s", u)
	}
}
