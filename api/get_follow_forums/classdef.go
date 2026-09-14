// Package getfollowforums 实现 aiotieba 的 get_follow_forums API。
//
// 对应 Python 包 aiotieba.api.get_follow_forums。
package getfollowforums

import (
	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper"
)

// FollowForum 关注吧信息。
type FollowForum struct {
	FID   int64  // 贴吧id
	FName string // 贴吧名
	Level int64  // 用户等级
	Exp   int64  // 经验值
}

// FollowForumFromJSON 对应 FollowForum.from_json。
func FollowForumFromJSON(m map[string]any) FollowForum {
	return FollowForum{
		FID:   helper.JSONInt(m, "id"),
		FName: helper.JSONStr(m, "name"),
		Level: helper.JSONInt(m, "level_id"),
		Exp:   helper.JSONInt(m, "cur_score"),
	}
}

// FollowForums 用户关注贴吧列表。
type FollowForums struct {
	classdef.Containers[*FollowForum]

	HasMore bool  // 是否还有下一页
	Err     error // 捕获的异常
}

// FollowForumsFromJSON 对应 FollowForums.from_json。
func FollowForumsFromJSON(m map[string]any) FollowForums {
	var forums FollowForums
	if forumList := helper.JSONMap(m, "forum_list"); forumList != nil {
		for _, group := range []string{"non-gconforum", "gconforum"} {
			for _, item := range helper.JSONSlice(forumList, group) {
				if im, ok := item.(map[string]any); ok {
					f := FollowForumFromJSON(im)
					forums.Objs = append(forums.Objs, &f)
				}
			}
		}
		forums.HasMore = helper.JSONBool(m, "has_more")
	}
	return forums
}
