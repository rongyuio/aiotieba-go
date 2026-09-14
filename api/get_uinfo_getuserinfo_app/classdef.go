// Package getuserinfoapp 实现 aiotieba 的 get_uinfo_getuserinfo_app API。
//
// 对应 Python 包 aiotieba.api.get_uinfo_getuserinfo_app。
package getuserinfoapp

import (
	"strconv"
	"strings"

	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/protobuf"
)

// UserInfoGuinfoApp 用户信息。
type UserInfoGuinfoApp struct {
	UserID      int64        // user_id
	Portrait    string       // portrait
	UserName    string       // 用户名
	NickNameOld string       // 旧版昵称
	Gender      enums.Gender // 性别
	IsVIP       bool         // 是否超级会员
	IsGod       bool         // 是否大神
}

// UserInfoGuinfoAppFromProto 对应 UserInfo_guinfo_app.from_proto。
func UserInfoGuinfoAppFromProto(p *protobuf.User) UserInfoGuinfoApp {
	portrait := p.GetPortrait()
	// portrait 携带 "?..." 查询后缀，客户端会将其去除。
	if strings.Contains(portrait, "?") && len(portrait) > 13 {
		portrait = portrait[:len(portrait)-13]
	}
	return UserInfoGuinfoApp{
		UserID:      p.GetId(),
		Portrait:    portrait,
		UserName:    p.GetName(),
		NickNameOld: p.GetNameShow(),
		Gender:      enums.GenderFrom(int(p.GetSex())),
		IsVIP:       p.GetVipInfo().GetVStatus() != 0,
		IsGod:       p.GetNewGodData().GetStatus() != 0,
	}
}

// NickName 用户昵称。
func (u UserInfoGuinfoApp) NickName() string { return u.NickNameOld }

// String 对应 __str__。
func (u UserInfoGuinfoApp) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// LogName 用于在日志中记录用户信息。
func (u UserInfoGuinfoApp) LogName() string {
	switch {
	case u.UserName != "":
		return u.UserName
	case u.Portrait != "":
		return u.NickNameOld + "/" + u.Portrait
	default:
		return strconv.FormatInt(u.UserID, 10)
	}
}

// Valid 对应 __bool__。
func (u UserInfoGuinfoApp) Valid() bool { return u.UserID != 0 }
