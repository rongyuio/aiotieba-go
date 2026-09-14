package login

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
//
// 返回账号信息和 tbs token。
func ParseBody(body []byte) (UserInfoLogin, string, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return UserInfoLogin{}, "", err
	}
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return UserInfoLogin{}, "", &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}

	user := UserInfoLoginFromJSON(helper.JSONMap(res, "user"))
	tbs := helper.JSONStr(helper.JSONMap(res, "anti"), "tbs")
	return user, tbs, nil
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   consts.AppBaseHost,
		Path:   "/c/s/login",
	}
}

// Request 执行 app 表单请求，对应 request。
func Request(ctx context.Context, httpCore *core.HttpCore) (UserInfoLogin, string, error) {
	data := []crypto.Param{
		{Key: "_client_version", Value: consts.LatestVersion},
		{Key: "bdusstoken", Value: httpCore.Account.BDUSS()},
	}

	resp, err := httpCore.AppForm(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return UserInfoLogin{}, "", err
	}
	return ParseBody(resp.Body())
}
