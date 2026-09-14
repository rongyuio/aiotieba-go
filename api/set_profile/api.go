// Package setprofile implements the set_profile API of aiotieba.
//
// It mirrors the Python package aiotieba.api.set_profile.
package setprofile

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// ParseBody mirrors parse_body.
func ParseBody(body []byte) error {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return err
	}
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}
	return nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.AppBaseHost, Path: "/c/c/profile/modify"}
}

// Request mirrors request.
func Request(ctx context.Context, httpCore *core.HttpCore, nickName, sign string, gender enums.Gender) error {
	data := []crypto.Param{
		{Key: "BDUSS", Value: httpCore.Account.BDUSS()},
		{Key: "intro", Value: sign},
		{Key: "nick_name", Value: nickName},
		{Key: "sex", Value: int(gender)},
	}

	req, err := httpCore.PackFormRequest(ctx, RequestURL(), data)
	if err != nil {
		return err
	}
	body, err := httpCore.NetCore.SendRequest(req)
	if err != nil {
		return err
	}
	return ParseBody(body)
}
