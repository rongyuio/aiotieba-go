package getblocks

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
func ParseBody(body []byte) (Blocks, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return Blocks{}, err
	}
	if code := helper.JSONInt(res, "no"); code != 0 {
		return Blocks{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error")}
	}
	return BlocksFromJSON(res)
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/mo/q/bawublock"}
}

// Request 执行网页端 GET 请求，对应 request。
func Request(ctx context.Context, httpCore *core.HttpCore, fid int64, name string, pn int64) (Blocks, error) {
	params := []crypto.Param{
		{Key: "fn", Value: "-"},
		{Key: "fid", Value: fid},
		{Key: "word", Value: name},
		{Key: "is_ajax", Value: 1},
		{Key: "pn", Value: pn},
	}
	resp, err := httpCore.WebGet(params, nil).SetContext(ctx).Get(RequestURL().String())
	if err != nil {
		return Blocks{}, err
	}
	return ParseBody(resp.Body())
}
