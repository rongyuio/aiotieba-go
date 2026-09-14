package core

import (
	"context"
	"io"
	"net/url"
	"strings"
	"testing"

	"github.com/rongyuio/aiotieba-go/config"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// The expected multipart body was captured from aiohttp 3.14.3 running the
// exact HttpCore.pack_proto_request code of the Python client.
const wantProtoBody = "---*_r1999\r\n" +
	"Content-Disposition: form-data; name=\"data\"; filename=\"file\"\r\n" +
	"\r\n" +
	"\x08\x02\x12\x03abc" +
	"\r\n---*_r1999--\r\n"

func TestMultipartProtoBodyMatchesAiohttp(t *testing.T) {
	got := string(multipartProtoBody([]byte("\x08\x02\x12\x03abc")))
	if got != wantProtoBody {
		t.Errorf("multipart body mismatch:\n got %q\nwant %q", got, wantProtoBody)
	}
}

func TestEncodeFormMatchesUrlencode(t *testing.T) {
	// urllib.parse.urlencode([...], doseq=True) with quote_plus.
	got := EncodeForm([]crypto.Param{
		{Key: "kw", Value: "天堂鸡汤"},
		{Key: "pn", Value: 1},
		{Key: "a b", Value: "x/y"},
		{Key: "sign", Value: "deadbeef"},
	})
	const want = "kw=%E5%A4%A9%E5%A0%82%E9%B8%A1%E6%B1%A4&pn=1&a+b=x%2Fy&sign=deadbeef"
	if got != want {
		t.Errorf("EncodeForm = %q, want %q", got, want)
	}
}

func testNetCore() *NetCore {
	return NewNetCore(nil, config.DefaultTimeoutConfig())
}

func newTestCore(t *testing.T, bduss string) *HttpCore {
	t.Helper()
	account, err := NewAccount(bduss, "")
	if err != nil {
		t.Fatalf("NewAccount: %v", err)
	}
	return NewHttpCore(account, testNetCore())
}

func TestPackProtoRequest(t *testing.T) {
	h := newTestCore(t, "")
	u, err := url.Parse("http://tiebac.baidu.com/c/f/frs/page?cmd=301001")
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}

	req, err := h.PackProtoRequest(context.Background(), u, []byte("\x08\x02\x12\x03abc"))
	if err != nil {
		t.Fatalf("PackProtoRequest: %v", err)
	}

	if req.Method != "POST" {
		t.Errorf("method = %s, want POST", req.Method)
	}
	if got := req.Header.Get("Content-Type"); got != "multipart/form-data; boundary=-*_r1999" {
		t.Errorf("content-type = %q", got)
	}
	if got := req.Header.Get("x_bd_data_type"); got != "protobuf" {
		t.Errorf("x_bd_data_type = %q", got)
	}
	if got := req.Header.Get("User-Agent"); got != "aiotieba/4.7.2" {
		t.Errorf("user-agent = %q", got)
	}
	if req.Host != "tiebac.baidu.com" {
		t.Errorf("host = %q", req.Host)
	}
	if req.URL.RawQuery != "cmd=301001" {
		t.Errorf("query = %q, want cmd=301001", req.URL.RawQuery)
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("reading body: %v", err)
	}
	if string(body) != wantProtoBody {
		t.Errorf("body mismatch:\n got %q\nwant %q", body, wantProtoBody)
	}
}

func TestPackFormRequestSignsWithAppSalt(t *testing.T) {
	h := newTestCore(t, "")
	u, _ := url.Parse("http://tiebac.baidu.com/c/c/bawu/commit")

	req, err := h.PackFormRequest(context.Background(), u, []crypto.Param{{Key: "a", Value: "1"}, {Key: "b", Value: 2}})
	if err != nil {
		t.Fatalf("PackFormRequest: %v", err)
	}
	if got := req.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
		t.Errorf("content-type = %q", got)
	}
	body, _ := io.ReadAll(req.Body)
	const want = "a=1&b=2&sign=42961b9881c2d7cb297e9498f9767789"
	if string(body) != want {
		t.Errorf("body = %q, want %q", body, want)
	}
}

func TestWebCookiesAttachedByDomain(t *testing.T) {
	h := newTestCore(t, strings.Repeat("b", 192))
	u, _ := url.Parse("https://tieba.baidu.com/f/commit/post/add")

	req, err := h.PackWebGetRequest(context.Background(), u, []crypto.Param{{Key: "kw", Value: "v"}}, nil)
	if err != nil {
		t.Fatalf("PackWebGetRequest: %v", err)
	}

	got := map[string]string{}
	for _, ck := range req.Cookies() {
		got[ck.Name] = ck.Value
	}
	if got["BDUSS"] != strings.Repeat("b", 192) {
		t.Errorf("BDUSS cookie = %q", got["BDUSS"])
	}
	if _, ok := got["STOKEN"]; !ok {
		t.Error("STOKEN cookie is missing for tieba.baidu.com")
	}
	if req.URL.RawQuery != "kw=v" {
		t.Errorf("query = %q, want kw=v", req.URL.RawQuery)
	}
}

func TestDomainMatch(t *testing.T) {
	cases := []struct {
		host   string
		domain string
		want   bool
	}{
		{"tieba.baidu.com", "baidu.com", true},
		{"baidu.com", "baidu.com", true},
		{"tiebac.baidu.com", "baidu.com", true},
		{"tieba.baidu.com", "tieba.baidu.com", true},
		{"notbaidu.com", "baidu.com", false},
		{"anything", "", true},
	}
	for _, c := range cases {
		if got := domainMatch(c.host, c.domain); got != c.want {
			t.Errorf("domainMatch(%q, %q) = %v, want %v", c.host, c.domain, got, c.want)
		}
	}
}

func TestWebFormRequestUsesWebSessionHeaders(t *testing.T) {
	h := newTestCore(t, "")
	u, _ := url.Parse("https://tieba.baidu.com/f/commit/post/add")

	req, err := h.PackWebFormRequest(context.Background(), u, []crypto.Param{{Key: "kw", Value: "v"}},
		map[string]string{"X-Requested-With": "XMLHttpRequest"})
	if err != nil {
		t.Fatalf("PackWebFormRequest: %v", err)
	}
	if got := req.Header.Get("x_bd_data_type"); got != "" {
		t.Errorf("web session must not set x_bd_data_type, got %q", got)
	}
	if got := req.Header.Get("Cache-Control"); got != "no-cache" {
		t.Errorf("cache-control = %q, want no-cache", got)
	}
	if got := req.Header.Get("X-Requested-With"); got != "XMLHttpRequest" {
		t.Errorf("extra header missing, got %q", got)
	}
	if req.Host != "tieba.baidu.com" {
		t.Errorf("host = %q, want tieba.baidu.com", req.Host)
	}
}
