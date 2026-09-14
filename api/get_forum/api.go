package getforum

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
func ParseBody(body []byte) (Forum, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return Forum{}, err
	}
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return Forum{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}
	return ForumFromJSON(helper.JSONMap(res, "forum")), nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   consts.AppBaseHost,
		Path:   "/c/f/frs/frsBottom",
	}
}

// Request performs the app form request, mirroring request.
func Request(ctx context.Context, httpCore *core.HttpCore, fname string) (Forum, error) {
	data := []crypto.Param{{Key: "kw", Value: fname}}

	req, err := httpCore.PackFormRequest(ctx, RequestURL(), data)
	if err != nil {
		return Forum{}, err
	}
	body, err := httpCore.NetCore.SendRequest(req)
	if err != nil {
		return Forum{}, err
	}
	return ParseBody(body)
}
