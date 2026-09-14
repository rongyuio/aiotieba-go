// Package getfans implements the get_fans API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_fans.
package getfans

import (
	"github.com/rongyuio/aiotieba/api/classdef"
	"github.com/rongyuio/aiotieba/helper"
)

// Fan mirrors Fan.
type Fan struct {
	UserID      int64
	Portrait    string
	UserName    string
	NickNameNew string
}

// FanFromJSON mirrors Fan.from_json.
func FanFromJSON(m map[string]any) Fan {
	return Fan{
		UserID:      helper.JSONInt(m, "id"),
		Portrait:    classdef.TrimPortrait(helper.JSONStr(m, "portrait")),
		UserName:    helper.JSONStr(m, "name"),
		NickNameNew: helper.JSONStr(m, "name_show"),
	}
}

// NickName mirrors the nick_name property.
func (f Fan) NickName() string { return f.NickNameNew }

// ShowName mirrors the show_name property.
func (f Fan) ShowName() string {
	if f.NickNameNew != "" {
		return f.NickNameNew
	}
	return f.UserName
}

// PageFan mirrors Page_fan.
type PageFan struct {
	PageSize    int64
	CurrentPage int64
	TotalPage   int64
	TotalCount  int64
	HasMore     bool
	HasPrev     bool
}

// PageFanFromJSON mirrors Page_fan.from_json.
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

// Fans mirrors Fans.
type Fans struct {
	classdef.Containers[*Fan]

	Page PageFan
	Err  error
}

// FansFromJSON mirrors Fans.from_json.
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

// HasMore mirrors the has_more property.
func (f Fans) HasMore() bool { return f.Page.HasMore }
