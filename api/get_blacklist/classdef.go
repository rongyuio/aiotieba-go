// Package getblacklist implements the get_blacklist API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_blacklist.
package getblacklist

import (
	"github.com/rongyuio/aiotieba/api/classdef"
	"github.com/rongyuio/aiotieba/enums"
	"github.com/rongyuio/aiotieba/helper"
)

// BlacklistUser mirrors BlacklistUser.
type BlacklistUser struct {
	UserID      int64
	Portrait    string
	UserName    string
	NickNameNew string
	BType       enums.BlacklistType
}

// BlacklistUserFromJSON mirrors BlacklistUser.from_json.
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

// NickName mirrors the nick_name property.
func (u BlacklistUser) NickName() string { return u.NickNameNew }

// ShowName mirrors the show_name property.
func (u BlacklistUser) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// BlacklistUsers mirrors BlacklistUsers.
type BlacklistUsers struct {
	classdef.Containers[*BlacklistUser]
	Err error
}

// BlacklistUsersFromJSON mirrors BlacklistUsers.from_json.
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
