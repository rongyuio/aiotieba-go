// Package setnicknameold 实现 aiotieba 的 set_nickname_old API。
//
// 对应 Python 包 aiotieba.api.set_nickname_old。
package setnicknameold

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
)

// ParseBody 对应 parse_body。
func ParseBody(body []byte) error {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return err
	}
	if code := helper.JSONInt(res, "no"); code != 0 {
		return &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error")}
	}
	return nil
}

// RequestURL 返回该 API 的请求地址。
//
// 参数通过查询字符串传递，请求体为空，对应 Python 模块。
func RequestURL(nickName string) *url.URL {
	q := url.Values{}
	q.Set("nickname", nickName)
	q.Set("tbs", "1")
	return &url.URL{
		Scheme:   "https",
		Host:     consts.WebBaseHost,
		Path:     "/mo/q/submit/modifyNickname",
		RawQuery: q.Encode(),
	}
}

// Request 执行网页端表单请求，对应 request。
func Request(ctx context.Context, httpCore *core.HttpCore, nickName string) error {
	resp, err := httpCore.WebForm(nil).SetContext(ctx).Post(RequestURL(nickName).String())
	if err != nil {
		return err
	}
	return ParseBody(resp.Body())
}
