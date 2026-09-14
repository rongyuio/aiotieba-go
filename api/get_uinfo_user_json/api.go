package getuserjson

import (
	"context"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/rongyuio/aiotieba/consts"
	"github.com/rongyuio/aiotieba/core"
	"github.com/rongyuio/aiotieba/exception"
	"github.com/rongyuio/aiotieba/helper"
	"github.com/rongyuio/aiotieba/helper/crypto"
)

// ParseBody mirrors parse_body.
func ParseBody(body []byte) (UserInfoJSON, error) {
	if len(body) == 0 {
		return UserInfoJSON{}, &exception.TiebaValueError{Msg: "Empty body"}
	}
	// The Python client decodes with errors="ignore".
	res, err := helper.ParseJSONMap([]byte(ignoreInvalidUTF8(string(body))))
	if err != nil {
		return UserInfoJSON{}, err
	}
	return UserInfoJSONFromJSON(helper.JSONMap(res, "creator")), nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "http", Host: consts.WebBaseHost, Path: "/i/sys/user_json"}
}

// Request mirrors request.
func Request(ctx context.Context, httpCore *core.HttpCore, userName string) (UserInfoJSON, error) {
	params := []crypto.Param{
		{Key: "un", Value: userName},
		{Key: "ie", Value: "utf-8"},
	}

	req, err := httpCore.PackWebGetRequest(ctx, RequestURL(), params, nil)
	if err != nil {
		return UserInfoJSON{}, err
	}
	body, err := httpCore.SendWeb(req)
	if err != nil {
		return UserInfoJSON{}, err
	}
	return ParseBody(body)
}

// ignoreInvalidUTF8 drops invalid UTF-8 sequences, mirroring
// `bytes.decode("utf-8", errors="ignore")`.
func ignoreInvalidUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}

	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			i++
			continue
		}
		b.WriteString(s[i : i+size])
		i += size
	}
	return b.String()
}
