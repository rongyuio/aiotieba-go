package getunblockappeals

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
func ParseBody(body []byte) (Appeals, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return Appeals{}, err
	}
	if code := helper.JSONInt(res, "no"); code != 0 {
		return Appeals{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error")}
	}
	return AppealsFromJSON(helper.JSONMap(res, "data")), nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/mo/q/getBawuAppealList"}
}

// Request performs the web form request, mirroring request.
func Request(ctx context.Context, httpCore *core.HttpCore, fid, pn, rn int64) (Appeals, error) {
	data := []crypto.Param{
		{Key: "fn", Value: "-"},
		{Key: "fid", Value: fid},
		{Key: "pn", Value: pn},
		{Key: "rn", Value: rn},
		{Key: "tbs", Value: httpCore.Account.Tbs()},
	}
	req, err := httpCore.PackWebFormRequest(ctx, RequestURL(), data, nil)
	if err != nil {
		return Appeals{}, err
	}
	body, err := httpCore.SendWeb(req)
	if err != nil {
		return Appeals{}, err
	}
	return ParseBody(body)
}
