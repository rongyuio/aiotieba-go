package getblacklist

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
func ParseBody(body []byte) (BlacklistUsers, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return BlacklistUsers{}, err
	}
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return BlacklistUsers{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}
	return BlacklistUsersFromJSON(res), nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.AppBaseHost, Path: "/c/u/user/userBlackPage"}
}

// Request performs the app form request, mirroring request.
func Request(ctx context.Context, httpCore *core.HttpCore) (BlacklistUsers, error) {
	data := []crypto.Param{
		{Key: "BDUSS", Value: httpCore.Account.BDUSS()},
		{Key: "_client_version", Value: consts.LatestVersion},
	}
	req, err := httpCore.PackFormRequest(ctx, RequestURL(), data)
	if err != nil {
		return BlacklistUsers{}, err
	}
	body, err := httpCore.NetCore.SendRequest(req)
	if err != nil {
		return BlacklistUsers{}, err
	}
	return ParseBody(body)
}
