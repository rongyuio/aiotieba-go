// Package sync 实现 aiotieba 的 sync API。
//
// 对应 Python 包 aiotieba.api.sync。
package sync

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// ParseBody 解析 JSON 响应，对应 parse_body。
//
// 返回 client id 与 sample id。
func ParseBody(body []byte) (string, string, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return "", "", err
	}
	if code := helper.JSONInt(res, "error_code"); code != 0 {
		return "", "", &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error_msg")}
	}

	clientID := helper.JSONStr(helper.JSONMap(res, "client"), "client_id")
	sampleID := helper.JSONStr(helper.JSONMap(res, "wl_config"), "sample_id")
	return clientID, sampleID, nil
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   consts.AppBaseHost,
		Path:   "/c/s/sync",
	}
}

// Request 执行 app 表单请求，对应 request。
func Request(ctx context.Context, httpCore *core.HttpCore) (string, string, error) {
	cuidGalaxy2, err := httpCore.Account.CuidGalaxy2()
	if err != nil {
		return "", "", err
	}

	data := []crypto.Param{
		{Key: "BDUSS", Value: httpCore.Account.BDUSS()},
		{Key: "_client_version", Value: consts.LatestVersion},
		{Key: "cuid", Value: cuidGalaxy2},
	}

	resp, err := httpCore.AppForm(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return "", "", err
	}
	return ParseBody(resp.Body())
}
