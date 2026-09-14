package getusercontentpc

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
func ParseBody(body []byte) (PcUserPosts, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return PcUserPosts{}, err
	}
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return PcUserPosts{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}
	return PcUserPostsFromJSON(helper.JSONMap(res, "data")), nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/c/u/feed/myThread"}
}

// Request performs the web get request, mirroring request.
func Request(ctx context.Context, httpCore *core.HttpCore, portrait string, pn, rn int64) (PcUserPosts, error) {
	params := []crypto.Param{
		{Key: "pn", Value: pn},
		{Key: "rn", Value: rn},
		{Key: "portrait", Value: portrait},
		{Key: "type", Value: 2},
		{Key: "subapp_type", Value: "pc"},
		{Key: "_client_type", Value: 20},
	}
	params = crypto.Sign(params, []byte(crypto.PCSalt))

	req, err := httpCore.PackWebGetRequest(ctx, RequestURL(), params, nil)
	if err != nil {
		return PcUserPosts{}, err
	}
	body, err := httpCore.SendWeb(req)
	if err != nil {
		return PcUserPosts{}, err
	}
	return ParseBody(body)
}
