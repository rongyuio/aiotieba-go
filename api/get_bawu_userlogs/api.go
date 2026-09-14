package getbawuuserlogs

import (
	"context"
	"net/url"
	"time"

	"github.com/rongyuio/aiotieba/consts"
	"github.com/rongyuio/aiotieba/core"
	"github.com/rongyuio/aiotieba/enums"
	"github.com/rongyuio/aiotieba/helper/crypto"
	"github.com/rongyuio/aiotieba/helper/htmlutil"
)

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/bawu2/platform/listUserLog"}
}

// Request performs the web get request, mirroring request.
func Request(
	ctx context.Context,
	httpCore *core.HttpCore,
	fname string,
	pn int64,
	searchValue string,
	searchType enums.BawuSearchType,
	startDT, endDT *time.Time,
	opType int64,
) (BawuUserLogs, error) {
	params := []crypto.Param{
		{Key: "word", Value: fname},
		{Key: "pn", Value: pn},
		{Key: "ie", Value: "utf-8"},
	}
	if opType != 0 {
		params = append(params, crypto.Param{Key: "op_type", Value: opType})
	}
	if searchValue != "" {
		if searchType == enums.BawuSearchUser {
			params = append(params,
				crypto.Param{Key: "svalue", Value: url.QueryEscape(searchValue)},
				crypto.Param{Key: "stype", Value: "post_uname"},
			)
		} else {
			params = append(params,
				crypto.Param{Key: "svalue", Value: searchValue},
				crypto.Param{Key: "stype", Value: "op_uname"},
			)
		}
	}
	if startDT != nil {
		begin := startDT.Unix()
		end := time.Now().Unix()
		if endDT != nil {
			end = endDT.Unix()
		}
		params = append(params,
			crypto.Param{Key: "end", Value: end},
			crypto.Param{Key: "begin", Value: begin},
		)
	}

	req, err := httpCore.PackWebGetRequest(ctx, RequestURL(), params, nil)
	if err != nil {
		return BawuUserLogs{}, err
	}
	body, err := httpCore.SendWeb(req)
	if err != nil {
		return BawuUserLogs{}, err
	}

	soup, err := htmlutil.Parse(body)
	if err != nil {
		return BawuUserLogs{}, err
	}
	return BawuUserLogsFromXML(soup), nil
}
