// Package getselffollowforums 实现 aiotieba 的 get_self_follow_forums API。
//
// 对应 Python 包 aiotieba.api.get_self_follow_forums。
package getselffollowforums

import (
	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper"
)

// SelfFollowForum 吧基本信息。
type SelfFollowForum struct {
	FID      int64  // 贴吧id
	FName    string // 贴吧名
	Level    int64  // 用户等级
	IsSigned bool   // 是否已签到
}

// SelfFollowForumFromJSON 对应 SelfFollowForum.from_json。
func SelfFollowForumFromJSON(m map[string]any) SelfFollowForum {
	return SelfFollowForum{
		FID:      helper.JSONInt(m, "forum_id"),
		FName:    helper.JSONStr(m, "forum_name"),
		Level:    helper.JSONInt(m, "level_id"),
		IsSigned: helper.JSONBool(m, "is_sign"),
	}
}

// SelfFollowForums 本账号关注贴吧列表。
type SelfFollowForums struct {
	classdef.Containers[*SelfFollowForum]

	HasMore bool  // 是否还有下一页
	Err     error // 捕获的异常
}

// SelfFollowForumsFromJSON 对应 SelfFollowForums.from_json。
func SelfFollowForumsFromJSON(m map[string]any) SelfFollowForums {
	var forums SelfFollowForums
	for _, item := range helper.JSONSlice(m, "like_forum") {
		if im, ok := item.(map[string]any); ok {
			f := SelfFollowForumFromJSON(im)
			forums.Objs = append(forums.Objs, &f)
		}
	}
	forums.HasMore = helper.JSONBool(m, "like_forum_has_more")
	return forums
}
