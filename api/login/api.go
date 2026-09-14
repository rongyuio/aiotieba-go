package login

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
//
// It returns the account information and the tbs token.
func ParseBody(body []byte) (UserInfoLogin, string, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return UserInfoLogin{}, "", err
	}
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return UserInfoLogin{}, "", &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}

	user := UserInfoLoginFromJSON(helper.JSONMap(res, "user"))
	tbs := helper.JSONStr(helper.JSONMap(res, "anti"), "tbs")
	return user, tbs, nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   consts.AppBaseHost,
		Path:   "/c/s/login",
	}
}

// Request performs the app form request, mirroring request.
func Request(ctx context.Context, httpCore *core.HttpCore) (UserInfoLogin, string, error) {
	data := []crypto.Param{
		{Key: "_client_version", Value: consts.LatestVersion},
		{Key: "bdusstoken", Value: httpCore.Account.BDUSS()},
	}

	req, err := httpCore.PackFormRequest(ctx, RequestURL(), data)
	if err != nil {
		return UserInfoLogin{}, "", err
	}
	body, err := httpCore.NetCore.SendRequest(req)
	if err != nil {
		return UserInfoLogin{}, "", err
	}
	return ParseBody(body)
}
