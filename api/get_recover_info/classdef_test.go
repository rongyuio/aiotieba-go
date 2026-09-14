package getrecoverinfo

import (
	"testing"

	"github.com/rongyuio/aiotieba-go/helper"
)

func TestContentsRIFromJSON(t *testing.T) {
	m, err := helper.ParseJSONMap([]byte(`{
		"content_detail": [
			{"type": 1, "value": "hello"},
			{"type": 3},
			{"type": 99, "x": "y"}
		],
		"all_pics": [
			{"url": "https://x/aabbccddeeff00112233445566778899.png", "width": 100, "height": 50}
		]
	}`))
	if err != nil {
		t.Fatalf("ParseJSONMap: %v", err)
	}

	c := ContentsRIFromJSON(m)
	if len(c.Texts) != 1 || c.Text() != "hello" {
		t.Errorf("Text = %q, want hello", c.Text())
	}
	if len(c.Imgs) != 1 {
		t.Fatalf("len(Imgs) = %d, want 1", len(c.Imgs))
	}
	if c.Imgs[0].Hash != "aabbccddeeff00112233445566778899" {
		t.Errorf("Imgs[0].Hash = %q", c.Imgs[0].Hash)
	}
	if c.Imgs[0].ShowWidth != 100 || c.Imgs[0].ShowHeight != 50 {
		t.Errorf("Imgs[0] = %+v", c.Imgs[0])
	}
	// text + unknown + image = 3 fragments
	if len(c.Objs) != 3 {
		t.Errorf("len(Objs) = %d, want 3", len(c.Objs))
	}
}
