// Package signforums implements the sign_forums API of aiotieba.
//
// It mirrors the Python package aiotieba.api.sign_forums.
package signforums

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// SubappType is the value sent in the body and the Subapp-Type header.
const SubappType = "hybrid"

// ParseBody mirrors parse_body.
//
// Besides the usual error_code/error_msg pair the endpoint may nest a second
// failure envelope under `error`.
func ParseBody(body []byte) error {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return err
	}
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}
	if helper.HasJSONKey(res, "error") {
		errObj := helper.JSONMap(res, "error")
		if code := helper.JSONInt(errObj, "errno"); code != 0 {
			return &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(errObj, "errmsg")}
		}
	}
	return nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/c/c/forum/msign"}
}

// Request mirrors request.
func Request(ctx context.Context, httpCore *core.HttpCore) error {
	data := []crypto.Param{
		{Key: "_client_version", Value: consts.LatestVersion},
		{Key: "subapp_type", Value: SubappType},
	}

	req, err := httpCore.PackWebFormRequest(ctx, RequestURL(), data, map[string]string{"Subapp-Type": SubappType})
	if err != nil {
		return err
	}
	body, err := httpCore.NetCore.SendRequest(req)
	if err != nil {
		return err
	}
	return ParseBody(body)
}
