// Package setnicknameold implements the set_nickname_old API of aiotieba.
//
// It mirrors the Python package aiotieba.api.set_nickname_old.
package setnicknameold

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba/consts"
	"github.com/rongyuio/aiotieba/core"
	"github.com/rongyuio/aiotieba/exception"
	"github.com/rongyuio/aiotieba/helper"
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
//
// The parameters travel in the query string and the request body is empty,
// mirroring the Python module.
func RequestURL(nickName string) *url.URL {
	q := url.Values{}
	q.Set("nickname", nickName)
	q.Set("tbs", "1")
	return &url.URL{
		Scheme:   "https",
		Host:     consts.WebBaseHost,
		Path:     "/mo/q/submit/modifyNickname",
		RawQuery: q.Encode(),
	}
}

// Request mirrors request.
func Request(ctx context.Context, httpCore *core.HttpCore, nickName string) error {
	req, err := httpCore.PackWebFormRequest(ctx, RequestURL(nickName), nil, nil)
	if err != nil {
		return err
	}
	body, err := httpCore.NetCore.SendRequest(req)
	if err != nil {
		return err
	}
	return ParseBody(body)
}
