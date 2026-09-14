// Package getforum implements the get_forum API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_forum.
package getforum

import (
	"github.com/rongyuio/aiotieba/helper"
)

// Forum is the information of a forum. It mirrors
// aiotieba.api.get_forum._classdef.Forum.
type Forum struct {
	FID   int64
	FName string

	Category    string
	Subcategory string

	SmallAvatar string
	Slogan      string
	MemberNum   int64
	PostNum     int64
	ThreadNum   int64

	HasBawu bool

	Err error
}

// ForumFromJSON mirrors Forum.from_json.
func ForumFromJSON(data map[string]any) Forum {
	return Forum{
		FID:         helper.JSONInt(data, "id"),
		FName:       helper.JSONStr(data, "name"),
		Category:    helper.JSONStr(data, "first_class"),
		Subcategory: helper.JSONStr(data, "second_class"),
		SmallAvatar: helper.JSONStr(data, "avatar"),
		Slogan:      helper.JSONStr(data, "slogan"),
		MemberNum:   helper.JSONInt(data, "member_num"),
		PostNum:     helper.JSONInt(data, "post_num"),
		ThreadNum:   helper.JSONInt(data, "thread_num"),
		HasBawu:     helper.HasJSONKey(data, "managers"),
	}
}
