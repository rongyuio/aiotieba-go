package getfid

import (
	"errors"
	"testing"

	"github.com/rongyuio/aiotieba/exception"
)

func TestParseBody(t *testing.T) {
	fid, err := ParseBody([]byte(`{"no":0,"error":"","data":{"fid":12345}}`))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if fid != 12345 {
		t.Errorf("fid = %d, want 12345", fid)
	}
}

func TestParseBodyLargeFID(t *testing.T) {
	// fids above 2^53 must keep full precision.
	fid, err := ParseBody([]byte(`{"no":0,"data":{"fid":9007199254740993}}`))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if fid != 9007199254740993 {
		t.Errorf("fid = %d, want 9007199254740993", fid)
	}
}

func TestParseBodyServerError(t *testing.T) {
	_, err := ParseBody([]byte(`{"no":340006,"error":"boom"}`))
	var serverErr *exception.TiebaServerError
	if !errors.As(err, &serverErr) {
		t.Fatalf("error = %v (%T), want *exception.TiebaServerError", err, err)
	}
	if serverErr.Code != 340006 || serverErr.Msg != "boom" {
		t.Errorf("server error = %+v", serverErr)
	}
}

func TestParseBodyZeroFID(t *testing.T) {
	_, err := ParseBody([]byte(`{"no":0,"data":{"fid":0}}`))
	var valueErr *exception.TiebaValueError
	if !errors.As(err, &valueErr) {
		t.Fatalf("error = %v (%T), want *exception.TiebaValueError", err, err)
	}
}

func TestRequestURL(t *testing.T) {
	u := RequestURL()
	if u.Scheme != "http" || u.Host != "tieba.baidu.com" || u.Path != "/f/commit/share/fnameShareApi" {
		t.Errorf("url = %s", u)
	}
}
