package getuserjson

import (
	"context"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// ParseBody 解析响应，对应 parse_body。
func ParseBody(body []byte) (UserInfoJSON, error) {
	if len(body) == 0 {
		return UserInfoJSON{}, &exception.TiebaValueError{Msg: "Empty body"}
	}
	// Python 客户端以 errors="ignore" 解码。
	res, err := helper.ParseJSONMap([]byte(ignoreInvalidUTF8(string(body))))
	if err != nil {
		return UserInfoJSON{}, err
	}
	return UserInfoJSONFromJSON(helper.JSONMap(res, "creator")), nil
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{Scheme: "http", Host: consts.WebBaseHost, Path: "/i/sys/user_json"}
}

// Request 执行网页端 GET 请求，对应 request。
func Request(ctx context.Context, httpCore *core.HttpCore, userName string) (UserInfoJSON, error) {
	params := []crypto.Param{
		{Key: "un", Value: userName},
		{Key: "ie", Value: "utf-8"},
	}

	resp, err := httpCore.WebGet(params, nil).SetContext(ctx).Get(RequestURL().String())
	if err != nil {
		return UserInfoJSON{}, err
	}
	return ParseBody(resp.Body())
}

// ignoreInvalidUTF8 丢弃非法的 UTF-8 序列，对应
// `bytes.decode("utf-8", errors="ignore")`。
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
