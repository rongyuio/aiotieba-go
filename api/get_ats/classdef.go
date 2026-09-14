// Package getats implements the get_ats API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_ats.
package getats

import (
	"github.com/rongyuio/aiotieba/api/classdef"
	"github.com/rongyuio/aiotieba/enums"
	"github.com/rongyuio/aiotieba/helper"
)

// PageAt is the page information. It mirrors Page_at.
type PageAt struct {
	CurrentPage int64
	HasMore     bool
	HasPrev     bool
}

// PageAtFromJSON mirrors Page_at.from_json.
func PageAtFromJSON(m map[string]any) PageAt {
	return PageAt{
		CurrentPage: helper.JSONInt(m, "current_page"),
		HasMore:     helper.JSONBool(m, "has_more"),
		HasPrev:     helper.JSONBool(m, "has_prev"),
	}
}

// UserInfoAt mirrors UserInfo_at.
type UserInfoAt struct {
	UserID      int64
	Portrait    string
	UserName    string
	NickNameNew string
	PrivLike    enums.PrivLike
	PrivReply   enums.PrivReply
}

// UserInfoAtFromJSON mirrors UserInfo_at.from_json.
func UserInfoAtFromJSON(m map[string]any) UserInfoAt {
	u := UserInfoAt{
		UserID:      helper.JSONInt(m, "id"),
		Portrait:    classdef.TrimPortrait(helper.JSONStr(m, "portrait")),
		UserName:    helper.JSONStr(m, "name"),
		NickNameNew: helper.JSONStr(m, "name_show"),
	}
	if priv := helper.JSONMap(m, "priv_sets"); priv != nil {
		u.PrivLike = enums.PrivLike(helper.JSONInt(priv, "like"))
		u.PrivReply = enums.PrivReply(helper.JSONInt(priv, "reply"))
	} else {
		u.PrivLike = enums.PrivLikePublic
		u.PrivReply = enums.PrivReplyAll
	}
	return u
}

// NickName mirrors the nick_name property.
func (u UserInfoAt) NickName() string { return u.NickNameNew }

// ShowName mirrors the show_name property.
func (u UserInfoAt) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// At mirrors the At dataclass.
type At struct {
	Text       string
	FName      string
	TID        int64
	PID        int64
	User       UserInfoAt
	IsComment  bool
	IsThread   bool
	CreateTime int64
}

// AtFromJSON mirrors At.from_json.
func AtFromJSON(m map[string]any) At {
	return At{
		Text:       helper.JSONStr(m, "content"),
		FName:      helper.JSONStr(m, "fname"),
		TID:        helper.JSONInt(m, "thread_id"),
		PID:        helper.JSONInt(m, "post_id"),
		User:       UserInfoAtFromJSON(helper.JSONMap(m, "replyer")),
		IsComment:  helper.JSONBool(m, "is_floor"),
		IsThread:   helper.JSONBool(m, "is_first_post"),
		CreateTime: helper.JSONInt(m, "time"),
	}
}

// AuthorID mirrors the author_id property.
func (a At) AuthorID() int64 { return a.User.UserID }

// Ats mirrors Ats.
type Ats struct {
	classdef.Containers[*At]

	Page PageAt
	Err  error
}

// AtsFromJSON mirrors Ats.from_json.
func AtsFromJSON(m map[string]any) Ats {
	var ats Ats
	for _, item := range helper.JSONSlice(m, "at_list") {
		if im, ok := item.(map[string]any); ok {
			at := AtFromJSON(im)
			ats.Objs = append(ats.Objs, &at)
		}
	}
	ats.Page = PageAtFromJSON(helper.JSONMap(m, "page"))
	return ats
}

// HasMore mirrors the has_more property.
func (a Ats) HasMore() bool { return a.Page.HasMore }
