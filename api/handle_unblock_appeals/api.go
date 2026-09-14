// Package handleunblockappeals implements the handle_unblock_appeals API of aiotieba.
//
// It mirrors the Python package aiotieba.api.handle_unblock_appeals.
package handleunblockappeals

import (
	"context"
	"net/url"
	"strconv"

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
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/mo/q/multiAppealhandle"}
}

// Request mirrors request.
//
// refuse selects between rejecting (status 2) and accepting (status 1) the
// appeals.
func Request(ctx context.Context, httpCore *core.HttpCore, fid int64, appealIDs []int64, refuse bool) error {
	status := 1
	if refuse {
		status = 2
	}

	data := make([]crypto.Param, 0, len(appealIDs)+5)
	data = append(data,
		crypto.Param{Key: "fn", Value: "-"},
		crypto.Param{Key: "fid", Value: fid},
	)
	for i, appealID := range appealIDs {
		data = append(data, crypto.Param{Key: "appeal_list[" + strconv.Itoa(i) + "]", Value: appealID})
	}
	data = append(data,
		crypto.Param{Key: "refuse_reason", Value: "_"},
		crypto.Param{Key: "status", Value: status},
		crypto.Param{Key: "tbs", Value: httpCore.Account.Tbs()},
	)

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
