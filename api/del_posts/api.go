// Package delposts 实现 aiotieba 的 del_posts API。
//
// 对应 Python 包 aiotieba.api.del_posts。
package delposts

import (
	"context"
	"net/url"
	"strconv"
	"strings"

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
	return &url.URL{Scheme: "https", Host: consts.AppBaseHost, Path: "/c/c/bawu/multiDelPost"}
}

// Request 执行 app 表单请求，对应 request。
//
// block 选择删除模式：普通删除或删除并封禁。
func Request(ctx context.Context, httpCore *core.HttpCore, fid, tid int64, pids []int64, block bool) error {
	kind := 1
	if block {
		kind = 2
	}

	data := []crypto.Param{
		{Key: "BDUSS", Value: httpCore.Account.BDUSS()},
		{Key: "forum_id", Value: fid},
		{Key: "post_ids", Value: joinInt64(pids)},
		{Key: "tbs", Value: httpCore.Account.Tbs()},
		{Key: "thread_id", Value: tid},
		{Key: "type", Value: kind},
	}

	resp, err := httpCore.AppForm(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return err
	}
	return ParseBody(resp.Body())
}

func joinInt64(values []int64) string {
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = strconv.FormatInt(v, 10)
	}
	return strings.Join(parts, ",")
}
