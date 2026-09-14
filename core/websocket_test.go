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

	if got := m.GetRecordID(); got != 1 {
		t.Errorf("initial record id = %d, want 1", got)
	}
	m.GID2MID[m.PrivGID].Update(7)
	if got := m.GetMsgID(m.PrivGID); got != 0 {
		t.Errorf("GetMsgID after one update = %d, want 0 (LastID)", got)
	}
	m.GID2MID[m.PrivGID].Update(9)
	if got := m.GetMsgID(m.PrivGID); got != 7 {
		t.Errorf("GetMsgID after two updates = %d, want 7", got)
	}
	if got := m.GetRecordID(); got != 701 {
		t.Errorf("record id = %d, want 701", got)
	}
	if got := m.GetMsgID(1234); got != 0 {
		t.Errorf("GetMsgID of an unknown group = %d, want 0", got)
	}
}

func TestWsCoreStartsClosed(t *testing.T) {
	a := wsTestAccount(t)
	w := NewWsCore(a, testNetCore())
	if w.Status() != 0 {
		t.Errorf("initial status = %d, want 0 (closed)", w.Status())
	}
	if _, err := w.Send([]byte("x"), 1); err == nil {
		t.Error("Send on a closed core: want error, got nil")
	}
}
