package core

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	xproxy "golang.org/x/net/proxy"

	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// wsURL 是贴吧 IM 服务的 websocket 地址。Python 客户端使用明文连接并附带非标准的握手头。
const wsURL = "ws://im.tieba.baidu.com:8000"

// wsPath 是升级请求的 request-target。Python 的 yarl.URL.build 未指定 path，HTTP 请求默认用 "/"。
const wsPath = "/"

// wsHandshakeGUID 是 RFC 6455 规定的握手校验常量。
const wsHandshakeGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

// WebsocketCallback 处理 cmd 已注册处理器的帧。
type WebsocketCallback func(w *WsCore, data []byte, reqID int)

// PackWsBytes 打包数据并添加9字节头部。
//
// 参数:
//
//	account 贴吧的用户参数容器
//	data 待发送的websocket数据
//	cmd 请求的cmd类型
//	reqID 请求的id
//	compress 是否需要gzip压缩
//	encrypt 是否需要aes加密
//
// 为 data 添加 9 字节贴吧 websocket 头部，对应 pack_ws_bytes。
// 布局为：1 字节 flag、4 字节大端 cmd、4 字节大端 req_id，随后是负载。
// flag 第 6 位表示 gzip，第 7 位表示 AES-ECB + PKCS7。
func PackWsBytes(account *Account, data []byte, cmd, reqID int, compress, encrypt bool) ([]byte, error) {
	flag := byte(0x08)

	if compress {
		flag |= 0b0100_0000
		compressed, err := gzipCompress(data)
		if err != nil {
			return nil, fmt.Errorf("compressing websocket payload: %w", err)
		}
		data = compressed
	}
	if encrypt {
		flag |= 0b1000_0000
		key, err := account.AESECBKey()
		if err != nil {
			return nil, err
		}
		encrypted, err := crypto.ECBEncrypt(key, data)
		if err != nil {
			return nil, fmt.Errorf("encrypting websocket payload: %w", err)
		}
		data = encrypted
	}

	out := make([]byte, 0, 9+len(data))
	out = append(out, flag)
	out = binary.BigEndian.AppendUint32(out, uint32(cmd))
	out = binary.BigEndian.AppendUint32(out, uint32(reqID))
	out = append(out, data...)
	return out, nil
}

// ParseWsBytes 对 websocket 返回数据进行解包。
//
// 参数:
//
//	account 贴吧的用户参数容器
//	data 接收到的websocket数据
//
// 解包贴吧 websocket 帧，对应 parse_ws_bytes。
func ParseWsBytes(account *Account, data []byte) ([]byte, int, int, error) {
	if len(data) < 9 {
		return nil, 0, 0, fmt.Errorf("websocket frame too short: %d bytes", len(data))
	}
	flag := data[0]
	cmd := int(binary.BigEndian.Uint32(data[1:5]))
	reqID := int(binary.BigEndian.Uint32(data[5:9]))
	body := data[9:]

	if flag&0b1000_0000 != 0 {
		key, err := account.AESECBKey()
		if err != nil {
			return nil, 0, 0, err
		}
		body, err = crypto.ECBDecrypt(key, body)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("decrypting websocket payload: %w", err)
		}
	}
	if flag&0b0100_0000 != 0 {
		decompressed, err := gzipDecompress(body)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("decompressing websocket payload: %w", err)
		}
		body = decompressed
	}

	return body, cmd, reqID, nil
}

// WsResponse websocket 响应，用于等待一次 websocket 请求的返回数据，对应 aiotieba.core.websocket.WsResponse。
type WsResponse struct {
	ch          chan []byte
	reqID       int
	readTimeout time.Duration
	cancel      func()
}

// ReqID 返回该响应的请求 id。
func (r *WsResponse) ReqID() int { return r.reqID }

// Read 读取 websocket 响应，读取超时返回 os.ErrDeadlineExceeded。
//
// 等待响应负载，读取超时返回 os.ErrDeadlineExceeded。
func (r *WsResponse) Read() ([]byte, error) {
	timer := time.NewTimer(r.readTimeout)
	defer timer.Stop()

	select {
	case data, ok := <-r.ch:
		if !ok {
			return nil, errors.New("core: websocket response was cancelled")
		}
		return data, nil
	case <-timer.C:
		r.cancel()
		return nil, fmt.Errorf("timeout to read: %w", os.ErrDeadlineExceeded)
	}
}

type wsWaiter struct {
	mu          sync.Mutex
	reqID       int
	readTimeout time.Duration
	pending     map[int]chan []byte
}

func newWsWaiter(readTimeout time.Duration) *wsWaiter {
	return &wsWaiter{
		reqID:       int(time.Now().Unix()),
		readTimeout: readTimeout,
		pending:     map[int]chan []byte{},
	}
}

func (w *wsWaiter) new() (int, chan []byte, func()) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.reqID++
	id := w.reqID
	ch := make(chan []byte, 1)
	w.pending[id] = ch
	return id, ch, func() { w.cancel(id) }
}

func (w *wsWaiter) cancel(reqID int) {
	w.mu.Lock()
	ch, ok := w.pending[reqID]
	delete(w.pending, reqID)
	w.mu.Unlock()
	if ok {
		close(ch)
	}
}

func (w *wsWaiter) setDone(reqID int, data []byte) {
	w.mu.Lock()
	ch, ok := w.pending[reqID]
	delete(w.pending, reqID)
	w.mu.Unlock()
	if ok {
		ch <- data
		close(ch)
	}
}

func (w *wsWaiter) cancelAll() {
	w.mu.Lock()
	chans := make([]chan []byte, 0, len(w.pending))
	for id, ch := range w.pending {
		chans = append(chans, ch)
		delete(w.pending, id)
	}
	w.mu.Unlock()
	for _, ch := range chans {
		close(ch)
	}
}

// sendConfig 保存 Send 解析后的选项。
type sendConfig struct {
	compress bool
	encrypt  bool
}

// SendOption 用于定制 Send。
type SendOption func(*sendConfig)

// WithCompress 发送前对负载做 gzip 压缩。
func WithCompress() SendOption {
	return func(c *sendConfig) { c.compress = true }
}

// WithoutEncrypt 关闭负载的 AES-ECB 加密。
func WithoutEncrypt() SendOption {
	return func(c *sendConfig) { c.encrypt = false }
}

// WsCore 保存 websocket 接口相关状态的核心容器，对应 aiotieba.core.websocket.WsCore。
type WsCore struct {
	Account *Account
	NetCore *NetCore

	mu        sync.Mutex
	writeMu   sync.Mutex
	callbacks map[int]WebsocketCallback
	conn      *wsConn
	waiter    *wsWaiter
	midMgr    *MsgIDManager
	status    enums.WsStatus
	done      chan struct{}
}

// NewWsCore 创建一个处于关闭状态的 websocket 会话。
func NewWsCore(account *Account, netCore *NetCore) *WsCore {
	return &WsCore{
		Account:   account,
		NetCore:   netCore,
		callbacks: map[int]WebsocketCallback{},
		status:    enums.WsStatusClosed,
	}
}

// SetAccount 替换会话的账号。
func (w *WsCore) SetAccount(a *Account) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.Account = a
}

func (w *WsCore) account() *Account {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.Account
}

// Status 当前的 websocket 状态。
func (w *WsCore) Status() enums.WsStatus {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.status
}

func (w *WsCore) setStatus(s enums.WsStatus) {
	w.mu.Lock()
	w.status = s
	w.mu.Unlock()
}

// RegisterCallback 为携带 cmd 的帧注册处理器，对应 Python 客户端的 callbacks 字典。
func (w *WsCore) RegisterCallback(cmd int, cb WebsocketCallback) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.callbacks[cmd] = cb
}

// MsgIDManager 返回 msg id 管理器，未连接时为 nil。
func (w *WsCore) MsgIDManager() *MsgIDManager {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.midMgr
}

// Connect 建立 websocket 连接，握手失败返回 exception.HTTPStatusError。
//
// 拨号连接 websocket 端点，对应 WsCore.connect。
//
// 握手自己实现而不用 gorilla：贴吧 IM 要求携带非标准的 Sec-WebSocket-Extensions 头
// （im_version=2.3），而 gorilla 把它列入保留头、拒绝调用方设置（client.go 的
// "duplicate header not allowed"），并且它的 *Conn 只能由它自己的握手中产生，
// 没有办法"自己握手 + 复用它的帧层"。Python 客户端同样是手写握手
// （aiohttp.ClientRequest 构造升级请求后再换上 WebSocket parser），这里与其对齐。
func (w *WsCore) Connect(ctx context.Context) error {
	w.mu.Lock()
	w.status = enums.WsStatusConnecting
	w.waiter = newWsWaiter(w.NetCore.Timeout().WSRead)
	w.midMgr = NewMsgIDManager()
	w.mu.Unlock()

	conn, err := w.dialWebsocket(ctx)
	if err != nil {
		w.setStatus(enums.WsStatusClosed)
		return err
	}

	done := make(chan struct{})
	w.mu.Lock()
	w.conn = conn
	w.done = done
	w.status = enums.WsStatusOpen
	w.mu.Unlock()

	go w.readLoop(conn, done)
	return nil
}

// dialWebsocket 拨号并完成 websocket 握手。
//
// 参数:
//
//	ctx 取消拨号与握手
//
// 状态码不是 101 时返回 exception.HTTPStatusError。
func (w *WsCore) dialWebsocket(ctx context.Context) (*wsConn, error) {
	u, err := url.Parse(wsURL)
	if err != nil {
		return nil, fmt.Errorf("parsing websocket url: %w", err)
	}

	timeout := w.NetCore.Timeout()
	conn, target, proxyAuth, err := w.dialUpstream(ctx, u)
	if err != nil {
		return nil, err
	}

	// 握手失败必须把连接关掉。
	established := false
	defer func() {
		if !established {
			_ = conn.Close()
		}
	}()

	if timeout.HTTPConnect > 0 {
		_ = conn.SetDeadline(time.Now().Add(timeout.HTTPConnect))
		defer func() { _ = conn.SetDeadline(time.Time{}) }()
	}

	c, err := wsHandshake(conn, u.Host, target, proxyAuth, timeout.WSSend)
	if err != nil {
		return nil, err
	}
	established = true
	return c, nil
}

// wsHandshake 在已建立的连接上完成 websocket 握手，对应 Python 的
// aiohttp.ClientRequest + req2res 那段。
//
// 参数:
//
//	conn 已建立的 TCP 连接
//	host Host 头的值，同时也是被请求的 IM 服务地址
//	target 请求行的 request-target，直连与 socks 用 origin-form，HTTP 代理用 absolute-form
//	proxyAuth 需要附加的代理认证头行，没有则为空串
//	writeTimeout 控制帧的写超时
func wsHandshake(conn net.Conn, host, target, proxyAuth string, writeTimeout time.Duration) (*wsConn, error) {
	// 握手与后续的帧读取必须共用同一个 bufio.Reader：http.ReadResponse 可能已经把
	// 紧跟响应的帧字节读进了缓冲区，另起一个 reader 会丢数据。
	br := bufio.NewReaderSize(conn, 4096)

	keyBytes := make([]byte, 16)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, fmt.Errorf("generating websocket key: %w", err)
	}
	key := base64.StdEncoding.EncodeToString(keyBytes)

	var req strings.Builder
	fmt.Fprintf(&req, "GET %s HTTP/1.1\r\n", target)
	fmt.Fprintf(&req, "Host: %s\r\n", host)
	req.WriteString("Upgrade: websocket\r\n")
	req.WriteString("Connection: upgrade\r\n")
	// 贴吧 IM 服务要求的非标准握手头。
	req.WriteString("Sec-WebSocket-Extensions: im_version=2.3\r\n")
	req.WriteString("Sec-WebSocket-Version: 13\r\n")
	fmt.Fprintf(&req, "Sec-WebSocket-Key: %s\r\n", key)
	req.WriteString("Accept-Encoding: gzip\r\n")
	req.WriteString(proxyAuth)
	req.WriteString("\r\n")

	if _, err := io.WriteString(conn, req.String()); err != nil {
		return nil, fmt.Errorf("writing websocket handshake: %w", err)
	}

	resp, err := http.ReadResponse(br, &http.Request{Method: http.MethodGet})
	if err != nil {
		return nil, fmt.Errorf("reading websocket handshake: %w", err)
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		return nil, &exception.HTTPStatusError{Code: resp.StatusCode, Msg: resp.Status}
	}
	if got, want := resp.Header.Get("Sec-WebSocket-Accept"), wsAcceptKey(key); got != want {
		return nil, fmt.Errorf("websocket handshake: Sec-WebSocket-Accept = %q, want %q", got, want)
	}

	return newWsConn(conn, br, writeTimeout), nil
}

// wsAcceptKey 按 RFC 6455 4.2.2 计算 Sec-WebSocket-Accept。
func wsAcceptKey(key string) string {
	h := sha1.New()
	io.WriteString(h, key)
	io.WriteString(h, wsHandshakeGUID)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// dialUpstream 建立到底层服务的 TCP 连接，按代理配置决定是否经过代理。
//
// 返回的 target 是握手请求行的 request-target：直连与 socks 代理用 origin-form（"/"），
// HTTP 代理用 absolute-form（"http://host:port/"）；proxyAuth 是需要附加的认证头行。
func (w *WsCore) dialUpstream(ctx context.Context, target *url.URL) (conn net.Conn, reqTarget string, proxyAuth string, err error) {
	dial := w.NetCore.transport.DialContext
	proxyCfg := w.NetCore.Proxy()

	if proxyCfg == nil || proxyCfg.URL == nil {
		c, derr := dial(ctx, "tcp", target.Host)
		return c, wsPath, "", derr
	}

	p := proxyCfg.URL
	switch p.Scheme {
	case "http", "https":
		c, derr := dial(ctx, "tcp", p.Host)
		if derr != nil {
			return nil, "", "", fmt.Errorf("dialing proxy %s: %w", p.Host, derr)
		}
		if proxyCfg.Auth != nil {
			cred := base64.StdEncoding.EncodeToString([]byte(proxyCfg.Auth.Login + ":" + proxyCfg.Auth.Password))
			proxyAuth = "Proxy-Authorization: Basic " + cred + "\r\n"
		}
		// 明文 websocket 经 HTTP 代理时用绝对形式的 request-target。
		return c, "http://" + target.Host + wsPath, proxyAuth, nil

	case "socks5", "socks5h":
		var auth *xproxy.Auth
		if proxyCfg.Auth != nil {
			auth = &xproxy.Auth{User: proxyCfg.Auth.Login, Password: proxyCfg.Auth.Password}
		}
		d, derr := xproxy.SOCKS5("tcp", p.Host, auth, contextDialer{dial})
		if derr != nil {
			return nil, "", "", fmt.Errorf("creating socks5 dialer: %w", derr)
		}
		c, derr := d.Dial("tcp", target.Host)
		if derr != nil {
			return nil, "", "", fmt.Errorf("dialing through socks5 proxy %s: %w", p.Host, derr)
		}
		return c, wsPath, "", nil

	default:
		return nil, "", "", fmt.Errorf("unsupported proxy scheme %q for websocket", p.Scheme)
	}
}

// contextDialer 把 http.Transport 的 DialContext 适配成 x/net/proxy 需要的拨号器。
type contextDialer struct {
	dial func(ctx context.Context, network, addr string) (net.Conn, error)
}

// Dial 实现 x/net/proxy.Dialer。
func (d contextDialer) Dial(network, addr string) (net.Conn, error) {
	return d.dial(context.Background(), network, addr)
}

// DialContext 让 x/net/proxy 走带上下文的拨号路径。
func (d contextDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return d.dial(ctx, network, addr)
}

// Close 终止 websocket 会话。
func (w *WsCore) Close() error {
	w.mu.Lock()
	conn := w.conn
	waiter := w.waiter
	status := w.status
	w.mu.Unlock()

	if conn == nil {
		w.setStatus(enums.WsStatusClosed)
		return nil
	}

	if status == enums.WsStatusOpen {
		w.writeMu.Lock()
		_ = conn.SetWriteDeadline(time.Now().Add(w.NetCore.Timeout().WSClose))
		_ = conn.WriteClose()
		_ = conn.SetWriteDeadline(time.Time{})
		w.writeMu.Unlock()
	}
	w.setStatus(enums.WsStatusClosed)
	if waiter != nil {
		waiter.cancelAll()
	}

	err := conn.Close()
	w.mu.Lock()
	w.conn = nil
	w.mu.Unlock()
	return err
}

func (w *WsCore) readLoop(conn *wsConn, done chan struct{}) {
	defer close(done)

	account := w.account()
	for {
		msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		data, cmd, reqID, err := ParseWsBytes(account, msg)
		if err != nil {
			continue
		}

		w.mu.Lock()
		cb := w.callbacks[cmd]
		waiter := w.waiter
		w.mu.Unlock()

		if cb == nil {
			if waiter != nil {
				waiter.setDone(reqID, data)
			}
			continue
		}
		go cb(w, data, reqID)
	}

	w.setStatus(enums.WsStatusClosed)
	w.mu.Lock()
	waiter := w.waiter
	w.mu.Unlock()
	if waiter != nil {
		waiter.cancelAll()
	}
}

// Send 将 protobuf 序列化结果打包发送，并返回响应对象；发送超时返回 os.ErrDeadlineExceeded。
//
// 参数:
//
//	data 待发送的数据
//	cmd 请求的cmd类型
//	opts 可选配置，见 SendOption
//
// 打包并发送 data，返回响应对象，对应 WsCore.send。
func (w *WsCore) Send(data []byte, cmd int, opts ...SendOption) (*WsResponse, error) {
	cfg := sendConfig{encrypt: true}
	for _, opt := range opts {
		opt(&cfg)
	}

	w.mu.Lock()
	conn := w.conn
	waiter := w.waiter
	w.mu.Unlock()

	if conn == nil || waiter == nil {
		return nil, errors.New("core: websocket is not connected")
	}

	reqID, ch, cancel := waiter.new()
	payload, err := PackWsBytes(w.account(), data, cmd, reqID, cfg.compress, cfg.encrypt)
	if err != nil {
		cancel()
		return nil, err
	}

	if err := w.write(conn, payload); err != nil {
		cancel()
		return nil, err
	}

	return &WsResponse{
		ch:          ch,
		reqID:       reqID,
		readTimeout: w.NetCore.Timeout().WSRead,
		cancel:      cancel,
	}, nil
}

func (w *WsCore) write(conn *wsConn, payload []byte) error {
	w.writeMu.Lock()
	defer w.writeMu.Unlock()

	if d := w.NetCore.Timeout().WSSend; d > 0 {
		if err := conn.SetWriteDeadline(time.Now().Add(d)); err != nil {
			return err
		}
		defer func() { _ = conn.SetWriteDeadline(time.Time{}) }()
	}

	if err := conn.WriteMessage(payload); err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return fmt.Errorf("timeout to send: %w", os.ErrDeadlineExceeded)
		}
		return fmt.Errorf("writing websocket frame: %w", err)
	}
	return nil
}

func gzipCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw, err := gzip.NewWriterLevel(&buf, 6)
	if err != nil {
		return nil, err
	}
	if _, err := zw.Write(data); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func gzipDecompress(data []byte) ([]byte, error) {
	zr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	return io.ReadAll(zr)
}
