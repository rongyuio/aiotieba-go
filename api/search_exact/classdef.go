// Package searchexact 实现 aiotieba 的 search_exact API。
//
// 对应 Python 包 aiotieba.api.search_exact。
package searchexact

import (
	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper"
)

// ExactSearch 搜索结果。
type ExactSearch struct {
	Text       string // 文本内容
	Title      string // 标题内容
	FName      string // 所在贴吧名
	TID        int64  // 所在主题帖id
	PID        int64  // 回复id
	ShowName   string // 发布者的显示名称
	IsComment  bool   // 是否楼中楼
	CreateTime int64  // 创建时间
}

// ExactSearchFromJSON 对应 ExactSearch.from_json。
func ExactSearchFromJSON(m map[string]any) ExactSearch {
	return ExactSearch{
		Text:       helper.JSONStr(m, "content"),
		Title:      helper.JSONStr(m, "title"),
		FName:      helper.JSONStr(m, "fname"),
		TID:        helper.JSONInt(m, "tid"),
		PID:        helper.JSONInt(m, "pid"),
		ShowName:   helper.JSONStr(helper.JSONMap(m, "author"), "name_show"),
		IsComment:  helper.JSONBool(m, "is_floor"),
		CreateTime: helper.JSONInt(m, "time"),
	}
}

// PageExsch 页信息。
type PageExsch struct {
	PageSize    int64 // 页大小
	CurrentPage int64 // 当前页码
	TotalPage   int64 // 总页码
	TotalCount  int64 // 总计数
	HasMore     bool  // 是否有后继页
	HasPrev     bool  // 是否有前驱页
}

// PageExschFromJSON 对应 Page_exsch.from_json。
func PageExschFromJSON(m map[string]any) PageExsch {
	return PageExsch{
		PageSize:    helper.JSONInt(m, "page_size"),
		CurrentPage: helper.JSONInt(m, "current_page"),
		TotalPage:   helper.JSONInt(m, "total_page"),
		TotalCount:  helper.JSONInt(m, "total_count"),
		HasMore:     helper.JSONBool(m, "has_more"),
		HasPrev:     helper.JSONBool(m, "has_prev"),
	}
}

// ExactSearches 搜索结果列表。
type ExactSearches struct {
	classdef.Containers[*ExactSearch]

	Page PageExsch // 页信息
	Err  error     // 捕获的异常
}

// ExactSearchesFromJSON 对应 ExactSearches.from_json。
func ExactSearchesFromJSON(m map[string]any) ExactSearches {
	var searches ExactSearches
	for _, item := range helper.JSONSlice(m, "post_list") {
		if im, ok := item.(map[string]any); ok {
			s := ExactSearchFromJSON(im)
			searches.Objs = append(searches.Objs, &s)
		}
	}
	searches.Page = PageExschFromJSON(helper.JSONMap(m, "page"))
	return searches
}

// HasMore 是否还有下一页。
func (e ExactSearches) HasMore() bool { return e.Page.HasMore }
