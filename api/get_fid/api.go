// Package getfid 实现 aiotieba 的 get_fid API。
//
// 对应 Python 包 aiotieba.api.get_fid。
package getfid

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// ParseBody 从网页端 JSON 响应中提取 fid，对应 parse_body。
func ParseBody(body []byte) (int64, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return 0, err
	}
	if code := helper.JSONInt(res, "no"); code != 0 {
		return 0, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error")}
	}

	fid := helper.JSONInt(helper.JSONMap(res, "data"), "fid")
	if fid == 0 {
		return 0, &exception.TiebaValueError{Msg: "fid is 0"}
	}
	return fid, nil
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   consts.WebBaseHost,
		Path:   "/f/commit/share/fnameShareApi",
	}
}

// Request 执行网页端请求，对应 request。
func Request(ctx context.Context, httpCore *core.HttpCore, fname string) (int64, error) {
	params := []crypto.Param{
		{Key: "fname", Value: fname},
		{Key: "ie", Value: "utf-8"},
	}
	resp, err := httpCore.WebGet(params, nil).SetContext(ctx).Get(RequestURL().String())
	if err != nil {
		return 0, err
	}
	return ParseBody(resp.Body())
}
