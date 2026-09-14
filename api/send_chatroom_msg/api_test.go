package sendchatroommsg

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/helper"
)

// The getBDUKFromUserID vectors were produced by the Python BLCPCore using the
// same fixed key and IV.

func TestGetBDUKFromUserID(t *testing.T) {
	tests := []struct{ userID, want string }{
		{"4444444", "-zNRlmJddhADRMSDiPFO0g"},
		{"123", "xZhOYgSe8wj46ZRgsWPI4Q"},
	}
	for _, tc := range tests {
		if got := BDUKFromUserID(tc.userID); got != tc.want {
			t.Errorf("BDUKFromUserID(%q) = %q, want %q", tc.userID, got, tc.want)
		}
	}
}

func newBLCP(t *testing.T) *core.BLCPCore {
	t.Helper()
	account, err := core.NewAccount(strings.Repeat("b", 192), "")
	if err != nil {
		t.Fatalf("NewAccount: %v", err)
	}
	return core.NewBLCPCore(account, nil, 100)
}

func TestConstructRequestDataStructure(t *testing.T) {
	blcp := newBLCP(t)

	data := ConstructRequestData(blcp, 111, 222, 4444444, 333, "名字", "tb.1.portrait", "正文", 999, 7, true, 12, nil, -1)

	// The fixed fields.
	if got := data["appid"]; got != appID {
		t.Errorf("appid = %v, want %d", got, appID)
	}
	if got := data["sdk_version"]; got != sdkVersion {
		t.Errorf("sdk_version = %v, want %d", got, sdkVersion)
	}
	if got := data["method"]; got != methodID {
		t.Errorf("method = %v, want %d", got, methodID)
	}
	if got := data["type"]; got != 81 {
		t.Errorf("type = %v, want 81", got)
	}
	if got := data["uk"]; got != int64(222) {
		t.Errorf("uk = %v, want 222", got)
	}
	if got := data["origin_id"]; got != int64(333) {
		t.Errorf("origin_id = %v, want 333", got)
	}
	if got := data["token"]; got != strings.Repeat("b", 192) {
		t.Errorf("token = %q, want the BDUSS", got)
	}

	// content is a triple-nested JSON string: {"text": "<json string>"} where
	// the inner text object itself carries ext/content_body as JSON strings.
	content, ok := data["content"].(string)
	if !ok {
		t.Fatalf("content is %T, want string", data["content"])
	}
	outer, err := helper.ParseJSONMap([]byte(content))
	if err != nil {
		t.Fatalf("parsing outer content: %v", err)
	}

	textStr := helper.JSONStr(outer, "text")
	text, err := helper.ParseJSONMap([]byte(textStr))
	if err != nil {
		t.Fatalf("parsing inner text: %v", err)
	}
	if got := helper.JSONStr(text, "baidu_uk"); got != "-zNRlmJddhADRMSDiPFO0g" {
		t.Errorf("baidu_uk = %q, want the BDUK of 4444444", got)
	}
	if got := helper.JSONStr(text, "room_id"); got != "111" {
		t.Errorf("room_id = %q, want \"111\"", got)
	}
	if got := helper.JSONStr(text, "vip"); got != "1" {
		t.Errorf("vip = %q, want \"1\"", got)
	}
	if got := helper.JSONStr(text, "content_body"); got != `{"text":"正文"}` {
		t.Errorf("content_body = %q", got)
	}

	extStr := helper.JSONStr(text, "ext")
	ext, err := helper.ParseJSONMap([]byte(extStr))
	if err != nil {
		t.Fatalf("parsing ext: %v", err)
	}
	if got := helper.JSONInt(ext, "forum_id"); got != 999 {
		t.Errorf("ext.forum_id = %d, want 999", got)
	}
	if got := helper.JSONInt(ext, "level"); got != 7 {
		t.Errorf("ext.level = %d, want 7", got)
	}
	if got := helper.JSONStr(ext, "user_name"); got != "名字" {
		t.Errorf("ext.user_name = %q", got)
	}
	// robot == -1 produces an empty content object.
	if contentField := ext["content"]; contentField == nil {
		t.Error("ext.content is missing")
	} else if m, ok := contentField.(map[string]any); !ok || len(m) != 0 {
		t.Errorf("ext.content = %v, want an empty object", contentField)
	}

	// The msg_key starts with the BDUK of the user id.
	msgKey, _ := data["msg_key"].(string)
	if !strings.HasPrefix(msgKey, "-zNRlmJddhADRMSDiPFO0g") {
		t.Errorf("msg_key = %q, want a BDUK prefix", msgKey)
	}

	// app_safe_ext is a JSON string carrying the z_id.
	appSafeExt, _ := data["app_safe_ext"].(string)
	if !strings.Contains(appSafeExt, `"zid":""`) {
		t.Errorf("app_safe_ext = %q, want an empty zid", appSafeExt)
	}
}

func TestConstructRequestDataRobot(t *testing.T) {
	blcp := newBLCP(t)
	data := ConstructRequestData(blcp, 1, 2, 3, 4, "n", "tb.1.p", "t", 5, 1, false, 1, nil, 42)

	content, _ := data["content"].(string)
	outer, _ := helper.ParseJSONMap([]byte(content))
	text, _ := helper.ParseJSONMap([]byte(helper.JSONStr(outer, "text")))
	ext, _ := helper.ParseJSONMap([]byte(helper.JSONStr(text, "ext")))

	contentField, ok := ext["content"].(map[string]any)
	if !ok {
		t.Fatalf("ext.content = %v (%T), want an object", ext["content"], ext["content"])
	}
	robotParams, ok := contentField["robot_params"].(map[string]any)
	if !ok {
		t.Fatalf("robot_params = %v", contentField["robot_params"])
	}
	if got := helper.JSONInt(robotParams, "type"); got != 42 {
		t.Errorf("robot_params.type = %d, want 42", got)
	}
}

func TestMarshalJSONCompact(t *testing.T) {
	// MarshalJSON must not add separators, matching json.dumps defaults.
	got := marshalJSON(map[string]any{"a": 1, "b": "x"})
	var m map[string]any
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("marshalJSON produced invalid JSON: %v", err)
	}
	if got != `{"a":1,"b":"x"}` {
		t.Errorf("marshalJSON = %q", got)
	}
}
