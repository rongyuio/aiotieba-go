// Package move implements the move API of aiotieba.
//
// It mirrors the Python package aiotieba.api.move.
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

// threadEntry is one element of the `threads` JSON array.
//
// The field order matters because the JSON text itself is part of the signed
// payload.
type threadEntry struct {
	ThreadID  int64 `json:"thread_id"`
	FromTabID int64 `json:"from_tab_id"`
	ToTabID   int64 `json:"to_tab_id"`
}

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
	return &url.URL{Scheme: "https", Host: consts.AppBaseHost, Path: "/c/c/bawu/moveTabThread"}
}

// Request mirrors request.
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
