package getuserinfousercard

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// subappType 与 clientType 随请求发送，对应 Python 模块。
const (
	subappType = "pc"
	clientType = 20
)

// ParseBody 解析响应，对应 parse_body。
func ParseBody(body []byte) (UserInfoUC, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return UserInfoUC{}, err
	}
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return UserInfoUC{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}
	return UserInfoUCFromJSON(helper.JSONMap(helper.JSONMap(res, "data"), "user_info")), nil
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/c/u/pc/userCard"}
}

// Request 执行网页端 GET 请求，对应 request。
//
// 参数使用 PC_SALT 而非 APP_SALT 签名。
func Request(ctx context.Context, httpCore *core.HttpCore, portrait string) (UserInfoUC, error) {
	params := crypto.Sign([]crypto.Param{
		{Key: "portrait", Value: portrait},
		{Key: "subapp_type", Value: subappType},
		{Key: "_client_type", Value: clientType},
	}, []byte(crypto.PCSalt))

	resp, err := httpCore.WebGet(params, nil).SetContext(ctx).Get(RequestURL().String())
	if err != nil {
		return UserInfoUC{}, err
	}
	return ParseBody(resp.Body())
}
