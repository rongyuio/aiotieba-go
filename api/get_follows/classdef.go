// Package getfollows 实现 aiotieba 的 get_follows API。
//
// 对应 Python 包 aiotieba.api.get_follows。
package getfollows

import (
	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper"
)

// Follow 用户信息。
type Follow struct {
	UserID      int64  // user_id
	Portrait    string // portrait
	UserName    string // 用户名
	NickNameNew string // 新版昵称
}

// FollowFromJSON 对应 Follow.from_json。
func FollowFromJSON(m map[string]any) Follow {
	return Follow{
		UserID:      helper.JSONInt(m, "id"),
		Portrait:    classdef.TrimPortrait(helper.JSONStr(m, "portrait")),
		UserName:    helper.JSONStr(m, "name"),
		NickNameNew: helper.JSONStr(m, "name_show"),
	}
}

// NickName 用户昵称。
func (f Follow) NickName() string { return f.NickNameNew }

// ShowName 显示名称。
func (f Follow) ShowName() string {
	if f.NickNameNew != "" {
		return f.NickNameNew
	}
	return f.UserName
}

// PageFollow 页信息。
type PageFollow struct {
	CurrentPage int64 // 当前页码
	TotalCount  int64 // 总计数
	HasMore     bool  // 是否有后继页
	HasPrev     bool  // 是否有前驱页
}

// PageFollowFromJSON 对应 Page_follow.from_json。
func PageFollowFromJSON(m map[string]any) PageFollow {
	return PageFollow{
		CurrentPage: helper.JSONInt(m, "pn"),
		TotalCount:  helper.JSONInt(m, "total_follow_num"),
		HasMore:     helper.JSONBool(m, "has_more"),
		HasPrev:     helper.JSONInt(m, "pn") > 1,
	}
}

// Follows 粉丝列表。
type Follows struct {
	classdef.Containers[*Follow]

	Page PageFollow // 页信息
	Err  error      // 捕获的异常
}

// FollowsFromJSON 对应 Follows.from_json。
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

// HasMore 是否还有下一页。
func (f Follows) HasMore() bool { return f.Page.HasMore }
