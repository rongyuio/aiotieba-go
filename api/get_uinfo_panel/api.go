package getuinfopanel

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba/consts"
	"github.com/rongyuio/aiotieba/core"
	"github.com/rongyuio/aiotieba/exception"
	"github.com/rongyuio/aiotieba/helper"
	"github.com/rongyuio/aiotieba/helper/crypto"
)

// ParseBody mirrors parse_body.
func ParseBody(body []byte) (UserInfoPanel, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return UserInfoPanel{}, err
	}
	if code := helper.JSONInt(res, "no"); code != 0 {
		return UserInfoPanel{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error")}
	}
	return UserInfoPanelFromJSON(helper.JSONMap(res, "data")), nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/home/get/panel"}
}

// Request mirrors request.
func Request(ctx context.Context, httpCore *core.HttpCore, nameOrPortrait string) (UserInfoPanel, error) {
	key := "un"
	if helper.IsPortrait(nameOrPortrait) {
		key = "id"
	}
	params := []crypto.Param{{Key: key, Value: nameOrPortrait}}

	req, err := httpCore.PackWebGetRequest(ctx, RequestURL(), params, nil)
	if err != nil {
		return UserInfoPanel{}, err
	}
	body, err := httpCore.SendWeb(req)
	if err != nil {
		return UserInfoPanel{}, err
	}
	return ParseBody(body)
}
