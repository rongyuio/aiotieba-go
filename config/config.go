// Package config defines the proxy and timeout configuration used by the client.
//
// It mirrors the Python module aiotieba.config.
package config

import (
	"net/url"
	"os"
	"time"
)

// BasicAuth carries the credentials of a proxy.
type BasicAuth struct {
	Login    string
	Password string
}

// ProxyConfig configures an optional HTTP(S) proxy.
//
// A nil URL disables the proxy.
type ProxyConfig struct {
	URL  *url.URL
	Auth *BasicAuth
}

// NewProxyConfig parses raw into a ProxyConfig. An empty raw disables the proxy.
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

// FromEnv builds a ProxyConfig from the standard proxy environment variables.
//
// It checks http_proxy/HTTP_PROXY first, then https_proxy/HTTPS_PROXY.
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

// TimeoutConfig holds every timeout used by the client.
//
// All durations are absolute values; the zero value is not usable, use
// DefaultTimeoutConfig instead.
type TimeoutConfig struct {
	// HTTPAcquireConn is the timeout to acquire an idle connection from the pool.
	HTTPAcquireConn time.Duration
	// HTTPRead is the timeout from sending a request to reading the whole response.
	HTTPRead time.Duration
	// HTTPConnect is the timeout to establish a new socket connection.
	HTTPConnect time.Duration
	// HTTPKeepAlive is how long an idle HTTP connection is kept.
	HTTPKeepAlive time.Duration
	// WSSend is the timeout to send a websocket frame.
	WSSend time.Duration
	// WSRead is the timeout from sending a websocket frame to receiving its response.
	WSRead time.Duration
	// WSClose is the timeout to wait for a websocket connection to close.
	WSClose time.Duration
	// WSKeepAlive closes the websocket connection after this idle period.
	WSKeepAlive time.Duration
	// WSHeartbeat is the interval between heartbeats; 0 disables heartbeats.
	WSHeartbeat time.Duration
	// DNSTTL is the local DNS cache lifetime.
	DNSTTL time.Duration
}

// DefaultTimeoutConfig returns the same defaults as Python's TimeoutConfig.
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
