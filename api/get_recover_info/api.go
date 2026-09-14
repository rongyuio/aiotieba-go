package getrecoverinfo

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
func ParseBody(body []byte) (RecoverInfo, error) {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return RecoverInfo{}, err
	}
	if code := helper.JSONInt(res, "no"); code != 0 {
		return RecoverInfo{}, &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error")}
	}
	return RecoverInfoFromJSON(helper.JSONMap(res, "data")), nil
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/mo/q/bawu/getRecoverInfo"}
}

// Request 执行网页端 GET 请求，对应 request。
func Request(ctx context.Context, httpCore *core.HttpCore, fid, tid, pid int64) (RecoverInfo, error) {
	subType := 1
	if pid != 0 {
		subType = 2
	}
	params := []crypto.Param{
		{Key: "forum_id", Value: fid},
		{Key: "thread_id", Value: tid},
		{Key: "post_id", Value: pid},
		{Key: "type", Value: 1},
		{Key: "sub_type", Value: subType},
	}
	resp, err := httpCore.WebGet(params, nil).SetContext(ctx).Get(RequestURL().String())
	if err != nil {
		return RecoverInfo{}, err
	}
	return ParseBody(resp.Body())
}
