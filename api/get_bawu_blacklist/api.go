package getbawublacklist

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
	"github.com/rongyuio/aiotieba-go/helper/htmlutil"
)

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/bawu2/platform/listBlackUser"}
}

// Request performs the web get request, mirroring request.
func Request(ctx context.Context, httpCore *core.HttpCore, fname string, pn int64) (BawuBlacklistUsers, error) {
	params := []crypto.Param{
		{Key: "word", Value: fname},
		{Key: "pn", Value: pn},
	}
	resp, err := httpCore.WebGet(params, nil).SetContext(ctx).Get(RequestURL().String())
	if err != nil {
		return BawuBlacklistUsers{}, err
	}

	soup, err := htmlutil.Parse(resp.Body())
	if err != nil {
		return BawuBlacklistUsers{}, err
	}
	return BawuBlacklistUsersFromXML(soup), nil
}
