package getuserinfousercard

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba/consts"
	"github.com/rongyuio/aiotieba/core"
	"github.com/rongyuio/aiotieba/exception"
	"github.com/rongyuio/aiotieba/helper"
	"github.com/rongyuio/aiotieba/helper/crypto"
)

// subappType and clientType are sent with the request, mirroring the Python
// module.
const (
	subappType = "pc"
	clientType = 20
)

// ParseBody mirrors parse_body.
func ParseBody(body []byte) (UserInfoUC, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return UserInfoUC{}, err
	}
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return UserInfoUC{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}
	return UserInfoUCFromJSON(helper.JSONMap(helper.JSONMap(res, "data"), "user_info")), nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/c/u/pc/userCard"}
}

// Request mirrors request.
//
// The parameters are signed with PC_SALT rather than APP_SALT.
func Request(ctx context.Context, httpCore *core.HttpCore, portrait string) (UserInfoUC, error) {
	params := crypto.Sign([]crypto.Param{
		{Key: "portrait", Value: portrait},
		{Key: "subapp_type", Value: subappType},
		{Key: "_client_type", Value: clientType},
	}, []byte(crypto.PCSalt))

	req, err := httpCore.PackWebGetRequest(ctx, RequestURL(), params, nil)
	if err != nil {
		return UserInfoUC{}, err
	}
	body, err := httpCore.SendWeb(req)
	if err != nil {
		return UserInfoUC{}, err
	}
	return ParseBody(body)
}
