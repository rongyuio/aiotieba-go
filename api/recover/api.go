// Package recover implements the recover API of aiotieba.
//
// It mirrors the Python package aiotieba.api.recover.
package recover

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba/consts"
	"github.com/rongyuio/aiotieba/core"
	"github.com/rongyuio/aiotieba/exception"
	"github.com/rongyuio/aiotieba/helper"
	"github.com/rongyuio/aiotieba/helper/crypto"
)

// ParseBody mirrors parse_body.
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

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/mo/q/bawurecoverthread"}
}

// Request mirrors request.
//
// A zero pid recovers the whole thread.
func Request(ctx context.Context, httpCore *core.HttpCore, fid, tid, pid int64, isHide bool) error {
	typeList := 0
	if pid != 0 {
		typeList = 1
	}

	data := []crypto.Param{
		{Key: "tbs", Value: httpCore.Account.Tbs()},
		{Key: "fn", Value: "-"},
		{Key: "fid", Value: fid},
		{Key: "tid_list[]", Value: tid},
		{Key: "pid_list[]", Value: pid},
		{Key: "type_list[]", Value: typeList},
		{Key: "is_frs_mask_list[]", Value: helper.BoolInt(isHide)},
	}

	req, err := httpCore.PackWebFormRequest(ctx, RequestURL(), data, nil)
	if err != nil {
		return err
	}
	body, err := httpCore.NetCore.SendRequest(req)
	if err != nil {
		return err
	}
	return ParseBody(body)
}
