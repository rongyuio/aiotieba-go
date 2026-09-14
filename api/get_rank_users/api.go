package getrankusers

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
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/f/like/furank"}
}

// Request performs the web get request, mirroring request.
func Request(ctx context.Context, httpCore *core.HttpCore, fname string, pn int64) (RankUsers, error) {
	params := []crypto.Param{
		{Key: "kw", Value: fname},
		{Key: "pn", Value: pn},
		{Key: "ie", Value: "utf-8"},
	}
	req, err := httpCore.PackWebGetRequest(ctx, RequestURL(), params, nil)
	if err != nil {
		return RankUsers{}, err
	}
	body, err := httpCore.SendWeb(req)
	if err != nil {
		return RankUsers{}, err
	}

	soup, err := htmlutil.Parse(body)
	if err != nil {
		return RankUsers{}, err
	}
	return RankUsersFromXML(soup), nil
}
