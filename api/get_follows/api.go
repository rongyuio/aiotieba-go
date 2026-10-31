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
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.AppBaseHost, Path: "/c/u/follow/followList"}
}

// Request 执行 app 请求，对应 request。
//
// 客户端以 GET 发起，并固定带上 tab=0 与 follow_list_switch=1。
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
	resp, err := httpCore.AppGet(data, nil).SetContext(ctx).Get(RequestURL().String())
	if err != nil {
		return Follows{}, err
	}
	return ParseBody(resp.Body())
}
