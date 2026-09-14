// Package exception 定义库中的异常与响应类型。
//
// 对应 Python 模块 aiotieba.exception。
package exception

import "fmt"

// TiebaServerError 贴吧服务器异常。
type TiebaServerError struct {
	Code int // 错误码
	Msg  string
}

func (e *TiebaServerError) Error() string {
	return fmt.Sprintf("tieba server error: code=%d msg=%s", e.Code, e.Msg)
}

// HTTPStatusError 错误的状态码。
type HTTPStatusError struct {
	Code int // 状态码
	Msg  string
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("unexpected http status: code=%d msg=%s", e.Code, e.Msg)
}

// TiebaValueError 意外的字段值。
type TiebaValueError struct {
	Msg string // 错误描述
}

func (e *TiebaValueError) Error() string {
	if e.Msg == "" {
		return "unexpected field value"
	}
	return "unexpected field value: " + e.Msg
}

// ContentTypeError 无法解析响应头中的 content-type。
type ContentTypeError struct {
	Msg string // 错误描述
}

func (e *ContentTypeError) Error() string {
	if e.Msg == "" {
		return "cannot parse content-type"
	}
	return "cannot parse content-type: " + e.Msg
}

// BoolResponse bool 返回值，不是内置 bool 的子类，可能不支持部分 bool 操作。
//
// Err 为 nil 表示成功。
type BoolResponse struct {
	Err error // 捕获的异常
}

// OK 报告底层操作是否成功。
func (r BoolResponse) OK() bool { return r.Err == nil }

// IntResponse int 返回值，是内置 int 的子类。
type IntResponse struct {
	Value int   // 返回值
	Err   error // 捕获的异常
}

// StrResponse str 返回值，是内置 str 的子类。
type StrResponse struct {
	Value string // 返回值
	Err   error  // 捕获的异常
}
