// Package getuserinfousercard implements the get_uinfo_userCard API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_uinfo_userCard.
package getuserinfousercard

import (
	"strings"

	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/helper"
)

// UserInfoUC is the user information returned by the PC user card. It mirrors
// aiotieba.api.get_uinfo_userCard._classdef.UserInfo_uc.
type UserInfoUC struct {
	Portrait    string
	NickNameNew string
	TiebaUID    int64

	Gender    enums.Gender
	Age       float64
	AgreeNum  int64
	FanNum    int64
	FollowNum int64

	Sign string
	IP   string
}

// UserInfoUCFromJSON mirrors UserInfo_uc.from_json.
func UserInfoUCFromJSON(data map[string]any) UserInfoUC {
	portrait := helper.JSONStr(data, "portrait")
	// The portrait carries a "?..." query suffix that the client strips.
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

// NickName mirrors the nick_name property.
func (u UserInfoUC) NickName() string { return u.NickNameNew }

// ShowName mirrors the show_name property.
func (u UserInfoUC) ShowName() string { return u.NickNameNew }

// String mirrors __str__.
func (u UserInfoUC) String() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.Portrait
}

// LogName mirrors the log_name property.
func (u UserInfoUC) LogName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.Portrait
}

// Valid mirrors __bool__.
func (u UserInfoUC) Valid() bool { return u.Portrait != "" }
