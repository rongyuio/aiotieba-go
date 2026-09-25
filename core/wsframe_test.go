package core

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"strings"
	"testing"
	"time"
)

// memConn 是测试用的内存 net.Conn：读来自 in，写追加到 out。
type memConn struct {
	in  *bytes.Reader
	out bytes.Buffer
}

func (c *memConn) Read(p []byte) (int, error)       { return c.in.Read(p) }
func (c *memConn) Write(p []byte) (int, error)      { return c.out.Write(p) }
func (c *memConn) Close() error                     { return nil }
func (c *memConn) LocalAddr() net.Addr              { return &net.TCPAddr{} }
func (c *memConn) RemoteAddr() net.Addr             { return &net.TCPAddr{} }
func (c *memConn) SetDeadline(time.Time) error      { return nil }
func (c *memConn) SetReadDeadline(time.Time) error  { return nil }
func (c *memConn) SetWriteDeadline(time.Time) error { return nil }

func newTestWsConn(stream []byte) (*wsConn, *memConn) {
	mc := &memConn{in: bytes.NewReader(stream)}
	return newWsConn(mc, bufio.NewReader(mc.in), time.Second), mc
}

// serverFrame 构造一个服务端帧（不掩码）。
func serverFrame(fin bool, opcode byte, payload []byte) []byte {
	var hdr []byte
	b0 := opcode
	if fin {
		b0 |= 0x80
	}
	switch size := len(payload); {
	case size < 126:
		hdr = []byte{b0, byte(size)}
	case size <= 0xFFFF:
		hdr = []byte{b0, 126, 0, 0}
		binary.BigEndian.PutUint16(hdr[2:4], uint16(size))
	default:
		hdr = []byte{b0, 127, 0, 0, 0, 0, 0, 0, 0, 0}
		binary.BigEndian.PutUint64(hdr[2:10], uint64(size))
	}
	return append(hdr, payload...)
}

// parseClientFrame 解析一个客户端帧（带掩码），返回 opcode、解掩码后的负载与剩余字节。
func parseClientFrame(t *testing.T, raw []byte) (byte, []byte, []byte) {
	t.Helper()
	if len(raw) < 2 {
		t.Fatalf("client frame too short: %d bytes", len(raw))
	}
	fin := raw[0]&0x80 != 0
	opcode := raw[0] & 0x0F
	masked := raw[1]&0x80 != 0
	if !fin {
		t.Error("client frame FIN = 0, want 1")
	}
	if !masked {
		t.Error("client frame MASK = 0, want 1")
	}

	pos := 2
	size := int64(raw[1] & 0x7F)
	switch size {
	case 126:
		size = int64(binary.BigEndian.Uint16(raw[pos : pos+2]))
		pos += 2
	case 127:
		size = int64(binary.BigEndian.Uint64(raw[pos : pos+8]))
		pos += 8
	}
	key := raw[pos : pos+4]
	pos += 4

	payload := make([]byte, size)
	for i := range payload {
		payload[i] = raw[pos+i] ^ key[i%4]
	}
	pos += int(size)
	return opcode, payload, raw[pos:]
}

// TestReadFrameRoundTrip 覆盖服务端方向（不掩码）的读取。
func TestReadFrameRoundTrip(t *testing.T) {
	for _, size := range []int{0, 1, 125, 126, 127, 65535, 65536, 70000} {
		payload := bytes.Repeat([]byte{0xAB}, size)

		fin, opcode, got, err := readFrame(bufio.NewReader(bytes.NewReader(serverFrame(true, wsOpBinary, payload))))
		if err != nil {
			t.Fatalf("size %d: readFrame: %v", size, err)
		}
		if !fin {
			t.Errorf("size %d: FIN = false", size)
		}
		if opcode != wsOpBinary {
			t.Errorf("size %d: opcode = %#x", size, opcode)
		}
		if !bytes.Equal(got, payload) {
			t.Errorf("size %d: payload mismatch (%d vs %d bytes)", size, len(got), len(payload))
		}
	}
}

// TestWriteFrameRoundTrip 覆盖客户端方向（必须掩码）的写出。
func TestWriteFrameRoundTrip(t *testing.T) {
	for _, size := range []int{0, 1, 125, 126, 127, 65535, 65536, 70000} {
		payload := bytes.Repeat([]byte{0xCD}, size)

		var buf bytes.Buffer
		if err := writeFrame(&buf, wsOpBinary, payload); err != nil {
			t.Fatalf("size %d: writeFrame: %v", size, err)
		}

		opcode, got, rest := parseClientFrame(t, buf.Bytes())
		if opcode != wsOpBinary {
			t.Errorf("size %d: opcode = %#x", size, opcode)
		}
		if len(rest) != 0 {
			t.Errorf("size %d: %d trailing bytes", size, len(rest))
		}
		if !bytes.Equal(got, payload) {
			t.Errorf("size %d: payload mismatch (%d vs %d bytes)", size, len(got), len(payload))
		}
	}
}

func TestWriteFrameHeaderShape(t *testing.T) {
	// 长度编码的三个区间必须落在正确的形式上。
	for _, tc := range []struct {
		size    int
		lenByte byte
		extLen  int // 扩展长度占用的字节数
	}{
		{0, 0x80, 0},
		{125, 0x80 | 125, 0},
		{126, 0x80 | 126, 2},
		{65535, 0x80 | 126, 2},
		{65536, 0x80 | 127, 8},
	} {
		var buf bytes.Buffer
		if err := writeFrame(&buf, wsOpBinary, make([]byte, tc.size)); err != nil {
			t.Fatalf("size %d: %v", tc.size, err)
		}
		raw := buf.Bytes()
		if raw[0] != 0x80|wsOpBinary {
			t.Errorf("size %d: first byte = %#x, want %#x", tc.size, raw[0], 0x80|wsOpBinary)
		}
		if raw[1] != tc.lenByte {
			t.Errorf("size %d: length byte = %#x, want %#x", tc.size, raw[1], tc.lenByte)
		}
		want := 2 + tc.extLen + 4 + tc.size
		if len(raw) != want {
			t.Errorf("size %d: frame = %d bytes, want %d", tc.size, len(raw), want)
		}
	}
}

func TestWriteFrameMasksPayload(t *testing.T) {
	payload := []byte("the quick brown fox")
	var buf bytes.Buffer
	if err := writeFrame(&buf, wsOpBinary, payload); err != nil {
		t.Fatalf("writeFrame: %v", err)
	}

	raw := buf.Bytes()
	if raw[1]&0x80 == 0 {
		t.Fatal("MASK bit is not set")
	}
	if bytes.Contains(raw, payload) {
		t.Error("payload appears unmasked on the wire")
	}

	_, got, rest := parseClientFrame(t, raw)
	if len(rest) != 0 {
		t.Errorf("%d trailing bytes", len(rest))
	}
	if !bytes.Equal(got, payload) {
		t.Errorf("payload = %q, want %q", got, payload)
	}
}

func TestWriteFrameUsesFreshMaskKey(t *testing.T) {
	// 掩码键必须是随机的：同样的负载两次编码不应得到相同的字节。
	payload := []byte("same payload")
	var a, b bytes.Buffer
	if err := writeFrame(&a, wsOpBinary, payload); err != nil {
		t.Fatal(err)
	}
	if err := writeFrame(&b, wsOpBinary, payload); err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(a.Bytes(), b.Bytes()) {
		t.Error("two frames with the same payload encoded identically; mask key is not random")
	}
}

func TestReadFrameRejectsMaskedServerFrame(t *testing.T) {
	// 服务端帧不得掩码。
	frame := serverFrame(true, wsOpBinary, []byte("hi"))
	frame[1] |= 0x80 // 强行打上 MASK

	_, _, _, err := readFrame(bufio.NewReader(bytes.NewReader(frame)))
	if err == nil || !strings.Contains(err.Error(), "must not be masked") {
		t.Fatalf("err = %v, want a mask violation", err)
	}
}

func TestReadFrameRejectsReservedBits(t *testing.T) {
	frame := serverFrame(true, wsOpBinary, []byte("hi"))
	frame[0] |= 0x40 // RSV1

	_, _, _, err := readFrame(bufio.NewReader(bytes.NewReader(frame)))
	if err == nil || !strings.Contains(err.Error(), "reserved bits") {
		t.Fatalf("err = %v, want a reserved-bits violation", err)
	}
}

func TestReadFrameRejectsOversizedFrame(t *testing.T) {
	hdr := []byte{0x80 | wsOpBinary, 127, 0, 0, 0, 0, 0, 0, 0, 0}
	binary.BigEndian.PutUint64(hdr[2:10], uint64(wsMaxMessageSize)+1)

	_, _, _, err := readFrame(bufio.NewReader(bytes.NewReader(hdr)))
	if err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("err = %v, want a size violation", err)
	}
}

func TestReadMessageAssemblesFragments(t *testing.T) {
	stream := append(serverFrame(false, wsOpBinary, []byte("first ")), nil...)
	stream = append(stream, serverFrame(false, wsOpContinuation, []byte("second "))...)
	stream = append(stream, serverFrame(true, wsOpContinuation, []byte("third"))...)

	conn, _ := newTestWsConn(stream)
	msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	if string(msg) != "first second third" {
		t.Errorf("msg = %q", msg)
	}
}

func TestReadMessageRepliesPingWithPong(t *testing.T) {
	stream := append(serverFrame(true, wsOpPing, []byte("beat")), nil...)
	stream = append(stream, serverFrame(true, wsOpBinary, []byte("payload"))...)

	conn, mc := newTestWsConn(stream)
	msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	if string(msg) != "payload" {
		t.Errorf("msg = %q, want the binary frame after the ping", msg)
	}

	opcode, payload, rest := parseClientFrame(t, mc.out.Bytes())
	if opcode != wsOpPong {
		t.Errorf("opcode = %#x, want pong", opcode)
	}
	if string(payload) != "beat" {
		t.Errorf("pong payload = %q, want the ping payload", payload)
	}
	if len(rest) != 0 {
		t.Errorf("%d trailing bytes after the pong", len(rest))
	}
}

func TestReadMessageIgnoresPong(t *testing.T) {
	stream := append(serverFrame(true, wsOpPong, []byte("unsolicited")), nil...)
	stream = append(stream, serverFrame(true, wsOpBinary, []byte("payload"))...)

	conn, mc := newTestWsConn(stream)
	msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	if string(msg) != "payload" {
		t.Errorf("msg = %q", msg)
	}
	if mc.out.Len() != 0 {
		t.Error("an unsolicited pong should not be answered")
	}
}

func TestReadMessageOnCloseEchoesAndStops(t *testing.T) {
	conn, mc := newTestWsConn(serverFrame(true, wsOpClose, []byte{0x03, 0xE8}))

	_, err := conn.ReadMessage()
	if !errors.Is(err, errWsClosedByPeer) {
		t.Fatalf("err = %v, want errWsClosedByPeer", err)
	}

	opcode, payload, _ := parseClientFrame(t, mc.out.Bytes())
	if opcode != wsOpClose {
		t.Errorf("opcode = %#x, want close", opcode)
	}
	if !bytes.Equal(payload, []byte{0x03, 0xE8}) {
		t.Errorf("close payload = %v, want the peer's status echoed", payload)
	}
}

func TestReadMessageRejectsUnexpectedContinuation(t *testing.T) {
	conn, _ := newTestWsConn(serverFrame(true, wsOpContinuation, []byte("stray")))

	_, err := conn.ReadMessage()
	if err == nil || !strings.Contains(err.Error(), "unexpected") {
		t.Fatalf("err = %v, want an unexpected-continuation error", err)
	}
}

func TestReadMessageRejectsNewMessageWhileFragmenting(t *testing.T) {
	stream := append(serverFrame(false, wsOpBinary, []byte("part")), nil...)
	stream = append(stream, serverFrame(true, wsOpBinary, []byte("restart"))...)

	conn, _ := newTestWsConn(stream)
	_, err := conn.ReadMessage()
	if err == nil || !strings.Contains(err.Error(), "before the previous") {
		t.Fatalf("err = %v, want a fragmentation-order error", err)
	}
}

func TestReadMessageRejectsOversizedMessage(t *testing.T) {
	// 单个帧在限额内，但多帧累加后超出。
	chunk := wsMaxMessageSize/2 + 1
	stream := append(serverFrame(false, wsOpBinary, make([]byte, chunk)), nil...)
	stream = append(stream, serverFrame(true, wsOpContinuation, make([]byte, chunk))...)

	conn, _ := newTestWsConn(stream)
	_, err := conn.ReadMessage()
	if err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("err = %v, want a size violation", err)
	}
}
