package getblocks

import (
	"testing"

	"github.com/rongyuio/aiotieba/helper"
)

func TestBlocksFromJSON(t *testing.T) {
	m, err := helper.ParseJSONMap([]byte(`{
		"data": {
			"content": "<li><a attr-uid=\"1\" attr-un=\"u1\" attr-nn=\"n1\" attr-blockday=\"3\"></a></li><li><a attr-uid=\"2\" attr-un=\"u2\" attr-nn=\"n2\" attr-blockday=\"0\"></a></li>",
			"page": {"size": 10, "pn": 1, "total_page": 3, "total_count": 25, "have_next": "1"}
		}
	}`))
	if err != nil {
		t.Fatalf("ParseJSONMap: %v", err)
	}

	blocks, err := BlocksFromJSON(m)
	if err != nil {
		t.Fatalf("BlocksFromJSON: %v", err)
	}
	if len(blocks.Objs) != 2 {
		t.Fatalf("len(Objs) = %d, want 2", len(blocks.Objs))
	}
	if blocks.Objs[0].UserID != 1 || blocks.Objs[0].UserName != "u1" || blocks.Objs[0].NickNameOld != "n1" || blocks.Objs[0].Day != 3 {
		t.Errorf("Objs[0] = %+v", *blocks.Objs[0])
	}
	if blocks.Page.PageSize != 10 || blocks.Page.CurrentPage != 1 || blocks.Page.TotalPage != 3 || blocks.Page.TotalCount != 25 {
		t.Errorf("Page = %+v", blocks.Page)
	}
	if !blocks.HasMore() {
		t.Error("HasMore = false, want true")
	}
}
