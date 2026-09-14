// Package followforum implements the follow_forum API of aiotieba.
//
// It mirrors the Python package aiotieba.api.follow_forum.
package followforum

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

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
		if code := helper.JSONInt(helper.JSONMap(res, "error"), "errno"); code != 0 {
			return &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(helper.JSONMap(res, "error"), "errmsg")}
		}
	}
	return nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.AppBaseHost, Path: "/c/c/forum/like"}
}

// Request mirrors request.
func Request(ctx context.Context, httpCore *core.HttpCore, fid int64) error {
	data := []crypto.Param{
		{Key: "BDUSS", Value: httpCore.Account.BDUSS()},
		{Key: "fid", Value: fid},
		{Key: "tbs", Value: httpCore.Account.Tbs()},
	}

	resp, err := httpCore.AppForm(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return err
	}
	return ParseBody(resp.Body())
}
