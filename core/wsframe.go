package core

import (
	"bufio"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// 本文件实现 WebSocket 帧层（RFC 6455）。
//
// 只覆盖贴吧 IM 客户端用到的子集：二进制消息、客户端掩码、分片重组、ping/pong 与 close。
// 之所以自己实现而不继续用 gorilla，是因为握手必须携带非标准的 Sec-WebSocket-Extensions 头，
// 而 gorilla 把它列为保留头、硬性拒绝调用方设置，且它的 *Conn 只能由它自己的握手中产生，
// 没有办法"自己握手 + 复用它的帧层"。详见 WsCore.Connect 的说明。

const (
	wsOpContinuation byte = 0x0
	wsOpText         byte = 0x1
	wsOpBinary       byte = 0x2
	wsOpClose        byte = 0x8
	wsOpPing         byte = 0x9
	wsOpPong         byte = 0xA
)

// wsMaxMessageSize 是单条消息的上限，对齐 Python 客户端的 4 MiB。
const wsMaxMessageSize = 4 * 1024 * 1024

// errWsClosedByPeer 表示对端发送了 close 帧。
var errWsClosedByPeer = errors.New("core: websocket closed by peer")

// wsConn 是一条已完成握手的连接上的 websocket 帧通道。
type wsConn struct {
	conn net.Conn
	br   *bufio.Reader

	// writeTimeout 用于控制帧（pong / close）的写超时；数据帧的截止时间由调用方设置。
	writeTimeout time.Duration

	// wmu 串行化所有写操作。数据帧与控制帧可能来自不同 goroutine（读循环要回 pong），
	// 而 websocket 不允许两帧的字节交错。
	wmu sync.Mutex
}

func newWsConn(conn net.Conn, br *bufio.Reader, writeTimeout time.Duration) *wsConn {
	return &wsConn{conn: conn, br: br, writeTimeout: writeTimeout}
}

// SetReadDeadline 设置底层连接的读截止时间。
func (c *wsConn) SetReadDeadline(t time.Time) error { return c.conn.SetReadDeadline(t) }

// SetWriteDeadline 设置底层连接的写截止时间。
func (c *wsConn) SetWriteDeadline(t time.Time) error { return c.conn.SetWriteDeadline(t) }

// Close 关闭底层连接。
func (c *wsConn) Close() error { return c.conn.Close() }

// WriteMessage 以单个二进制帧发送 payload。
func (c *wsConn) WriteMessage(payload []byte) error {
	c.wmu.Lock()
	defer c.wmu.Unlock()
	return writeFrame(c.conn, wsOpBinary, payload)
}

// WriteClose 发送一个 close 帧。
func (c *wsConn) WriteClose() error {
	c.wmu.Lock()
	defer c.wmu.Unlock()
	if d := c.writeTimeout; d > 0 {
		_ = c.conn.SetWriteDeadline(time.Now().Add(d))
		defer func() { _ = c.conn.SetWriteDeadline(time.Time{}) }()
	}
	return writeFrame(c.conn, wsOpClose, nil)
}

// ReadMessage 读取下一条完整消息。
//
// 分片会被重组；ping 会就地回以 pong；对端 close 返回 errWsClosedByPeer。
func (c *wsConn) ReadMessage() ([]byte, error) {
	var (
		msg        []byte
		inFragment bool
	)

	for {
		fin, opcode, payload, err := readFrame(c.br)
		if err != nil {
			return nil, err
		}

		switch opcode {
		case wsOpPing:
			c.reply(wsOpPong, payload)
			continue
		case wsOpPong:
			continue
		case wsOpClose:
			// 按 RFC 6455 回一个 close，然后结束。
			c.reply(wsOpClose, payload)
			return nil, errWsClosedByPeer
		case wsOpContinuation:
			if !inFragment {
				return nil, errors.New("core: unexpected websocket continuation frame")
			}
		case wsOpBinary, wsOpText:
			if inFragment {
				return nil, errors.New("core: new websocket message before the previous one finished")
			}
			inFragment = true
		default:
			return nil, fmt.Errorf("core: unsupported websocket opcode %#x", opcode)
		}

		msg = append(msg, payload...)
		if len(msg) > wsMaxMessageSize {
			return nil, fmt.Errorf("core: websocket message exceeds %d bytes", wsMaxMessageSize)
		}

		if fin {
			return msg, nil
		}
	}
}

// reply 发送一个控制帧，失败只忽略——对端已不可用时不值得让读循环报错。
func (c *wsConn) reply(opcode byte, payload []byte) {
	c.wmu.Lock()
	defer c.wmu.Unlock()
	if d := c.writeTimeout; d > 0 {
		_ = c.conn.SetWriteDeadline(time.Now().Add(d))
		defer func() { _ = c.conn.SetWriteDeadline(time.Time{}) }()
	}
	_ = writeFrame(c.conn, opcode, payload)
}

// writeFrame 写出一个终帧（FIN=1）。客户端发出的帧必须掩码。
func writeFrame(w io.Writer, opcode byte, payload []byte) error {
	// 头部最长 14 字节：2 固定 + 8 扩展长度 + 4 掩码键。
	var (
		hdr [14]byte
		n   int
	)
	hdr[0] = 0x80 | opcode
	hdr[1] = 0x80 // MASK=1
	n = 2

	switch size := len(payload); {
	case size < 126:
		hdr[1] |= byte(size)
	case size <= 0xFFFF:
		hdr[1] |= 126
		binary.BigEndian.PutUint16(hdr[2:4], uint16(size))
		n = 4
	default:
		hdr[1] |= 127
		binary.BigEndian.PutUint64(hdr[2:10], uint64(size))
		n = 10
	}

	var key [4]byte
	if _, err := rand.Read(key[:]); err != nil {
		return fmt.Errorf("core: generating websocket mask key: %w", err)
	}
	copy(hdr[n:n+4], key[:])
	n += 4

	// 头与负载一次写出，避免两次系统调用把一帧拆开。
	buf := make([]byte, n+len(payload))
	copy(buf, hdr[:n])
	for i, b := range payload {
		buf[n+i] = b ^ key[i%4]
	}

	if _, err := w.Write(buf); err != nil {
		return err
	}
	return nil
}

// readFrame 读取一个帧。返回的 payload 已解掩码。
//
// 服务端发来的帧不得掩码（RFC 6455 5.1），否则按协议错误处理。
func readFrame(r io.Reader) (fin bool, opcode byte, payload []byte, err error) {
	var h [2]byte
	if _, err = io.ReadFull(r, h[:]); err != nil {
		return false, 0, nil, err
	}

	if h[0]&0x70 != 0 {
		return false, 0, nil, errors.New("core: websocket frame has reserved bits set")
	}
	fin = h[0]&0x80 != 0
	opcode = h[0] & 0x0F

	masked := h[1]&0x80 != 0
	if masked {
		return false, 0, nil, errors.New("core: server websocket frame must not be masked")
	}

	size := int64(h[1] & 0x7F)
	switch size {
	case 126:
		var ext [2]byte
		if _, err = io.ReadFull(r, ext[:]); err != nil {
			return false, 0, nil, err
		}
		size = int64(binary.BigEndian.Uint16(ext[:]))
	case 127:
		var ext [8]byte
		if _, err = io.ReadFull(r, ext[:]); err != nil {
			return false, 0, nil, err
		}
		u := binary.BigEndian.Uint64(ext[:])
		if u > wsMaxMessageSize {
			return false, 0, nil, fmt.Errorf("core: websocket frame exceeds %d bytes", wsMaxMessageSize)
		}
		size = int64(u)
	}
	if size > wsMaxMessageSize {
		return false, 0, nil, fmt.Errorf("core: websocket frame exceeds %d bytes", wsMaxMessageSize)
	}

	payload = make([]byte, size)
	if _, err = io.ReadFull(r, payload); err != nil {
		return false, 0, nil, err
	}
	return fin, opcode, payload, nil
}
