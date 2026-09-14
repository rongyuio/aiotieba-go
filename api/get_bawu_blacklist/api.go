package getbawublacklist

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba/consts"
	"github.com/rongyuio/aiotieba/core"
	"github.com/rongyuio/aiotieba/helper/crypto"
	"github.com/rongyuio/aiotieba/helper/htmlutil"
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
	req, err := httpCore.PackWebGetRequest(ctx, RequestURL(), params, nil)
	if err != nil {
		return BawuBlacklistUsers{}, err
	}
	body, err := httpCore.SendWeb(req)
	if err != nil {
		return BawuBlacklistUsers{}, err
	}

	soup, err := htmlutil.Parse(body)
	if err != nil {
		return BawuBlacklistUsers{}, err
	}
	return BawuBlacklistUsersFromXML(soup), nil
}
