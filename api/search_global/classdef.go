// Package searchglobal 实现 aiotieba 的 search_global API。
//
// 对应 Python 包 aiotieba.api.search_global。
package searchglobal

import (
	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper"
)

// SearchGlobal 全吧搜索结果。
type SearchGlobal struct {
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

// SearchGlobalFromJSON 对应 SearchGlobal.from_json。
func SearchGlobalFromJSON(m map[string]any) SearchGlobal {
	user := helper.JSONMap(m, "user")
	createTime := helper.JSONInt(m, "create_time")
	if createTime == 0 {
		createTime = helper.JSONInt(m, "time")
	}
	authorShowName := helper.JSONStr(user, "show_nickname")
	if authorShowName == "" {
		authorShowName = helper.JSONStr(user, "user_name")
	}
	// post_num 会以 "1.2W" 形式缩写，该接口用的是英文后缀。
	postNum, _ := helper.JSONRaw(m, "post_num")
	return SearchGlobal{
		TID:            helper.JSONInt(m, "tid"),
		PID:            helper.JSONInt(m, "pid"),
		Title:          helper.JSONStr(m, "title"),
		Content:        helper.JSONStr(m, "content"),
		CreateTime:     createTime,
		ForumID:        helper.JSONInt(m, "forum_id"),
		ForumName:      helper.JSONStr(m, "forum_name"),
		PostNum:        helper.TbNum2Int(postNum, "W"),
		PbURL:          helper.JSONStr(m, "pb_url"),
		AuthorID:       helper.JSONInt(user, "user_id"),
		AuthorName:     helper.JSONStr(user, "user_name"),
		AuthorShowName: authorShowName,
	}
}

// SearchGlobals 全吧搜索结果列表。
type SearchGlobals struct {
	classdef.Containers[*SearchGlobal]

	HasMore     bool  // 是否还有下一页
	CurrentPage int64 // 当前页码
	Err         error // 捕获的异常
}

// SearchGlobalsFromJSON 对应 SearchGlobals.from_json。
func SearchGlobalsFromJSON(m map[string]any) SearchGlobals {
	var searches SearchGlobals
	for _, item := range helper.JSONSlice(m, "post_list") {
		if im, ok := item.(map[string]any); ok {
			p := SearchGlobalFromJSON(im)
			searches.Objs = append(searches.Objs, &p)
		}
	}
	searches.HasMore = helper.JSONBool(m, "has_more")
	searches.CurrentPage = helper.JSONInt(m, "current_page")
	return searches
}
