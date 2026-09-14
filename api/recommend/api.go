// Package recommend implements the recommend API of aiotieba.
//
// It mirrors the Python package aiotieba.api.recommend.
package recommend

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
// The endpoint reports its own outcome under `data.is_push_success`, which must
// be 1.
func ParseBody(body []byte) error {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return err
	}
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}

	data := helper.JSONMap(res, "data")
	if code := helper.JSONInt(data, "is_push_success"); code != 1 {
		return &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(data, "msg")}
	}
	return nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.AppBaseHost, Path: "/c/c/bawu/pushRecomToPersonalized"}
}

// Request mirrors request.
func Request(ctx context.Context, httpCore *core.HttpCore, fid, tid int64) error {
	data := []crypto.Param{
		{Key: "BDUSS", Value: httpCore.Account.BDUSS()},
		{Key: "forum_id", Value: fid},
		{Key: "thread_id", Value: tid},
	}

	resp, err := httpCore.AppForm(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return err
	}
	return ParseBody(resp.Body())
}
