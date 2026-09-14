// Package getfans 实现 aiotieba 的 get_fans API。
//
// 对应 Python 包 aiotieba.api.get_fans。
package getfans

import (
	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper"
)

// Fan 用户信息。
type Fan struct {
	UserID      int64  // user_id
	Portrait    string // portrait
	UserName    string // 用户名
	NickNameNew string // 新版昵称
}

// FanFromJSON 对应 Fan.from_json。
func FanFromJSON(m map[string]any) Fan {
	return Fan{
		UserID:      helper.JSONInt(m, "id"),
		Portrait:    classdef.TrimPortrait(helper.JSONStr(m, "portrait")),
		UserName:    helper.JSONStr(m, "name"),
		NickNameNew: helper.JSONStr(m, "name_show"),
	}
}

// NickName 用户昵称。
func (f Fan) NickName() string { return f.NickNameNew }

// ShowName 显示名称。
func (f Fan) ShowName() string {
	if f.NickNameNew != "" {
		return f.NickNameNew
	}
	return f.UserName
}

// PageFan 页信息。
type PageFan struct {
	PageSize    int64 // 页大小
	CurrentPage int64 // 当前页码
	TotalPage   int64 // 总页码
	TotalCount  int64 // 总计数
	HasMore     bool  // 是否有后继页
	HasPrev     bool  // 是否有前驱页
}

// PageFanFromJSON 对应 Page_fan.from_json。
func PageFanFromJSON(m map[string]any) PageFan {
	return PageFan{
		PageSize:    helper.JSONInt(m, "page_size"),
		CurrentPage: helper.JSONInt(m, "current_page"),
		TotalPage:   helper.JSONInt(m, "total_page"),
		TotalCount:  helper.JSONInt(m, "total_count"),
		HasMore:     helper.JSONBool(m, "has_more"),
		HasPrev:     helper.JSONBool(m, "has_prev"),
	}
}

// Fans 粉丝列表。
type Fans struct {
	classdef.Containers[*Fan]

	Page PageFan // 页信息
	Err  error   // 捕获的异常
}

// FansFromJSON 对应 Fans.from_json。
func FansFromJSON(m map[string]any) Fans {
	var fans Fans
	for _, item := range helper.JSONSlice(m, "user_list") {
		if im, ok := item.(map[string]any); ok {
			f := FanFromJSON(im)
			fans.Objs = append(fans.Objs, &f)
		}
	}
	fans.Page = PageFanFromJSON(helper.JSONMap(m, "page"))
	return fans
}

// HasMore 是否还有下一页。
func (f Fans) HasMore() bool { return f.Page.HasMore }
