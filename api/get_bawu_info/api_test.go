package getbawuinfo

import (
	"bytes"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba/exception"

	pb "github.com/rongyuio/aiotieba/api/get_bawu_info/protobuf"
	commonpb "github.com/rongyuio/aiotieba/protobuf"
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
	got := PackProto(12345)
	if want := loadHex(t, "req.hex"); !bytes.Equal(got, want) {
		t.Errorf("PackProto = %x\n         want %x", got, want)
	}
}

func TestParseBodyMatchesPython(t *testing.T) {
	info, err := ParseBody(loadHex(t, "response.hex"))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}

	// The wire order is 小吧主 → 吧主 → 图片小编, but the `all` list follows the
	// fixed extract order 吧主 → 小吧主 → 图片小编.
	wantIDs := []int64{111, 222, 223, 555}
	if len(info.All) != len(wantIDs) {
		t.Fatalf("len(All) = %d, want %d", len(info.All), len(wantIDs))
	}
	for i, want := range wantIDs {
		if info.All[i].UserID != want {
			t.Errorf("All[%d].UserID = %d, want %d", i, info.All[i].UserID, want)
		}
	}

	if len(info.Admin) != 1 || info.Admin[0].UserID != 111 {
		t.Errorf("admin = %+v", info.Admin)
	}
	if len(info.Manager) != 2 || info.Manager[0].UserID != 222 || info.Manager[1].UserID != 223 {
		t.Errorf("manager = %+v", info.Manager)
	}
	if len(info.ImageEditor) != 1 || info.ImageEditor[0].UserID != 555 {
		t.Errorf("image editor = %+v", info.ImageEditor)
	}
	// The remaining buckets stay empty.
	if len(info.VoiceEditor)+len(info.VideoEditor)+len(info.BroadcastEditor)+
		len(info.JournalChiefEditor)+len(info.JournalEditor)+len(info.ProfessAdmin)+len(info.FourthAdmin) != 0 {
		t.Errorf("unexpected non-empty buckets = %+v", info)
	}

	// The portrait is NOT trimmed for this endpoint.
	admin := info.Admin[0]
	if admin.Portrait != "tb.1.a1" || admin.UserName != "a1" || admin.NickNameNew != "a1_昵称" {
		t.Errorf("admin fields = %+v", admin)
	}
	if admin.Level != 1 {
		t.Errorf("admin level = %d, want 1", admin.Level)
	}
	if got := admin.ShowName(); got != "a1_昵称" {
		t.Errorf("ShowName() = %q", got)
	}
	if got := admin.LogName(); got != "a1" {
		t.Errorf("LogName() = %q", got)
	}
	if !admin.Valid() {
		t.Error("Valid() = false, want true")
	}
}

func TestParseBodySurfacesServerError(t *testing.T) {
	res := &pb.GetBawuInfoResIdl{Error: &commonpb.Error{Errorno: 340006, Errmsg: "boom"}}
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
	if u.Scheme != "http" || u.Host != "tiebac.baidu.com" || u.Path != "/c/f/forum/getBawuInfo" {
		t.Errorf("url = %s", u)
	}
	if u.RawQuery != "cmd=301007" {
		t.Errorf("query = %q, want cmd=301007", u.RawQuery)
	}
}
