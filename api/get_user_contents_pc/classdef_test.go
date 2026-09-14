package getusercontentpc

import (
	"testing"

	"github.com/rongyuio/aiotieba-go/helper"
)

func TestContentsPcupFromJSON(t *testing.T) {
	m, err := helper.ParseJSONMap([]byte(`{
		"content": [
			{"type": 0, "text": "hello"},
			{"type": 1, "text": "http://x", "link": "http://x", "title": "t"},
			{"type": 10, "voice_md5": "abc", "during_time": 1500},
			{"type": 7, "x": 1}
		]
	}`))
	if err != nil {
		t.Fatalf("ParseJSONMap: %v", err)
	}

	c := ContentsPcupFromJSON(m)
	if c.Text() != "hellohttp://x" {
		t.Errorf("Text = %q, want hellohttp://x", c.Text())
	}
	if len(c.Links) != 1 {
		t.Errorf("len(Links) = %d, want 1", len(c.Links))
	}
	if c.Voice.MD5 != "abc" || c.Voice.Duration != 1.5 {
		t.Errorf("Voice = %+v", c.Voice)
	}
	// text + link + unknown = 3 fragments (voice is not appended to Objs)
	if len(c.Objs) != 3 {
		t.Errorf("len(Objs) = %d, want 3", len(c.Objs))
	}
}
