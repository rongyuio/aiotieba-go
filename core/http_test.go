package core

import (
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

func TestAppProtoRequest(t *testing.T) {
	h := newTestCore(t, "")
	req := h.AppProto([]byte("\x08\x02\x12\x03abc"))

	if got := req.Header.Get("Content-Type"); got != "multipart/form-data; boundary=-*_r1999" {
		t.Errorf("content-type = %q", got)
	}
	if got := h.appProto.Header.Get("x_bd_data_type"); got != "protobuf" {
		t.Errorf("x_bd_data_type = %q", got)
	}
	body, ok := req.Body.([]byte)
	if !ok {
		t.Fatalf("body type = %T, want []byte", req.Body)
	}
	if string(body) != wantProtoBody {
		t.Errorf("body mismatch:\n got %q\nwant %q", body, wantProtoBody)
	}
}

func TestAppFormSignsWithAppSalt(t *testing.T) {
	h := newTestCore(t, "")
	req := h.AppForm([]crypto.Param{{Key: "a", Value: "1"}, {Key: "b", Value: 2}})

	if got := req.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
		t.Errorf("content-type = %q", got)
	}
	body, ok := req.Body.(string)
	if !ok {
		t.Fatalf("body type = %T, want string", req.Body)
	}
	const want = "a=1&b=2&sign=42961b9881c2d7cb297e9498f9767789"
	if body != want {
		t.Errorf("body = %q, want %q", body, want)
	}
}

func TestWebCookiesAttached(t *testing.T) {
	h := newTestCore(t, strings.Repeat("b", 192))

	got := map[string]string{}
	for _, ck := range h.web.Cookies {
		got[ck.Name] = ck.Value
	}
	if got["BDUSS"] != strings.Repeat("b", 192) {
		t.Errorf("BDUSS cookie = %q", got["BDUSS"])
	}
	if _, ok := got["STOKEN"]; !ok {
		t.Error("STOKEN cookie is missing")
	}
}

func TestWebGetQueryParams(t *testing.T) {
	h := newTestCore(t, "")
	req := h.WebGet([]crypto.Param{{Key: "kw", Value: "v"}}, map[string]string{"Referer": "tieba.baidu.com"})

	if got := req.QueryParam.Get("kw"); got != "v" {
		t.Errorf("query kw = %q, want v", got)
	}
	if got := req.Header.Get("Referer"); got != "tieba.baidu.com" {
		t.Errorf("referer = %q", got)
	}
}

func TestWebFormUsesWebSessionHeaders(t *testing.T) {
	h := newTestCore(t, "")
	req := h.WebForm([]crypto.Param{{Key: "kw", Value: "v"}})

	if got := h.web.Header.Get("Cache-Control"); got != "no-cache" {
		t.Errorf("cache-control = %q, want no-cache", got)
	}
	if got := req.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
		t.Errorf("content-type = %q", got)
	}
	body, ok := req.Body.(string)
	if !ok {
		t.Fatalf("body type = %T, want string", req.Body)
	}
	if body != "kw=v" {
		t.Errorf("body = %q, want kw=v", body)
	}
}
