package getuinfoprofile

import (
	"bytes"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rongyuio/aiotieba-go/api/profile"
)

// The golden fixtures under ../testdata were produced by the Python bindings
// generated from the same .proto files.

func loadHex(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "testdata", name))
	if err != nil {
		t.Fatalf("reading testdata/%s: %v", name, err)
	}
	b, err := hex.DecodeString(strings.TrimSpace(string(raw)))
	if err != nil {
		t.Fatalf("decoding testdata/%s: %v", name, err)
	}
	return b
}

func TestPackProtoByPortraitMatchesPython(t *testing.T) {
	ref := profile.ByPortrait("tb.1.abcdefghijklmnopqrstuv?1739164613")
	got := PackProto(ref)
	if want := loadHex(t, "req_profile.hex"); !bytes.Equal(got, want) {
		t.Errorf("PackProto = %x\n         want %x", got, want)
	}
}

func TestPackProtoByUIDMatchesPython(t *testing.T) {
	got := PackProto(profile.ByUserID(4444444))
	if want := loadHex(t, "req_profile_uid.hex"); !bytes.Equal(got, want) {
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
