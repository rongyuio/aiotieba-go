package agree

import (
	"testing"
)

func TestObjType(t *testing.T) {
	tests := []struct {
		name      string
		pid       int64
		isComment bool
		want      int
	}{
		{"thread", 0, false, objTypeThread},
		{"thread even when comment", 0, true, objTypeThread},
		{"post", 123, false, objTypePost},
		{"comment", 123, true, objTypeComment},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ObjType(tc.pid, tc.isComment); got != tc.want {
				t.Errorf("ObjType(%d, %v) = %d, want %d", tc.pid, tc.isComment, got, tc.want)
			}
		})
	}
}

func TestAgreeType(t *testing.T) {
	if got := AgreeType(false); got != agreeTypePositive {
		t.Errorf("AgreeType(false) = %d, want %d", got, agreeTypePositive)
	}
	if got := AgreeType(true); got != agreeTypeNegative {
		t.Errorf("AgreeType(true) = %d, want %d", got, agreeTypeNegative)
	}
}

func TestParseBodyServerError(t *testing.T) {
	if err := ParseBody([]byte(`{"error_code":0}`)); err != nil {
		t.Errorf("ParseBody(ok) = %v, want nil", err)
	}
	if err := ParseBody([]byte(`{"error_code":340006,"error_msg":"boom"}`)); err == nil {
		t.Error("ParseBody(err) = nil, want an error")
	}
}

func TestRequestURL(t *testing.T) {
	u := RequestURL()
	if u.Host != "tiebac.baidu.com" || u.Path != "/c/c/agree/opAgree" {
		t.Errorf("url = %s", u)
	}
}
