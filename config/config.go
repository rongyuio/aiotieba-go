// Package config 定义客户端使用的代理与超时配置。
//
// 对应 Python 模块 aiotieba.config。
package config

import (
	"net/url"
	"os"
	"time"
)

// BasicAuth 代理认证信息。
type BasicAuth struct {
	Login    string // 用户名
	Password string // 密码
}

// ProxyConfig 代理配置。URL 为 nil 即禁用代理。
type ProxyConfig struct {
	URL  *url.URL   // 代理url
	Auth *BasicAuth // 代理认证
}

// NewProxyConfig 把 raw 解析为 ProxyConfig，raw 为空即禁用代理。
//
// 参数:
//
//	raw 代理url
//	auth 代理认证
func NewProxyConfig(raw string, auth *BasicAuth) (*ProxyConfig, error) {
	if raw == "" {
		return &ProxyConfig{Auth: auth}, nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	return &ProxyConfig{URL: u, Auth: auth}, nil
}

// FromEnv 从标准代理环境变量构建 ProxyConfig。
//
// 优先检查 http_proxy/HTTP_PROXY，其次是 https_proxy/HTTPS_PROXY。
func FromEnv() *ProxyConfig {
	raw := firstEnv("http_proxy", "HTTP_PROXY", "https_proxy", "HTTPS_PROXY")
	proxy, err := NewProxyConfig(raw, nil)
	if err != nil {
		return &ProxyConfig{}
	}
	return proxy
}

func firstEnv(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}

// TimeoutConfig 各种超时配置。所有时间均以秒为单位。
//
// 保存客户端使用的所有超时。所有时长都是绝对值，零值不可用，请使用 DefaultTimeoutConfig。
type TimeoutConfig struct {
	// HTTPAcquireConn 从连接池获取一个可用连接的超时时间。
	HTTPAcquireConn time.Duration
	// HTTPRead 从发送http请求到读取全部响应的超时时间。
	HTTPRead time.Duration
	// HTTPConnect 新建一个socket连接的超时时间。
	HTTPConnect time.Duration
	// HTTPKeepAlive http长连接的保持时间。
	HTTPKeepAlive time.Duration
	// WSSend websocket发送数据的超时时间。
	WSSend time.Duration
	// WSRead 从发送websocket数据到结束等待响应的超时时间。
	WSRead time.Duration
	// WSClose 等待websocket终止连接的时间。
	WSClose time.Duration
	// WSKeepAlive websocket在长达 WSKeepAlive 的时间内未发生IO则发送close信号关闭连接。
	WSKeepAlive time.Duration
	// WSHeartbeat websocket心跳间隔。为 0 则不发送心跳。
	WSHeartbeat time.Duration
	// DNSTTL dns的本地缓存超时时间。
	DNSTTL time.Duration
}

// DefaultTimeoutConfig 返回与 Python TimeoutConfig 相同的默认值。
func DefaultTimeoutConfig() TimeoutConfig {
	return TimeoutConfig{
		HTTPAcquireConn: 4 * time.Second,
		HTTPRead:        12 * time.Second,
		HTTPConnect:     3 * time.Second,
		HTTPKeepAlive:   30 * time.Second,
		WSSend:          3 * time.Second,
		WSRead:          8 * time.Second,
		WSClose:         10 * time.Second,
		WSKeepAlive:     300 * time.Second,
		WSHeartbeat:     0,
		DNSTTL:          600 * time.Second,
	}
}
