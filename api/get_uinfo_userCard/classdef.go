// Package getuserinfousercard 实现 aiotieba 的 get_uinfo_userCard API。
//
// 对应 Python 包 aiotieba.api.get_uinfo_userCard。
package getuserinfousercard

import (
	"strings"

	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/helper"
)

// UserInfoUC 用户信息。
type UserInfoUC struct {
	Portrait    string // portrait
	NickNameNew string // 新版昵称
	TiebaUID    int64  // 用户个人主页uid

	Gender    enums.Gender // 性别
	Age       float64      // 吧龄 以年为单位
	AgreeNum  int64        // 获赞数
	FanNum    int64        // 粉丝数
	FollowNum int64        // 关注数

	Sign string // 个性签名
	IP   string // ip归属地
}

// UserInfoUCFromJSON 对应 UserInfo_uc.from_json。
func UserInfoUCFromJSON(data map[string]any) UserInfoUC {
	portrait := helper.JSONStr(data, "portrait")
	// portrait 携带 "?..." 查询后缀，客户端会将其去除。
	if strings.Contains(portrait, "?") && len(portrait) > 13 {
		portrait = portrait[:len(portrait)-13]
	}

	var age float64
	if raw, ok := helper.JSONRaw(data, "tb_age"); ok {
		age = helper.AnyFloat64(raw)
	}

	return UserInfoUC{
		Portrait:    portrait,
		NickNameNew: helper.JSONStr(data, "name_show"),
		TiebaUID:    helper.JSONInt(data, "tieba_uid"),
		Gender:      enums.GenderFrom(int(helper.JSONInt(data, "sex"))),
		Age:         age,
		AgreeNum:    helper.JSONInt(data, "total_agree_num"),
		FanNum:      helper.JSONInt(data, "fans_num"),
		FollowNum:   helper.JSONInt(data, "concern_num"),
		Sign:        helper.JSONStr(data, "intro"),
		IP:          helper.JSONStr(data, "ip_address"),
	}
}

// NickName 用户昵称。
func (u UserInfoUC) NickName() string { return u.NickNameNew }

// ShowName 显示名称。
func (u UserInfoUC) ShowName() string { return u.NickNameNew }

// String 对应 __str__。
func (u UserInfoUC) String() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.Portrait
}

// LogName 用于在日志中记录用户信息。
func (u UserInfoUC) LogName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.Portrait
}

// Valid 对应 __bool__。
func (u UserInfoUC) Valid() bool { return u.Portrait != "" }
