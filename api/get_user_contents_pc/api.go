package getusercontentpc

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
func ParseBody(body []byte) (PcUserPosts, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return PcUserPosts{}, err
	}
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return PcUserPosts{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}
	return PcUserPostsFromJSON(helper.JSONMap(res, "data")), nil
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/c/u/feed/myThread"}
}

// Request 执行网页端 GET 请求，对应 request。
func Request(ctx context.Context, httpCore *core.HttpCore, portrait string, pn, rn int64) (PcUserPosts, error) {
	params := []crypto.Param{
		{Key: "pn", Value: pn},
		{Key: "rn", Value: rn},
		{Key: "portrait", Value: portrait},
		{Key: "type", Value: 2},
		{Key: "subapp_type", Value: "pc"},
		{Key: "_client_type", Value: 20},
	}
	params = crypto.Sign(params, []byte(crypto.PCSalt))

	resp, err := httpCore.WebGet(params, nil).SetContext(ctx).Get(RequestURL().String())
	if err != nil {
		return PcUserPosts{}, err
	}
	return ParseBody(resp.Body())
}
