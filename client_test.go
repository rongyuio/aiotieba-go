package aiotieba

import (
	"net/url"
	"strings"
	"testing"

	"github.com/rongyuio/aiotieba/config"
	"github.com/rongyuio/aiotieba/core"
	"github.com/rongyuio/aiotieba/enums"
	"github.com/rongyuio/aiotieba/helper"
)

func TestNewValidatesTokens(t *testing.T) {
	if _, err := New("bad", ""); err == nil {
		t.Error("New with a bad BDUSS: want error, got nil")
	}
	if _, err := New(strings.Repeat("b", 192), "bad"); err == nil {
		t.Error("New with a bad STOKEN: want error, got nil")
	}

	c, err := New(strings.Repeat("b", 192), strings.Repeat("s", 64))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer c.Close()

	if got := c.Account().BDUSS(); got != strings.Repeat("b", 192) {
		t.Errorf("BDUSS = %q", got)
	}
	if got := c.Account().STOKEN(); got != strings.Repeat("s", 64) {
		t.Errorf("STOKEN = %q", got)
	}
	if c.HTTPCore().Account != c.Account() {
		t.Error("HttpCore did not receive the account")
	}
	if c.WSCore().Account != c.Account() {
		t.Error("WsCore did not receive the account")
	}
	if c.BLCPCore().Account != c.Account() {
		t.Error("BLCPCore did not receive the account")
	}
	if c.User().Valid() {
		t.Error("User() of a fresh client must be the zero user")
	}
	if got := c.WSCore().Status(); got != enums.WsStatusClosed {
		t.Errorf("websocket status = %d, want closed", got)
	}
	if got := c.BLCPCore().Status(); got != -1 {
		t.Errorf("BLCP status = %d, want -1", got)
	}
}

func TestNewWithAccountOption(t *testing.T) {
	account, err := core.NewAccount(strings.Repeat("a", 192), "")
	if err != nil {
		t.Fatalf("NewAccount: %v", err)
	}
	// The account option overrides BDUSS and STOKEN.
	c, err := New("bogus-bduss", "bogus-stoken", WithAccount(account))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer c.Close()

	if c.Account() != account {
		t.Error("WithAccount did not take effect")
	}
}

func TestSetAccountPropagates(t *testing.T) {
	c, err := New("", "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer c.Close()

	next, err := core.NewAccount(strings.Repeat("c", 192), "")
	if err != nil {
		t.Fatalf("NewAccount: %v", err)
	}
	if err := c.SetAccount(next); err != nil {
		t.Fatalf("SetAccount: %v", err)
	}
	if c.Account() != next {
		t.Error("Account() did not change")
	}
	if c.HTTPCore().Account != next {
		t.Error("HttpCore did not change")
	}
	if c.WSCore().Account != next {
		t.Error("WsCore did not change")
	}
	if c.BLCPCore().Account != next {
		t.Error("BLCPCore did not change")
	}

	if err := c.SetAccount(nil); err == nil {
		t.Error("SetAccount(nil): want error, got nil")
	}
}

func TestProxyAndTimeoutOptions(t *testing.T) {
	proxy, err := config.NewProxyConfig("http://127.0.0.1:8888", &config.BasicAuth{Login: "u", Password: "p"})
	if err != nil {
		t.Fatalf("NewProxyConfig: %v", err)
	}
	timeout := config.DefaultTimeoutConfig()
	timeout.HTTPRead = 42 * 1e9 // 42s

	c, err := New("", "", WithProxy(proxy), WithTimeout(timeout))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer c.Close()

	if c.proxy != proxy {
		t.Error("WithProxy did not take effect")
	}
	if got := c.HTTPCore().NetCore.Proxy(); got != proxy {
		t.Error("NetCore did not receive the proxy")
	}
	if got := c.HTTPCore().NetCore.Timeout().HTTPRead; got != timeout.HTTPRead {
		t.Errorf("timeout = %v, want %v", got, timeout.HTTPRead)
	}
	if _, err := url.Parse("http://example.com"); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestDefaultGetThreadsArgs(t *testing.T) {
	args := DefaultGetThreadsArgs()
	if args.Pn != 1 || args.Rn != 30 {
		t.Errorf("pn/rn = %d/%d, want 1/30", args.Pn, args.Rn)
	}
	if args.Sort != enums.ThreadSortReply {
		t.Errorf("sort = %d, want %d (REPLY)", args.Sort, enums.ThreadSortReply)
	}
	if args.IsGood {
		t.Error("is_good = true, want false")
	}
}

func TestUserRef(t *testing.T) {
	if ref := ByUserID(42); ref.UserID != 42 || ref.Portrait != "" || ref.UserName != "" {
		t.Errorf("ByUserID = %+v", ref)
	}
	if ref := ByPortrait("tb.1.x"); ref.Portrait != "tb.1.x" || ref.IsZero() {
		t.Errorf("ByPortrait = %+v", ref)
	}
	if ref := ByUserName("name"); ref.UserName != "name" || ref.IsZero() {
		t.Errorf("ByUserName = %+v", ref)
	}
	if !(UserRef{}).IsZero() {
		t.Error("zero UserRef.IsZero() = false, want true")
	}
}

func TestIsSubset(t *testing.T) {
	tests := []struct {
		name string
		a, b enums.ReqUInfo
		want bool
	}{
		{"empty a", 0, enums.ReqUInfoBasic, true},
		{"equal", enums.ReqUInfoBasic, enums.ReqUInfoBasic, true},
		{"strict subset", enums.ReqUInfoPortrait, enums.ReqUInfoBasic, true},
		{"superset", enums.ReqUInfoAll, enums.ReqUInfoBasic, false},
		{"disjoint", enums.ReqUInfoOther, enums.ReqUInfoBasic, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isSubset(tc.a, tc.b); got != tc.want {
				t.Errorf("isSubset(%b, %b) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestGetUserInfoEmptyInput(t *testing.T) {
	c, err := New("", "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer c.Close()

	// An empty reference short-circuits before any request, like the Python
	// client, which logs a warning and returns an empty UserInfo.
	user, err := c.GetUserInfo(t.Context(), UserRef{}, enums.ReqUInfoAll)
	if err != nil {
		t.Fatalf("GetUserInfo: %v", err)
	}
	if user.UserID != 0 || user.Portrait != "" || user.UserName != "" {
		t.Errorf("user = %+v, want the zero value", user)
	}
}

func TestForumCacheShortCircuits(t *testing.T) {
	c, err := New("", "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer c.Close()

	// A cached entry must be served without touching the network.
	helper.DefaultForumInfoCache.AddForum("缓存吧", 987654321)

	fid, err := c.fetchFID(t.Context(), "缓存吧")
	if err != nil {
		t.Fatalf("fetchFID: %v", err)
	}
	if fid != 987654321 {
		t.Errorf("fid = %d, want 987654321", fid)
	}

	fname, err := c.fetchFName(t.Context(), 987654321)
	if err != nil {
		t.Fatalf("fetchFName: %v", err)
	}
	if fname != "缓存吧" {
		t.Errorf("fname = %q, want 缓存吧", fname)
	}
}

func TestForumRef(t *testing.T) {
	if ref := ByFName("吧名"); ref.FName != "吧名" || ref.FID != 0 {
		t.Errorf("ByFName = %+v", ref)
	}
	if ref := ByFID(42); ref.FID != 42 || ref.FName != "" {
		t.Errorf("ByFID = %+v", ref)
	}
}

func TestTryWebsocketDisabledByDefault(t *testing.T) {
	c, err := New("", "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer c.Close()

	if c.tryWS {
		t.Error("tryWS = true, want false by default")
	}
	// With try_ws disabled the helper must not touch the network.
	c.tryInitWebsocket(t.Context())
	if got := c.WSCore().Status(); got != enums.WsStatusClosed {
		t.Errorf("websocket status = %d, want closed", got)
	}
}
