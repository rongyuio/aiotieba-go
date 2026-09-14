// Package getfollowforumspc implements the get_follow_forums_pc API of
// aiotieba.
//
// It mirrors the Python package aiotieba.api.get_follow_forums_pc.
package getfollowforumspc

import (
	"github.com/rongyuio/aiotieba/api/classdef"
	"github.com/rongyuio/aiotieba/helper"
)

// PcFollowForum mirrors PcFollowForum.
type PcFollowForum struct {
	FID   int64
	FName string
	Level int64
}

// PcFollowForumFromJSON mirrors PcFollowForum.from_json.
func PcFollowForumFromJSON(m map[string]any) PcFollowForum {
	return PcFollowForum{
		FID:   helper.JSONInt(m, "forum_id"),
		FName: helper.JSONStr(m, "forum_name"),
		Level: helper.JSONInt(m, "level_id"),
	}
}

// PcFollowForums mirrors PcFollowForums.
type PcFollowForums struct {
	classdef.Containers[*PcFollowForum]

	HasMore bool
	Err     error
}

// PcFollowForumsFromJSON mirrors PcFollowForums.from_json.
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
