package core

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
	"github.com/rongyuio/aiotieba-go/logging"
)

// protoBoundary 是 protobuf 请求的 multipart 边界，Python 客户端中在
// HttpCore.pack_proto_request 里硬编码为该值。
const protoBoundary = "-*_r1999"

// HttpCore 保存 http 接口相关状态的核心容器，对应 aiotieba.core.http.HttpCore。
//
// app、appProto 与 web 都是共享 NetCore 连接池的 resty 客户端：app 与 appProto 面向
// 移动端 API（强制 Host 头），web 面向网页端 API（携带 BDUSS/STOKEN Cookie）。
type HttpCore struct {
	Account  *Account
	NetCore  *NetCore
	app      *resty.Client
	appProto *resty.Client
	web      *resty.Client
}

// NewHttpCore 构建 app、app_proto 与 web 三个 resty 客户端。
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
	// 仅对瞬时网络错误导致的幂等 GET 请求重试。POST 写操作（add_post、block 等）
	// 一律不重试，避免产生副作用。
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

// SetAccount 替换会话的账号，对应 HttpCore.set_account。
func (h *HttpCore) SetAccount(a *Account) {
	h.Account = a
	h.web.Cookies = []*http.Cookie{
		{Name: "BDUSS", Value: a.BDUSS(), Domain: "baidu.com", Path: "/"},
		{Name: "STOKEN", Value: a.STOKEN(), Domain: "tieba.baidu.com", Path: "/"},
	}
}

// AppForm 自动签名参数元组列表，并将其打包为移动端表单请求。
//
// 参数:
//
//	data 参数元组列表
//
// 使用 APP_SALT 对 data 签名后返回移动端表单请求。
func (h *HttpCore) AppForm(data []crypto.Param) *resty.Request {
	signed := crypto.Sign(data, []byte(crypto.AppSalt))
	return h.app.R().
		SetBody(EncodeForm(signed)).
		SetHeader("Content-Type", "application/x-www-form-urlencoded")
}

// AppProto 打包移动端 protobuf 请求。
//
// 参数:
//
//	data protobuf序列化后的二进制数据
//
// 使用固定边界返回移动端 protobuf multipart 请求。
func (h *HttpCore) AppProto(data []byte) *resty.Request {
	return h.appProto.R().
		SetBody(multipartProtoBody(data)).
		SetHeader("Content-Type", "multipart/form-data; boundary="+protoBoundary)
}

// WebGet 打包网页端参数请求，参数放在查询字符串中。
//
// 参数:
//
//	params 参数元组列表
//	extraHeaders 额外的请求头
//
// 返回网页端 GET 请求，参数放在查询字符串中。
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

// WebForm 打包网页端表单请求。
//
// 参数:
//
//	data 参数元组列表
//
// 返回网页端表单请求。
func (h *HttpCore) WebForm(data []crypto.Param) *resty.Request {
	return h.web.R().
		SetBody(EncodeForm(data)).
		SetHeader("Content-Type", "application/x-www-form-urlencoded")
}

// logRequest 以 debug 级别记录发出的请求元数据。它不会记录凭据：
// BDUSS/STOKEN 走 Cookie，不出现在 URL 或请求头中。
func logRequest(logger *zerolog.Logger) resty.RequestMiddleware {
	return func(_ *resty.Client, req *resty.Request) error {
		logger.Debug().Str("method", req.Method).Str("url", req.URL).Msg("http request")
		return nil
	}
}

// retryOnIdempotentNetworkError 报告请求是否应当重试：只有因传输层错误失败的
// 幂等 GET 请求才会重试；HTTP 错误码（HTTPStatusError）与 POST 写操作一律不重试。
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

// hostOverride 强制设置 app 请求的 Host 头。Go 的 http 客户端会忽略普通的
// "Host" 头，因此直接设置请求的 Host 字段。
func hostOverride(host string) resty.RequestMiddleware {
	return func(_ *resty.Client, req *resty.Request) error {
		if req.RawRequest != nil {
			req.RawRequest.Host = host
		}
		return nil
	}
}

// checkStatus 对非 200 响应返回错误，这样所有调用方都可以假定成功响应已经是 200 状态码。
func checkStatus(_ *resty.Client, resp *resty.Response) error {
	if resp.StatusCode() != http.StatusOK {
		return &exception.HTTPStatusError{Code: resp.StatusCode(), Msg: resp.Status()}
	}
	return nil
}

// paramsToMap 把参数转换为用于查询字符串的字符串映射。
func paramsToMap(params []crypto.Param) map[string]string {
	m := make(map[string]string, len(params))
	for _, p := range params {
		m[p.Key] = fmt.Sprintf("%v", p.Value)
	}
	return m
}

// EncodeForm 按顺序编码参数，对应 Python 客户端使用的
// urllib.parse.urlencode(doseq=True)（quote_plus）。
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
