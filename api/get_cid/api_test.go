package getcid

import (
	"errors"
	"testing"

	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
)

func TestParseBody(t *testing.T) {
	body := `{"error_code":0,"cates":[{"class_name":"原创","class_id":1},{"class_name":"转载","class_id":2}]}`
	cates, err := ParseBody([]byte(body))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if len(cates) != 2 {
		t.Fatalf("len(cates) = %d, want 2", len(cates))
	}
	if got := helper.JSONStr(cates[0], "class_name"); got != "原创" {
		t.Errorf("class name = %q, want 原创", got)
	}
	if got := helper.JSONInt(cates[1], "class_id"); got != 2 {
		t.Errorf("class id = %d, want 2", got)
	}
}

func TestParseBodyEmpty(t *testing.T) {
	cates, err := ParseBody([]byte(`{"error_code":0}`))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if len(cates) != 0 {
		t.Errorf("len(cates) = %d, want 0", len(cates))
	}
}

func TestParseBodyServerError(t *testing.T) {
	_, err := ParseBody([]byte(`{"error_code":340006,"error_msg":"boom"}`))
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
	if u.Host != "tiebac.baidu.com" || u.Path != "/c/c/bawu/goodlist" {
		t.Errorf("url = %s", u)
	}
}
