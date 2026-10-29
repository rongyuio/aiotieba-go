package getchatroomdetail

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// ParseBody 解析 JSON 响应，对应 parse_body。
func ParseBody(body []byte) (ChatroomDetail, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return ChatroomDetail{}, err
	}
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return ChatroomDetail{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}
	return ChatroomDetailFromJSON(helper.JSONMap(res, "data"))
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.AppBaseHost, Path: "/c/f/chat/getChatRoomDetail"}
}

// Request 执行网页端 GET 请求，对应 request。
//
// lastMsgID 与 lastTime 为 0 时不参与请求，与 Python 版一致。
func Request(
	ctx context.Context,
	httpCore *core.HttpCore,
	chatroomID, lastMsgID, lastTime int64,
) (ChatroomDetail, error) {
	params := []crypto.Param{
		{Key: "_client_version", Value: consts.ChatVersion},
		{Key: "chatroom_id", Value: chatroomID},
	}
	if lastMsgID != 0 {
		params = append(params, crypto.Param{Key: "last_msg_id", Value: lastMsgID})
	}
	if lastTime != 0 {
		params = append(params, crypto.Param{Key: "last_time", Value: lastTime})
	}
	params = crypto.Sign(params, []byte(crypto.AppSalt))

	resp, err := httpCore.WebGet(params, nil).SetContext(ctx).Get(RequestURL().String())
	if err != nil {
		return ChatroomDetail{}, err
	}
	return ParseBody(resp.Body())
}
