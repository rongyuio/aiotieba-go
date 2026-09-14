// Package signgrowth 实现 aiotieba 的 sign_growth API。
//
// 对应 Python 包 aiotieba.api.sign_growth，它同时提供该接口的网页端与 app 变体。
package signgrowth

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// ParseBodyWeb 对应 parse_body_web。
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

// RequestURLWeb 返回网页端变体的请求地址。
func RequestURLWeb() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/mo/q/usergrowth/commitUGTaskInfo"}
}

// RequestWeb 执行网页端表单请求，对应 request_web。
func RequestWeb(ctx context.Context, httpCore *core.HttpCore, actType string) error {
	data := []crypto.Param{
		{Key: "tbs", Value: httpCore.Account.Tbs()},
		{Key: "act_type", Value: actType},
		{Key: "cuid", Value: "-"},
	}

	resp, err := httpCore.WebForm(data).SetContext(ctx).Post(RequestURLWeb().String())
	if err != nil {
		return err
	}
	return ParseBodyWeb(resp.Body())
}

// ParseBodyApp 对应 parse_body_app。
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

// RequestURLApp 返回 app 变体的请求地址。
func RequestURLApp() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.AppBaseHost, Path: "/c/c/user/commitUGTaskInfo"}
}

// RequestApp 执行 app 表单请求，对应 request_app。
func RequestApp(ctx context.Context, httpCore *core.HttpCore, actType string) error {
	data := []crypto.Param{
		{Key: "BDUSS", Value: httpCore.Account.BDUSS()},
		{Key: "act_type", Value: actType},
		{Key: "cuid", Value: "-"},
		{Key: "tbs", Value: httpCore.Account.Tbs()},
	}

	resp, err := httpCore.AppForm(data).SetContext(ctx).Post(RequestURLApp().String())
	if err != nil {
		return err
	}
	return ParseBodyApp(resp.Body())
}
