package searchglobal

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// refererGlobal PC网页端搜索结果页 保留Referer可降低风控概率。
const refererGlobal = "https://tieba.baidu.com/f/search/res"

// ParseBody 解析响应体，将 JSON 转换为 GlobalSearches。
func ParseBody(body []byte) (GlobalSearches, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return GlobalSearches{}, err
	}
	if code := helper.JSONInt(res, "no"); code != 0 {
		return GlobalSearches{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error")}
	}
	return GlobalSearchesFromJSON(helper.JSONMap(res, "data")), nil
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/mo/q/search/thread"}
}

// Request 发起网页端全吧搜索的 GET 请求。
//
// pn 为页码，rn 为单页条目数，sort 为排序取值。
func Request(ctx context.Context, httpCore *core.HttpCore, word string, pn, rn, sort int64) (GlobalSearches, error) {
	params := []crypto.Param{
		{Key: "word", Value: word},
		{Key: "pn", Value: pn},
		{Key: "rn", Value: rn},
		{Key: "st", Value: sort},
		{Key: "tt", Value: 1},
		{Key: "subapp_type", Value: "pc"},
		{Key: "_client_type", Value: 20},
	}
	params = crypto.Sign(params, []byte(crypto.PCSalt))

	resp, err := httpCore.WebGet(params, map[string]string{"Referer": refererGlobal}).SetContext(ctx).Get(RequestURL().String())
	if err != nil {
		return GlobalSearches{}, err
	}
	return ParseBody(resp.Body())
}
