package getselfinfomoindex

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
func ParseBody(body []byte) (UserInfoMoindex, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return UserInfoMoindex{}, err
	}
	if code := helper.JSONInt(res, "no"); code != 0 {
		return UserInfoMoindex{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error")}
	}
	return UserInfoMoindexFromJSON(helper.JSONMap(res, "data")), nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/mo/q/newmoindex"}
}

// Request performs the web get request, mirroring request.
func Request(ctx context.Context, httpCore *core.HttpCore) (UserInfoMoindex, error) {
	params := []crypto.Param{{Key: "need_user", Value: 1}}
	req, err := httpCore.PackWebGetRequest(ctx, RequestURL(), params, nil)
	if err != nil {
		return UserInfoMoindex{}, err
	}
	body, err := httpCore.SendWeb(req)
	if err != nil {
		return UserInfoMoindex{}, err
	}
	return ParseBody(body)
}
