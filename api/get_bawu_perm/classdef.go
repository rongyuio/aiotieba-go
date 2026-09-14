// Package getbawuperm implements the get_bawu_perm API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_bawu_perm.
package getbawuperm

import (
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/helper"
)

// BawuPerm mirrors BawuPerm.
type BawuPerm struct {
	Perms enums.BawuPermType
	Err   error
}

// BawuPermFromJSON mirrors BawuPerm.from_json.
func BawuPermFromJSON(m map[string]any) BawuPerm {
	var perms enums.BawuPermType
	permSetting := helper.JSONMap(m, "perm_setting")
	for _, cate := range []string{"category_user", "category_thread"} {
		for _, item := range helper.JSONSlice(permSetting, cate) {
			im, ok := item.(map[string]any)
			if !ok || !helper.JSONBool(im, "switch") {
				continue
			}
			switch helper.JSONInt(im, "perm") - 2 {
			case 0:
				perms |= enums.BawuPermRecoverAppeal
			case 1:
				perms |= enums.BawuPermRecover
			case 2:
				perms |= enums.BawuPermUnblock
			case 3:
				perms |= enums.BawuPermUnblockAppeal
			}
		}
	}
	return BawuPerm{Perms: perms}
}
