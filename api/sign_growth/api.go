// Package signgrowth implements the sign_growth API of aiotieba.
//
// It mirrors the Python package aiotieba.api.sign_growth, which exposes both
// the web and the app variant of the endpoint.
package signgrowth

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba/consts"
	"github.com/rongyuio/aiotieba/core"
	"github.com/rongyuio/aiotieba/exception"
	"github.com/rongyuio/aiotieba/helper"
	"github.com/rongyuio/aiotieba/helper/crypto"
)

// ParseBodyWeb mirrors parse_body_web.
func ParseBodyWeb(body []byte) error {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return err
	}
	if code := helper.JSONInt(res, "no"); code != 0 {
		return &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error")}
	}
	return nil
}

// RequestURLWeb returns the endpoint of the web variant.
func RequestURLWeb() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/mo/q/usergrowth/commitUGTaskInfo"}
}

// RequestWeb mirrors request_web.
func RequestWeb(ctx context.Context, httpCore *core.HttpCore, actType string) error {
	data := []crypto.Param{
		{Key: "tbs", Value: httpCore.Account.Tbs()},
		{Key: "act_type", Value: actType},
		{Key: "cuid", Value: "-"},
	}

	req, err := httpCore.PackWebFormRequest(ctx, RequestURLWeb(), data, nil)
	if err != nil {
		return err
	}
	body, err := httpCore.NetCore.SendRequest(req)
	if err != nil {
		return err
	}
	return ParseBodyWeb(body)
}

// ParseBodyApp mirrors parse_body_app.
func ParseBodyApp(body []byte) error {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return err
	}
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}
	return nil
}

// RequestURLApp returns the endpoint of the app variant.
func RequestURLApp() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.AppBaseHost, Path: "/c/c/user/commitUGTaskInfo"}
}

// RequestApp mirrors request_app.
func RequestApp(ctx context.Context, httpCore *core.HttpCore, actType string) error {
	data := []crypto.Param{
		{Key: "BDUSS", Value: httpCore.Account.BDUSS()},
		{Key: "act_type", Value: actType},
		{Key: "cuid", Value: "-"},
		{Key: "tbs", Value: httpCore.Account.Tbs()},
	}

	req, err := httpCore.PackFormRequest(ctx, RequestURLApp(), data)
	if err != nil {
		return err
	}
	body, err := httpCore.NetCore.SendRequest(req)
	if err != nil {
		return err
	}
	return ParseBodyApp(body)
}
