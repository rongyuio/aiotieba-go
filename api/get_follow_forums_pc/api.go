package getfollowforumspc

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// ParseBody decodes the JSON response, mirroring parse_body.
func ParseBody(body []byte) (PcFollowForums, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return PcFollowForums{}, err
	}
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return PcFollowForums{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}
	return PcFollowForumsFromJSON(helper.JSONMap(res, "data")), nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/c/f/pc/myForumList"}
}

// Request performs the web get request, mirroring request.
func Request(ctx context.Context, httpCore *core.HttpCore, portrait string, pn, rn int64) (PcFollowForums, error) {
	params := []crypto.Param{
		{Key: "portrait", Value: portrait},
		{Key: "pn", Value: pn},
		{Key: "rn", Value: rn},
		{Key: "subapp_type", Value: "pc"},
		{Key: "_client_type", Value: 20},
	}
	params = crypto.Sign(params, []byte(crypto.PCSalt))

	req, err := httpCore.PackWebGetRequest(ctx, RequestURL(), params, nil)
	if err != nil {
		return PcFollowForums{}, err
	}
	body, err := httpCore.SendWeb(req)
	if err != nil {
		return PcFollowForums{}, err
	}
	return ParseBody(body)
}
