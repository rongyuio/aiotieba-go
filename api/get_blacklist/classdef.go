// Package getblacklist 实现 aiotieba 的 get_blacklist API。
//
// 对应 Python 包 aiotieba.api.get_blacklist。
package getblacklist

import (
	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/helper"
)

// BlacklistUser 用户信息。
type BlacklistUser struct {
	UserID      int64               // user_id
	Portrait    string              // portrait
	UserName    string              // 用户名
	NickNameNew string              // 新版昵称
	BType       enums.BlacklistType // 黑名单类型 FOLLOW禁止关注 INTERACT禁止互动 CHAT禁止私信
}

// BlacklistUserFromJSON 对应 BlacklistUser.from_json。
func BlacklistUserFromJSON(m map[string]any) BlacklistUser {
	u := BlacklistUser{
		UserID:      helper.JSONInt(m, "uid"),
		Portrait:    classdef.TrimPortrait(helper.JSONStr(m, "portrait")),
		UserName:    helper.JSONStr(m, "user_name"),
		NickNameNew: helper.JSONStr(m, "name_show"),
	}
	if perm := helper.JSONMap(m, "perm_list"); perm != nil {
		if helper.JSONBool(perm, "follow") {
			u.BType |= enums.BlacklistFollow
		}
		if helper.JSONBool(perm, "chat") {
			u.BType |= enums.BlacklistChat
		}
		if helper.JSONBool(perm, "interact") {
			u.BType |= enums.BlacklistInteract
		}
	}
	return u
}

// NickName 用户昵称。
func (u BlacklistUser) NickName() string { return u.NickNameNew }

// ShowName 显示名称。
func (u BlacklistUser) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// BlacklistUsers 新版用户黑名单列表。
type BlacklistUsers struct {
	classdef.Containers[*BlacklistUser]
	Err error // 捕获的异常
}

// BlacklistUsersFromJSON 对应 BlacklistUsers.from_json。
func BlacklistUsersFromJSON(m map[string]any) BlacklistUsers {
	var users BlacklistUsers
	for _, item := range helper.JSONSlice(m, "user_perm_list") {
		if im, ok := item.(map[string]any); ok {
			u := BlacklistUserFromJSON(im)
			users.Objs = append(users.Objs, &u)
		}
	}
	return users
}
