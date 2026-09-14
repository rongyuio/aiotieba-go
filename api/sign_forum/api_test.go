package signforum

import (
	"errors"
	"testing"

	"github.com/rongyuio/aiotieba-go/exception"
)

func TestParseBodySuccess(t *testing.T) {
	if err := ParseBody([]byte(`{"error_code":0,"user_info":{"sign_bonus_point":8}}`)); err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
}

func TestParseBodyAlreadySigned(t *testing.T) {
	// A zero sign bonus means the forum was already signed today.
	err := ParseBody([]byte(`{"error_code":0,"user_info":{"sign_bonus_point":0}}`))
	var valueErr *exception.TiebaValueError
	if !errors.As(err, &valueErr) {
		t.Fatalf("error = %v (%T), want *exception.TiebaValueError", err, err)
	}
	if valueErr.Msg != "sign_bonus_point is 0" {
		t.Errorf("message = %q", valueErr.Msg)
	}
}

func TestParseBodyServerError(t *testing.T) {
	err := ParseBody([]byte(`{"error_code":340006,"error_msg":"boom"}`))
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
	if u.Host != "tiebac.baidu.com" || u.Path != "/c/c/forum/sign" {
		t.Errorf("url = %s", u)
	}
}
