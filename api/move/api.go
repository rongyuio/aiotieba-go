// Package move 实现 aiotieba 的 move API。
//
// 对应 Python 包 aiotieba.api.move。
package move

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// threadEntry 是 `threads` JSON 数组中的一个元素。
//
// 字段顺序很重要，因为 JSON 文本本身是签名载荷的一部分。
type threadEntry struct {
	ThreadID  int64 `json:"thread_id"`
	FromTabID int64 `json:"from_tab_id"`
	ToTabID   int64 `json:"to_tab_id"`
}

// ParseBody 对应 parse_body。
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
	return &url.URL{Scheme: "https", Host: consts.AppBaseHost, Path: "/c/c/bawu/moveTabThread"}
}

// Request 执行 app 表单请求，对应 request。
func Request(ctx context.Context, httpCore *core.HttpCore, fid, tid, toTabID, fromTabID int64) error {
	threads := helper.PackJSON([]threadEntry{{
		ThreadID:  tid,
		FromTabID: fromTabID,
		ToTabID:   toTabID,
	}})

	data := []crypto.Param{
		{Key: "BDUSS", Value: httpCore.Account.BDUSS()},
		{Key: "_client_version", Value: consts.LatestVersion},
		{Key: "forum_id", Value: fid},
		{Key: "tbs", Value: httpCore.Account.Tbs()},
		{Key: "threads", Value: threads},
	}

	resp, err := httpCore.AppForm(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return err
	}
	return ParseBody(resp.Body())
}
