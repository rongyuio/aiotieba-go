// Package getuserforuminfo 实现 aiotieba 的 get_user_forum_info API。
//
// 对应 Python 包 aiotieba.api.get_user_forum_info。
package getuserforuminfo

import (
	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper"
)

// UserInfoUF 用户信息。
type UserInfoUF struct {
	UserID   int64  // user_id
	Portrait string // portrait
	ShowName string // 显示名称
	IsLike   bool   // 是否已关注该用户
}

// UserInfoUFFromJSON 对应 UserInfo_uf.from_json。
func UserInfoUFFromJSON(m map[string]any) UserInfoUF {
	return UserInfoUF{
		ShowName: helper.JSONStr(m, "name"),
		UserID:   helper.JSONInt(m, "id"),
		Portrait: classdef.TrimPortrait(helper.JSONStr(m, "portrait")),
		IsLike:   helper.JSONBool(m, "is_like"),
	}
}

// UserForumInfo 用户在吧内的信息。
type UserForumInfo struct {
	User              UserInfoUF // 用户信息
	FName             string     // 贴吧名
	SmallAvatar       string     // 吧头像(小)
	IsFollow          bool       // 是否已关注该吧
	FollowDays        int64      // 关注天数
	SignDays          int64      // 签到天数
	ThreadNum         int64      // 本吧发帖数
	DayPostNum        int64      // 今日发帖数
	MemberRank        int64      // 吧内排名
	DaySignRank       int64      // 今日签到排名
	Level             int64      // 等级
	LevelName         string     // 本吧头衔名称
	Exp               int64      // 当前经验
	LevelupExp        int64      // 升级经验
	RoleName          string     // 吧务名称
	Identify          string     // 身份标识
	HighLightSignDays int64      // 连续签到天数
	Err               error      // 捕获的异常
}

// UserForumInfoFromJSON 对应 UserForumInfo.from_json。
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
