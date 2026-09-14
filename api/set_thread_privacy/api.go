// Package setthreadprivacy 实现 aiotieba 的 set_thread_privacy API。
//
// 对应 Python 包 aiotieba.api.set_thread_privacy。
package setthreadprivacy

import (
	"context"
	"net/url"
	"strconv"

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
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}
	return nil
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.AppBaseHost, Path: "/c/c/thread/setPrivacy"}
}

// Request 执行 app 表单请求，对应 request。
func Request(ctx context.Context, httpCore *core.HttpCore, fid, tid, pid int64, isHide bool) error {
	data := []crypto.Param{
		{Key: "BDUSS", Value: httpCore.Account.BDUSS()},
		{Key: "forum_id", Value: fid},
		{Key: "is_hide", Value: strconv.Itoa(helper.BoolInt(isHide))},
		{Key: "post_id", Value: pid},
		{Key: "thread_id", Value: tid},
	}

	resp, err := httpCore.AppForm(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return err
	}
	return ParseBody(resp.Body())
}
