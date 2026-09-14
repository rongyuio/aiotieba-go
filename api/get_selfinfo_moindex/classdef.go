// Package getselfinfomoindex 实现 aiotieba 的 get_selfinfo_moindex API。
//
// 对应 Python 包 aiotieba.api.get_selfinfo_moindex。
package getselfinfomoindex

import (
	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/helper"
)

// UserInfoMoindex 用户信息。
type UserInfoMoindex struct {
	UserID    int64        // user_id
	Portrait  string       // portrait
	UserName  string       // 用户名
	Gender    enums.Gender // 性别
	PostNum   int64        // 发帖数
	FanNum    int64        // 粉丝数
	FollowNum int64        // 关注数
	ForumNum  int64        // 关注贴吧数
	Sign      string       // 个性签名
	IsVIP     bool         // 是否超级会员
}

// UserInfoMoindexFromJSON 对应 UserInfo_moindex.from_json。
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
