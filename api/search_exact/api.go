package searchexact

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// ParseBody 解析 JSON 响应，对应 parse_body。
func ParseBody(body []byte) (ExactSearches, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return ExactSearches{}, err
	}
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return ExactSearches{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}
	return ExactSearchesFromJSON(res), nil
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{Scheme: "http", Host: consts.AppBaseHost, Path: "/c/s/searchpost"}
}

// Request 执行 app 表单请求，对应 request。
func Request(ctx context.Context, httpCore *core.HttpCore, fname, query string, pn, rn int64, searchType enums.SearchType, onlyThread bool) (ExactSearches, error) {
	data := []crypto.Param{
		{Key: "_client_version", Value: consts.LatestVersion},
		{Key: "kw", Value: fname},
		{Key: "only_thread", Value: helper.BoolInt(onlyThread)},
		{Key: "pn", Value: pn},
		{Key: "rn", Value: rn},
		{Key: "sm", Value: int(searchType)},
		{Key: "word", Value: query},
	}
	resp, err := httpCore.AppForm(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return ExactSearches{}, err
	}
	return ParseBody(resp.Body())
}
