// Package recover 实现 aiotieba 的 recover API。
//
// 对应 Python 包 aiotieba.api.recover。
package recover

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// ParseBody 解析响应，对应 parse_body。
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
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/mo/q/bawurecoverthread"}
}

// Request 执行网页端表单请求，对应 request。
//
// pid 为零时恢复整个主题帖。
func Request(ctx context.Context, httpCore *core.HttpCore, fid, tid, pid int64, isHide bool) error {
	typeList := 0
	if pid != 0 {
		typeList = 1
	}

	data := []crypto.Param{
		{Key: "tbs", Value: httpCore.Account.Tbs()},
		{Key: "fn", Value: "-"},
		{Key: "fid", Value: fid},
		{Key: "tid_list[]", Value: tid},
		{Key: "pid_list[]", Value: pid},
		{Key: "type_list[]", Value: typeList},
		{Key: "is_frs_mask_list[]", Value: helper.BoolInt(isHide)},
	}

	resp, err := httpCore.WebForm(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return err
	}
	return ParseBody(resp.Body())
}
