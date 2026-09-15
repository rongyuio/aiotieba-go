package logging

import (
	"errors"
	"testing"
)

// fakeServerError 模拟库内异常，参数与 TiebaServerError 一致。
type fakeServerError struct {
	code int
	msg  string
}

func (e fakeServerError) Error() string { return "tieba server error" }
func (e fakeServerError) PyArgs() []any { return []any{e.code, e.msg} }

// fakeValueError 模拟单参数的库内异常。
type fakeValueError struct{ msg string }

func (e fakeValueError) Error() string { return "unexpected field value" }
func (e fakeValueError) PyArgs() []any { return []any{e.msg} }

// fakeForumRef 模拟 ForumRef / UserRef 的可读渲染。
type fakeForumRef struct{ fname string }

func (r fakeForumRef) PyRepr() string { return PyRepr(r.fname) }

func TestPyRepr(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{"nil", nil, "None"},
		{"true", true, "True"},
		{"false", false, "False"},
		{"int", 42, "42"},
		{"negative int64", int64(-3), "-3"},
		{"uint", uint(7), "7"},
		{"float", 1.5, "1.5"},
		{"plain string", "盗墓笔记", "'盗墓笔记'"},
		{"single quote", "it's", `"it's"`},
		{"single and double quote", `it's "x"`, `'it\'s "x"'`},
		{"backslash", `a\b`, `'a\\b'`},
		{"bytes", []byte("ab"), "'ab'"},
		{"slice", []int{1, 2}, "[1, 2]"},
		{"map", map[string]int{"a": 1}, "{'a': 1}"},
		{"nil pointer", (*int)(nil), "None"},
		{"py reprer", fakeForumRef{fname: "天堂鸡汤"}, "'天堂鸡汤'"},
		{"error", errors.New("boom"), "boom"},
		{"exception args", fakeServerError{code: 340011}, "(340011, '')"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := PyRepr(tc.in); got != tc.want {
				t.Errorf("PyRepr(%#v) = %s, want %s", tc.in, got, tc.want)
			}
		})
	}
}

func TestPyReprEnumLike(t *testing.T) {
	// 具名整型（枚举）必须走 Kind 分支渲染为数字，而不是掉进 default。
	type bawuType int32
	if got := PyRepr(bawuType(3)); got != "3" {
		t.Errorf("PyRepr(bawuType(3)) = %s, want 3", got)
	}
	t.Run("pointer to enum", func(t *testing.T) {
		v := bawuType(4)
		if got := PyRepr(&v); got != "4" {
			t.Errorf("PyRepr(&bawuType(4)) = %s, want 4", got)
		}
	})
}

func TestPyArgs(t *testing.T) {
	tests := []struct {
		name string
		in   []any
		want string
	}{
		{"empty", nil, "args=() kwargs={}"},
		{"single", []any{"盗墓笔记"}, "args=('盗墓笔记',) kwargs={}"},
		{"multiple", []any{int64(123), "x"}, "args=(123, 'x') kwargs={}"},
		{
			"kwargs only",
			[]any{PyKw{Name: "fname", Value: "盗墓笔记"}},
			"args=() kwargs={'fname': '盗墓笔记'}",
		},
		{
			"mixed",
			[]any{int64(123), PyKw{Name: "pn", Value: 2}},
			"args=(123,) kwargs={'pn': 2}",
		},
		{
			"multiple kwargs",
			[]any{PyKw{Name: "pn", Value: 2}, PyKw{Name: "rn", Value: 30}},
			"args=() kwargs={'pn': 2, 'rn': 30}",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := PyArgs(tc.in...); got != tc.want {
				t.Errorf("PyArgs(%v) = %s, want %s", tc.in, got, tc.want)
			}
		})
	}
}

func TestPyErr(t *testing.T) {
	tests := []struct {
		name string
		in   error
		want string
	}{
		{"nil", nil, "None"},
		{"plain error", errors.New("boom"), "boom"},
		// 与示例日志逐字一致: (340011, '')
		{"server error", fakeServerError{code: 340011}, "(340011, '')"},
		{"server error with msg", fakeServerError{code: 4, msg: "x"}, "(4, 'x')"},
		{"single arg exception", fakeValueError{msg: "sign_bonus_point is 0"}, "'sign_bonus_point is 0'"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := PyErr(tc.in); got != tc.want {
				t.Errorf("PyErr(%v) = %s, want %s", tc.in, got, tc.want)
			}
		})
	}
}
