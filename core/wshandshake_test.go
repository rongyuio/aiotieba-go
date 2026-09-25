package core

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/rongyuio/aiotieba-go/exception"
)

// TestWsAcceptKeyRFC6455Vector 用 RFC 6455 第 5.7 节的官方示例校验 accept 计算。
func TestWsAcceptKeyRFC6455Vector(t *testing.T) {
	const (
		key  = "dGhlIHNhbXBsZSBub25jZQ=="
		want = "s3pPLMBiTxaQ9kYGzzhZRbK+xOo="
	)
	if got := wsAcceptKey(key); got != want {
		t.Errorf("wsAcceptKey(%q) = %q, want %q", key, got, want)
	}
}

// newFakeWsServer 在 net.Pipe 的另一端扮演 IM 服务端：读到握手请求后交给 respond 写回响应。
func newFakeWsServer(t *testing.T, respond func(req string, w io.Writer)) (client net.Conn, stop func()) {
	t.Helper()
	client, server := net.Pipe()
	served := make(chan struct{})
	go func() {
		defer close(served)
		defer server.Close()
		req, err := readHandshakeRequest(bufio.NewReader(server))
		if err != nil {
			return
		}
		respond(req, server)
	}()
	return client, func() { _ = server.Close(); <-served }
}

func readHandshakeRequest(br *bufio.Reader) (string, error) {
	var sb strings.Builder
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			return sb.String(), err
		}
		sb.WriteString(line)
		if line == "\r\n" {
			return sb.String(), nil
		}
	}
}

func requestHeader(req, name string) string {
	for _, line := range strings.Split(req, "\r\n") {
		k, v, ok := strings.Cut(line, ":")
		if ok && strings.EqualFold(strings.TrimSpace(k), name) {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func switchingProtocols(req string) string {
	return fmt.Sprintf(
		"HTTP/1.1 101 Switching Protocols\r\n"+
			"Upgrade: websocket\r\n"+
			"Connection: Upgrade\r\n"+
			"Sec-WebSocket-Accept: %s\r\n\r\n",
		wsAcceptKey(requestHeader(req, "Sec-WebSocket-Key")))
}

func TestWsHandshakeSendsExpectedRequest(t *testing.T) {
	const host = "im.tieba.baidu.com:8000"

	client, stop := newFakeWsServer(t, func(req string, w io.Writer) {
		if !strings.HasPrefix(req, "GET / HTTP/1.1\r\n") {
			t.Errorf("request line = %q", firstLine(req))
		}
		for _, tc := range []struct{ name, want string }{
			{"Host", host},
			{"Upgrade", "websocket"},
			{"Connection", "upgrade"},
			// 这一条是整个修复的核心：gorilla 不允许调用方设置它。
			{"Sec-WebSocket-Extensions", "im_version=2.3"},
			{"Sec-WebSocket-Version", "13"},
			{"Accept-Encoding", "gzip"},
		} {
			if got := requestHeader(req, tc.name); got != tc.want {
				t.Errorf("%s = %q, want %q", tc.name, got, tc.want)
			}
		}
		// Sec-WebSocket-Key 必须是 16 字节的 base64。
		if got := requestHeader(req, "Sec-WebSocket-Key"); len(got) != 24 {
			t.Errorf("Sec-WebSocket-Key = %q, want a 24-character base64 of 16 bytes", got)
		}
		_, _ = io.WriteString(w, switchingProtocols(req))
	})
	defer stop()

	conn, err := wsHandshake(client, host, wsPath, "", time.Second)
	if err != nil {
		t.Fatalf("wsHandshake: %v", err)
	}
	defer conn.Close()
}

func firstLine(s string) string {
	if i := strings.Index(s, "\r\n"); i >= 0 {
		return s[:i]
	}
	return s
}

func TestWsHandshakePropagatesProxyAuth(t *testing.T) {
	const auth = "Proxy-Authorization: Basic dXNlcjpwYXNz\r\n"

	client, stop := newFakeWsServer(t, func(req string, w io.Writer) {
		if got := requestHeader(req, "Proxy-Authorization"); got != "Basic dXNlcjpwYXNz" {
			t.Errorf("Proxy-Authorization = %q", got)
		}
		_, _ = io.WriteString(w, switchingProtocols(req))
	})
	defer stop()

	conn, err := wsHandshake(client, "im.tieba.baidu.com:8000", "http://im.tieba.baidu.com:8000/", auth, time.Second)
	if err != nil {
		t.Fatalf("wsHandshake: %v", err)
	}
	defer conn.Close()
}

func TestWsHandshakeRejectsNon101(t *testing.T) {
	client, stop := newFakeWsServer(t, func(req string, w io.Writer) {
		_, _ = io.WriteString(w, "HTTP/1.1 400 Bad Request\r\nContent-Length: 0\r\n\r\n")
	})
	defer stop()

	_, err := wsHandshake(client, "im.tieba.baidu.com:8000", wsPath, "", time.Second)

	var statusErr *exception.HTTPStatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("err = %v (%T), want *exception.HTTPStatusError", err, err)
	}
	if statusErr.Code != 400 {
		t.Errorf("code = %d, want 400", statusErr.Code)
	}
}

func TestWsHandshakeRejectsWrongAccept(t *testing.T) {
	client, stop := newFakeWsServer(t, func(req string, w io.Writer) {
		_, _ = io.WriteString(w, "HTTP/1.1 101 Switching Protocols\r\n"+
			"Upgrade: websocket\r\nConnection: Upgrade\r\n"+
			"Sec-WebSocket-Accept: bm90LXRoZS1yaWdodC12YWx1ZQ==\r\n\r\n")
	})
	defer stop()

	_, err := wsHandshake(client, "im.tieba.baidu.com:8000", wsPath, "", time.Second)
	if err == nil || !strings.Contains(err.Error(), "Sec-WebSocket-Accept") {
		t.Fatalf("err = %v, want an accept mismatch", err)
	}
}

// TestWsHandshakeKeepsFramesAlreadyBuffered 是这次修复最容易踩的坑：
// 服务端把 101 响应和第一帧写在同一次 write 里，如果握手阶段另起一个 bufio.Reader，
// 帧就会被丢掉。这里断言握手完成后帧依然读得到。
func TestWsHandshakeKeepsFramesAlreadyBuffered(t *testing.T) {
	payload := []byte("already buffered")

	client, stop := newFakeWsServer(t, func(req string, w io.Writer) {
		// 响应与帧一次性写出。
		_, _ = w.Write(append([]byte(switchingProtocols(req)), serverFrame(true, wsOpBinary, payload)...))
	})
	defer stop()

	conn, err := wsHandshake(client, "im.tieba.baidu.com:8000", wsPath, "", time.Second)
	if err != nil {
		t.Fatalf("wsHandshake: %v", err)
	}
	defer conn.Close()

	msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	if string(msg) != string(payload) {
		t.Errorf("msg = %q, want %q", msg, payload)
	}
}
