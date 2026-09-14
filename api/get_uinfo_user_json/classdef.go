// Package getuserjson implements the get_uinfo_user_json API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_uinfo_user_json.
package getuserjson

import (
	"strconv"

	"github.com/rongyuio/aiotieba/helper"
)

// UserInfoJSON is the user information returned by /i/sys/user_json. It mirrors
// aiotieba.api.get_uinfo_user_json._classdef.UserInfo_json.
type UserInfoJSON struct {
	UserID   int64
	Portrait string
	UserName string
}

// UserInfoJSONFromJSON mirrors UserInfo_json.from_json.
//
// The endpoint does not echo the user name; the caller fills it in.
func UserInfoJSONFromJSON(data map[string]any) UserInfoJSON {
	return UserInfoJSON{
		UserID:   helper.JSONInt(data, "id"),
		Portrait: helper.JSONStr(data, "portrait"),
	}
}

// String mirrors __str__.
func (u UserInfoJSON) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// LogName mirrors the log_name property.
func (u UserInfoJSON) LogName() string { return u.String() }

// Valid mirrors __bool__.
func (u UserInfoJSON) Valid() bool { return u.UserID != 0 }
