package getfollowforums

import (
	"testing"

	"github.com/rongyuio/aiotieba/helper"
)

func TestFollowForumsFromJSON(t *testing.T) {
	m, err := helper.ParseJSONMap([]byte(`{
		"forum_list": {
			"non-gconforum": [{"id": 1, "name": "a", "level_id": 2, "cur_score": 3}],
			"gconforum": [{"id": 4, "name": "b", "level_id": 5, "cur_score": 6}]
		},
		"has_more": "1"
	}`))
	if err != nil {
		t.Fatalf("ParseJSONMap: %v", err)
	}

	got := FollowForumsFromJSON(m)
	if len(got.Objs) != 2 {
		t.Fatalf("len(Objs) = %d, want 2", len(got.Objs))
	}
	if got.Objs[0].FID != 1 || got.Objs[0].FName != "a" || got.Objs[0].Level != 2 || got.Objs[0].Exp != 3 {
		t.Errorf("Objs[0] = %+v", *got.Objs[0])
	}
	if got.Objs[1].FID != 4 || got.Objs[1].FName != "b" {
		t.Errorf("Objs[1] = %+v", *got.Objs[1])
	}
	if !got.HasMore {
		t.Error("HasMore = false, want true")
	}
}

func TestFollowForumsFromJSONEmpty(t *testing.T) {
	m, _ := helper.ParseJSONMap([]byte(`{}`))
	got := FollowForumsFromJSON(m)
	if len(got.Objs) != 0 || got.HasMore {
		t.Errorf("got %+v, want empty", got)
	}
}
