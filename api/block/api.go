// Package block 实现 aiotieba 的 block API。
//
// 对应 Python 包 aiotieba.api.block。
package block

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// standardBlockDays 是每个用户都可用的封禁时长；其他时长需要超级会员的循环封禁能力。
var standardBlockDays = map[int64]struct{}{1: {}, 3: {}, 10: {}}

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
	return &url.URL{Scheme: "https", Host: consts.AppBaseHost, Path: "/c/c/bawu/commitprison"}
}

// IsLoopBan 返回上报给服务端的 is_loop_ban 标志：标准时长所有用户都支持，
// 更长的时长需要超级会员的循环封禁能力。
func IsLoopBan(day int64) int {
	if _, ok := standardBlockDays[day]; ok {
		return 0
	}
	return 1
}

// Request 对应 request。
func Request(ctx context.Context, httpCore *core.HttpCore, fid int64, portrait string, day int64, reason string) error {
	data := []crypto.Param{
		{Key: "BDUSS", Value: httpCore.Account.BDUSS()},
		{Key: "day", Value: day},
		{Key: "fid", Value: fid},
		{Key: "is_loop_ban", Value: IsLoopBan(day)},
		{Key: "ntn", Value: "banid"},
		{Key: "portrait", Value: portrait},
		{Key: "reason", Value: reason},
		{Key: "tbs", Value: httpCore.Account.Tbs()},
		{Key: "word", Value: "-"},
		{Key: "z", Value: 6},
	}

	resp, err := httpCore.AppForm(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return err
	}
	return ParseBody(resp.Body())
}
