// Package joinchatroom 实现 aiotieba 的 join_chatroom API。
//
// 对应 Python 包 aiotieba.api.join_chatroom。该 API 仅可通过 BLCP 传输使用。
package joinchatroom

import (
	"fmt"

	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/helper"
)

// 加入聊天室所用的服务号与方法号，对应 BLCPServiceID.MCAST 与 BLCPMethodID.JOIN_CAST。
const (
	serviceID = int64(3)
	methodID  = int64(201)
)

// ConstructRequestData 构造加入聊天室的请求体，对应 request 里的 request_data。
//
// 上游请求体只有 method、mcast_id、appid、uk 四个字段：与发送消息不同，这里不注入
// client_logid 与 rpc。导出以便测试。
func ConstructRequestData(chatroomID, uk int64) map[string]any {
	return map[string]any{
		"method":   methodID,
		"mcast_id": chatroomID,
		"appid":    int64(core.ChatAppID),
		"uk":       uk,
	}
}

// Request 加入聊天室，对应 request。
//
// 先看帧头错误码，再看业务面 err_code，两者任一非零都视为加入失败。
func Request(blcpCore *core.BLCPCore, chatroomID int64) error {
	errorCode, errorText, payload, err := blcpCore.SendLcmRaw(
		serviceID, methodID, ConstructRequestData(chatroomID, blcpCore.UK()),
	)
	if err != nil {
		return err
	}
	if errorCode != 0 {
		return fmt.Errorf("aiotieba: 加入聊天室失败 code=%d text=%q", errorCode, errorText)
	}
	if code := helper.JSONInt(payload, "err_code"); code != 0 {
		return fmt.Errorf("aiotieba: 加入聊天室失败 err_code=%d", code)
	}
	return nil
}
