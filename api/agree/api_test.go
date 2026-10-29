package agree

import (
	"testing"
)

func TestContentType(t *testing.T) {
	// 取值来自上游：pid 非零时 1=回复 2=楼中楼，pid 为零时 3=主题帖。
	// 注意这与 enums.ContentType 的编号不是一回事。
	tests := []struct {
		name      string
		pid       int64
		isComment bool
		want      int
	}{
		{"thread", 0, false, 3},
		{"thread even when comment", 0, true, 3},
		{"post", 123, false, 1},
		{"comment", 123, true, 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ContentType(tc.pid, tc.isComment); got != tc.want {
				t.Errorf("ContentType(%d, %v) = %d, want %d", tc.pid, tc.isComment, got, tc.want)
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
