package getrecovers

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba/consts"
	"github.com/rongyuio/aiotieba/core"
	"github.com/rongyuio/aiotieba/exception"
	"github.com/rongyuio/aiotieba/helper"
	"github.com/rongyuio/aiotieba/helper/crypto"
)

// ParseBody decodes the JSON response, mirroring parse_body.
func ParseBody(body []byte) (Recovers, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return Recovers{}, err
	}
	if code := helper.JSONInt(res, "no"); code != 0 {
		return Recovers{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error")}
	}
	return RecoversFromJSON(helper.JSONMap(res, "data")), nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/mo/q/manage/getRecoverList"}
}

// Request performs the web get request, mirroring request. A zero userID omits
// the uid parameter.
func Request(ctx context.Context, httpCore *core.HttpCore, fid, userID, pn, rn int64) (Recovers, error) {
	params := []crypto.Param{
		{Key: "rn", Value: rn},
		{Key: "forum_id", Value: fid},
		{Key: "pn", Value: pn},
		{Key: "type", Value: 1},
		{Key: "sub_type", Value: 1},
	}
	if userID != 0 {
		params = append(params, crypto.Param{Key: "uid", Value: userID})
	}
	req, err := httpCore.PackWebGetRequest(ctx, RequestURL(), params, nil)
	if err != nil {
		return Recovers{}, err
	}
	body, err := httpCore.SendWeb(req)
	if err != nil {
		return Recovers{}, err
	}
	return ParseBody(body)
}
