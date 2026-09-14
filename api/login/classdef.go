// Package login 实现 aiotieba 的 login API。
//
// 对应 Python 包 aiotieba.api.login。
package login

import (
	"strconv"

	"github.com/rongyuio/aiotieba-go/helper"
)

// UserInfoLogin 用户信息。
type UserInfoLogin struct {
	UserID   int64  // user_id
	Portrait string // portrait
	UserName string // 用户名
}

// UserInfoLoginFromJSON 对应 UserInfo_login.from_json。
func UserInfoLoginFromJSON(data map[string]any) UserInfoLogin {
	return UserInfoLogin{
		UserID:   helper.JSONInt(data, "id"),
		Portrait: helper.JSONStr(data, "portrait"),
		UserName: helper.JSONStr(data, "name"),
	}
}

// String 对应 __str__。
func (u UserInfoLogin) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// Valid 对应 __bool__。
func (u UserInfoLogin) Valid() bool { return u.UserID != 0 }
