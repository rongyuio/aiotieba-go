package recommend

import (
	"errors"
	"testing"

	"github.com/rongyuio/aiotieba/exception"
)

func TestParseBodySuccess(t *testing.T) {
	if err := ParseBody([]byte(`{"error_code":0,"data":{"is_push_success":1}}`)); err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
}

func TestParseBodyPushFailure(t *testing.T) {
	// The endpoint reports its own outcome under data.is_push_success.
	err := ParseBody([]byte(`{"error_code":0,"data":{"is_push_success":0,"msg":"no permission"}}`))
	var serverErr *exception.TiebaServerError
	if !errors.As(err, &serverErr) {
		t.Fatalf("error = %v (%T), want *exception.TiebaServerError", err, err)
	}
	if serverErr.Code != 0 || serverErr.Msg != "no permission" {
		t.Errorf("server error = %+v", serverErr)
	}
}

func TestParseBodyTransportError(t *testing.T) {
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
	if u.Host != "tiebac.baidu.com" || u.Path != "/c/c/bawu/pushRecomToPersonalized" {
		t.Errorf("url = %s", u)
	}
}
