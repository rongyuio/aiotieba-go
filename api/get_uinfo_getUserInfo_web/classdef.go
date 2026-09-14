// Package getuserinfoweb 实现 aiotieba 的 get_uinfo_getUserInfo_web API。
//
// 对应 Python 包 aiotieba.api.get_uinfo_getUserInfo_web。
package getuserinfoweb

import (
	"strconv"

	"github.com/rongyuio/aiotieba-go/helper"
)

// UserInfoGuinfoWeb 用户信息。
type UserInfoGuinfoWeb struct {
	UserID      int64  // user_id
	Portrait    string // portrait
	UserName    string // 用户名
	NickNameNew string // 新版昵称
}

// UserInfoGuinfoWebFromJSON 对应 UserInfo_guinfo_web.from_json。
//
// 当服务端把 uid 回显为用户名时会丢弃该名字；比较是类型敏感的，对应 Python 的 `!=`。
func UserInfoGuinfoWebFromJSON(data map[string]any) UserInfoGuinfoWeb {
	rawUID, hasUID := helper.JSONRaw(data, "uid")
	userName := helper.JSONStr(data, "uname")
	if hasUID && helper.JSONScalarEqual(userName, rawUID) {
		userName = ""
	}

	return UserInfoGuinfoWeb{
		UserID:      helper.JSONInt(data, "uid"),
		Portrait:    helper.JSONStr(data, "portrait"),
		UserName:    userName,
		NickNameNew: helper.JSONStr(data, "show_nickname"),
	}
}

// NickName 用户昵称。
func (u UserInfoGuinfoWeb) NickName() string { return u.NickNameNew }

// ShowName 显示名称。
func (u UserInfoGuinfoWeb) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// String 对应 __str__。
func (u UserInfoGuinfoWeb) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// LogName 用于在日志中记录用户信息。
func (u UserInfoGuinfoWeb) LogName() string {
	switch {
	case u.UserName != "":
		return u.UserName
	case u.Portrait != "":
		return u.NickNameNew + "/" + u.Portrait
	default:
		return strconv.FormatInt(u.UserID, 10)
	}
}

// Valid 对应 __bool__。
func (u UserInfoGuinfoWeb) Valid() bool { return u.UserID != 0 }
