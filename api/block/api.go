// Package block implements the block API of aiotieba.
//
// It mirrors the Python package aiotieba.api.block.
package block

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba/consts"
	"github.com/rongyuio/aiotieba/core"
	"github.com/rongyuio/aiotieba/exception"
	"github.com/rongyuio/aiotieba/helper"
	"github.com/rongyuio/aiotieba/helper/crypto"
)

// standardBlockDays are the ban durations available to every user; anything
// else requires the SVIP loop-ban capability.
var standardBlockDays = map[int64]struct{}{1: {}, 3: {}, 10: {}}

// ParseBody mirrors parse_body.
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

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.AppBaseHost, Path: "/c/c/bawu/commitprison"}
}

// IsLoopBan returns the is_loop_ban flag reported to the server: the standard
// durations are supported by every user, longer ones require the SVIP loop-ban
// capability.
func IsLoopBan(day int64) int {
	if _, ok := standardBlockDays[day]; ok {
		return 0
	}
	return 1
}

// Request mirrors request.
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

	req, err := httpCore.PackFormRequest(ctx, RequestURL(), data)
	if err != nil {
		return err
	}
	body, err := httpCore.NetCore.SendRequest(req)
	if err != nil {
		return err
	}
	return ParseBody(body)
}
