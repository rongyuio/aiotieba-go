// Package delbawu 实现 aiotieba 的 del_bawu API。
//
// 对应 Python 包 aiotieba.api.del_bawu。
package delbawu

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/enums"
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
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/mo/q/bawuteamclear"}
}

// Request 对应 request。
func Request(ctx context.Context, httpCore *core.HttpCore, fid int64, portrait string, bawuType enums.BawuType) error {
	data := []crypto.Param{
		{Key: "fn", Value: "-"},
		{Key: "fid", Value: fid},
		{Key: "team_un", Value: "-"},
		{Key: "team_uid", Value: portrait},
		{Key: "bawu_type", Value: bawuType},
	}

	resp, err := httpCore.WebForm(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return err
	}
	return ParseBody(resp.Body())
}
