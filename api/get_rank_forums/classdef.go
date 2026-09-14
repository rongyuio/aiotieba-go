// Package getrankforums 实现 aiotieba 的 get_rank_forums API。
//
// 对应 Python 包 aiotieba.api.get_rank_forums。
package getrankforums

import (
	"strings"

	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper/htmlutil"
)

// RankForum 吧签到排名。
type RankForum struct {
	FName     string // 吧名
	SignNum   int64  // 签到用户数
	MemberNum int64  // 总用户数
	HasBawu   bool   // 是否有吧务
}

// RankForumFromXML 对应 RankForum.from_xml。
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

// PageRankForum 页信息。
type PageRankForum struct {
	CurrentPage int64 // 当前页码
	TotalPage   int64 // 总页码
	HasMore     bool  // 是否有后继页
	HasPrev     bool  // 是否有前驱页
}

// PageRankForumFromXML 对应 Page_rankforum.from_xml。
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

// RankForums 吧签到排名表。
type RankForums struct {
	classdef.Containers[*RankForum]

	Page PageRankForum // 页信息
	Err  error         // 捕获的异常
}

// RankForumsFromXML 对应 RankForums.from_xml。
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

// HasMore 是否还有下一页。
func (f RankForums) HasMore() bool { return f.Page.HasMore }
