// Package getrankforums implements the get_rank_forums API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_rank_forums.
package getrankforums

import (
	"strings"

	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper/htmlutil"
)

// RankForum mirrors RankForum.
type RankForum struct {
	FName     string
	SignNum   int64
	MemberNum int64
	HasBawu   bool
}

// RankForumFromXML mirrors RankForum.from_xml.
func RankForumFromXML(tr *htmlutil.Node) RankForum {
	var f RankForum

	rankIdx := htmlutil.FirstChildTag(tr, "td")
	fnameItem := htmlutil.NextSiblingTag(rankIdx, "td")
	f.FName = htmlutil.Text(fnameItem)

	signNumItem := htmlutil.NextSiblingTag(fnameItem, "td")
	f.SignNum = htmlutil.Atoi(htmlutil.Text(signNumItem))

	memberNumItem := htmlutil.NextSiblingTag(signNumItem, "td")
	f.MemberNum = htmlutil.Atoi(htmlutil.Text(memberNumItem))

	managerItem := htmlutil.NextSiblingTagClass(memberNumItem, "td", "clearfix")
	if managerDiv := htmlutil.FirstChildTag(managerItem, "div"); managerDiv != nil {
		classes := strings.Fields(htmlutil.Attr(managerDiv, "class"))
		if len(classes) > 0 {
			f.HasBawu = classes[0] != "no_bawu"
		}
	}
	return f
}

// PageRankForum mirrors Page_rankforum.
type PageRankForum struct {
	CurrentPage int64
	TotalPage   int64
	HasMore     bool
	HasPrev     bool
}

// PageRankForumFromXML mirrors Page_rankforum.from_xml.
func PageRankForumFromXML(soup *htmlutil.Node) PageRankForum {
	var p PageRankForum
	pages := htmlutil.FindClass(soup, "div", "pagination")
	if pages != nil {
		if span := htmlutil.FirstChildTag(pages, "span"); span != nil {
			p.CurrentPage = htmlutil.Atoi(htmlutil.Text(span))
		}
		as := htmlutil.FindAllTag(pages, "a")
		if len(as) > 0 {
			href := htmlutil.Attr(as[len(as)-1], "href")
			if idx := strings.LastIndex(href, "pn="); idx >= 0 {
				p.TotalPage = htmlutil.Atoi(href[idx+3:])
			}
		}
	}
	p.HasMore = p.CurrentPage < p.TotalPage
	p.HasPrev = p.CurrentPage > 1
	return p
}

// RankForums mirrors RankForums.
type RankForums struct {
	classdef.Containers[*RankForum]

	Page PageRankForum
	Err  error
}

// RankForumsFromXML mirrors RankForums.from_xml.
func RankForumsFromXML(soup *htmlutil.Node) RankForums {
	var forums RankForums
	if table := htmlutil.Find(soup, "table"); table != nil {
		for _, tr := range htmlutil.FindAllClass(table, "tr", "j_rank_row") {
			f := RankForumFromXML(tr)
			forums.Objs = append(forums.Objs, &f)
		}
	}
	forums.Page = PageRankForumFromXML(soup)
	return forums
}

// HasMore mirrors the has_more property.
func (f RankForums) HasMore() bool { return f.Page.HasMore }
