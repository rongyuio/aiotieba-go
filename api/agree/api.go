// Package agree implements the agree API of aiotieba.
//
// It mirrors the Python package aiotieba.api.agree.
package agree

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// Object types reported to the server.
const (
	objTypeThread  = 1
	objTypeComment = 2
	objTypePost    = 3

	// agreeTypePositive is the agree_type of a normal agreement.
	agreeTypePositive = 2
	// agreeTypeNegative is the agree_type of a disagreement.
	agreeTypeNegative = 5
)

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
	return &url.URL{Scheme: "https", Host: consts.AppBaseHost, Path: "/c/c/agree/opAgree"}
}

// ObjType returns the obj_type reported to the server. A zero pid targets the
// thread itself.
func ObjType(pid int64, isComment bool) int {
	if pid == 0 {
		return objTypeThread
	}
	if isComment {
		return objTypeComment
	}
	return objTypePost
}

// AgreeType returns the agree_type reported to the server.
func AgreeType(isDisagree bool) int {
	if isDisagree {
		return agreeTypeNegative
	}
	return agreeTypePositive
}

// Request mirrors request.
//
// A zero pid targets the thread itself.
func Request(ctx context.Context, httpCore *core.HttpCore, tid, pid int64, isComment, isDisagree, isUndo bool) error {
	cuidGalaxy2, err := httpCore.Account.CuidGalaxy2()
	if err != nil {
		return err
	}

	data := []crypto.Param{
		{Key: "BDUSS", Value: httpCore.Account.BDUSS()},
		{Key: "_client_version", Value: consts.LatestVersion},
		{Key: "agree_type", Value: AgreeType(isDisagree)},
		{Key: "cuid", Value: cuidGalaxy2},
		{Key: "obj_type", Value: ObjType(pid, isComment)},
		{Key: "op_type", Value: helper.BoolInt(isUndo)},
		{Key: "post_id", Value: pid},
		{Key: "tbs", Value: httpCore.Account.Tbs()},
		{Key: "thread_id", Value: tid},
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
