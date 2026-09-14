// Package dislikeforum implements the dislike_forum API of aiotieba.
//
// It mirrors the Python package aiotieba.api.dislike_forum.
package dislikeforum

import (
	"context"
	"net/url"
	"time"

	"github.com/rongyuio/aiotieba/consts"
	"github.com/rongyuio/aiotieba/core"
	"github.com/rongyuio/aiotieba/exception"
	"github.com/rongyuio/aiotieba/helper"
	"github.com/rongyuio/aiotieba/helper/crypto"
)

// dislikeEntry is one element of the `dislike` JSON array.
//
// The field order matters because the JSON text itself is part of the signed
// payload.
type dislikeEntry struct {
	TID         int   `json:"tid"`
	DislikeIDs  int   `json:"dislike_ids"`
	FID         int64 `json:"fid"`
	ClickTimeMs int64 `json:"click_time"`
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
	return &url.URL{Scheme: "https", Host: consts.AppBaseHost, Path: "/c/c/excellent/submitDislike"}
}

// Request mirrors request.
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
