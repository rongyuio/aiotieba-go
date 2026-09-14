package getselfinfomoindex

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// ParseBody 解析 JSON 响应，对应 parse_body。
func ParseBody(body []byte) (UserInfoMoindex, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return UserInfoMoindex{}, err
	}
	if code := helper.JSONInt(res, "no"); code != 0 {
		return UserInfoMoindex{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error")}
	}
	return UserInfoMoindexFromJSON(helper.JSONMap(res, "data")), nil
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/mo/q/newmoindex"}
}

// Request 执行网页端 GET 请求，对应 request。
func Request(ctx context.Context, httpCore *core.HttpCore) (UserInfoMoindex, error) {
	params := []crypto.Param{{Key: "need_user", Value: 1}}
	resp, err := httpCore.WebGet(params, nil).SetContext(ctx).Get(RequestURL().String())
	if err != nil {
		return UserInfoMoindex{}, err
	}
	return ParseBody(resp.Body())
}
