// Package getuserinfoweb implements the get_uinfo_getUserInfo_web API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_uinfo_getUserInfo_web.
package getuserinfoweb

import (
	"strconv"

	"github.com/rongyuio/aiotieba/helper"
)

// UserInfoGuinfoWeb is the user information returned by the web messaging
// endpoint. It mirrors
// aiotieba.api.get_uinfo_getUserInfo_web._classdef.UserInfo_guinfo_web.
type UserInfoGuinfoWeb struct {
	UserID      int64
	Portrait    string
	UserName    string
	NickNameNew string
}

// UserInfoGuinfoWebFromJSON mirrors UserInfo_guinfo_web.from_json.
//
// The name is dropped when the server echoes the uid as the user name; the
// comparison is type sensitive, mirroring Python's `!=`.
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

// NickName mirrors the nick_name property.
func (u UserInfoGuinfoWeb) NickName() string { return u.NickNameNew }

// ShowName mirrors the show_name property.
func (u UserInfoGuinfoWeb) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// String mirrors __str__.
func (u UserInfoGuinfoWeb) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// LogName mirrors the log_name property.
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

// Valid mirrors __bool__.
func (u UserInfoGuinfoWeb) Valid() bool { return u.UserID != 0 }
