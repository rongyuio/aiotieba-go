package fetchchatroommsg

import (
	"context"
	"math"
	"net/http"
	"time"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// pimBaseURL 是贴吧 IM 参数服务的地址。
const pimBaseURL = "https://pim.baidu.com"

// pimFetchMcastMsgPath 是拉取聊天室消息的路径。
const pimFetchMcastMsgPath = "/rest/3.0/im/fetch_mcast_msg_client"

// pimUA 是所有 pim.baidu.com 请求使用的 User-Agent。
const pimUA = "okhttp/3.11.0"

// 请求参数中的固定值，对应 Python 模块里的常量。
const (
	// msgIDEnd 表示拉取到最新一条为止，即 sys.maxsize。
	msgIDEnd = int64(math.MaxInt64)
	// msgCount 是单次拉取的消息条数。
	msgCount = int64(-60)
	// msgCategory 是消息类别。
	msgCategory = int64(4)
)

// ParseBody 解析 JSON 响应，对应 parse_body。
func ParseBody(body []byte) (ChatroomMessages, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return ChatroomMessages{}, err
	}
	if code := helper.JSONInt(res, "err_code"); code != 0 {
		return ChatroomMessages{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "err_msg")}
	}

	var messages ChatroomMessages
	for _, item := range helper.JSONSlice(res, "messages") {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		msg, err := ChatroomMessageFromJSON(m)
		if err != nil {
			return ChatroomMessages{}, err
		}
		messages.Objs = append(messages.Objs, msg)
	}
	return messages, nil
}

// Request 拉取聊天室消息，对应 request。
func Request(ctx context.Context, httpCore *core.HttpCore, chatroomID int64) (ChatroomMessages, error) {
	cuidGalaxy2, err := httpCore.Account.CuidGalaxy2()
	if err != nil {
		return ChatroomMessages{}, err
	}

	data := []crypto.Param{
		{Key: "appid", Value: core.ChatAppID},
		{Key: "mcast_id", Value: chatroomID},
		{Key: "msgid_end", Value: msgIDEnd},
		{Key: "count", Value: msgCount},
		{Key: "category", Value: msgCategory},
		{Key: "app_version", Value: consts.ChatVersion},
		{Key: "sdk_version", Value: core.ChatSDKVersion},
		{Key: "device_id", Value: cuidGalaxy2},
		{Key: "timestamp", Value: time.Now().Unix()},
	}
	signed := crypto.Sign(data, nil)

	payload := make(map[string]any, len(signed))
	for _, p := range signed {
		payload[p.Key] = p.Value
	}

	body, err := postPIM(ctx, httpCore, pimFetchMcastMsgPath, payload)
	if err != nil {
		return ChatroomMessages{}, err
	}
	return ParseBody(body)
}

// postPIM 向 pim.baidu.com 提交 JSON 请求体并返回原始响应，对应 pack_pim_request + send。
func postPIM(ctx context.Context, httpCore *core.HttpCore, path string, payload map[string]any) ([]byte, error) {
	resp, err := httpCore.NetCore.NewRestyClient().R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", pimUA).
		SetCookie(&http.Cookie{Name: "BDUSS", Value: httpCore.Account.BDUSS()}).
		SetBody(payload).
		Post(pimBaseURL + path)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, &exception.HTTPStatusError{Code: resp.StatusCode(), Msg: resp.Status()}
	}
	return resp.Body(), nil
}
