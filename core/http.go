package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
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

// CallRecord 描述一次 app 接口调用的结果。
//
// 真实客户端的每个请求都会在 CommonReq 里申报「上一次调用」的四个字段
// （m_api / m_logid / m_cost / m_result），本类型承载这四个值。
//
// 证据来自分享链路的抓包：thread/share 请求申报的 m_api 与 m_logid，
// 与紧邻其前的 getShareInfo 调用的路径、以及该调用响应里的 logid 逐字相同。
// 说明这是一条真实可校验的调用链，而不是可以随便编的常量。
type CallRecord struct {
	// API 是上一次调用的路径，含查询串，不含前导斜杠与主机名。
	// 例如 c/f/forum/getShareInfo、c/c/thread/share?cmd=309480&format=protobuf。
	API string

	// LogID 是上一次调用服务端返回的 logid。
	//
	// 只对返回 JSON 的接口可得：响应体里的 "logid" 字段。
	// protobuf 响应的 logid 来源尚未查清，此时该字段为空。
	LogID string

	// Cost 是上一次调用的往返耗时，毫秒，保留 6 位小数。
	Cost string

	// Result 是上一次调用的错误码，成功时为 "0"。
	Result string
}

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

	// lastCall 是最近一次 app / appProto 调用的记录，由 recordCall 中间件更新，
	// 供调用方填充 CommonReq 的 m_* 字段。web 调用不参与该链条。
	lastCall   *CallRecord
	lastCallMu sync.Mutex
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
	// 记录中间件依赖 h，因此必须在 h 创建之后再注册。
	app.OnAfterResponse(h.recordCall)
	appProto.OnAfterResponse(h.recordCall)
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

// recordCall 是 app / appProto 的响应中间件，把本次调用记入 lastCall。
//
// 它不返回错误：记录失败不应影响请求本身的成败。
func (h *HttpCore) recordCall(_ *resty.Client, resp *resty.Response) error {
	if resp == nil || resp.Request == nil {
		return nil
	}
	rec := CallRecord{
		API:    apiPath(resp.Request.URL),
		Cost:   formatCost(resp.Time()),
		LogID:  extractLogID(resp),
		Result: extractErrno(resp),
	}
	h.lastCallMu.Lock()
	h.lastCall = &rec
	h.lastCallMu.Unlock()
	return nil
}

// LastCall 返回最近一次 app / appProto 调用的记录。
//
// 尚未发生过 app 调用时返回零值，此时各字段为空字符串，
// 调用方据此决定是否申报 m_* 字段。
func (h *HttpCore) LastCall() CallRecord {
	h.lastCallMu.Lock()
	defer h.lastCallMu.Unlock()
	if h.lastCall == nil {
		return CallRecord{}
	}
	return *h.lastCall
}

// apiPath 把完整 URL 转成 m_api 使用的形态：去掉 scheme、主机与前导斜杠，保留查询串。
//
// 抓包中该字段形如 c/f/forum/getShareInfo（form 接口无查询串）与
// c/c/thread/share?cmd=309480&format=protobuf（protobuf 接口带查询串）。
func apiPath(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	p := strings.TrimPrefix(u.Path, "/")
	if u.RawQuery != "" {
		return p + "?" + u.RawQuery
	}
	return p
}

// formatCost 把往返耗时格式化为毫秒、6 位小数，与抓包中 m_cost 的形态一致。
func formatCost(d time.Duration) string {
	return fmt.Sprintf("%.6f", float64(d.Nanoseconds())/1e6)
}

// extractLogID 从 JSON 响应体里取 logid 字段。
//
// 只对返回 JSON 的接口有效（如 getShareInfo）。protobuf 响应的 logid 来源尚未查清，
// 此时返回空串，调用方需自行决定是否申报该字段。
//
// 注意 logid 在响应里是**字符串**（"0269126890"），带前导零，不能用 json.Number
// 接收 —— 那个类型只接受 JSON 数字，遇到字符串会直接报错并丢掉整个字段。
func extractLogID(resp *resty.Response) string {
	body := resp.Body()
	if len(body) == 0 || body[0] != '{' {
		return ""
	}
	var doc struct {
		LogID string `json:"logid"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		return ""
	}
	return doc.LogID
}

// extractErrno 从响应里取错误码。
//
// JSON 响应读 error_code（贴吧 form 接口的约定，是 JSON 数字），
// protobuf 响应尚未实现解析，返回空串。
func extractErrno(resp *resty.Response) string {
	body := resp.Body()
	if len(body) == 0 || body[0] != '{' {
		return ""
	}
	var doc struct {
		ErrorCode json.Number `json:"error_code"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		return ""
	}
	return doc.ErrorCode.String()
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
