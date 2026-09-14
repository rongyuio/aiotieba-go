package searchglobal

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba/consts"
	"github.com/rongyuio/aiotieba/core"
	"github.com/rongyuio/aiotieba/exception"
	"github.com/rongyuio/aiotieba/helper"
	"github.com/rongyuio/aiotieba/helper/crypto"
)

// refererGlobal mirrors REFERER_GLOBAL.
const refererGlobal = "https://tieba.baidu.com/f/search/res"

// ParseBody decodes the JSON response, mirroring parse_body.
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

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/mo/q/search/thread"}
}

// Request performs the web get request, mirroring request.
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

	req, err := httpCore.PackWebGetRequest(ctx, RequestURL(), params, map[string]string{"Referer": refererGlobal})
	if err != nil {
		return GlobalSearches{}, err
	}
	body, err := httpCore.SendWeb(req)
	if err != nil {
		return GlobalSearches{}, err
	}
	return ParseBody(body)
}
