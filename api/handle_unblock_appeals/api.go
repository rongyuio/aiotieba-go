// Package handleunblockappeals 实现 aiotieba 的 handle_unblock_appeals API。
//
// 对应 Python 包 aiotieba.api.handle_unblock_appeals。
package handleunblockappeals

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
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/mo/q/multiAppealhandle"}
}

// Request 执行网页端表单请求，对应 request。
//
// 请求体与 bawu-user-appeal.js 一致：fn、fid、appeal_list（数组）、
// refuse_reason、status。refuse 为 true 时 status=2 表示驳回，否则 1 表示通过。
func Request(ctx context.Context, httpCore *core.HttpCore, fname string, fid int64, appealIDs []int64, refuse bool) error {
	status := 1
	refuseReason := "已解封"
	if refuse {
		status = 2
		refuseReason = "_"
	}

	data := make([]crypto.Param, 0, len(appealIDs)+5)
	data = append(data,
		crypto.Param{Key: "fn", Value: fname},
		crypto.Param{Key: "fid", Value: fid},
	)
	for i, appealID := range appealIDs {
		data = append(data, crypto.Param{Key: "appeal_list[" + strconv.Itoa(i) + "]", Value: appealID})
	}
	data = append(data,
		crypto.Param{Key: "refuse_reason", Value: refuseReason},
		crypto.Param{Key: "status", Value: status},
		crypto.Param{Key: "tbs", Value: httpCore.Account.Tbs()},
	)

	resp, err := httpCore.WebForm(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return err
	}
	return ParseBody(resp.Body())
}
