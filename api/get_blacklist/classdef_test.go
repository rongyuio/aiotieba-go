package getblacklist

import (
	"testing"

	"github.com/rongyuio/aiotieba/enums"
	"github.com/rongyuio/aiotieba/helper"
)

func TestBlacklistUserFromJSON(t *testing.T) {
	m, err := helper.ParseJSONMap([]byte(`{
		"uid": "123",
		"portrait": "tb.1.abc?t=1234567890",
		"user_name": "u",
		"name_show": "n",
		"perm_list": {"follow": 1, "chat": 0, "interact": 1}
	}`))
	if err != nil {
		t.Fatalf("ParseJSONMap: %v", err)
	}

	got := BlacklistUserFromJSON(m)
	if got.UserID != 123 {
		t.Errorf("UserID = %d, want 123", got.UserID)
	}
	if got.Portrait != "tb.1.abc" {
		t.Errorf("Portrait = %q, want tb.1.abc", got.Portrait)
	}
	if got.BType != enums.BlacklistFollow|enums.BlacklistInteract {
		t.Errorf("BType = %v, want FOLLOW|INTERACT", got.BType)
	}
	if got.ShowName() != "n" {
		t.Errorf("ShowName = %q, want n", got.ShowName())
	}
}
