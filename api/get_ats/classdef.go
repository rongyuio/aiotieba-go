// Package getats 实现 aiotieba 的 get_ats API。
//
// 对应 Python 包 aiotieba.api.get_ats。
package getats

import (
	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/helper"
)

// PageAt 页信息。
type PageAt struct {
	CurrentPage int64 // 当前页码
	HasMore     bool  // 是否有后继页
	HasPrev     bool  // 是否有前驱页
}

// PageAtFromJSON 对应 Page_at.from_json。
func PageAtFromJSON(m map[string]any) PageAt {
	return PageAt{
		CurrentPage: helper.JSONInt(m, "current_page"),
		HasMore:     helper.JSONBool(m, "has_more"),
		HasPrev:     helper.JSONBool(m, "has_prev"),
	}
}

// UserInfoAt 用户信息。
type UserInfoAt struct {
	UserID      int64           // user_id
	Portrait    string          // portrait
	UserName    string          // 用户名
	NickNameNew string          // 新版昵称
	PrivLike    enums.PrivLike  // 关注吧列表的公开状态
	PrivReply   enums.PrivReply // 帖子评论权限
}

// UserInfoAtFromJSON 对应 UserInfo_at.from_json。
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

// NickName 用户昵称。
func (u UserInfoAt) NickName() string { return u.NickNameNew }

// ShowName 显示名称。
func (u UserInfoAt) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// At @信息。
type At struct {
	Text       string     // 文本内容
	FName      string     // 所在贴吧名
	TID        int64      // 所在主题帖id
	PID        int64      // 回复id
	User       UserInfoAt // 发布者的用户信息
	IsComment  bool       // 是否楼中楼
	IsThread   bool       // 是否主题帖
	CreateTime int64      // 创建时间
}

// AtFromJSON 对应 At.from_json。
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

// AuthorID 发布者的user_id。
func (a At) AuthorID() int64 { return a.User.UserID }

// Ats @信息列表。
type Ats struct {
	classdef.Containers[*At]

	Page PageAt // 页信息
	Err  error  // 捕获的异常
}

// AtsFromJSON 对应 Ats.from_json。
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

// HasMore 是否还有下一页。
func (a Ats) HasMore() bool { return a.Page.HasMore }
