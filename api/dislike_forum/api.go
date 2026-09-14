// Package dislikeforum 实现 aiotieba 的 dislike_forum API。
//
// 对应 Python 包 aiotieba.api.dislike_forum。
package dislikeforum

import (
	"context"
	"net/url"
	"time"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// dislikeEntry 是 `dislike` JSON 数组中的一个元素。
//
// 字段顺序很重要，因为 JSON 文本本身是签名载荷的一部分。
type dislikeEntry struct {
	TID         int   `json:"tid"`
	DislikeIDs  int   `json:"dislike_ids"`
	FID         int64 `json:"fid"`
	ClickTimeMs int64 `json:"click_time"`
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
	return &url.URL{Scheme: "https", Host: consts.AppBaseHost, Path: "/c/c/excellent/submitDislike"}
}

// Request 执行 app 表单请求，对应 request。
func Request(ctx context.Context, httpCore *core.HttpCore, fid int64) error {
	dislike := helper.PackJSON([]dislikeEntry{{
		TID:         1,
		DislikeIDs:  7,
		FID:         fid,
		ClickTimeMs: time.Now().UnixMilli(),
	}})

	data := []crypto.Param{
		{Key: "BDUSS", Value: httpCore.Account.BDUSS()},
		{Key: "_client_version", Value: consts.LatestVersion},
		{Key: "dislike", Value: dislike},
		{Key: "dislike_from", Value: "homepage"},
	}

	resp, err := httpCore.AppForm(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return err
	}
	return ParseBody(resp.Body())
}
