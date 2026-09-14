// Package login implements the login API of aiotieba.
//
// It mirrors the Python package aiotieba.api.login.
package login

import (
	"strconv"

	"github.com/rongyuio/aiotieba/helper"
)

// UserInfoLogin is the information of the logged in account. It mirrors
// aiotieba.api.login._classdef.UserInfo_login.
type UserInfoLogin struct {
	UserID   int64
	Portrait string
	UserName string
}

// UserInfoLoginFromJSON mirrors UserInfo_login.from_json.
func UserInfoLoginFromJSON(data map[string]any) UserInfoLogin {
	return UserInfoLogin{
		UserID:   helper.JSONInt(data, "id"),
		Portrait: helper.JSONStr(data, "portrait"),
		UserName: helper.JSONStr(data, "name"),
	}
}

// String mirrors __str__.
func (u UserInfoLogin) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// Valid mirrors __bool__.
func (u UserInfoLogin) Valid() bool { return u.UserID != 0 }
