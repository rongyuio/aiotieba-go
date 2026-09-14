// Package delbawu implements the del_bawu API of aiotieba.
//
// It mirrors the Python package aiotieba.api.del_bawu.
package delbawu

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba/consts"
	"github.com/rongyuio/aiotieba/core"
	"github.com/rongyuio/aiotieba/enums"
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
	if code := helper.JSONInt(res, "no"); code != 0 {
		return &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error")}
	}
	return nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/mo/q/bawuteamclear"}
}

// Request mirrors request.
func Request(ctx context.Context, httpCore *core.HttpCore, fid int64, portrait string, bawuType enums.BawuType) error {
	data := []crypto.Param{
		{Key: "fn", Value: "-"},
		{Key: "fid", Value: fid},
		{Key: "team_un", Value: "-"},
		{Key: "team_uid", Value: portrait},
		{Key: "bawu_type", Value: bawuType},
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
