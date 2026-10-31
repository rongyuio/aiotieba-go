package getbawulist

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
func ParseBody(body []byte) (BawuList, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return BawuList{}, err
	}
	if code := helper.JSONInt(res, "no"); code != 0 {
		return BawuList{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error")}
	}
	return BawuListFromJSON(helper.JSONMap(res, "data")), nil
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/mo/q/bawulist"}
}

// Request 执行网页端 GET 请求，对应 request。
func Request(ctx context.Context, httpCore *core.HttpCore, fname string, fid int64) (BawuList, error) {
	params := []crypto.Param{
		{Key: "fn", Value: fname},
		{Key: "fid", Value: fid},
		{Key: "pn", Value: 1},
		{Key: "is_ajax", Value: 1},
		{Key: "_client_version", Value: consts.LatestVersion},
		{Key: "_client_type", Value: "2"},
	}
	resp, err := httpCore.WebGet(params, nil).SetContext(ctx).Get(RequestURL().String())
	if err != nil {
		return BawuList{}, err
	}
	return ParseBody(resp.Body())
}
