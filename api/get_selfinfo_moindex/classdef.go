// Package getselfinfomoindex implements the get_selfinfo_moindex API of
// aiotieba.
//
// It mirrors the Python package aiotieba.api.get_selfinfo_moindex.
package getselfinfomoindex

import (
	"github.com/rongyuio/aiotieba/api/classdef"
	"github.com/rongyuio/aiotieba/enums"
	"github.com/rongyuio/aiotieba/helper"
)

// UserInfoMoindex mirrors UserInfo_moindex.
type UserInfoMoindex struct {
	UserID    int64
	Portrait  string
	UserName  string
	Gender    enums.Gender
	PostNum   int64
	FanNum    int64
	FollowNum int64
	ForumNum  int64
	Sign      string
	IsVIP     bool
}

// UserInfoMoindexFromJSON mirrors UserInfo_moindex.from_json.
func UserInfoMoindexFromJSON(m map[string]any) UserInfoMoindex {
	u := UserInfoMoindex{
		UserID:    helper.JSONInt(m, "id"),
		Portrait:  classdef.TrimPortrait(helper.JSONStr(m, "portrait")),
		UserName:  helper.JSONStr(m, "name"),
		Gender:    enums.GenderFrom(int(helper.JSONInt(m, "user_sex"))),
		PostNum:   helper.JSONInt(m, "post_num"),
		FanNum:    helper.JSONInt(m, "fans_num"),
		FollowNum: helper.JSONInt(m, "concern_num"),
		ForumNum:  helper.JSONInt(m, "like_forum_num"),
		Sign:      helper.JSONStr(m, "intro"),
	}
	if vip := helper.JSONMap(m, "vipInfo"); vip != nil {
		u.IsVIP = helper.JSONInt(vip, "v_status") == 3
	}
	return u
}
