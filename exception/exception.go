// Package exception defines the error and response types of the library.
//
// It mirrors the Python module aiotieba.exception.
package exception

import "fmt"

// TiebaServerError is raised when the Tieba server reports an error.
type TiebaServerError struct {
	Code int
	Msg  string
}

func (e *TiebaServerError) Error() string {
	return fmt.Sprintf("tieba server error: code=%d msg=%s", e.Code, e.Msg)
}

// HTTPStatusError is raised when an HTTP status code is unexpected.
type HTTPStatusError struct {
	Code int
	Msg  string
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("unexpected http status: code=%d msg=%s", e.Code, e.Msg)
}

// TiebaValueError signals an unexpected field value.
type TiebaValueError struct {
	Msg string
}

func (e *TiebaValueError) Error() string {
	if e.Msg == "" {
		return "unexpected field value"
	}
	return "unexpected field value: " + e.Msg
}

// ContentTypeError signals that the content-type header could not be parsed.
type ContentTypeError struct {
	Msg string
}

func (e *ContentTypeError) Error() string {
	if e.Msg == "" {
		return "cannot parse content-type"
	}
	return "cannot parse content-type: " + e.Msg
}

// BoolResponse mirrors Python's BoolResponse: a boolean result that also
// carries a captured error. A nil Err means success.
type BoolResponse struct {
	Err error
}

// OK reports whether the underlying operation succeeded.
func (r BoolResponse) OK() bool { return r.Err == nil }

// IntResponse mirrors Python's IntResponse.
type IntResponse struct {
	Value int
	Err   error
}

// StrResponse mirrors Python's StrResponse.
type StrResponse struct {
	Value string
	Err   error
}
