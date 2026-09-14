package getuserinfoweb

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// ParseBody 对应 parse_body。
func ParseBody(body []byte) (UserInfoGuinfoWeb, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return UserInfoGuinfoWeb{}, err
	}
	if code := helper.JSONInt(res, "errno"); code != 0 {
		return UserInfoGuinfoWeb{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "errmsg")}
	}
	return UserInfoGuinfoWebFromJSON(helper.JSONMap(res, "chatUser")), nil
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{Scheme: "http", Host: consts.WebBaseHost, Path: "/im/pcmsg/query/getUserInfo"}
}

// Request 对应 request。
//
// 该接口需要 BDUSS cookie。
func Request(ctx context.Context, httpCore *core.HttpCore, userID int64) (UserInfoGuinfoWeb, error) {
	params := []crypto.Param{{Key: "chatUid", Value: userID}}

	resp, err := httpCore.WebGet(params, nil).SetContext(ctx).Get(RequestURL().String())
	if err != nil {
		return UserInfoGuinfoWeb{}, err
	}
	return ParseBody(resp.Body())
}
