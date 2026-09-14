package gethomepage

import (
	"bytes"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The golden fixtures under ../testdata were produced by the Python bindings
// generated from the same .proto files.

func TestPackProtoMatchesPython(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "testdata", "req_homepage.hex"))
	if err != nil {
		t.Fatalf("reading testdata/req_homepage.hex: %v", err)
	}
	want, err := hex.DecodeString(strings.TrimSpace(string(raw)))
	if err != nil {
		t.Fatalf("decoding the fixture: %v", err)
	}

	got := PackProto(4444444, 2)
	if !bytes.Equal(got, want) {
		t.Errorf("PackProto = %x\n         want %x", got, want)
	}
}

func TestRequestURL(t *testing.T) {
	u := RequestURL()
	if u.Scheme != "http" || u.Host != "tiebac.baidu.com" || u.Path != "/c/u/user/profile" {
		t.Errorf("url = %s", u)
	}
	if u.RawQuery != "cmd=303012" {
		t.Errorf("query = %q, want cmd=303012", u.RawQuery)
	}
}
