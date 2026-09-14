package getbawuperm

import (
	"testing"

	"github.com/rongyuio/aiotieba/enums"
	"github.com/rongyuio/aiotieba/helper"
)

func TestBawuPermFromJSON(t *testing.T) {
	m, err := helper.ParseJSONMap([]byte(`{
		"perm_setting": {
			"category_user": [
				{"switch": true, "perm": 4},
				{"switch": false, "perm": 5},
				{"switch": true, "perm": 2}
			],
			"category_thread": [
				{"switch": true, "perm": 3}
			]
		}
	}`))
	if err != nil {
		t.Fatalf("ParseJSONMap: %v", err)
	}

	got := BawuPermFromJSON(m)
	// perm 4 -> idx 2 -> UNBLOCK, perm 2 -> idx 0 -> RECOVER_APPEAL,
	// perm 3 -> idx 1 -> RECOVER. The switched-off perm 5 is ignored.
	want := enums.BawuPermUnblock | enums.BawuPermRecoverAppeal | enums.BawuPermRecover
	if got.Perms != want {
		t.Errorf("Perms = %v, want %v", got.Perms, want)
	}
}
