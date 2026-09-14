package core

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
	"github.com/rongyuio/aiotieba-go/logging"
)

// protoBoundary is the multipart boundary of the protobuf requests. It is
// hard-coded in the Python client (HttpCore.pack_proto_request).
const protoBoundary = "-*_r1999"

// HttpCore keeps the state of the HTTP interfaces. It mirrors
// aiotieba.core.http.HttpCore.
//
// Each of app, appProto and web is a resty client that shares the NetCore
// connection pool. app and appProto target the mobile API (with the forced
// Host header), while web targets the web API (with the BDUSS/STOKEN cookies).
type HttpCore struct {
	Account  *Account
	NetCore  *NetCore
	app      *resty.Client
	appProto *resty.Client
	web      *resty.Client
}

// NewHttpCore builds the app, app_proto and web resty clients.
func NewHttpCore(account *Account, netCore *NetCore) *HttpCore {
	userAgent := "aiotieba/" + consts.Version
	logger := logging.GetLogger()

	app := netCore.NewRestyClient()
	app.SetHeader("User-Agent", userAgent)
	app.OnBeforeRequest(logRequest(logger))
	app.OnBeforeRequest(hostOverride(consts.AppBaseHost))
	app.OnAfterResponse(checkStatus)

	appProto := netCore.NewRestyClient()
	appProto.SetHeader("User-Agent", userAgent)
	appProto.SetHeader("x_bd_data_type", "protobuf")
	appProto.OnBeforeRequest(logRequest(logger))
	appProto.OnBeforeRequest(hostOverride(consts.AppBaseHost))
	appProto.OnAfterResponse(checkStatus)

	web := netCore.NewRestyClient()
	web.SetHeader("User-Agent", userAgent)
	web.SetHeader("Cache-Control", "no-cache")
	web.OnBeforeRequest(logRequest(logger))
	web.OnAfterResponse(checkStatus)
	// Retry idempotent GET requests on transient network errors only. POST
	// writes (add_post, block, ...) are never retried to avoid side effects.
	web.SetRetryCount(2)
	web.SetRetryWaitTime(200 * time.Millisecond)
	web.SetRetryMaxWaitTime(2 * time.Second)
	web.AddRetryCondition(retryOnIdempotentNetworkError)
	if jar, err := cookiejar.New(nil); err == nil {
		web.SetCookieJar(jar)
	}

	h := &HttpCore{
		Account:  account,
		NetCore:  netCore,
		app:      app,
		appProto: appProto,
		web:      web,
	}
	h.SetAccount(account)
	return h
}

// SetAccount swaps the account of the session, mirroring HttpCore.set_account.
func (h *HttpCore) SetAccount(a *Account) {
	h.Account = a
	h.web.Cookies = []*http.Cookie{
		{Name: "BDUSS", Value: a.BDUSS(), Domain: "baidu.com", Path: "/"},
		{Name: "STOKEN", Value: a.STOKEN(), Domain: "tieba.baidu.com", Path: "/"},
	}
}

// AppForm signs data with APP_SALT and returns an app form request.
func (h *HttpCore) AppForm(data []crypto.Param) *resty.Request {
	signed := crypto.Sign(data, []byte(crypto.AppSalt))
	return h.app.R().
		SetBody(EncodeForm(signed)).
		SetHeader("Content-Type", "application/x-www-form-urlencoded")
}

// AppProto returns an app protobuf multipart request with the fixed boundary.
func (h *HttpCore) AppProto(data []byte) *resty.Request {
	return h.appProto.R().
		SetBody(multipartProtoBody(data)).
		SetHeader("Content-Type", "multipart/form-data; boundary="+protoBoundary)
}

// WebGet returns a web GET request with the params in the query string.
func (h *HttpCore) WebGet(params []crypto.Param, extraHeaders map[string]string) *resty.Request {
	req := h.web.R()
	if len(extraHeaders) > 0 {
		req.SetHeaders(extraHeaders)
	}
	if len(params) > 0 {
		req.SetQueryParams(paramsToMap(params))
	}
	return req
}

// WebForm returns a web form request.
func (h *HttpCore) WebForm(data []crypto.Param) *resty.Request {
	return h.web.R().
		SetBody(EncodeForm(data)).
		SetHeader("Content-Type", "application/x-www-form-urlencoded")
}

// logRequest records the outgoing request metadata at debug level. It never
// logs credentials: BDUSS/STOKEN travel in cookies, not in the URL or headers.
func logRequest(logger *slog.Logger) resty.RequestMiddleware {
	return func(_ *resty.Client, req *resty.Request) error {
		logger.Debug("http request", "method", req.Method, "url", req.URL)
		return nil
	}
}

// retryOnIdempotentNetworkError reports whether a request should be retried:
// only idempotent GET requests that failed with a transport error are retried.
// HTTP error codes (HTTPStatusError) and POST writes are never retried.
func retryOnIdempotentNetworkError(resp *resty.Response, err error) bool {
	if resp != nil && resp.Request != nil && resp.Request.Method != http.MethodGet {
		return false
	}
	var statusErr *exception.HTTPStatusError
	if errors.As(err, &statusErr) {
		return false
	}
	return err != nil
}

// hostOverride forces the Host header on app requests. The Go http client
// ignores a plain "Host" header, so the request Host field is set directly.
func hostOverride(host string) resty.RequestMiddleware {
	return func(_ *resty.Client, req *resty.Request) error {
		if req.RawRequest != nil {
			req.RawRequest.Host = host
		}
		return nil
	}
}

// checkStatus returns an error for non-200 responses, so every caller can
// assume a successful response already carries the 200 status.
func checkStatus(_ *resty.Client, resp *resty.Response) error {
	if resp.StatusCode() != http.StatusOK {
		return &exception.HTTPStatusError{Code: resp.StatusCode(), Msg: resp.Status()}
	}
	return nil
}

// paramsToMap converts the params to a string map for the query string.
func paramsToMap(params []crypto.Param) map[string]string {
	m := make(map[string]string, len(params))
	for _, p := range params {
		m[p.Key] = fmt.Sprintf("%v", p.Value)
	}
	return m
}

// EncodeForm encodes params in order, mirroring urllib.parse.urlencode with
// doseq=True (quote_plus) as used by the Python client.
func EncodeForm(params []crypto.Param) string {
	var b strings.Builder
	for i, p := range params {
		if i > 0 {
			b.WriteByte('&')
		}
		b.WriteString(url.QueryEscape(p.Key))
		b.WriteByte('=')
		b.WriteString(url.QueryEscape(fmt.Sprintf("%v", p.Value)))
	}
	return b.String()
}

func multipartProtoBody(data []byte) []byte {
	var buf bytes.Buffer
	buf.Grow(len(data) + 128)
	buf.WriteString("--" + protoBoundary + "\r\n")
	buf.WriteString(`Content-Disposition: form-data; name="data"; filename="file"`)
	buf.WriteString("\r\n\r\n")
	buf.Write(data)
	buf.WriteString("\r\n--" + protoBoundary + "--\r\n")
	return buf.Bytes()
}
