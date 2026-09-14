// Package getbawublacklist 实现 aiotieba 的 get_bawu_blacklist API。
//
// 对应 Python 包 aiotieba.api.get_bawu_blacklist。
package getbawublacklist

import (
	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper/htmlutil"
)

// BawuBlacklistUser 用户信息。
type BawuBlacklistUser struct {
	UserID   int64  // user_id
	Portrait string // portrait
	UserName string // 用户名
}

// BawuBlacklistUserFromXML 对应 BawuBlacklistUser.from_xml。
func BawuBlacklistUserFromXML(td *htmlutil.Node) BawuBlacklistUser {
	var u BawuBlacklistUser
	// 对应 data_tag.previous_sibling.input。
	if prev := td.PrevSibling; prev != nil {
		if input := htmlutil.FirstChildTag(prev, "input"); input != nil {
			u.UserName = htmlutil.Attr(input, "data-user-name")
			u.UserID = htmlutil.Atoi(htmlutil.Attr(input, "data-user-id"))
		}
	}
	if a := htmlutil.FirstChildTag(td, "a"); a != nil {
		if href := htmlutil.Attr(a, "href"); len(href) > 17 {
			u.Portrait = href[14 : len(href)-17]
		}
	}
	return u
}

// PageBwBlacklist 页信息。
type PageBwBlacklist struct {
	CurrentPage int64 // 当前页码
	TotalPage   int64 // 总页码
	TotalCount  int64 // 总计数
	HasMore     bool  // 是否有后继页
	HasPrev     bool  // 是否有前驱页
}

// PageBwBlacklistFromXML 对应 Page_bwblacklist.from_xml。
func PageBwBlacklistFromXML(soup *htmlutil.Node) PageBwBlacklist {
	var p PageBwBlacklist

	if bread := htmlutil.FindClass(soup, "div", "breadcrumbs"); bread != nil {
		if em := htmlutil.FirstChildTag(bread, "em"); em != nil {
			p.TotalCount = htmlutil.Atoi(htmlutil.Text(em))
		}
	}

	var pageTag *htmlutil.Node
	if pag := htmlutil.FindClass(soup, "div", "tbui_pagination"); pag != nil {
		pageTag = htmlutil.FindClass(pag, "li", "active")
	}

	if pageTag == nil {
		if p.TotalCount != 0 {
			p.CurrentPage, p.TotalPage = 1, 1
		}
	} else {
		p.CurrentPage = htmlutil.Atoi(htmlutil.Text(pageTag))
		if parent := pageTag.Parent; parent != nil && parent.NextSibling != nil {
			t := htmlutil.Text(parent.NextSibling)
			if len(t) >= 2 {
				p.TotalPage = htmlutil.Atoi(t[1 : len(t)-1])
			}
		}
	}
	p.HasMore = p.CurrentPage < p.TotalPage
	p.HasPrev = p.CurrentPage > 1
	return p
}

// BawuBlacklistUsers 吧务黑名单列表。
type BawuBlacklistUsers struct {
	classdef.Containers[*BawuBlacklistUser]

	Page PageBwBlacklist // 页信息
	Err  error           // 捕获的异常
}

// BawuBlacklistUsersFromXML 对应 BawuBlacklistUsers.from_xml。
func BawuBlacklistUsersFromXML(soup *htmlutil.Node) BawuBlacklistUsers {
	var users BawuBlacklistUsers
	for _, td := range htmlutil.FindAllClass(soup, "td", "left_cell") {
		u := BawuBlacklistUserFromXML(td)
		users.Objs = append(users.Objs, &u)
	}
	users.Page = PageBwBlacklistFromXML(soup)
	return users
}

// HasMore 是否还有下一页。
func (u BawuBlacklistUsers) HasMore() bool { return u.Page.HasMore }
