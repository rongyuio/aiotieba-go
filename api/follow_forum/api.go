// Package followforum 实现 aiotieba 的 follow_forum API。
//
// 对应 Python 包 aiotieba.api.follow_forum。
package followforum

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

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
		if code := helper.JSONInt(helper.JSONMap(res, "error"), "errno"); code != 0 {
			return &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(helper.JSONMap(res, "error"), "errmsg")}
		}
	}
	return nil
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.AppBaseHost, Path: "/c/c/forum/like"}
}

// Request 执行 app 表单请求，对应 request。
func Request(ctx context.Context, httpCore *core.HttpCore, fid int64) error {
	data := []crypto.Param{
		{Key: "BDUSS", Value: httpCore.Account.BDUSS()},
		{Key: "fid", Value: fid},
		{Key: "tbs", Value: httpCore.Account.Tbs()},
	}

	resp, err := httpCore.AppForm(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return err
	}
	return ParseBody(resp.Body())
}
