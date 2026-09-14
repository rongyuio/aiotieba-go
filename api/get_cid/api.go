// Package getcid 实现 aiotieba 的 get_cid API。
//
// 对应 Python 包 aiotieba.api.get_cid。
package getcid

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// Cate 是 API 返回的精品分类列表中的一项。
type Cate map[string]any

// ParseBody 解析响应，对应 parse_body。
func ParseBody(body []byte) ([]Cate, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return nil, err
	}
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return nil, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}

	raw := helper.JSONSlice(res, "cates")
	cates := make([]Cate, 0, len(raw))
	for _, item := range raw {
		if entry, ok := item.(map[string]any); ok {
			cates = append(cates, Cate(entry))
		}
	}
	return cates, nil
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.AppBaseHost, Path: "/c/c/bawu/goodlist"}
}

// Request 执行 app 表单请求，对应 request。
func Request(ctx context.Context, httpCore *core.HttpCore, fname string) ([]Cate, error) {
	data := []crypto.Param{
		{Key: "BDUSS", Value: httpCore.Account.BDUSS()},
		{Key: "word", Value: fname},
	}

	resp, err := httpCore.AppForm(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return nil, err
	}
	return ParseBody(resp.Body())
}
