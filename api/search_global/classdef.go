// Package searchglobal implements the search_global API of aiotieba.
//
// It mirrors the Python package aiotieba.api.search_global.
package searchglobal

import (
	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper"
)

// GlobalSearchPost mirrors GlobalSearchPost.
type GlobalSearchPost struct {
	TID            int64
	PID            int64
	Title          string
	Content        string
	CreateTime     int64
	ForumID        int64
	ForumName      string
	PostNum        int64
	PbURL          string
	AuthorID       int64
	AuthorName     string
	AuthorShowName string
}

// GlobalSearchPostFromJSON mirrors GlobalSearchPost.from_json.
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

// GlobalSearches mirrors GlobalSearches.
type GlobalSearches struct {
	classdef.Containers[*GlobalSearchPost]

	HasMore     bool
	CurrentPage int64
	Err         error
}

// GlobalSearchesFromJSON mirrors GlobalSearches.from_json.
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
