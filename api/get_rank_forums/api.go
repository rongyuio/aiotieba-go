package getrankforums

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
	"github.com/rongyuio/aiotieba-go/helper/htmlutil"
)

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/sign/index"}
}

// Request performs the web get request, mirroring request.
func Request(ctx context.Context, httpCore *core.HttpCore, fname string, pn int64, rankType enums.RankForumType) (RankForums, error) {
	params := []crypto.Param{
		{Key: "kw", Value: fname},
		{Key: "type", Value: int(rankType)},
		{Key: "pn", Value: pn},
		{Key: "ie", Value: "utf-8"},
	}
	resp, err := httpCore.WebGet(params, nil).SetContext(ctx).Get(RequestURL().String())
	if err != nil {
		return RankForums{}, err
	}

	soup, err := htmlutil.Parse(resp.Body())
	if err != nil {
		return RankForums{}, err
	}
	return RankForumsFromXML(soup), nil
}
