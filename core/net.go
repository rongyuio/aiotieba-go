package core

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/rongyuio/aiotieba-go/config"
	"github.com/rongyuio/aiotieba-go/exception"
)

// NetCore 网络请求相关容器，持有共享连接池与代理、超时配置，对应 aiotieba.core.net.NetCore。
//
// HTTP 传输层由 HttpCore 构建的各 resty 客户端共享，使每个会话复用同一个连接池。
type NetCore struct {
	transport *http.Transport
	proxy     *config.ProxyConfig
	timeout   config.TimeoutConfig

	// client 是过渡用的 http.Client，供 Do/SendRequest 使用，直到 api 各包迁移到 resty 客户端。
	client *http.Client
}

// NewNetCore 构建共享的 HTTP 传输层。
//
// 参数:
//
//	proxy 代理配置
//	timeout 超时配置
//
// proxy 为 nil 时禁用代理。
func NewNetCore(proxy *config.ProxyConfig, timeout config.TimeoutConfig) *NetCore {
	if proxy == nil {
		proxy = &config.ProxyConfig{}
	}

	dialer := &net.Dialer{
		Timeout:   timeout.HTTPConnect,
		KeepAlive: timeout.HTTPKeepAlive,
	}
	transport := &http.Transport{
		Proxy:                 proxyFunc(proxy),
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     false,
		MaxIdleConns:          0, // unlimited, mirrors aiohttp limit=0
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       timeout.HTTPKeepAlive,
		TLSHandshakeTimeout:   timeout.HTTPConnect,
		ResponseHeaderTimeout: timeout.HTTPRead,
		ExpectContinueTimeout: time.Second,
		// 贴吧的 app API 走明文 HTTP，且 Python 客户端禁用了校验（aiohttp ssl=False），
		// 因此这里同样不校验证书。
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS12,
		},
	}

	return &NetCore{
		transport: transport,
		proxy:     proxy,
		timeout:   timeout,
		client:    &http.Client{Transport: transport, Timeout: timeout.HTTPRead},
	}
}

// Proxy 返回代理配置。
func (n *NetCore) Proxy() *config.ProxyConfig { return n.proxy }

// Timeout 返回超时配置。
func (n *NetCore) Timeout() config.TimeoutConfig { return n.timeout }

// Transport 返回各 resty 客户端共享的 HTTP 传输层。
func (n *NetCore) Transport() *http.Transport { return n.transport }

// NewRestyClient 构建共享 NetCore 连接池的 resty 客户端。
// TLS、代理与 keep-alive 设置都由共享的传输层承载。
func (n *NetCore) NewRestyClient() *resty.Client {
	c := resty.New()
	c.SetTransport(n.transport)
	c.SetTimeout(n.timeout.HTTPRead)
	return c
}

// Do 发送 req 并返回原始响应，不检查状态码。
//
// Deprecated: 仅在迁移到 resty 期间保留，api 各包应改用 HttpCore 提供的 resty 客户端。
func (n *NetCore) Do(req *http.Request) (*http.Response, error) {
	resp, err := n.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending %s %s: %w", req.Method, req.URL.Redacted(), err)
	}
	return resp, nil
}

// SendRequest 简单发送 http 请求 不包含重定向和身份验证功能。
//
// 参数:
//
//	req 待发送的请求
//
// 期望返回 200 状态码并返回完整 body，对应 NetCore.send_request。Go 会自动加上
// "Accept-Encoding: gzip" 并透明解压响应，因此 body 始终是解码后的内容。
//
// Deprecated: 见 Do。
func (n *NetCore) SendRequest(req *http.Request) ([]byte, error) {
	resp, err := n.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &exception.HTTPStatusError{Code: resp.StatusCode, Msg: resp.Status}
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	return body, nil
}

// Close 释放连接池中的空闲连接。
func (n *NetCore) Close() {
	n.client.CloseIdleConnections()
}

// proxyFunc 把 ProxyConfig 转换为代理函数。它始终返回非 nil 函数，避免调用方回退到环境变量。
func proxyFunc(p *config.ProxyConfig) func(*http.Request) (*url.URL, error) {
	if p == nil || p.URL == nil {
		return func(*http.Request) (*url.URL, error) { return nil, nil }
	}
	u := *p.URL
	if p.Auth != nil {
		u.User = url.UserPassword(p.Auth.Login, p.Auth.Password)
	}
	return http.ProxyURL(&u)
}
