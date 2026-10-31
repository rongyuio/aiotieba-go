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

// ParseBody 解析响应体，将 JSON 转换为 SearchGlobals。
func ParseBody(body []byte) (SearchGlobals, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return SearchGlobals{}, err
	}
	if code := helper.JSONInt(res, "no"); code != 0 {
		return SearchGlobals{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error")}
	}
	return SearchGlobalsFromJSON(helper.JSONMap(res, "data")), nil
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/mo/q/search/thread"}
}

// Request 发起网页端搜索的 GET 请求。
//
// fname 非空时限定在该吧内搜索，为空则搜索全站。
// pn 为页码，rn 为单页条目数，sort 为排序取值。
func Request(ctx context.Context, httpCore *core.HttpCore, fname, word string, pn, rn, sort int64) (SearchGlobals, error) {
	params := []crypto.Param{
		{Key: "word", Value: word},
		{Key: "pn", Value: pn},
		{Key: "rn", Value: rn},
		{Key: "st", Value: sort},
		{Key: "tt", Value: 1},
		{Key: "ct", Value: 2},
		{Key: "cv", Value: consts.LatestVersion},
	}
	if fname != "" {
		params = append(params, crypto.Param{Key: "fname", Value: fname})
	}
	params = crypto.Sign(params, []byte(crypto.PCSalt))

	resp, err := httpCore.WebGet(params, map[string]string{"Referer": refererGlobal}).SetContext(ctx).Get(RequestURL().String())
	if err != nil {
		return SearchGlobals{}, err
	}
	return ParseBody(resp.Body())
}
