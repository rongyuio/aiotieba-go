package followforum

import (
	"errors"
	"testing"

	"github.com/rongyuio/aiotieba/exception"
)

func TestParseBodySuccess(t *testing.T) {
	tests := []string{
		`{"error_code":0}`,
		`{"error_code":0,"error":{"errno":0}}`,
	}
	for _, body := range tests {
		if err := ParseBody([]byte(body)); err != nil {
			t.Errorf("ParseBody(%s) = %v, want nil", body, err)
		}
	}
}

func TestParseBodyNestedError(t *testing.T) {
	// The endpoint may nest a second failure envelope under `error`.
	err := ParseBody([]byte(`{"error_code":0,"error":{"errno":340006,"errmsg":"boom"}}`))
	var serverErr *exception.TiebaServerError
	if !errors.As(err, &serverErr) {
		t.Fatalf("error = %v (%T), want *exception.TiebaServerError", err, err)
	}
	if serverErr.Code != 340006 || serverErr.Msg != "boom" {
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
	if u.Host != "tiebac.baidu.com" || u.Path != "/c/c/forum/like" {
		t.Errorf("url = %s", u)
	}
}
