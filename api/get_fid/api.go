// Package getfid implements the get_fid API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_fid.
package getfid

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// ParseBody extracts the fid from the web JSON response, mirroring parse_body.
func ParseBody(body []byte) (int64, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return 0, err
	}
	if code := helper.JSONInt(res, "no"); code != 0 {
		return 0, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error")}
	}

	fid := helper.JSONInt(helper.JSONMap(res, "data"), "fid")
	if fid == 0 {
		return 0, &exception.TiebaValueError{Msg: "fid is 0"}
	}
	return fid, nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   consts.WebBaseHost,
		Path:   "/f/commit/share/fnameShareApi",
	}
}

// Request performs the web request, mirroring request.
func Request(ctx context.Context, httpCore *core.HttpCore, fname string) (int64, error) {
	params := []crypto.Param{
		{Key: "fname", Value: fname},
		{Key: "ie", Value: "utf-8"},
	}
	req, err := httpCore.PackWebGetRequest(ctx, RequestURL(), params, nil)
	if err != nil {
		return 0, err
	}
	body, err := httpCore.SendWeb(req)
	if err != nil {
		return 0, err
	}
	return ParseBody(body)
}
