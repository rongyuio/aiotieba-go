// Package getselffollowforums implements the get_self_follow_forums API of
// aiotieba.
//
// It mirrors the Python package aiotieba.api.get_self_follow_forums.
package getselffollowforums

import (
	"github.com/rongyuio/aiotieba/api/classdef"
	"github.com/rongyuio/aiotieba/helper"
)

// SelfFollowForum mirrors SelfFollowForum.
type SelfFollowForum struct {
	FID      int64
	FName    string
	Level    int64
	IsSigned bool
}

// SelfFollowForumFromJSON mirrors SelfFollowForum.from_json.
func SelfFollowForumFromJSON(m map[string]any) SelfFollowForum {
	return SelfFollowForum{
		FID:      helper.JSONInt(m, "forum_id"),
		FName:    helper.JSONStr(m, "forum_name"),
		Level:    helper.JSONInt(m, "level_id"),
		IsSigned: helper.JSONBool(m, "is_sign"),
	}
}

// SelfFollowForums mirrors SelfFollowForums.
type SelfFollowForums struct {
	classdef.Containers[*SelfFollowForum]

	HasMore bool
	Err     error
}

// SelfFollowForumsFromJSON mirrors SelfFollowForums.from_json.
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
