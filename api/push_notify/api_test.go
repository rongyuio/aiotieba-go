package pushnotify

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The golden fixtures under testdata/ were produced by the Python bindings
// generated from the same .proto files.

func TestParseBodyMatchesPython(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "response.hex"))
	if err != nil {
		t.Fatalf("reading testdata/response.hex: %v", err)
	}
	body, err := hex.DecodeString(strings.TrimSpace(string(raw)))
	if err != nil {
		t.Fatalf("decoding the fixture: %v", err)
	}

	notifies, err := ParseBody(body)
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if len(notifies) != 2 {
		t.Fatalf("len(notifies) = %d, want 2", len(notifies))
	}

	first := notifies[0]
	if first.NoteType != 1 || first.GroupID != 111 || first.GroupType != 3 || first.MsgID != 222 {
		t.Errorf("first = %+v", first)
	}
	if first.CreateTime != 1700000000 {
		t.Errorf("create time = %d, want 1700000000", first.CreateTime)
	}

	// An empty et string becomes a zero create time.
	second := notifies[1]
	if second.GroupID != 333 || second.NoteType != 10 {
		t.Errorf("second = %+v", second)
	}
	if second.CreateTime != 0 {
		t.Errorf("create time = %d, want 0 (empty et)", second.CreateTime)
	}
}

func TestCMD(t *testing.T) {
	if CMD != 202006 {
		t.Errorf("CMD = %d, want 202006", CMD)
	}
}
