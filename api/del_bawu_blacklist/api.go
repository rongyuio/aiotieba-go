// Package delbawublacklist 实现 aiotieba 的 del_bawu_blacklist API。
//
// 对应 Python 包 aiotieba.api.del_bawu_blacklist。
package delbawublacklist

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
func ParseBody(body []byte) error {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return err
	}
	if code := helper.JSONInt(res, "errno"); code != 0 {
		return &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "errmsg")}
	}
	return nil
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/bawu2/platform/cancelBlack"}
}

// Request 对应 request。
func Request(ctx context.Context, httpCore *core.HttpCore, fname string, userID int64) error {
	data := []crypto.Param{
		{Key: "word", Value: fname},
		{Key: "tbs", Value: httpCore.Account.Tbs()},
		{Key: "list[]", Value: userID},
		{Key: "ie", Value: "utf-8"},
	}

	resp, err := httpCore.WebForm(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return err
	}
	return ParseBody(resp.Body())
}
