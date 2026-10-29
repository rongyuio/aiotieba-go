package fetchchatroommsg

import (
	"encoding/json"
	"errors"
	"strconv"
	"testing"

	"github.com/rongyuio/aiotieba-go/exception"
)

// buildBody 构造一条消息的响应体。
//
// content 在真实响应里是字符串，其 text 成员又是一段转义后的 JSON 字符串，
// 这里用 strconv.Quote 还原同样的双层结构。
func buildBody(innerContent string) []byte {
	content := `{"text":` + strconv.Quote(innerContent) + `}`
	body, err := json.Marshal(map[string]any{
		"err_code": 0,
		"messages": []any{
			map[string]any{
				"msgid":       100,
				"type":        1,
				"from_user":   123456,
				"create_time": 1700000000,
				"content":     content,
			},
		},
	})
	if err != nil {
		panic(err)
	}
	return body
}

func TestParseBody(t *testing.T) {
	messages, err := ParseBody(buildBody(`{"message":"hi","robot_params":{"scene":"x"}}`))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if messages.Len() != 1 {
		t.Fatalf("Len = %d, want 1", messages.Len())
	}

	msg := messages.Objs[0]
	if msg.MsgID != 100 || msg.MsgType != 1 || msg.UserID != 123456 || msg.CreateTime != 1700000000 {
		t.Errorf("message = %+v", msg)
	}
	if msg.Content["message"] != "hi" {
		t.Errorf("content message = %v", msg.Content["message"])
	}
	if params, ok := msg.Content["robot_params"].(map[string]any); !ok || params["scene"] != "x" {
		t.Errorf("content robot_params = %v", msg.Content["robot_params"])
	}
}

func TestParseBodyEmpty(t *testing.T) {
	messages, err := ParseBody([]byte(`{"err_code":0,"messages":[]}`))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if !messages.Empty() {
		t.Errorf("Len = %d, want 0", messages.Len())
	}
}

func TestParseBodyServerError(t *testing.T) {
	_, err := ParseBody([]byte(`{"err_code":1507,"err_msg":"no"}`))
	var serverErr *exception.TiebaServerError
	if !errors.As(err, &serverErr) {
		t.Fatalf("error = %v (%T), want *exception.TiebaServerError", err, err)
	}
	if serverErr.Code != 1507 || serverErr.Msg != "no" {
		t.Errorf("server error = %+v", serverErr)
	}
}

func TestParseBodyInvalidContent(t *testing.T) {
	if _, err := ParseBody(buildBody(`not-json`)); err == nil {
		t.Error("ParseBody accepted content that is not JSON")
	}
}

func TestParseContentNeedsTextMember(t *testing.T) {
	if _, err := ParseContent(`{"other":1}`); err == nil {
		t.Error("ParseContent accepted content without a text member")
	}
}
