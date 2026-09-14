package getblocks

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba/consts"
	"github.com/rongyuio/aiotieba/core"
	"github.com/rongyuio/aiotieba/exception"
	"github.com/rongyuio/aiotieba/helper"
	"github.com/rongyuio/aiotieba/helper/crypto"
)

// ParseBody decodes the JSON response, mirroring parse_body.
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

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/mo/q/bawublock"}
}

// Request performs the web get request, mirroring request.
func Request(ctx context.Context, httpCore *core.HttpCore, fid int64, name string, pn int64) (Blocks, error) {
	params := []crypto.Param{
		{Key: "fn", Value: "-"},
		{Key: "fid", Value: fid},
		{Key: "word", Value: name},
		{Key: "is_ajax", Value: 1},
		{Key: "pn", Value: pn},
	}
	req, err := httpCore.PackWebGetRequest(ctx, RequestURL(), params, nil)
	if err != nil {
		return Blocks{}, err
	}
	body, err := httpCore.SendWeb(req)
	if err != nil {
		return Blocks{}, err
	}
	return ParseBody(body)
}
