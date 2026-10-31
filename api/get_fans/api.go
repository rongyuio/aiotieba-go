package getfans

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
func ParseBody(body []byte) (Fans, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return Fans{}, err
	}
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return Fans{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}
	return FansFromJSON(res), nil
}

// RequestURL 返回该 API 的请求地址。
//
// 该接口挂在网页主域上，不走移动端 API 域。
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/c/u/fans/page"}
}

// Request 执行 app 请求，对应 request。
//
// 该接口挂在网页主域上，必须走 web 会话：app 会话有 hostOverride 中间件，
// 会把 Host 强制改写成移动端 API 域，导致服务端返回空列表。
// 签名仍使用 APP 盐值，与客户端一致。
func Request(ctx context.Context, httpCore *core.HttpCore, userID, pn int64) (Fans, error) {
	data := []crypto.Param{
		{Key: "pn", Value: pn},
		{Key: "uid", Value: userID},
		{Key: "follow_list_switch", Value: 1},
		{Key: "subapp_type", Value: "client_fe"},
		{Key: "_client_type", Value: "2"},
		{Key: "_client_version", Value: consts.LatestVersion},
	}
	signed := crypto.Sign(data, []byte(crypto.AppSalt))
	resp, err := httpCore.WebGet(signed, nil).SetContext(ctx).Get(RequestURL().String())
	if err != nil {
		return Fans{}, err
	}
	return ParseBody(resp.Body())
}
