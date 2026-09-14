package getuserinfoweb

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
func ParseBody(body []byte) (UserInfoGuinfoWeb, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return UserInfoGuinfoWeb{}, err
	}
	if code := helper.JSONInt(res, "errno"); code != 0 {
		return UserInfoGuinfoWeb{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "errmsg")}
	}
	return UserInfoGuinfoWebFromJSON(helper.JSONMap(res, "chatUser")), nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "http", Host: consts.WebBaseHost, Path: "/im/pcmsg/query/getUserInfo"}
}

// Request mirrors request.
//
// The endpoint requires the BDUSS cookie.
func Request(ctx context.Context, httpCore *core.HttpCore, userID int64) (UserInfoGuinfoWeb, error) {
	params := []crypto.Param{{Key: "chatUid", Value: userID}}

	req, err := httpCore.PackWebGetRequest(ctx, RequestURL(), params, nil)
	if err != nil {
		return UserInfoGuinfoWeb{}, err
	}
	body, err := httpCore.SendWeb(req)
	if err != nil {
		return UserInfoGuinfoWeb{}, err
	}
	return ParseBody(body)
}
