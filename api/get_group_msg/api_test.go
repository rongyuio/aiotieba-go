package getgroupmsg

import (
	"bytes"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba/core"
	"github.com/rongyuio/aiotieba/exception"

	pb "github.com/rongyuio/aiotieba/api/get_group_msg/protobuf"
	commonpb "github.com/rongyuio/aiotieba/protobuf"
)

// The golden fixtures under testdata/ were produced by the Python bindings
// generated from the same .proto files.

const testCuid = "0123456789abcdef"

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

// newAccount returns an account whose cuid matches the fixture.
func newAccount(t *testing.T) *core.Account {
	t.Helper()
	account, err := core.NewAccount(strings.Repeat("b", 192), "")
	if err != nil {
		t.Fatalf("NewAccount: %v", err)
	}
	account.SetCuid(testCuid)
	return account
}

func TestPackProtoMatchesPython(t *testing.T) {
	got := PackProto(newAccount(t), []int64{111, 333}, []int64{222, 444}, 1)
	if want := loadHex(t, "req.hex"); !bytes.Equal(got, want) {
		t.Errorf("PackProto = %x\n         want %x", got, want)
	}
}

func TestPackProtoZipsToShortest(t *testing.T) {
	// Python uses zip(strict=False): the extra id is dropped.
	req := &pb.GetGroupMsgReqIdl{}
	if err := proto.Unmarshal(PackProto(newAccount(t), []int64{1, 2, 3}, []int64{9}, 1), req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := len(req.GetData().GetGroupMids()); got != 1 {
		t.Fatalf("len(groupMids) = %d, want 1", got)
	}
	if req.GetData().GetGroupMids()[0].GetGroupId() != 1 {
		t.Errorf("group id = %d, want 1", req.GetData().GetGroupMids()[0].GetGroupId())
	}
	if req.GetCuid() != testCuid+cuidSuffix {
		t.Errorf("cuid = %q, want %q", req.GetCuid(), testCuid+cuidSuffix)
	}
	if req.GetData().GetGettype() != "1" {
		t.Errorf("gettype = %q, want \"1\"", req.GetData().GetGettype())
	}
}

func TestParseBodyMatchesPython(t *testing.T) {
	groups, err := ParseBody(loadHex(t, "response.hex"))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if groups.Len() != 1 {
		t.Fatalf("len(groups) = %d, want 1", groups.Len())
	}

	group := groups.Objs[0]
	if group.GroupID != 111 || group.GroupType != 3 {
		t.Errorf("group = %+v", group)
	}
	if len(group.Messages) != 1 {
		t.Fatalf("len(messages) = %d, want 1", len(group.Messages))
	}

	msg := group.Messages[0]
	if msg.MsgID != 222 || msg.MsgType != 1 || msg.Text != "消息内容" {
		t.Errorf("message = %+v", msg)
	}
	if msg.CreateTime != 1700000000 {
		t.Errorf("create time = %d", msg.CreateTime)
	}
	// The sender portrait keeps its "?..." suffix stripping.
	if msg.User.UserID != 4444444 || msg.User.Portrait != "tb.1.abcdefghijklmnopqrst" {
		t.Errorf("user = %+v", msg.User)
	}
	if got := msg.User.String(); got != "某个用户名" {
		t.Errorf("String() = %q", got)
	}
	if got := msg.User.LogName(); got != "某个用户名" {
		t.Errorf("LogName() = %q", got)
	}
	if !msg.User.Valid() {
		t.Error("Valid() = false, want true")
	}
}

func TestParseBodySurfacesServerError(t *testing.T) {
	res := &pb.GetGroupMsgResIdl{Error: &commonpb.Error{Errorno: 340006, Errmsg: "boom"}}
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

func TestCMD(t *testing.T) {
	if CMD != 202003 {
		t.Errorf("CMD = %d, want 202003", CMD)
	}
}
