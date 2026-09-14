package getmemberusers

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
	"github.com/rongyuio/aiotieba-go/helper/htmlutil"
)

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/bawu2/platform/listMemberInfo"}
}

// Request 执行网页端 GET 请求，对应 request。
func Request(ctx context.Context, httpCore *core.HttpCore, fname string, pn int64) (MemberUsers, error) {
	params := []crypto.Param{
		{Key: "word", Value: fname},
		{Key: "pn", Value: pn},
		{Key: "ie", Value: "utf-8"},
	}
	resp, err := httpCore.WebGet(params, nil).SetContext(ctx).Get(RequestURL().String())
	if err != nil {
		return MemberUsers{}, err
	}

	soup, err := htmlutil.Parse(resp.Body())
	if err != nil {
		return MemberUsers{}, err
	}
	return MemberUsersFromXML(soup), nil
}
