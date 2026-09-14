// Package signforums 实现 aiotieba 的 sign_forums API。
//
// 对应 Python 包 aiotieba.api.sign_forums。
package signforums

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// SubappType 是请求体与 Subapp-Type 头中发送的值。
const SubappType = "hybrid"

// ParseBody 对应 parse_body。
//
// 除常规的 error_code/error_msg 之外，该接口还可能在 `error` 下嵌套第二层错误信息。
func ParseBody(body []byte) error {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return err
	}
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}
	if helper.HasJSONKey(res, "error") {
		errObj := helper.JSONMap(res, "error")
		if code := helper.JSONInt(errObj, "errno"); code != 0 {
			return &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(errObj, "errmsg")}
		}
	}
	return nil
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/c/c/forum/msign"}
}

// Request 执行网页端表单请求，对应 request。
func Request(ctx context.Context, httpCore *core.HttpCore) error {
	data := []crypto.Param{
		{Key: "_client_version", Value: consts.LatestVersion},
		{Key: "subapp_type", Value: SubappType},
	}

	resp, err := httpCore.WebForm(data).SetHeader("Subapp-Type", SubappType).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return err
	}
	return ParseBody(resp.Body())
}
