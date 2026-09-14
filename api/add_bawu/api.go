// Package addbawu implements the add_bawu API of aiotieba.
//
// It mirrors the Python package aiotieba.api.add_bawu.
package addbawu

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
	if code := helper.JSONInt(res, "no"); code != 0 {
		return &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error")}
	}
	return nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/mo/q/bawuteamadd"}
}

// Request mirrors request.
func Request(ctx context.Context, httpCore *core.HttpCore, fid int64, userName string, bawuType enums.BawuType) error {
	data := []crypto.Param{
		{Key: "fn", Value: "-"},
		{Key: "fid", Value: fid},
		{Key: "team_un", Value: userName},
		{Key: "type", Value: bawuType},
		{Key: "tbs", Value: httpCore.Account.Tbs()},
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
