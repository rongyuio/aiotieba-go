// Package getuserinfoapp implements the get_uinfo_getuserinfo_app API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_uinfo_getuserinfo_app.
package getuserinfoapp

import (
	"strconv"
	"strings"

	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/protobuf"
)

// UserInfoGuinfoApp is the user information returned by the app endpoint. It
// mirrors aiotieba.api.get_uinfo_getuserinfo_app._classdef.UserInfo_guinfo_app.
type UserInfoGuinfoApp struct {
	UserID      int64
	Portrait    string
	UserName    string
	NickNameOld string
	Gender      enums.Gender
	IsVIP       bool
	IsGod       bool
}

// UserInfoGuinfoAppFromProto mirrors UserInfo_guinfo_app.from_proto.
func UserInfoGuinfoAppFromProto(p *protobuf.User) UserInfoGuinfoApp {
	portrait := p.GetPortrait()
	// The portrait carries a "?..." query suffix that the client strips.
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

// NickName returns the user nickname, mirroring the nick_name property.
func (u UserInfoGuinfoApp) NickName() string { return u.NickNameOld }

// String mirrors __str__.
func (u UserInfoGuinfoApp) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// LogName mirrors the log_name property.
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

// Valid mirrors __bool__.
func (u UserInfoGuinfoApp) Valid() bool { return u.UserID != 0 }
