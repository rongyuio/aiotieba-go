// Package getuserjson 实现 aiotieba 的 get_uinfo_user_json API。
//
// 对应 Python 包 aiotieba.api.get_uinfo_user_json。
package getuserjson

import (
	"strconv"

	"github.com/rongyuio/aiotieba-go/helper"
)

// UserInfoJSON 用户信息。
type UserInfoJSON struct {
	UserID   int64  // user_id
	Portrait string // portrait
	UserName string // 用户名
}

// UserInfoJSONFromJSON 对应 UserInfo_json.from_json。
//
// 该端点不回显用户名；由调用方填充。
func UserInfoJSONFromJSON(data map[string]any) UserInfoJSON {
	return UserInfoJSON{
		UserID:   helper.JSONInt(data, "id"),
		Portrait: helper.JSONStr(data, "portrait"),
	}
}

// String 对应 __str__。
func (u UserInfoJSON) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// LogName 用于在日志中记录用户信息。
func (u UserInfoJSON) LogName() string { return u.String() }

// Valid 对应 __bool__。
func (u UserInfoJSON) Valid() bool { return u.UserID != 0 }
