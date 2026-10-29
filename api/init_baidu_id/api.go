// Package initbaiduid 实现 aiotieba 的 init_baidu_id API。
//
// 对应 Python 包 aiotieba.api.init_baidu_id。
package initbaiduid

import (
	"context"
	"net/http"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// CookieKey 是承载 BAIDUID 的 cookie 名。
const CookieKey = "BAIDUID"

// ParseBody 解析 JSON 响应，对应 parse_body。
func ParseBody(body []byte) error {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return err
	}
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}
	return nil
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.AppBaseHost, Path: "/c/c/forum/statClientStartTimes"}
}

// Request 执行请求并从 Set-Cookie 里取出 BAIDUID，对应 request。
//
// 该接口只用来种 cookie：响应体固定，真正有价值的是 BAIDUID。
func Request(ctx context.Context, httpCore *core.HttpCore) (string, error) {
	data := []crypto.Param{{Key: "type", Value: "1"}}

	resp, err := httpCore.AppForm(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return "", err
	}
	if resp.StatusCode() != http.StatusOK {
		return "", &exception.HTTPStatusError{Code: resp.StatusCode(), Msg: resp.Status()}
	}
	if err := ParseBody(resp.Body()); err != nil {
		return "", err
	}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == CookieKey {
			return cookie.Value, nil
		}
	}
	return "", &exception.TiebaValueError{Msg: "missing " + CookieKey + " in response cookies"}
}
