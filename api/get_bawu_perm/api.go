package getbawuperm

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
func ParseBody(body []byte) (BawuPerm, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return BawuPerm{}, err
	}
	if code := helper.JSONInt(res, "no"); code != 0 {
		return BawuPerm{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error")}
	}
	return BawuPermFromJSON(helper.JSONMap(res, "data")), nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/mo/q/getAuthToolPerm"}
}

// Request performs the web get request, mirroring request.
func Request(ctx context.Context, httpCore *core.HttpCore, fid int64, portrait string) (BawuPerm, error) {
	params := []crypto.Param{
		{Key: "forum_id", Value: fid},
		{Key: "portrait", Value: portrait},
	}
	resp, err := httpCore.WebGet(params, nil).SetContext(ctx).Get(RequestURL().String())
	if err != nil {
		return BawuPerm{}, err
	}
	return ParseBody(resp.Body())
}
