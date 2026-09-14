// Package getselfinfoinitnickname implements the get_selfinfo_initNickname API
// of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_selfinfo_initNickname.
package getselfinfoinitnickname

import "github.com/rongyuio/aiotieba/helper"

// UserInfoSelfinit mirrors UserInfo_selfinit.
type UserInfoSelfinit struct {
	UserName    string
	NickNameOld string
	TiebaUID    int64
}

// UserInfoSelfinitFromJSON mirrors UserInfo_selfinit.from_json.
func UserInfoSelfinitFromJSON(m map[string]any) UserInfoSelfinit {
	return UserInfoSelfinit{
		UserName:    helper.JSONStr(m, "user_name"),
		NickNameOld: helper.JSONStr(m, "name_show"),
		TiebaUID:    helper.JSONInt(m, "tieba_uid"),
	}
}

// NickName mirrors the nick_name property.
func (u UserInfoSelfinit) NickName() string { return u.NickNameOld }
