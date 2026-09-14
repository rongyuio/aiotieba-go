// Package getfollowforumspc 实现 aiotieba 的 get_follow_forums_pc API。
//
// 对应 Python 包 aiotieba.api.get_follow_forums_pc。
package getfollowforumspc

import (
	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper"
)

// PcFollowForum 关注吧信息。
type PcFollowForum struct {
	FID   int64  // 贴吧id
	FName string // 贴吧名
	Level int64  // 用户等级
}

// PcFollowForumFromJSON 对应 PcFollowForum.from_json。
func PcFollowForumFromJSON(m map[string]any) PcFollowForum {
	return PcFollowForum{
		FID:   helper.JSONInt(m, "forum_id"),
		FName: helper.JSONStr(m, "forum_name"),
		Level: helper.JSONInt(m, "level_id"),
	}
}

// PcFollowForums 用户关注贴吧列表。
type PcFollowForums struct {
	classdef.Containers[*PcFollowForum]

	HasMore bool  // 是否还有下一页
	Err     error // 捕获的异常
}

// PcFollowForumsFromJSON 对应 PcFollowForums.from_json。
func PcFollowForumsFromJSON(m map[string]any) PcFollowForums {
	var forums PcFollowForums
	for _, item := range helper.JSONSlice(m, "like") {
		if im, ok := item.(map[string]any); ok {
			f := PcFollowForumFromJSON(im)
			forums.Objs = append(forums.Objs, &f)
		}
	}
	forums.HasMore = helper.JSONBool(m, "has_more")
	return forums
}
