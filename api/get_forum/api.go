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

// ParseBody 解析 JSON 响应，对应 parse_body。
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

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   consts.AppBaseHost,
		Path:   "/c/f/frs/frsBottom",
	}
}

// Request 执行 app 表单请求，对应 request。
func Request(ctx context.Context, httpCore *core.HttpCore, fname string) (Forum, error) {
	data := []crypto.Param{{Key: "kw", Value: fname}}

	resp, err := httpCore.AppForm(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return Forum{}, err
	}
	return ParseBody(resp.Body())
}
