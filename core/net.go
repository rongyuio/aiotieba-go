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

// NetCore owns the shared connection pool together with the proxy and timeout
// configuration. It mirrors aiotieba.core.net.NetCore.
//
// The HTTP transport is shared between the resty clients built by HttpCore so
// that every session reuses the same connection pool.
type NetCore struct {
	transport *http.Transport
	proxy     *config.ProxyConfig
	timeout   config.TimeoutConfig

	// client is the transitional http.Client used by Do/SendRequest until the
	// api packages are migrated to the resty clients.
	client *http.Client
}

// NewNetCore builds the shared HTTP transport. A nil proxy disables proxying.
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
		// Tieba serves the app API over plain HTTP and the Python client
		// disables verification (aiohttp ssl=False), so certificates are not
		// verified here either.
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

// Proxy returns the proxy configuration.
func (n *NetCore) Proxy() *config.ProxyConfig { return n.proxy }

// Timeout returns the timeout configuration.
func (n *NetCore) Timeout() config.TimeoutConfig { return n.timeout }

// Transport returns the shared HTTP transport used by the resty clients.
func (n *NetCore) Transport() *http.Transport { return n.transport }

// NewRestyClient builds a resty client that shares the NetCore connection pool.
// The TLS, proxy and keep-alive settings are carried by the shared transport.
func (n *NetCore) NewRestyClient() *resty.Client {
	c := resty.New()
	c.SetTransport(n.transport)
	c.SetTimeout(n.timeout.HTTPRead)
	return c
}

// Do sends req and returns the raw response without checking the status code.
//
// Deprecated: it is kept only during the migration to resty. API packages should
// use the resty clients provided by HttpCore instead.
func (n *NetCore) Do(req *http.Request) (*http.Response, error) {
	resp, err := n.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending %s %s: %w", req.Method, req.URL.Redacted(), err)
	}
	return resp, nil
}

// SendRequest sends req, expects a 200 status and returns the whole body.
//
// It mirrors NetCore.send_request. Go adds "Accept-Encoding: gzip" on its own
// and transparently decompresses the response, so the body is always decoded.
//
// Deprecated: see Do.
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

// Close releases the idle connections of the pool.
func (n *NetCore) Close() {
	n.client.CloseIdleConnections()
}

// proxyFunc converts a ProxyConfig into a proxy function. It always returns a
// non-nil function so that callers never fall back to the environment.
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
