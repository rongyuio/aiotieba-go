// Package searchglobal 实现 aiotieba 的 search_global API。
//
// 对应 Python 包 aiotieba.api.search_global。
package searchglobal

import (
	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper"
)

// GlobalSearchPost 全吧搜索结果。
type GlobalSearchPost struct {
	TID            int64  // 所在主题帖id
	PID            int64  // 回复id
	Title          string // 标题
	Content        string // 正文
	CreateTime     int64  // 创建时间
	ForumID        int64  // 所在贴吧fid
	ForumName      string // 所在贴吧名
	PostNum        int64  // 回复数
	PbURL          string // 帖子页相对路径
	AuthorID       int64  // 作者user_id
	AuthorName     string // 作者用户名
	AuthorShowName string // 作者显示昵称
}

// GlobalSearchPostFromJSON 对应 GlobalSearchPost.from_json。
func GlobalSearchPostFromJSON(m map[string]any) GlobalSearchPost {
	user := helper.JSONMap(m, "user")
	createTime := helper.JSONInt(m, "create_time")
	if createTime == 0 {
		createTime = helper.JSONInt(m, "time")
	}
	authorShowName := helper.JSONStr(user, "show_nickname")
	if authorShowName == "" {
		authorShowName = helper.JSONStr(user, "user_name")
	}
	return GlobalSearchPost{
		TID:            helper.JSONInt(m, "tid"),
		PID:            helper.JSONInt(m, "pid"),
		Title:          helper.JSONStr(m, "title"),
		Content:        helper.JSONStr(m, "content"),
		CreateTime:     createTime,
		ForumID:        helper.JSONInt(m, "forum_id"),
		ForumName:      helper.JSONStr(m, "forum_name"),
		PostNum:        helper.JSONInt(m, "post_num"),
		PbURL:          helper.JSONStr(m, "pb_url"),
		AuthorID:       helper.JSONInt(user, "user_id"),
		AuthorName:     helper.JSONStr(user, "user_name"),
		AuthorShowName: authorShowName,
	}
}

// GlobalSearches 全吧搜索结果列表。
type GlobalSearches struct {
	classdef.Containers[*GlobalSearchPost]

	HasMore     bool  // 是否还有下一页
	CurrentPage int64 // 当前页码
	Err         error // 捕获的异常
}

// GlobalSearchesFromJSON 对应 GlobalSearches.from_json。
func GlobalSearchesFromJSON(m map[string]any) GlobalSearches {
	var searches GlobalSearches
	for _, item := range helper.JSONSlice(m, "post_list") {
		if im, ok := item.(map[string]any); ok {
			p := GlobalSearchPostFromJSON(im)
			searches.Objs = append(searches.Objs, &p)
		}
	}
	searches.HasMore = helper.JSONBool(m, "has_more")
	searches.CurrentPage = helper.JSONInt(m, "current_page")
	return searches
}
