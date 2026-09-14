package core

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	mrand "math/rand/v2"
	"net"
	"os"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba/consts"
	"github.com/rongyuio/aiotieba/helper/crypto"
	"github.com/rongyuio/aiotieba/protobuf"
)

// Constants of the Tieba chat service, mirroring aiotieba.core.blcp.
const (
	// ChatAppID is the app id of the Tieba chat service.
	ChatAppID = 10773430
	// ChatSDKVersion is the sdk version of the Tieba chat service.
	ChatSDKVersion = 11250036

	blcpHost      = "common.lcs.baidu.com"
	blcpPort      = "443"
	blcpMagic     = "lcp\x01"
	blcpLoginFrom = "1008550l"
	blcpUA        = "900_1600_android_12.68.1.0_240"
	blcpSDKVer    = "3460016"
)

// BLCPData is one frame of the BLCP protocol. It mirrors
// aiotieba.core.blcp.BLCPData.
type BLCPData struct {
	ServiceID     int64
	MethodID      int64
	RPCBody       []byte
	LCMBody       []byte
	Timestamp     int64
	IfRequest     bool
	CorrelationID int64
	IsNotify      bool

	body *BLCPBody // decoded Lcm part, filled by the reader
}

// NewBLCPData creates a frame with a random correlation id.
func NewBLCPData(serviceID, methodID int64) *BLCPData {
	return &BLCPData{
		ServiceID:     serviceID,
		MethodID:      methodID,
		Timestamp:     time.Now().UnixMilli(),
		IfRequest:     true,
		CorrelationID: mrand.Int64N(maxInt64),
	}
}

const maxInt64 = int64(^uint64(0) >> 1)

// Bytes encodes the frame. It mirrors BLCPData.toBytes.
func (d *BLCPData) Bytes() []byte {
	out := make([]byte, 0, 8+len(d.RPCBody)+len(d.LCMBody))
	out = append(out, blcpMagic...)
	out = binary.BigEndian.AppendUint32(out, uint32(len(d.RPCBody)+len(d.LCMBody)))
	out = binary.BigEndian.AppendUint32(out, uint32(len(d.RPCBody)))
	out = append(out, d.RPCBody...)
	out = append(out, d.LCMBody...)
	return out
}

// BLCPBody is the Lcm part of a BLCP frame: either a protobuf RpcData or a
// decoded JSON object. It mirrors the dual return type of
// ClientBLCPResponses.parseBLCPResponse.
type BLCPBody struct {
	RPC  *protobuf.RpcData
	JSON map[string]any
}

// ErrorMsg returns the error message of a protobuf Lcm response, or the
// "error_msg" field of a JSON body.
func (b *BLCPBody) ErrorMsg() string {
	if b.RPC != nil && b.RPC.GetLcmResponse() != nil {
		return b.RPC.GetLcmResponse().GetErrorMsg()
	}
	if v, ok := b.JSON["error_msg"]; ok {
		return fmt.Sprint(v)
	}
	return ""
}

// JSONValue returns a JSON field converted to string.
func (b *BLCPBody) JSONValue(key string) (string, bool) {
	if b.JSON == nil {
		return "", false
	}
	v, ok := b.JSON[key]
	if !ok {
		return "", false
	}
	return fmt.Sprint(v), true
}

// ParseBLCPResponse decodes a raw BLCP frame into its RpcMeta and Lcm body.
//
// It mirrors the try-chain of ClientBLCPResponses.parseBLCPResponse:
// protobuf, gzip+protobuf, JSON, gzip+JSON.
func ParseBLCPResponse(data []byte) (*protobuf.RpcMeta, *BLCPBody, error) {
	if len(data) < 4 || string(data[:4]) != blcpMagic {
		return nil, nil, errors.New("core: not a BLCP frame")
	}
	rest := data[4:]
	if len(rest) < 8 {
		return nil, nil, errors.New("core: truncated BLCP frame")
	}
	b1 := int(binary.BigEndian.Uint32(rest[0:4]))
	b2 := int(binary.BigEndian.Uint32(rest[4:8]))
	rest = rest[8:]
	if b2 > len(rest) || b1 > len(rest) || b2 > b1 {
		return nil, nil, errors.New("core: inconsistent BLCP frame lengths")
	}

	rpc := &protobuf.RpcMeta{}
	if err := proto.Unmarshal(rest[:b2], rpc); err != nil {
		return nil, nil, fmt.Errorf("core: decoding the BLCP RpcMeta: %w", err)
	}

	raw := rest[b2:b1]
	if body, ok := decodeLCM(raw); ok {
		return rpc, body, nil
	}
	if unzipped, err := gzipDecompress(raw); err == nil {
		if body, ok := decodeLCM(unzipped); ok {
			return rpc, body, nil
		}
	}
	if v, ok := decodeJSON(raw); ok {
		return rpc, &BLCPBody{JSON: v}, nil
	}
	if unzipped, err := gzipDecompress(raw); err == nil {
		if v, ok := decodeJSON(unzipped); ok {
			return rpc, &BLCPBody{JSON: v}, nil
		}
	}
	return nil, nil, errors.New("core: cannot decode the BLCP Lcm body")
}

func decodeLCM(raw []byte) (*BLCPBody, bool) {
	body := &protobuf.RpcData{}
	if err := proto.Unmarshal(raw, body); err != nil {
		return nil, false
	}
	// Accept only messages that actually carry one of the RpcData fields so
	// that JSON bodies are not mistaken for protobuf.
	if body.GetLcmRequest() == nil && body.GetLcmResponse() == nil && body.GetLcmNotify() == nil {
		return nil, false
	}
	return &BLCPBody{RPC: body}, true
}

func decodeJSON(raw []byte) (map[string]any, bool) {
	var v map[string]any
	if err := json.Unmarshal(raw, &v); err != nil || v == nil {
		return nil, false
	}
	return v, true
}

// ClientBLCPResponses reads BLCP frames from a connection. It mirrors
// aiotieba.core.blcp.ClientBLCPResponses.
type ClientBLCPResponses struct {
	reader *bufio.Reader
	conn   net.Conn
}

// NewClientBLCPResponses wraps conn.
func NewClientBLCPResponses(conn net.Conn) *ClientBLCPResponses {
	return &ClientBLCPResponses{reader: bufio.NewReader(conn), conn: conn}
}

// Next reads and parses one frame. It returns an error (io.EOF) when the
// connection is closed.
func (c *ClientBLCPResponses) Next() (*BLCPData, *BLCPBody, error) {
	raw, err := c.readFrame()
	if err != nil {
		return nil, nil, err
	}
	rpc, body, err := ParseBLCPResponse(raw)
	if err != nil {
		return nil, nil, err
	}

	frame := NewBLCPData(rpc.GetResponse().GetServiceId(), rpc.GetResponse().GetMethodId())
	frame.CorrelationID = rpc.GetCorrelationId()
	frame.RPCBody = marshal(rpc)
	frame.IfRequest = false
	frame.body = body

	// Heartbeats are filtered out, mirroring the Python condition.
	if rpc.GetNotify() != nil && !(frame.MethodID == 3 && frame.ServiceID == 1) {
		frame.IsNotify = true
	}
	if body.RPC != nil {
		frame.LCMBody = marshal(body.RPC)
	} else {
		frame.LCMBody = marshalJSON(body.JSON)
	}
	return frame, body, nil
}

func (c *ClientBLCPResponses) readFrame() ([]byte, error) {
	if err := c.readUntilMagic(); err != nil {
		return nil, err
	}

	lengthBytes := make([]byte, 8)
	if _, err := io.ReadFull(c.reader, lengthBytes); err != nil {
		return nil, err
	}
	allLength := int(binary.BigEndian.Uint32(lengthBytes[0:4]))
	body := make([]byte, allLength)
	if _, err := io.ReadFull(c.reader, body); err != nil {
		return nil, err
	}

	out := make([]byte, 0, 4+8+allLength)
	out = append(out, blcpMagic...)
	out = append(out, lengthBytes...)
	out = append(out, body...)
	return out, nil
}

func (c *ClientBLCPResponses) readUntilMagic() error {
	var window [4]byte
	for {
		b, err := c.reader.ReadByte()
		if err != nil {
			return err
		}
		window[0], window[1], window[2], window[3] = window[1], window[2], window[3], b
		if window[0] == 'l' && window[1] == 'c' && window[2] == 'p' && window[3] == 1 {
			return nil
		}
	}
}

// BLCPResponse waits for the response of one BLCP request. It mirrors
// aiotieba.core.blcp.BLCPResponse.
type BLCPResponse struct {
	ch          chan *BLCPData
	reqID       int64
	readTimeout time.Duration
	cancel      func()
}

// Read waits for the response frame and returns os.ErrDeadlineExceeded on
// timeout.
func (r *BLCPResponse) Read() (*BLCPData, error) {
	timer := time.NewTimer(r.readTimeout)
	defer timer.Stop()

	select {
	case data, ok := <-r.ch:
		if !ok {
			return nil, errors.New("core: BLCP response was cancelled")
		}
		return data, nil
	case <-timer.C:
		r.cancel()
		return nil, fmt.Errorf("timeout to read: %w", os.ErrDeadlineExceeded)
	}
}

type blcpWaiter struct {
	mu          sync.Mutex
	reqID       int64
	readTimeout time.Duration
	pending     map[int64]chan *BLCPData
}

func newBLCPWaiter(readTimeout time.Duration) *blcpWaiter {
	return &blcpWaiter{
		reqID:       time.Now().Unix(),
		readTimeout: readTimeout,
		pending:     map[int64]chan *BLCPData{},
	}
}

func (w *blcpWaiter) new(reqID int64) (int64, *BLCPResponse) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if reqID == 0 {
		w.reqID++
	} else {
		w.reqID = reqID
	}
	id := w.reqID
	ch := make(chan *BLCPData, 1)
	w.pending[id] = ch
	return id, &BLCPResponse{
		ch:          ch,
		reqID:       id,
		readTimeout: w.readTimeout,
		cancel:      func() { w.cancel(id) },
	}
}

func (w *blcpWaiter) cancel(reqID int64) {
	w.mu.Lock()
	ch, ok := w.pending[reqID]
	delete(w.pending, reqID)
	w.mu.Unlock()
	if ok {
		close(ch)
	}
}

func (w *blcpWaiter) setDone(reqID int64, data *BLCPData) {
	w.mu.Lock()
	ch, ok := w.pending[reqID]
	delete(w.pending, reqID)
	w.mu.Unlock()
	if ok {
		ch <- data
		close(ch)
	}
}

func (w *blcpWaiter) cancelAll() {
	w.mu.Lock()
	chans := make([]chan *BLCPData, 0, len(w.pending))
	for id, ch := range w.pending {
		chans = append(chans, ch)
		delete(w.pending, id)
	}
	w.mu.Unlock()
	for _, ch := range chans {
		close(ch)
	}
}

// BLCPCore is a BLCP session bound to one account. It mirrors
// aiotieba.core.blcp.BLCPCore.
type BLCPCore struct {
	Account *Account
	NetCore *NetCore

	mu        sync.Mutex
	writeMu   sync.Mutex
	conn      net.Conn
	responses *ClientBLCPResponses
	waiter    *blcpWaiter
	status    int

	triggerID int64
	uk        string
	bduk      string
	loginID   string

	messages chan *BLCPBody
	cancelHB context.CancelFunc
}

// NewBLCPCore creates a disconnected BLCP session.
func NewBLCPCore(account *Account, netCore *NetCore, maxQueueLength int) *BLCPCore {
	if maxQueueLength <= 0 {
		maxQueueLength = 100
	}
	return &BLCPCore{
		Account:  account,
		NetCore:  netCore,
		status:   -1,
		messages: make(chan *BLCPBody, maxQueueLength),
	}
}

// SetAccount swaps the account of the session.
func (b *BLCPCore) SetAccount(a *Account) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Account = a
}

// Status returns 1 when logged in, 0 after connecting and -1 when disconnected.
func (b *BLCPCore) Status() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.status
}

// TriggerID returns the trigger id assigned by the server.
func (b *BLCPCore) TriggerID() int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.triggerID
}

// UK returns the uk assigned by the server.
func (b *BLCPCore) UK() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.uk
}

// BDUK returns the bduk assigned by the server.
func (b *BLCPCore) BDUK() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.bduk
}

// LoginID returns the login id assigned by the server.
func (b *BLCPCore) LoginID() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.loginID
}

// Messages exposes the queue of chat notifications.
func (b *BLCPCore) Messages() <-chan *BLCPBody { return b.messages }

// Connect opens the TLS connection, mirroring BLCPCore.connect.
func (b *BLCPCore) Connect(ctx context.Context) error {
	b.mu.Lock()
	b.waiter = newBLCPWaiter(5 * time.Second)
	b.mu.Unlock()

	dialer := &net.Dialer{Timeout: b.NetCore.Timeout().HTTPConnect}
	// The Python client connects over IPv4 with check_hostname disabled while
	// still validating the chain; Go performs the standard verification.
	conn, err := tls.DialWithDialer(dialer, "tcp4", net.JoinHostPort(blcpHost, blcpPort), &tls.Config{
		ServerName: blcpHost,
		MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		return fmt.Errorf("core: dialing the BLCP endpoint: %w", err)
	}

	responses := NewClientBLCPResponses(conn)
	b.mu.Lock()
	b.conn = conn
	b.responses = responses
	b.status = 0
	b.mu.Unlock()

	go b.dispatch(responses, conn)
	_ = ctx
	return nil
}

// Close terminates the BLCP session.
func (b *BLCPCore) Close() error {
	b.mu.Lock()
	conn := b.conn
	waiter := b.waiter
	cancel := b.cancelHB
	b.cancelHB = nil
	b.status = -1
	b.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if waiter != nil {
		waiter.cancelAll()
	}
	if conn != nil {
		return conn.Close()
	}
	return nil
}

// Login performs the three step BLCP handshake, mirroring BLCPCore.login.
func (b *BLCPCore) Login(ctx context.Context) error {
	cuidGalaxy2, err := b.Account.CuidGalaxy2()
	if err != nil {
		return err
	}
	token := b.GenerateLCMToken(ctx, cuidGalaxy2)

	// Step 1: LCM handshake.
	loginReq := NewBLCPData(1, 1)
	loginReq.RPCBody = BuildRPCBody(1, 1, loginReq.CorrelationID, 0, 1)
	loginReq.LCMBody = marshal(&protobuf.RpcData{
		LcmRequest: &protobuf.LcmRequest{
			LogId: loginReq.CorrelationID,
			Token: token,
			Common: &protobuf.Common{
				Cuid:       cuidGalaxy2,
				Device:     "android",
				AppId:      fmt.Sprint(ChatAppID),
				AppVersion: consts.ChatVersion,
				SdkVersion: blcpSDKVer,
				Network:    "wifi",
			},
			Timestamp: time.Now().UnixMilli(),
			StartType: -1,
			ConnType:  1,
		},
	})

	rpc, body, err := b.exchange(loginReq)
	if err != nil {
		return fmt.Errorf("core: BLCP handshake step 1: %w", err)
	}
	if rpc.GetResponse().GetErrorText() != "success" || body.ErrorMsg() != "success" {
		return errors.New("core: BLCP handshake step 1 failed")
	}

	// Step 2: parameter login (JSON).
	enuid, err := crypto.Enuid(cuidGalaxy2)
	if err != nil {
		return err
	}
	c3Aid, err := b.Account.C3Aid()
	if err != nil {
		return err
	}
	step2 := NewBLCPData(4, 1)
	step2.RPCBody = BuildRPCBody(4, 1, step2.CorrelationID, 0, 1)
	step2.LCMBody = marshalJSON(map[string]any{
		"params": map[string]any{
			"appname": "tieba",
			"sid":     b.Account.SampleID(),
			"ua":      blcpUA,
			"uid":     enuid,
			"cfrom":   blcpLoginFrom,
			"from":    blcpLoginFrom,
			"network": "1_-1",
			"p_sv":    "32",
			"mps":     "",
			"mpv":     "1",
			"c3_aid":  c3Aid,
			"type_id": "0",
		},
		"filter": map[string]any{
			"aps":     map[string]any{"cpu_abi": "armeabi-v7a"},
			"command": map[string]any{"step": "0"},
		},
	})

	rpc, body, err = b.exchange(step2)
	if err != nil {
		return fmt.Errorf("core: BLCP handshake step 2: %w", err)
	}
	if v, ok := body.JSONValue("errno"); rpc.GetResponse().GetErrorText() != "success" || !ok || v != "0" {
		return errors.New("core: BLCP handshake step 2 failed")
	}

	// Step 3: account login (JSON).
	step3 := NewBLCPData(2, 50)
	step3.CorrelationID = 2000003149381050
	step3.RPCBody = BuildRPCBody(2, 50, step3.CorrelationID, 0, 1)
	step3.LCMBody = marshalJSON(map[string]any{
		"method":            50,
		"appid":             ChatAppID,
		"device_id":         "android_" + cuidGalaxy2,
		"account_type":      1,
		"token":             b.Account.BDUSS(),
		"version":           4,
		"sdk_version":       ChatSDKVersion,
		"app_version":       consts.ChatVersion,
		"app_open_type":     0,
		"client_identifier": string(marshalJSON(map[string]any{"zid": "", "version_code": ""})),
		"tail":              0,
		"timeout":           10,
		"cookie":            "",
		"device_info": map[string]any{
			"app_version": consts.ChatVersion,
			"os_version":  "32",
			"platform":    "android",
			"appid":       fmt.Sprint(ChatAppID),
			"from":        blcpLoginFrom,
			"cfrom":       blcpLoginFrom,
		},
		"rpc":          string(marshalJSON(map[string]any{"rpc_retry_time": 0})),
		"user_type":    0,
		"client_logid": time.Now().UnixMilli() * 1000,
	})

	rpc, body, err = b.exchange(step3)
	if err != nil {
		return fmt.Errorf("core: BLCP handshake step 3: %w", err)
	}
	if v, ok := body.JSONValue("err_code"); rpc.GetResponse().GetErrorText() != "success" || !ok || v != "0" {
		return errors.New("core: BLCP handshake step 3 failed")
	}

	reply := body.JSON
	if triggers, ok := reply["trigger_id"].([]any); ok && len(triggers) > 0 {
		if v, ok := triggers[0].(float64); ok {
			b.triggerID = int64(v)
		}
	}
	if v, ok := reply["uk"]; ok {
		b.uk = fmt.Sprint(v)
	}
	if v, ok := reply["bd_uid"]; ok {
		b.bduk = fmt.Sprint(v)
	}
	if v, ok := reply["login_id"]; ok {
		b.loginID = fmt.Sprint(v)
	}

	b.mu.Lock()
	b.status = 1
	b.mu.Unlock()
	return nil
}

// exchange sends a request frame and waits for the matching response.
func (b *BLCPCore) exchange(req *BLCPData) (*protobuf.RpcMeta, *BLCPBody, error) {
	b.mu.Lock()
	conn := b.conn
	waiter := b.waiter
	b.mu.Unlock()
	if conn == nil || waiter == nil {
		return nil, nil, errors.New("core: BLCP is not connected")
	}

	_, resp := waiter.new(req.CorrelationID)
	if err := b.writeFrame(conn, req); err != nil {
		resp.cancel()
		return nil, nil, err
	}
	frame, err := resp.Read()
	if err != nil {
		return nil, nil, err
	}

	rpc := &protobuf.RpcMeta{}
	if err := proto.Unmarshal(frame.RPCBody, rpc); err != nil {
		return nil, nil, fmt.Errorf("core: decoding the BLCP RpcMeta: %w", err)
	}
	return rpc, frame.body, nil
}

// SendLcm performs a BLCP Lcm request and returns the decoded JSON response,
// mirroring the request helpers of the chatroom APIs.
//
// The client_logid and rpc fields are injected into requestData, matching the
// Python send_request helper of send_chatroom_msg.
func (b *BLCPCore) SendLcm(serviceID, methodID int64, requestData map[string]any) (map[string]any, error) {
	req := NewBLCPData(serviceID, methodID)
	req.RPCBody = BuildRPCBody(serviceID, methodID, req.CorrelationID, 0, 1)

	requestData["client_logid"] = req.CorrelationID
	requestData["rpc"] = string(marshalJSON(map[string]any{"rpc_retry_time": 0}))
	req.LCMBody = marshalJSON(requestData)

	rpc, body, err := b.exchange(req)
	if err != nil {
		return nil, err
	}
	if rpc.GetResponse().GetErrorText() != "success" {
		return nil, fmt.Errorf("core: BLCP lcm request failed: %s", rpc.GetResponse().GetErrorText())
	}
	if body == nil || body.JSON == nil {
		return nil, errors.New("core: BLCP lcm response has no JSON body")
	}
	return body.JSON, nil
}

// Heartbeat sends a keep-alive frame, mirroring BLCPCore.heartbeat.
func (b *BLCPCore) Heartbeat() error {
	b.mu.Lock()
	conn := b.conn
	b.mu.Unlock()
	if conn == nil {
		return errors.New("core: BLCP is not connected")
	}

	req := NewBLCPData(1, 3)
	req.RPCBody = BuildRPCBody(1, 3, req.CorrelationID, 0, 1)
	req.LCMBody = marshal(&protobuf.RpcData{
		LcmRequest: &protobuf.LcmRequest{
			LogId:     req.CorrelationID,
			Timestamp: req.Timestamp,
		},
	})
	return b.writeFrame(conn, req)
}

func (b *BLCPCore) dispatch(responses *ClientBLCPResponses, conn net.Conn) {
	for {
		frame, body, err := responses.Next()
		if err != nil {
			break
		}
		if frame == nil {
			continue
		}

		b.mu.Lock()
		waiter := b.waiter
		b.mu.Unlock()
		if waiter != nil {
			waiter.setDone(frame.CorrelationID, frame)
		}

		// The upstream re-encodes and re-parses the frame here; the already
		// decoded body is reused instead.
		if !frame.IsNotify {
			continue
		}
		select {
		case b.messages <- body:
		default:
			// Queue full: drop the oldest message, mirroring the Python code.
			select {
			case <-b.messages:
			default:
			}
			select {
			case b.messages <- body:
			default:
			}
		}
	}

	b.mu.Lock()
	b.status = -1
	b.mu.Unlock()
}

func (b *BLCPCore) writeFrame(conn net.Conn, frame *BLCPData) error {
	b.writeMu.Lock()
	defer b.writeMu.Unlock()
	if _, err := conn.Write(frame.Bytes()); err != nil {
		return fmt.Errorf("core: writing the BLCP frame: %w", err)
	}
	return nil
}

// BuildRPCBody builds the RpcMeta preamble of a request, mirroring
// BLCPCore.buildRpcBody.
func BuildRPCBody(serviceID, methodID, correlationID int64, compressType, needCommon int32) []byte {
	meta := &protobuf.RpcMeta{
		Request: &protobuf.RpcRequestMeta{
			LogId:      correlationID,
			ServiceId:  serviceID,
			MethodId:   methodID,
			NeedCommon: needCommon,
			EventList: []*protobuf.EventTimestamp{
				{Event: "CLCPReqBegin", TimestampMs: time.Now().UnixMilli()},
			},
		},
		CorrelationId:      correlationID,
		CompressType:       &compressType,
		AcceptCompressType: 1,
	}
	return marshal(meta)
}

// GetBDUKFromUserID derives the bduk of a user id, mirroring
// BLCPCore.getBDUKfromUserId: AES-CBC with a fixed key and iv, then base64url
// without padding.
func GetBDUKFromUserID(userID string) (string, error) {
	encrypted, err := crypto.CBCEncrypt([]byte("AFD311832EDEEAEF"), []byte("2011121211143000"), []byte(userID))
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(encrypted), nil
}

// GetMsgKey builds a msg key, mirroring BLCPCore.getmsgkey.
func GetMsgKey(bduk string) string {
	return bduk + fmt.Sprint(time.Now().UnixMilli()*1000) + fmt.Sprint(mrand.Int64())
}

func marshal(msg proto.Message) []byte {
	out, _ := proto.Marshal(msg)
	return out
}

func marshalJSON(v any) []byte {
	out, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return out
}
