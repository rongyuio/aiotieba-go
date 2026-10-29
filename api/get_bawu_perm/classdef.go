// Package getbawuperm 实现 aiotieba 的 get_bawu_perm API。
//
// 对应 Python 包 aiotieba.api.get_bawu_perm。
package getbawuperm

import (
	"time"

	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/helper"
)

// UserInfoPerm 吧务权限查询的目标用户。
type UserInfoPerm struct {
	UserName    string         // 用户名
	NickNameOld string         // 旧版昵称
	Level       int64          // 吧内等级
	BawuType    enums.BawuType // 吧务类型
	BeginTime   time.Time      // 上任时间
}

// UserInfoPermFromJSON 对应 UserInfo_perm.from_json。
func UserInfoPermFromJSON(m map[string]any) UserInfoPerm {
	beginTime := helper.DefaultDatetime()
	if ts := helper.JSONInt(m, "create_time"); ts != 0 {
		beginTime = time.Unix(ts, 0)
	}
	return UserInfoPerm{
		UserName:    helper.JSONStr(m, "user_name"),
		NickNameOld: helper.JSONStr(m, "user_nickname"),
		Level:       helper.JSONInt(m, "level_id"),
		BawuType:    enums.BawuTypeFrom(helper.JSONStr(m, "role_type")),
		BeginTime:   beginTime,
	}
}

// BawuPerm 吧务已分配的权限。
type BawuPerm struct {
	User  UserInfoPerm       // 目标用户
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
	return BawuPerm{
		User:  UserInfoPermFromJSON(helper.JSONMap(m, "user")),
		Perms: perms,
	}
}
