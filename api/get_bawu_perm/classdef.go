// Package getbawuperm 实现 aiotieba 的 get_bawu_perm API。
//
// 对应 Python 包 aiotieba.api.get_bawu_perm。
package getbawuperm

import (
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/helper"
)

// BawuPerm 吧务已分配的权限。
type BawuPerm struct {
	Perms enums.BawuPermType // 吧务已分配的权限
	Err   error              // 捕获的异常
}

// BawuPermFromJSON 对应 BawuPerm.from_json。
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
