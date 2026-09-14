// Package getuserforuminfo implements the get_user_forum_info API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_user_forum_info.
package getuserforuminfo

import (
	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper"
)

// UserInfoUF mirrors UserInfo_uf.
type UserInfoUF struct {
	UserID   int64
	Portrait string
	ShowName string
	IsLike   bool
}

// UserInfoUFFromJSON mirrors UserInfo_uf.from_json.
func UserInfoUFFromJSON(m map[string]any) UserInfoUF {
	return UserInfoUF{
		ShowName: helper.JSONStr(m, "name"),
		UserID:   helper.JSONInt(m, "id"),
		Portrait: classdef.TrimPortrait(helper.JSONStr(m, "portrait")),
		IsLike:   helper.JSONBool(m, "is_like"),
	}
}

// UserForumInfo mirrors UserForumInfo.
type UserForumInfo struct {
	User              UserInfoUF
	FName             string
	SmallAvatar       string
	IsFollow          bool
	FollowDays        int64
	SignDays          int64
	ThreadNum         int64
	DayPostNum        int64
	MemberRank        int64
	DaySignRank       int64
	Level             int64
	LevelName         string
	Exp               int64
	LevelupExp        int64
	RoleName          string
	Identify          string
	HighLightSignDays int64
	Err               error
}

// UserForumInfoFromJSON mirrors UserForumInfo.from_json.
func UserForumInfoFromJSON(data map[string]any) UserForumInfo {
	user := UserInfoUFFromJSON(helper.JSONMap(data, "user_info"))
	uf := helper.JSONMap(data, "user_forum_info")
	forum := helper.JSONMap(data, "forum_info")

	return UserForumInfo{
		User:              user,
		IsFollow:          helper.JSONBool(uf, "is_follow"),
		FollowDays:        helper.JSONInt(uf, "follow_days"),
		SignDays:          helper.JSONInt(uf, "sign_days"),
		ThreadNum:         helper.JSONInt(uf, "thread_num"),
		DayPostNum:        helper.JSONInt(uf, "day_post_num"),
		MemberRank:        helper.JSONInt(uf, "member_no"),
		DaySignRank:       helper.JSONInt(uf, "day_sign_no"),
		Level:             helper.JSONInt(uf, "level_id"),
		LevelName:         helper.JSONStr(uf, "level_name"),
		Exp:               helper.JSONInt(uf, "cur_score"),
		LevelupExp:        helper.JSONInt(uf, "levelup_score"),
		RoleName:          helper.JSONStr(uf, "role_name"),
		Identify:          helper.JSONStr(uf, "identify"),
		HighLightSignDays: helper.JSONInt(uf, "high_light_sign_days"),
		FName:             helper.JSONStr(forum, "forum_name"),
		SmallAvatar:       helper.JSONStr(forum, "forum_avatar"),
	}
}
