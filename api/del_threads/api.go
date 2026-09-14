// Package delthreads implements the del_threads API of aiotieba.
//
// It mirrors the Python package aiotieba.api.del_threads.
package delthreads

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
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
	return &url.URL{Scheme: "https", Host: consts.AppBaseHost, Path: "/c/c/bawu/multiDelThread"}
}

// Request mirrors request.
//
// block selects the removal mode: a normal delete or a delete plus ban.
func Request(ctx context.Context, httpCore *core.HttpCore, fid int64, tids []int64, block bool) error {
	kind := 1
	if block {
		kind = 2
	}

	parts := make([]string, len(tids))
	for i, tid := range tids {
		parts[i] = strconv.FormatInt(tid, 10)
	}

	data := []crypto.Param{
		{Key: "BDUSS", Value: httpCore.Account.BDUSS()},
		{Key: "forum_id", Value: fid},
		{Key: "tbs", Value: httpCore.Account.Tbs()},
		{Key: "thread_ids", Value: strings.Join(parts, ",")},
		{Key: "type", Value: kind},
	}

	resp, err := httpCore.AppForm(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return err
	}
	return ParseBody(resp.Body())
}
