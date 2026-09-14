// Package sync implements the sync API of aiotieba.
//
// It mirrors the Python package aiotieba.api.sync.
package sync

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// ParseBody decodes the JSON response, mirroring parse_body.
//
// It returns the client id and the sample id.
func ParseBody(body []byte) (string, string, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return "", "", err
	}
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return "", "", &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}

	clientID := helper.JSONStr(helper.JSONMap(res, "client"), "client_id")
	sampleID := helper.JSONStr(helper.JSONMap(res, "wl_config"), "sample_id")
	return clientID, sampleID, nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   consts.AppBaseHost,
		Path:   "/c/s/sync",
	}
}

// Request performs the app form request, mirroring request.
func Request(ctx context.Context, httpCore *core.HttpCore) (string, string, error) {
	cuidGalaxy2, err := httpCore.Account.CuidGalaxy2()
	if err != nil {
		return "", "", err
	}

	data := []crypto.Param{
		{Key: "BDUSS", Value: httpCore.Account.BDUSS()},
		{Key: "_client_version", Value: consts.LatestVersion},
		{Key: "cuid", Value: cuidGalaxy2},
	}

	resp, err := httpCore.AppForm(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return "", "", err
	}
	return ParseBody(resp.Body())
}
