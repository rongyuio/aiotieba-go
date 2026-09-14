package getuserforuminfo

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
func ParseBody(body []byte) (UserForumInfo, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return UserForumInfo{}, err
	}
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		msg := helper.JSONStr(res, "error_msg")
		if msg == "" {
			msg = helper.JSONStr(res, "error")
		}
		if msg == "" {
			msg = helper.JSONStr(res, "errmsg")
		}
		return UserForumInfo{}, &exception.TiebaServerError{Code: int(code), Msg: msg}
	}
	return UserForumInfoFromJSON(helper.JSONMap(res, "data")), nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.AppBaseHost, Path: "/c/f/forum/getUserForumLevelInfo"}
}

// Request performs the app form request, mirroring request.
func Request(ctx context.Context, httpCore *core.HttpCore, fid int64, friendPortrait string) (UserForumInfo, error) {
	data := []crypto.Param{
		{Key: "BDUSS", Value: httpCore.Account.BDUSS()},
		{Key: "_client_version", Value: consts.LatestVersion},
		{Key: "forum_id", Value: fid},
		{Key: "friend_portrait", Value: friendPortrait},
	}
	req, err := httpCore.PackFormRequest(ctx, RequestURL(), data)
	if err != nil {
		return UserForumInfo{}, err
	}
	body, err := httpCore.NetCore.SendRequest(req)
	if err != nil {
		return UserForumInfo{}, err
	}
	return ParseBody(body)
}
