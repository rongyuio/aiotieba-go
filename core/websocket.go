package core

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// wsURL 是贴吧 IM 服务的 websocket 地址。Python 客户端使用明文连接并附带非标准的握手头。
const wsURL = "ws://im.tieba.baidu.com:8000"

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
	conn      *websocket.Conn
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

// Connect 建立 weboscket 连接，握手失败返回 exception.HTTPStatusError。
//
// 拨号连接 websocket 端点，对应 WsCore.connect。
func (w *WsCore) Connect(ctx context.Context) error {
	w.mu.Lock()
	w.status = enums.WsStatusConnecting
	w.waiter = newWsWaiter(w.NetCore.Timeout().WSRead)
	w.midMgr = NewMsgIDManager()
	w.mu.Unlock()

	header := http.Header{}
	// 贴吧 IM 服务要求的非标准握手头。
	header.Set("Sec-WebSocket-Extensions", "im_version=2.3")

	dialer := &websocket.Dialer{
		HandshakeTimeout: w.NetCore.Timeout().HTTPConnect,
		Proxy:            proxyFunc(w.NetCore.Proxy()),
		ReadBufferSize:   4096,
		WriteBufferSize:  4096,
	}

	conn, resp, err := dialer.DialContext(ctx, wsURL, header)
	if err != nil {
		w.setStatus(enums.WsStatusClosed)
		if resp != nil {
			return &exception.HTTPStatusError{Code: resp.StatusCode, Msg: resp.Status}
		}
		return fmt.Errorf("dialing websocket: %w", err)
	}
	conn.SetReadLimit(4 * 1024 * 1024)

	done := make(chan struct{})
	w.mu.Lock()
	w.conn = conn
	w.done = done
	w.status = enums.WsStatusOpen
	w.mu.Unlock()

	go w.readLoop(conn, done)
	return nil
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
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
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

func (w *WsCore) readLoop(conn *websocket.Conn, done chan struct{}) {
	defer close(done)

	account := w.account()
	for {
		_, msg, err := conn.ReadMessage()
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

func (w *WsCore) write(conn *websocket.Conn, payload []byte) error {
	w.writeMu.Lock()
	defer w.writeMu.Unlock()

	if d := w.NetCore.Timeout().WSSend; d > 0 {
		if err := conn.SetWriteDeadline(time.Now().Add(d)); err != nil {
			return err
		}
		defer func() { _ = conn.SetWriteDeadline(time.Time{}) }()
	}

	if err := conn.WriteMessage(websocket.BinaryMessage, payload); err != nil {
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
