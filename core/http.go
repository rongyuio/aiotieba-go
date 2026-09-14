package core

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// protoBoundary is the multipart boundary of the protobuf requests. It is
// hard-coded in the Python client (HttpCore.pack_proto_request).
const protoBoundary = "-*_r1999"

// HttpContainer keeps the headers (and cookies) of one request family.
type HttpContainer struct {
	headers map[string]string
	cookies []*http.Cookie
	jar     http.CookieJar
}

func newContainer(headers map[string]string, withJar bool) *HttpContainer {
	c := &HttpContainer{headers: headers}
	if withJar {
		c.jar, _ = cookiejar.New(nil)
	}
	return c
}

// cookiesFor returns the cookies that apply to u, mirroring
// aiohttp.CookieJar.filter_cookies.
func (c *HttpContainer) cookiesFor(u *url.URL) []*http.Cookie {
	host := u.Hostname()
	var out []*http.Cookie
	for _, ck := range c.cookies {
		if domainMatch(host, ck.Domain) {
			out = append(out, ck)
		}
	}
	if c.jar != nil {
		out = append(out, c.jar.Cookies(u)...)
	}
	return out
}

func domainMatch(host, domain string) bool {
	if domain == "" {
		return true
	}
	return host == domain || strings.HasSuffix(host, "."+domain)
}

// HttpCore keeps the state of the HTTP interfaces. It mirrors
// aiotieba.core.http.HttpCore.
type HttpCore struct {
	Account  *Account
	NetCore  *NetCore
	App      *HttpContainer
	AppProto *HttpContainer
	Web      *HttpContainer
}

// NewHttpCore builds the app, app_proto and web session containers.
func NewHttpCore(account *Account, netCore *NetCore) *HttpCore {
	userAgent := "aiotieba/" + consts.Version
	h := &HttpCore{
		NetCore: netCore,
		App: newContainer(map[string]string{
			"User-Agent": userAgent,
		}, false),
		AppProto: newContainer(map[string]string{
			"User-Agent":     userAgent,
			"x_bd_data_type": "protobuf",
		}, false),
		Web: newContainer(map[string]string{
			"User-Agent":    userAgent,
			"Cache-Control": "no-cache",
		}, true),
	}
	h.SetAccount(account)
	return h
}

// SetAccount swaps the account of the session, mirroring HttpCore.set_account.
func (h *HttpCore) SetAccount(a *Account) {
	h.Account = a
	h.Web.cookies = []*http.Cookie{
		{Name: "BDUSS", Value: a.BDUSS(), Domain: "baidu.com", Path: "/"},
		{Name: "STOKEN", Value: a.STOKEN(), Domain: "tieba.baidu.com", Path: "/"},
	}
}

// PackFormRequest signs data with APP_SALT and packs a mobile form request.
func (h *HttpCore) PackFormRequest(ctx context.Context, u *url.URL, data []crypto.Param) (*http.Request, error) {
	signed := crypto.Sign(data, []byte(crypto.AppSalt))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), strings.NewReader(EncodeForm(signed)))
	if err != nil {
		return nil, fmt.Errorf("building form request: %w", err)
	}
	applyHeaders(req, h.App.headers, nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Host = consts.AppBaseHost
	return req, nil
}

// PackProtoRequest packs a mobile protobuf request.
func (h *HttpCore) PackProtoRequest(ctx context.Context, u *url.URL, data []byte) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(multipartProtoBody(data)))
	if err != nil {
		return nil, fmt.Errorf("building protobuf request: %w", err)
	}
	applyHeaders(req, h.AppProto.headers, nil)
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+protoBoundary)
	req.Host = consts.AppBaseHost
	return req, nil
}

// PackWebGetRequest packs a web GET request with the params in the query string.
func (h *HttpCore) PackWebGetRequest(ctx context.Context, u *url.URL, params []crypto.Param, extraHeaders map[string]string) (*http.Request, error) {
	target := *u
	target.RawQuery = EncodeForm(params)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("building web get request: %w", err)
	}
	applyHeaders(req, h.Web.headers, extraHeaders)
	for _, ck := range h.Web.cookiesFor(&target) {
		req.AddCookie(ck)
	}
	return req, nil
}

// PackWebFormRequest packs a web form request.
func (h *HttpCore) PackWebFormRequest(ctx context.Context, u *url.URL, data []crypto.Param, extraHeaders map[string]string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), strings.NewReader(EncodeForm(data)))
	if err != nil {
		return nil, fmt.Errorf("building web form request: %w", err)
	}
	applyHeaders(req, h.Web.headers, extraHeaders)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, ck := range h.Web.cookiesFor(u) {
		req.AddCookie(ck)
	}
	return req, nil
}

// SendProto sends req over the app session.
func (h *HttpCore) SendProto(req *http.Request) ([]byte, error) {
	return h.NetCore.SendRequest(req)
}

// SendWeb sends req over the web session and stores the response cookies in the
// web cookie jar, mirroring aiohttp.CookieJar.
func (h *HttpCore) SendWeb(req *http.Request) ([]byte, error) {
	resp, err := h.NetCore.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &exception.HTTPStatusError{Code: resp.StatusCode, Msg: resp.Status}
	}
	if h.Web.jar != nil {
		if setCookies := resp.Cookies(); len(setCookies) > 0 {
			h.Web.jar.SetCookies(req.URL, setCookies)
		}
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	return body, nil
}

func applyHeaders(req *http.Request, base, extra map[string]string) {
	for k, v := range base {
		req.Header.Set(k, v)
	}
	for k, v := range extra {
		req.Header.Set(k, v)
	}
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
