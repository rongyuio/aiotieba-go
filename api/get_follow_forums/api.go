package getfollowforums

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
func ParseBody(body []byte) (FollowForums, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return FollowForums{}, err
	}
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return FollowForums{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}
	return FollowForumsFromJSON(res), nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.AppBaseHost, Path: "/c/f/forum/like"}
}

// Request performs the app form request, mirroring request.
func Request(ctx context.Context, httpCore *core.HttpCore, userID, pn, rn int64) (FollowForums, error) {
	data := []crypto.Param{
		{Key: "BDUSS", Value: httpCore.Account.BDUSS()},
		{Key: "_client_version", Value: consts.LatestVersion},
		{Key: "friend_uid", Value: userID},
		{Key: "page_no", Value: pn},
		{Key: "page_size", Value: rn},
	}
	req, err := httpCore.PackFormRequest(ctx, RequestURL(), data)
	if err != nil {
		return FollowForums{}, err
	}
	body, err := httpCore.NetCore.SendRequest(req)
	if err != nil {
		return FollowForums{}, err
	}
	return ParseBody(body)
}
