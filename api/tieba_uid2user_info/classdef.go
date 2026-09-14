// Package tiebauid2userinfo 实现 aiotieba 的 tieba_uid2user_info API。
//
// 对应 Python 包 aiotieba.api.tieba_uid2user_info。
package tiebauid2userinfo

import (
	"strconv"
	"strings"

	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/protobuf"
)

// UserInfoTUid 用户信息。
type UserInfoTUid struct {
	UserID      int64  // user_id
	Portrait    string // portrait
	UserName    string // 用户名
	NickNameNew string // 新版昵称
	TiebaUID    int64  // 用户个人主页uid

	Age   float64 // 吧龄
	Sign  string  // 个性签名
	IsGod bool    // 是否大神
}

// UserInfoTUidFromProto 对应 UserInfo_TUid.from_proto。
func UserInfoTUidFromProto(p *protobuf.User) UserInfoTUid {
	portrait := p.GetPortrait()
	// portrait 携带 "?..." 查询后缀，客户端会将其去除。
	if strings.Contains(portrait, "?") && len(portrait) > 13 {
		portrait = portrait[:len(portrait)-13]
	}

	return UserInfoTUid{
		UserID:      p.GetId(),
		Portrait:    portrait,
		UserName:    p.GetName(),
		NickNameNew: p.GetNameShow(),
		TiebaUID:    helper.AnyInt64(p.GetTiebaUid()),
		Age:         helper.AnyFloat64(p.GetTbAge()),
		Sign:        p.GetIntro(),
		IsGod:       p.GetNewGodData().GetStatus() != 0,
	}
}

// NickName 用户昵称。
func (u UserInfoTUid) NickName() string { return u.NickNameNew }

// ShowName 显示名称。
func (u UserInfoTUid) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// String 对应 __str__。
func (u UserInfoTUid) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// LogName 用于在日志中记录用户信息。
func (u UserInfoTUid) LogName() string {
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
func (u UserInfoTUid) Valid() bool { return u.UserID != 0 }
