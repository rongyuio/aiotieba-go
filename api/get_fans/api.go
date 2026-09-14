package getfans

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
func ParseBody(body []byte) (Fans, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return Fans{}, err
	}
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return Fans{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}
	return FansFromJSON(res), nil
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.AppBaseHost, Path: "/c/u/fans/page"}
}

// Request 执行 app 表单请求，对应 request。
func Request(ctx context.Context, httpCore *core.HttpCore, userID, pn int64) (Fans, error) {
	data := []crypto.Param{
		{Key: "BDUSS", Value: httpCore.Account.BDUSS()},
		{Key: "_client_version", Value: consts.LatestVersion},
		{Key: "pn", Value: pn},
		{Key: "uid", Value: userID},
	}
	resp, err := httpCore.AppForm(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return Fans{}, err
	}
	return ParseBody(resp.Body())
}
