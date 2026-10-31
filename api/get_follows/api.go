package getfollows

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
func ParseBody(body []byte) (Follows, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return Follows{}, err
	}
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return Follows{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}
	return FollowsFromJSON(res), nil
}

// RequestURL 返回该 API 的请求地址。
//
// 该接口挂在网页主域上，不走移动端 API 域。
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/c/u/follow/followList"}
}

// Request 执行 app 请求，对应 request。
//
// 客户端以 GET 发起，固定带 tab=0 与 follow_list_switch=1。
// 走 web 会话的理由同 get_fans：该路径位于网页主域，app 会话的 hostOverride
// 会把 Host 改写掉，服务端返回 110001。
func Request(ctx context.Context, httpCore *core.HttpCore, userID, pn int64) (Follows, error) {
	data := []crypto.Param{
		{Key: "pn", Value: pn},
		{Key: "tab", Value: 0},
		{Key: "uid", Value: userID},
		{Key: "follow_list_switch", Value: 1},
		{Key: "subapp_type", Value: "client_fe"},
		{Key: "_client_type", Value: "2"},
		{Key: "_client_version", Value: consts.LatestVersion},
	}
	signed := crypto.Sign(data, []byte(crypto.AppSalt))
	resp, err := httpCore.WebGet(signed, nil).SetContext(ctx).Get(RequestURL().String())
	if err != nil {
		return Follows{}, err
	}
	return ParseBody(resp.Body())
}
