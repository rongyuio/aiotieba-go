package getselffollowforums

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
func ParseBody(body []byte) (SelfFollowForums, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return SelfFollowForums{}, err
	}
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return SelfFollowForums{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}
	return SelfFollowForumsFromJSON(res), nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/c/f/forum/forumGuide"}
}

// Request performs the web form request, mirroring request.
func Request(ctx context.Context, httpCore *core.HttpCore, pn, rn int64) (SelfFollowForums, error) {
	data := []crypto.Param{
		{Key: "tbs", Value: httpCore.Account.Tbs()},
		{Key: "sort_type", Value: 3},
		{Key: "call_from", Value: 3},
		{Key: "page_no", Value: pn},
		{Key: "res_num", Value: rn},
	}
	resp, err := httpCore.WebForm(data).SetHeader("Subapp-Type", "hybrid").SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return SelfFollowForums{}, err
	}
	return ParseBody(resp.Body())
}
