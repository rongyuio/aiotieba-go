// Package searchexact implements the search_exact API of aiotieba.
//
// It mirrors the Python package aiotieba.api.search_exact.
package searchexact

import (
	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper"
)

// ExactSearch mirrors ExactSearch.
type ExactSearch struct {
	Text       string
	Title      string
	FName      string
	TID        int64
	PID        int64
	ShowName   string
	IsComment  bool
	CreateTime int64
}

// ExactSearchFromJSON mirrors ExactSearch.from_json.
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

// PageExsch mirrors Page_exsch.
type PageExsch struct {
	PageSize    int64
	CurrentPage int64
	TotalPage   int64
	TotalCount  int64
	HasMore     bool
	HasPrev     bool
}

// PageExschFromJSON mirrors Page_exsch.from_json.
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

// ExactSearches mirrors ExactSearches.
type ExactSearches struct {
	classdef.Containers[*ExactSearch]

	Page PageExsch
	Err  error
}

// ExactSearchesFromJSON mirrors ExactSearches.from_json.
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

// HasMore mirrors the has_more property.
func (e ExactSearches) HasMore() bool { return e.Page.HasMore }
