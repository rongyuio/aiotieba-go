// Package tiebauid2userinfo implements the tieba_uid2user_info API of aiotieba.
//
// It mirrors the Python package aiotieba.api.tieba_uid2user_info.
package tiebauid2userinfo

import (
	"strconv"
	"strings"

	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/protobuf"
)

// UserInfoTUid is the user information looked up by tieba uid. It mirrors
// aiotieba.api.tieba_uid2user_info._classdef.UserInfo_TUid.
type UserInfoTUid struct {
	UserID      int64
	Portrait    string
	UserName    string
	NickNameNew string
	TiebaUID    int64

	Age   float64
	Sign  string
	IsGod bool
}

// UserInfoTUidFromProto mirrors UserInfo_TUid.from_proto.
func UserInfoTUidFromProto(p *protobuf.User) UserInfoTUid {
	portrait := p.GetPortrait()
	// The portrait carries a "?..." query suffix that the client strips.
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

// NickName mirrors the nick_name property.
func (u UserInfoTUid) NickName() string { return u.NickNameNew }

// ShowName mirrors the show_name property.
func (u UserInfoTUid) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// String mirrors __str__.
func (u UserInfoTUid) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// LogName mirrors the log_name property.
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

// Valid mirrors __bool__.
func (u UserInfoTUid) Valid() bool { return u.UserID != 0 }
