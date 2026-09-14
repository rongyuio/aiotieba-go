package gettabmap

import (
	"bytes"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/get_tab_map/protobuf"
	commonpb "github.com/rongyuio/aiotieba-go/protobuf"
)

// The golden fixtures under testdata/ were produced by the Python bindings
// generated from the same .proto files.

func loadHex(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("reading testdata/%s: %v", name, err)
	}
	b, err := hex.DecodeString(strings.TrimSpace(string(raw)))
	if err != nil {
		t.Fatalf("decoding testdata/%s: %v", name, err)
	}
	return b
}

func TestPackProtoMatchesPython(t *testing.T) {
	account, err := core.NewAccount(strings.Repeat("b", 192), "")
	if err != nil {
		t.Fatalf("NewAccount: %v", err)
	}

	got := PackProto(account, "天堂鸡汤")
	if want := loadHex(t, "req.hex"); !bytes.Equal(got, want) {
		t.Errorf("PackProto = %x\n         want %x", got, want)
	}
}

func TestParseBodyMatchesPython(t *testing.T) {
	tabMap, err := ParseBody(loadHex(t, "response.hex"))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if tabMap.Len() != 2 {
		t.Fatalf("len(map) = %d, want 2", tabMap.Len())
	}
	if id, ok := tabMap.Get("全部"); !ok || id != 1 {
		t.Errorf("全部 -> %d (%v), want 1", id, ok)
	}
	if id, ok := tabMap.Get("原创"); !ok || id != 3 {
		t.Errorf("原创 -> %d (%v), want 3", id, ok)
	}
	if _, ok := tabMap.Get("不存在"); ok {
		t.Error("Get(不存在) reported a hit")
	}
	if !tabMap.Valid() {
		t.Error("Valid() = false, want true")
	}
}

func TestParseBodySurfacesServerError(t *testing.T) {
	res := &pb.SearchPostForumResIdl{Error: &commonpb.Error{Errorno: 340006, Errmsg: "boom"}}
	raw, err := proto.Marshal(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	_, err = ParseBody(raw)
	var serverErr *exception.TiebaServerError
	if !errors.As(err, &serverErr) {
		t.Fatalf("error = %v (%T), want *exception.TiebaServerError", err, err)
	}
	if serverErr.Code != 340006 || serverErr.Msg != "boom" {
		t.Errorf("server error = %+v", serverErr)
	}
}

func TestRequestURL(t *testing.T) {
	u := RequestURL()
	if u.Scheme != "https" || u.Host != "tiebac.baidu.com" || u.Path != "/c/f/forum/searchPostForum" {
		t.Errorf("url = %s", u)
	}
	if u.RawQuery != "cmd=309466" {
		t.Errorf("query = %q, want cmd=309466", u.RawQuery)
	}
}
