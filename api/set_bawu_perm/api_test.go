package setbawuperm

import (
	"errors"
	"testing"

	"github.com/rongyuio/aiotieba/enums"
	"github.com/rongyuio/aiotieba/exception"
	"github.com/rongyuio/aiotieba/helper"
)

func TestPackPermSettingsNone(t *testing.T) {
	got := helper.PackJSON(PackPermSettings(enums.BawuPermNull))
	// The order of the perm2id table and the switch/perm field order are both
	// observable because the JSON text is part of the signed payload.
	want := `[{"switch":0,"perm":4},{"switch":0,"perm":5},{"switch":0,"perm":3},{"switch":0,"perm":2}]`
	if got != want {
		t.Errorf("perm settings = %s, want %s", got, want)
	}
}

func TestPackPermSettingsAll(t *testing.T) {
	got := helper.PackJSON(PackPermSettings(enums.BawuPermAll))
	want := `[{"switch":1,"perm":4},{"switch":1,"perm":5},{"switch":1,"perm":3},{"switch":1,"perm":2}]`
	if got != want {
		t.Errorf("perm settings = %s, want %s", got, want)
	}
}

func TestPackPermSettingsPartial(t *testing.T) {
	perms := enums.BawuPermUnblock | enums.BawuPermRecoverAppeal
	got := helper.PackJSON(PackPermSettings(perms))
	want := `[{"switch":1,"perm":4},{"switch":0,"perm":5},{"switch":0,"perm":3},{"switch":1,"perm":2}]`
	if got != want {
		t.Errorf("perm settings = %s, want %s", got, want)
	}
}

func TestParseBodyServerError(t *testing.T) {
	err := ParseBody([]byte(`{"no":340006,"error":"boom"}`))
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
	if u.Host != "tieba.baidu.com" || u.Path != "/mo/q/setAuthToolPerm" {
		t.Errorf("url = %s", u)
	}
}
