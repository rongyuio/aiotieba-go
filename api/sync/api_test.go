package sync

import (
	"errors"
	"testing"

	"github.com/rongyuio/aiotieba/exception"
)

func TestParseBody(t *testing.T) {
	clientID, sampleID, err := ParseBody([]byte(`{"error_code":0,"client":{"client_id":"cid"},"wl_config":{"sample_id":"sid"}}`))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if clientID != "cid" {
		t.Errorf("client id = %q, want cid", clientID)
	}
	if sampleID != "sid" {
		t.Errorf("sample id = %q, want sid", sampleID)
	}
}

func TestParseBodyServerError(t *testing.T) {
	_, _, err := ParseBody([]byte(`{"error_code":340006,"error_msg":"boom"}`))
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
	if u.Scheme != "https" || u.Host != "tiebac.baidu.com" || u.Path != "/c/s/sync" {
		t.Errorf("url = %s", u)
	}
}
