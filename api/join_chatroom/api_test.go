package joinchatroom

import (
	"strings"
	"testing"

	"github.com/rongyuio/aiotieba-go/config"
	"github.com/rongyuio/aiotieba-go/core"
)

func TestConstructRequestData(t *testing.T) {
	data := ConstructRequestData(456, 789)

	if len(data) != 4 {
		t.Fatalf("len(request data) = %d, want 4: %v", len(data), data)
	}
	want := map[string]any{
		"method":   int64(201),
		"mcast_id": int64(456),
		"appid":    int64(core.ChatAppID),
		"uk":       int64(789),
	}
	for k, v := range want {
		if data[k] != v {
			t.Errorf("%s = %v, want %v", k, data[k], v)
		}
	}

	for _, key := range []string{"client_logid", "rpc", "origin_id", "msg_key"} {
		if _, ok := data[key]; ok {
			t.Errorf("request data carries %q, which the upstream body does not", key)
		}
	}
}

func TestRequestWithoutConnection(t *testing.T) {
	account, err := core.NewAccount(strings.Repeat("b", 192), "")
	if err != nil {
		t.Fatalf("NewAccount: %v", err)
	}
	blcpCore := core.NewBLCPCore(account, core.NewNetCore(nil, config.TimeoutConfig{}), 10)

	if err := Request(blcpCore, 456); err == nil {
		t.Fatal("Request succeeded without a BLCP connection")
	}
}
