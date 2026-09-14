package getbawumemberlist

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
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/bawu2/platform/listMember"}
}

// Request performs the web get request, mirroring request.
func Request(ctx context.Context, httpCore *core.HttpCore, fname string, pn int64, searchValue string) (BawuListMemberUsers, error) {
	params := []crypto.Param{
		{Key: "word", Value: fname},
		{Key: "pn", Value: pn},
		{Key: "ie", Value: "utf-8"},
	}
	if searchValue != "" {
		params = append(params,
			crypto.Param{Key: "svalue", Value: url.QueryEscape(searchValue)},
			crypto.Param{Key: "stype", Value: "uname"},
		)
	}
	req, err := httpCore.PackWebGetRequest(ctx, RequestURL(), params, nil)
	if err != nil {
		return BawuListMemberUsers{}, err
	}
	body, err := httpCore.SendWeb(req)
	if err != nil {
		return BawuListMemberUsers{}, err
	}

	soup, err := htmlutil.Parse(body)
	if err != nil {
		return BawuListMemberUsers{}, err
	}
	return BawuListMemberUsersFromXML(soup), nil
}
