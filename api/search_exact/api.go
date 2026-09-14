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

// ParseBody decodes the JSON response, mirroring parse_body.
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

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "http", Host: consts.AppBaseHost, Path: "/c/s/searchpost"}
}

// Request performs the app form request, mirroring request.
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
	req, err := httpCore.PackFormRequest(ctx, RequestURL(), data)
	if err != nil {
		return ExactSearches{}, err
	}
	body, err := httpCore.NetCore.SendRequest(req)
	if err != nil {
		return ExactSearches{}, err
	}
	return ParseBody(body)
}
