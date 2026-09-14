package addblacklistold

import (
	"errors"
	"testing"

	"github.com/rongyuio/aiotieba-go/exception"
)

func TestParseBodySuccess(t *testing.T) {
	if err := ParseBody([]byte(`{"error_code":0,"errorno":0}`)); err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
}

func TestParseBodyDualEnvelope(t *testing.T) {
	tests := []struct {
		name string
		body string
		code int
		msg  string
	}{
		{"error_code", `{"error_code":340006,"error_msg":"first"}`, 340006, "first"},
		{"errorno", `{"error_code":0,"errorno":1,"errmsg":"second"}`, 1, "second"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ParseBody([]byte(tc.body))
			var serverErr *exception.TiebaServerError
			if !errors.As(err, &serverErr) {
				t.Fatalf("error = %v (%T), want *exception.TiebaServerError", err, err)
			}
			if serverErr.Code != tc.code || serverErr.Msg != tc.msg {
				t.Errorf("server error = %+v, want code=%d msg=%s", serverErr, tc.code, tc.msg)
			}
		})
	}
}

func TestRequestURL(t *testing.T) {
	u := RequestURL()
	if u.Host != "tiebac.baidu.com" || u.Path != "/c/c/user/userMuteAdd" {
		t.Errorf("url = %s", u)
	}
}
