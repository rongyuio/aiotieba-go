package getselfinfoinitnickname

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
func ParseBody(body []byte) (UserInfoSelfinit, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return UserInfoSelfinit{}, err
	}
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return UserInfoSelfinit{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}
	return UserInfoSelfinitFromJSON(helper.JSONMap(res, "user_info")), nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.AppBaseHost, Path: "/c/s/initNickname"}
}

// Request performs the app form request, mirroring request.
func Request(ctx context.Context, httpCore *core.HttpCore) (UserInfoSelfinit, error) {
	data := []crypto.Param{
		{Key: "BDUSS", Value: httpCore.Account.BDUSS()},
		{Key: "_client_version", Value: consts.LatestVersion},
	}
	req, err := httpCore.PackFormRequest(ctx, RequestURL(), data)
	if err != nil {
		return UserInfoSelfinit{}, err
	}
	body, err := httpCore.NetCore.SendRequest(req)
	if err != nil {
		return UserInfoSelfinit{}, err
	}
	return ParseBody(body)
}
