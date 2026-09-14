// Package getcid implements the get_cid API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_cid.
package getcid

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba/consts"
	"github.com/rongyuio/aiotieba/core"
	"github.com/rongyuio/aiotieba/exception"
	"github.com/rongyuio/aiotieba/helper"
	"github.com/rongyuio/aiotieba/helper/crypto"
)

// Cate is one entry of the good-category list returned by the API.
type Cate map[string]any

// ParseBody mirrors parse_body.
func ParseBody(body []byte) ([]Cate, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return nil, err
	}
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return nil, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}

	raw := helper.JSONSlice(res, "cates")
	cates := make([]Cate, 0, len(raw))
	for _, item := range raw {
		if entry, ok := item.(map[string]any); ok {
			cates = append(cates, Cate(entry))
		}
	}
	return cates, nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.AppBaseHost, Path: "/c/c/bawu/goodlist"}
}

// Request mirrors request.
func Request(ctx context.Context, httpCore *core.HttpCore, fname string) ([]Cate, error) {
	data := []crypto.Param{
		{Key: "BDUSS", Value: httpCore.Account.BDUSS()},
		{Key: "word", Value: fname},
	}

	req, err := httpCore.PackFormRequest(ctx, RequestURL(), data)
	if err != nil {
		return nil, err
	}
	body, err := httpCore.NetCore.SendRequest(req)
	if err != nil {
		return nil, err
	}
	return ParseBody(body)
}
