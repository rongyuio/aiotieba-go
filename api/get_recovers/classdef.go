// Package getrecovers 实现 aiotieba 的 get_recovers API。
//
// 对应 Python 包 aiotieba.api.get_recovers。
package getrecovers

import (
	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper"
)

// UserInfoRec 用户信息。
type UserInfoRec struct {
	UserName    string // 用户名
	Portrait    string // portrait
	NickNameNew string // 新版昵称
}

// UserInfoRecFromJSON 对应 UserInfo_rec.from_json。
func UserInfoRecFromJSON(m map[string]any) UserInfoRec {
	return UserInfoRec{
		Portrait:    classdef.TrimPortrait(helper.JSONStr(m, "portrait")),
		UserName:    helper.JSONStr(m, "user_name"),
		NickNameNew: helper.JSONStr(m, "user_nickname"),
	}
}

// NickName 用户昵称。
func (u UserInfoRec) NickName() string { return u.NickNameNew }

// ShowName 显示名称。
func (u UserInfoRec) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// Recover 待恢复帖子信息。
type Recover struct {
	Text       string      // 文本内容
	TID        int64       // 所在主题帖id
	PID        int64       // 回复id 若为主题帖则该字段为0
	User       UserInfoRec // 发布者的用户信息
	OpShowName string      // 操作人显示名称
	OpTime     int64       // 操作时间 10位时间戳 以秒为单位
	IsFloor    bool        // 是否为楼中楼
	IsHide     bool        // 是否为屏蔽
}

// RecoverFromJSON 对应 Recover.from_json。
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

// PageRecover 页信息。
type PageRecover struct {
	PageSize    int64 // 页大小
	CurrentPage int64 // 当前页码
	HasMore     bool  // 是否有后继页
	HasPrev     bool  // 是否有前驱页
}

// PageRecoverFromJSON 对应 Page_recover.from_json。
func PageRecoverFromJSON(m map[string]any) PageRecover {
	return PageRecover{
		PageSize:    helper.JSONInt(m, "rn"),
		CurrentPage: helper.JSONInt(m, "pn"),
		HasMore:     helper.JSONBool(m, "has_more"),
		HasPrev:     helper.JSONInt(m, "pn") > 1,
	}
}

// Recovers 待恢复帖子列表。
type Recovers struct {
	classdef.Containers[*Recover]

	Page PageRecover // 页信息
	Err  error       // 捕获的异常
}

// RecoversFromJSON 对应 Recovers.from_json。
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

// HasMore 是否还有下一页。
func (r Recovers) HasMore() bool { return r.Page.HasMore }
