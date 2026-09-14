// Package getfollows implements the get_follows API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_follows.
package getfollows

import (
	"github.com/rongyuio/aiotieba/api/classdef"
	"github.com/rongyuio/aiotieba/helper"
)

// Follow mirrors Follow.
type Follow struct {
	UserID      int64
	Portrait    string
	UserName    string
	NickNameNew string
}

// FollowFromJSON mirrors Follow.from_json.
func FollowFromJSON(m map[string]any) Follow {
	return Follow{
		UserID:      helper.JSONInt(m, "id"),
		Portrait:    classdef.TrimPortrait(helper.JSONStr(m, "portrait")),
		UserName:    helper.JSONStr(m, "name"),
		NickNameNew: helper.JSONStr(m, "name_show"),
	}
}

// NickName mirrors the nick_name property.
func (f Follow) NickName() string { return f.NickNameNew }

// ShowName mirrors the show_name property.
func (f Follow) ShowName() string {
	if f.NickNameNew != "" {
		return f.NickNameNew
	}
	return f.UserName
}

// PageFollow mirrors Page_follow.
type PageFollow struct {
	CurrentPage int64
	TotalCount  int64
	HasMore     bool
	HasPrev     bool
}

// PageFollowFromJSON mirrors Page_follow.from_json.
func PageFollowFromJSON(m map[string]any) PageFollow {
	return PageFollow{
		CurrentPage: helper.JSONInt(m, "pn"),
		TotalCount:  helper.JSONInt(m, "total_follow_num"),
		HasMore:     helper.JSONBool(m, "has_more"),
		HasPrev:     helper.JSONInt(m, "pn") > 1,
	}
}

// Follows mirrors Follows.
type Follows struct {
	classdef.Containers[*Follow]

	Page PageFollow
	Err  error
}

// FollowsFromJSON mirrors Follows.from_json.
func FollowsFromJSON(m map[string]any) Follows {
	var follows Follows
	for _, item := range helper.JSONSlice(m, "follow_list") {
		if im, ok := item.(map[string]any); ok {
			f := FollowFromJSON(im)
			follows.Objs = append(follows.Objs, &f)
		}
	}
	follows.Page = PageFollowFromJSON(m)
	return follows
}

// HasMore mirrors the has_more property.
func (f Follows) HasMore() bool { return f.Page.HasMore }
