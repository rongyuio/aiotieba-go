package core

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// The vectors below come from running the Python pack_ws_bytes implementation
// with a 31-byte zero AES-ECB seed.
func wsTestAccount(t *testing.T) *Account {
	t.Helper()
	a, err := NewAccount("", "")
	if err != nil {
		t.Fatalf("NewAccount: %v", err)
	}
	a.SetAESECBSecKey(make([]byte, 31))
	return a
}

func TestPackWsBytesMatchesPython(t *testing.T) {
	a := wsTestAccount(t)
	data := []byte("\x08\x02\x12\x03abc")

	cases := []struct {
		name     string
		compress bool
		encrypt  bool
		want     string
	}{
		{"encrypted", false, true, "88000497c900003039159389d983e992d5a0c590a8d40dc0ed"},
		{"plain", false, false, "08000497c90000303908021203616263"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := PackWsBytes(a, data, 301001, 12345, tc.compress, tc.encrypt)
			if err != nil {
				t.Fatalf("PackWsBytes: %v", err)
			}
			if hex.EncodeToString(got) != tc.want {
				t.Errorf("packed = %s, want %s", hex.EncodeToString(got), tc.want)
			}
		})
	}

	got, err := PackWsBytes(a, nil, 1, 0, false, false)
	if err != nil {
		t.Fatalf("PackWsBytes: %v", err)
	}
	if hex.EncodeToString(got) != "080000000100000000" {
		t.Errorf("empty frame = %s", hex.EncodeToString(got))
	}
}

func TestParseWsBytesRoundTrip(t *testing.T) {
	a := wsTestAccount(t)
	payload := []byte(`{"hello":"world"}`)

	cases := []struct {
		name     string
		compress bool
		encrypt  bool
	}{
		{"encrypted", false, true},
		{"encrypted+gzip", true, true},
		{"plain", false, false},
		{"gzip", true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			packed, err := PackWsBytes(a, payload, 301001, 999, tc.compress, tc.encrypt)
			if err != nil {
				t.Fatalf("PackWsBytes: %v", err)
			}
			got, cmd, reqID, err := ParseWsBytes(a, packed)
			if err != nil {
				t.Fatalf("ParseWsBytes: %v", err)
			}
			if !bytes.Equal(got, payload) {
				t.Errorf("payload = %q, want %q", got, payload)
			}
			if cmd != 301001 {
				t.Errorf("cmd = %d, want 301001", cmd)
			}
			if reqID != 999 {
				t.Errorf("reqID = %d, want 999", reqID)
			}
		})
	}
}

func TestParseWsBytesRejectsShortFrames(t *testing.T) {
	a := wsTestAccount(t)
	if _, _, _, err := ParseWsBytes(a, []byte{0x08, 0x00}); err == nil {
		t.Error("ParseWsBytes with a 2-byte frame: want error, got nil")
	}
}

func TestMsgIDManager(t *testing.T) {
	m := NewMsgIDManager()

	m.GID2MID[m.PrivGID].Update(7)
	if got := m.GetMsgID(m.PrivGID); got != 0 {
		t.Errorf("GetMsgID after one update = %d, want 0 (LastID)", got)
	}
	m.GID2MID[m.PrivGID].Update(9)
	if got := m.GetMsgID(m.PrivGID); got != 7 {
		t.Errorf("GetMsgID after two updates = %d, want 7", got)
	}
	if got := m.GetMsgID(1234); got != 0 {
		t.Errorf("GetMsgID of an unknown group = %d, want 0", got)
	}
}

// local_msg_id 必须始终大于已拉取的消息，且游标只增不减。
func TestPrivChatCursors(t *testing.T) {
	m := NewMsgIDManager()

	if got := m.GetLocalMsgID(1); got != 1 {
		t.Errorf("initial local msg id = %d, want 1", got)
	}
	m.UpdatePulledMsgID(1, 5)
	if got := m.GetLocalMsgID(1); got != 501 {
		t.Errorf("local msg id after pulling 5 = %d, want 501", got)
	}
	m.UpdateLocalMsgID(1, 600)
	if got := m.GetLocalMsgID(1); got != 601 {
		t.Errorf("local msg id after using 600 = %d, want 601", got)
	}
	m.UpdateLocalMsgID(1, 100)
	if got := m.GetCursors(1).LocalMsgID; got != 600 {
		t.Errorf("local msg id after a smaller update = %d, want 600", got)
	}
	m.UpdateSentMsgID(1, 300)
	m.UpdateReadMsgID(1, 400)
	if c := m.GetCursors(1); c.SentMsgID != 300 || c.ReadMsgID != 400 {
		t.Errorf("cursors = %+v", c)
	}
}

// 未知分组必须被登记，否则首条消息之后的 msg_id 会永远丢失。
func TestUpdateMsgIDRegistersUnknownGroup(t *testing.T) {
	m := NewMsgIDManager()

	m.UpdateMsgID(1234, 42)
	if got := m.GetMsgID(1234); got != 42 {
		t.Errorf("GetMsgID right after the first update = %d, want 42 (group registered)", got)
	}
	m.UpdateMsgID(1234, 99)
	if got := m.GetMsgID(1234); got != 42 {
		t.Errorf("GetMsgID after the second update = %d, want 42 (previous CurrID)", got)
	}
}

func TestWsCoreStartsClosed(t *testing.T) {
	a := wsTestAccount(t)
	w := NewWsCore(a, testNetCore(), "22.12.1.0")
	if w.Status() != 0 {
		t.Errorf("initial status = %d, want 0 (closed)", w.Status())
	}
	if _, err := w.Send([]byte("x"), 1); err == nil {
		t.Error("Send on a closed core: want error, got nil")
	}
}
