// Package fetchchatroommsg 实现 aiotieba 的 fetch_chatroom_msg API。
//
// 对应 Python 包 aiotieba.api.fetch_chatroom_msg。该 API 直连 pim.baidu.com，
// 不走 BLCP 连接。
package fetchchatroommsg

import (
	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper"
)

// ChatroomMessage 吧群消息。
type ChatroomMessage struct {
	MsgID      int64          // 消息id
	MsgType    int64          // 消息类型
	UserID     int64          // 发送者user_id 为0时表示系统消息
	CreateTime int64          // 发送时间 10位时间戳 以秒为单位
	Content    map[string]any // 消息内容 结构随消息类型变化
}

// ChatroomMessageFromJSON 对应 ChatroomMessage.from_json。
//
// content 是双层 JSON：外层字符串解码后取 text 字段，text 本身又是一段 JSON 字符串。
func ChatroomMessageFromJSON(data map[string]any) (ChatroomMessage, error) {
	content, err := ParseContent(helper.JSONStr(data, "content"))
	if err != nil {
		return ChatroomMessage{}, err
	}
	return ChatroomMessage{
		MsgID:      helper.JSONInt(data, "msgid"),
		MsgType:    helper.JSONInt(data, "type"),
		UserID:     helper.JSONInt(data, "from_user"),
		CreateTime: helper.JSONInt(data, "create_time"),
		Content:    content,
	}, nil
}

// ParseContent 把消息 content 字段解成最终的消息体。
//
// 对应 parse_json(parse_json(content)["text"])：content 本身是 JSON 字符串，
// 其 text 成员又是转义后的 JSON 字符串。
func ParseContent(raw string) (map[string]any, error) {
	outer, err := helper.ParseJSONMap([]byte(raw))
	if err != nil {
		return nil, err
	}
	return helper.ParseJSONMap([]byte(helper.JSONStr(outer, "text")))
}

// ChatroomMessages 吧群消息列表。
type ChatroomMessages struct {
	classdef.Containers[ChatroomMessage]
	Err error // 捕获的异常
}
