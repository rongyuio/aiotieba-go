// Package getselfinfoinitnickname 实现 aiotieba 的 get_selfinfo_initNickname API。
//
// 对应 Python 包 aiotieba.api.get_selfinfo_initNickname。
package getselfinfoinitnickname

import "github.com/rongyuio/aiotieba-go/helper"

// UserInfoSelfinit 用户信息。
type UserInfoSelfinit struct {
	UserName    string // 用户名
	NickNameOld string // 旧版昵称
	TiebaUID    int64  // 用户个人主页uid
}

// UserInfoSelfinitFromJSON 对应 UserInfo_selfinit.from_json。
func UserInfoSelfinitFromJSON(m map[string]any) UserInfoSelfinit {
	return UserInfoSelfinit{
		UserName:    helper.JSONStr(m, "user_name"),
		NickNameOld: helper.JSONStr(m, "name_show"),
		TiebaUID:    helper.JSONInt(m, "tieba_uid"),
	}
}

// NickName 用户昵称。
func (u UserInfoSelfinit) NickName() string { return u.NickNameOld }
