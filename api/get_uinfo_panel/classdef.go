// Package getuinfopanel 实现 aiotieba 的 get_uinfo_panel API。
//
// 对应 Python 包 aiotieba.api.get_uinfo_panel。
package getuinfopanel

import (
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/helper"
)

// tenThousandSuffix 是服务端用于大数计数器的中文 "万"（一万）后缀。
const tenThousandSuffix = "万"

// UserInfoPanel 用户信息。
type UserInfoPanel struct {
	Portrait    string       // portrait
	UserName    string       // 用户名
	NickNameNew string       // 新版昵称
	NickNameOld string       // 旧版昵称
	Gender      enums.Gender // 性别
	Age         float64      // 吧龄
	PostNum     int64        // 发帖数
	FanNum      int64        // 粉丝数
	IsVIP       bool         // 是否超级会员
}

// UserInfoPanelFromJSON 对应 UserInfo_panel.from_json。
func UserInfoPanelFromJSON(data map[string]any) UserInfoPanel {
	var gender enums.Gender
	switch helper.JSONStr(data, "sex") {
	case "male":
		gender = enums.GenderMale
	case "female":
		gender = enums.GenderFemale
	default:
		gender = enums.GenderUnknown
	}

	var age float64
	if tbAge := helper.JSONStr(data, "tb_age"); tbAge != "-" {
		age = helper.AnyFloat64(tbAge)
	}

	isVIP := helper.VipInfo2Bool(helper.JSONMap(data, "vipInfo"))

	return UserInfoPanel{
		Portrait:    helper.JSONStr(data, "portrait"),
		UserName:    helper.JSONStr(data, "name"),
		NickNameNew: helper.JSONStr(data, "show_nickname"),
		NickNameOld: helper.JSONStr(data, "name_show"),
		Gender:      gender,
		Age:         age,
		PostNum:     tbNum2Int(data, "post_num"),
		FanNum:      tbNum2Int(data, "followed_count"),
		IsVIP:       isVIP,
	}
}

// tbNum2Int 对应 _tbnum2int：形如 "1.2万" 的计数会变成 12000。
func tbNum2Int(data map[string]any, key string) int64 {
	raw, _ := helper.JSONRaw(data, key)
	return helper.TbNum2Int(raw, tenThousandSuffix)
}

// NickName 用户昵称。
func (u UserInfoPanel) NickName() string { return u.NickNameNew }

// ShowName 显示名称。
func (u UserInfoPanel) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// String 对应 __str__。
func (u UserInfoPanel) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	return u.Portrait
}

// LogName 用于在日志中记录用户信息。
func (u UserInfoPanel) LogName() string {
	if u.UserName != "" {
		return u.UserName
	}
	return u.NickNameNew + "/" + u.Portrait
}

// Valid 对应 __bool__，当 portrait 存在时为真。
func (u UserInfoPanel) Valid() bool { return u.Portrait != "" }
