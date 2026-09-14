// Package delbawublacklist implements the del_bawu_blacklist API of aiotieba.
//
// It mirrors the Python package aiotieba.api.del_bawu_blacklist.
package delbawublacklist

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
func ParseBody(body []byte) error {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return err
	}
	if code := helper.JSONInt(res, "errno"); code != 0 {
		return &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "errmsg")}
	}
	return nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/bawu2/platform/cancelBlack"}
}

// Request mirrors request.
func Request(ctx context.Context, httpCore *core.HttpCore, fname string, userID int64) error {
	data := []crypto.Param{
		{Key: "word", Value: fname},
		{Key: "tbs", Value: httpCore.Account.Tbs()},
		{Key: "list[]", Value: userID},
		{Key: "ie", Value: "utf-8"},
	}

	req, err := httpCore.PackWebFormRequest(ctx, RequestURL(), data, nil)
	if err != nil {
		return err
	}
	body, err := httpCore.NetCore.SendRequest(req)
	if err != nil {
		return err
	}
	return ParseBody(body)
}
