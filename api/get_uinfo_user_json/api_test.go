package getuserjson

import (
	"errors"
	"testing"

	"github.com/rongyuio/aiotieba-go/exception"
)

func TestParseBody(t *testing.T) {
	user, err := ParseBody([]byte(`{"creator":{"id":123,"portrait":"tb.1.x"}}`))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if user.UserID != 123 || user.Portrait != "tb.1.x" {
		t.Errorf("user = %+v", user)
	}
	// The endpoint does not echo the name; the caller fills it in.
	if user.UserName != "" {
		t.Errorf("user name = %q, want empty", user.UserName)
	}
	if got := user.String(); got != "tb.1.x" {
		t.Errorf("String() = %q", got)
	}
	if !user.Valid() {
		t.Error("Valid() = false, want true")
	}
}

func TestParseBodyEmpty(t *testing.T) {
	for _, body := range [][]byte{nil, {}} {
		_, err := ParseBody(body)
		var valueErr *exception.TiebaValueError
		if !errors.As(err, &valueErr) {
			t.Fatalf("error = %v (%T), want *exception.TiebaValueError", err, err)
		}
		if valueErr.Msg != "Empty body" {
			t.Errorf("message = %q, want Empty body", valueErr.Msg)
		}
	}
}

func TestParseBodyIgnoresInvalidUTF8(t *testing.T) {
	// A stray invalid byte must not break decoding, mirroring errors="ignore".
	body := append([]byte(`{"creator":{"id":1,"portrait":"tb.`), 0xFF)
	body = append(body, []byte(`1.x"}}`)...)

	user, err := ParseBody(body)
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if user.UserID != 1 || user.Portrait != "tb.1.x" {
		t.Errorf("user = %+v", user)
	}
}

func TestRequestURL(t *testing.T) {
	u := RequestURL()
	if u.Scheme != "http" || u.Host != "tieba.baidu.com" || u.Path != "/i/sys/user_json" {
		t.Errorf("url = %s", u)
	}
}
