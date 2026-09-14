// Package getfollowforums implements the get_follow_forums API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_follow_forums.
package getfollowforums

import (
	"github.com/rongyuio/aiotieba/api/classdef"
	"github.com/rongyuio/aiotieba/helper"
)

// FollowForum mirrors FollowForum.
type FollowForum struct {
	FID   int64
	FName string
	Level int64
	Exp   int64
}

// FollowForumFromJSON mirrors FollowForum.from_json.
func FollowForumFromJSON(m map[string]any) FollowForum {
	return FollowForum{
		FID:   helper.JSONInt(m, "id"),
		FName: helper.JSONStr(m, "name"),
		Level: helper.JSONInt(m, "level_id"),
		Exp:   helper.JSONInt(m, "cur_score"),
	}
}

// FollowForums mirrors FollowForums.
type FollowForums struct {
	classdef.Containers[*FollowForum]

	HasMore bool
	Err     error
}

// FollowForumsFromJSON mirrors FollowForums.from_json.
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
