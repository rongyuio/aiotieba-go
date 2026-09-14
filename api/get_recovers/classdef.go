// Package getrecovers implements the get_recovers API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_recovers.
package getrecovers

import (
	"github.com/rongyuio/aiotieba/api/classdef"
	"github.com/rongyuio/aiotieba/helper"
)

// UserInfoRec mirrors UserInfo_rec.
type UserInfoRec struct {
	UserName    string
	Portrait    string
	NickNameNew string
}

// UserInfoRecFromJSON mirrors UserInfo_rec.from_json.
func UserInfoRecFromJSON(m map[string]any) UserInfoRec {
	return UserInfoRec{
		Portrait:    classdef.TrimPortrait(helper.JSONStr(m, "portrait")),
		UserName:    helper.JSONStr(m, "user_name"),
		NickNameNew: helper.JSONStr(m, "user_nickname"),
	}
}

// NickName mirrors the nick_name property.
func (u UserInfoRec) NickName() string { return u.NickNameNew }

// ShowName mirrors the show_name property.
func (u UserInfoRec) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// Recover mirrors Recover.
type Recover struct {
	Text       string
	TID        int64
	PID        int64
	User       UserInfoRec
	OpShowName string
	OpTime     int64
	IsFloor    bool
	IsHide     bool
}

// RecoverFromJSON mirrors Recover.from_json.
func RecoverFromJSON(m map[string]any) Recover {
	threadInfo := helper.JSONMap(m, "thread_info")
	r := Recover{TID: helper.JSONInt(threadInfo, "tid")}

	if postInfo := helper.JSONMap(m, "post_info"); postInfo != nil {
		r.Text = helper.JSONStr(postInfo, "abstract")
		r.PID = helper.JSONInt(postInfo, "pid")
		r.User = UserInfoRecFromJSON(postInfo)
	} else {
		r.Text = helper.JSONStr(threadInfo, "abstract")
		r.User = UserInfoRecFromJSON(threadInfo)
	}

	r.IsFloor = helper.JSONBool(m, "is_foor") // 百度 code review 的装饰性字段名
	r.IsHide = helper.JSONBool(m, "is_frs_mask")

	if op := helper.JSONMap(m, "op_info"); op != nil {
		r.OpShowName = helper.JSONStr(op, "name")
		r.OpTime = helper.JSONInt(op, "time")
	}
	return r
}

// PageRecover mirrors Page_recover.
type PageRecover struct {
	PageSize    int64
	CurrentPage int64
	HasMore     bool
	HasPrev     bool
}

// PageRecoverFromJSON mirrors Page_recover.from_json.
func PageRecoverFromJSON(m map[string]any) PageRecover {
	return PageRecover{
		PageSize:    helper.JSONInt(m, "rn"),
		CurrentPage: helper.JSONInt(m, "pn"),
		HasMore:     helper.JSONBool(m, "has_more"),
		HasPrev:     helper.JSONInt(m, "pn") > 1,
	}
}

// Recovers mirrors Recovers.
type Recovers struct {
	classdef.Containers[*Recover]

	Page PageRecover
	Err  error
}

// RecoversFromJSON mirrors Recovers.from_json.
func RecoversFromJSON(m map[string]any) Recovers {
	var recovers Recovers
	for _, item := range helper.JSONSlice(m, "thread_list") {
		if im, ok := item.(map[string]any); ok {
			r := RecoverFromJSON(im)
			recovers.Objs = append(recovers.Objs, &r)
		}
	}
	recovers.Page = PageRecoverFromJSON(helper.JSONMap(m, "page"))
	return recovers
}

// HasMore mirrors the has_more property.
func (r Recovers) HasMore() bool { return r.Page.HasMore }
