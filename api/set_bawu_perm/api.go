// Package setbawuperm implements the set_bawu_perm API of aiotieba.
//
// It mirrors the Python package aiotieba.api.set_bawu_perm.
package setbawuperm

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// permSetting is one element of the `perm_setting` JSON array.
//
// The field order matters because the JSON text itself is part of the signed
// payload.
type permSetting struct {
	Switch int `json:"switch"`
	Perm   int `json:"perm"`
}

// perm2id maps a permission to the id the server expects, mirroring the
// perm2id table of the Python module.
var perm2id = []struct {
	perm enums.BawuPermType
	id   int
}{
	{enums.BawuPermUnblock, 4},
	{enums.BawuPermUnblockAppeal, 5},
	{enums.BawuPermRecover, 3},
	{enums.BawuPermRecoverAppeal, 2},
}

// PackPermSettings mirrors pack_perm_settings.
func PackPermSettings(perms enums.BawuPermType) []permSetting {
	settings := make([]permSetting, 0, len(perm2id))
	for _, entry := range perm2id {
		settings = append(settings, permSetting{
			Switch: helper.BoolInt(perms&entry.perm != 0),
			Perm:   entry.id,
		})
	}
	return settings
}

// ParseBody mirrors parse_body.
func ParseBody(body []byte) error {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return err
	}
	if code := helper.JSONInt(res, "no"); code != 0 {
		return &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error")}
	}
	return nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/mo/q/setAuthToolPerm"}
}

// Request mirrors request.
func Request(ctx context.Context, httpCore *core.HttpCore, fid int64, portrait string, perms enums.BawuPermType) error {
	data := []crypto.Param{
		{Key: "forum_id", Value: fid},
		{Key: "auth_user_portrait", Value: portrait},
		{Key: "perm_setting", Value: helper.PackJSON(PackPermSettings(perms))},
	}

	resp, err := httpCore.WebForm(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return err
	}
	return ParseBody(resp.Body())
}
