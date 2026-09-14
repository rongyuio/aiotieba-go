// Package good 实现 aiotieba 的 good API。
//
// 对应 Python 包 aiotieba.api.good。
package good

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
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}
	return nil
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.AppBaseHost, Path: "/c/c/bawu/commitgood"}
}

// Request 对应 request。
func Request(ctx context.Context, httpCore *core.HttpCore, fname string, fid, tid, cid int64) error {
	data := []crypto.Param{
		{Key: "BDUSS", Value: httpCore.Account.BDUSS()},
		{Key: "cid", Value: cid},
		{Key: "fid", Value: fid},
		{Key: "ntn", Value: "set"},
		{Key: "tbs", Value: httpCore.Account.Tbs()},
		{Key: "word", Value: fname},
		{Key: "z", Value: tid},
	}

	resp, err := httpCore.AppForm(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return err
	}
	return ParseBody(resp.Body())
}
