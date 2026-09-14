// Package getbawublacklist implements the get_bawu_blacklist API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_bawu_blacklist.
package getbawublacklist

import (
	"github.com/rongyuio/aiotieba/api/classdef"
	"github.com/rongyuio/aiotieba/helper/htmlutil"
)

// BawuBlacklistUser mirrors BawuBlacklistUser.
type BawuBlacklistUser struct {
	UserID   int64
	Portrait string
	UserName string
}

// BawuBlacklistUserFromXML mirrors BawuBlacklistUser.from_xml.
func BawuBlacklistUserFromXML(td *htmlutil.Node) BawuBlacklistUser {
	var u BawuBlacklistUser
	// data_tag.previous_sibling.input
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

// PageBwBlacklist mirrors Page_bwblacklist.
type PageBwBlacklist struct {
	CurrentPage int64
	TotalPage   int64
	TotalCount  int64
	HasMore     bool
	HasPrev     bool
}

// PageBwBlacklistFromXML mirrors Page_bwblacklist.from_xml.
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

// BawuBlacklistUsers mirrors BawuBlacklistUsers.
type BawuBlacklistUsers struct {
	classdef.Containers[*BawuBlacklistUser]

	Page PageBwBlacklist
	Err  error
}

// BawuBlacklistUsersFromXML mirrors BawuBlacklistUsers.from_xml.
func BawuBlacklistUsersFromXML(soup *htmlutil.Node) BawuBlacklistUsers {
	var users BawuBlacklistUsers
	for _, td := range htmlutil.FindAllClass(soup, "td", "left_cell") {
		u := BawuBlacklistUserFromXML(td)
		users.Objs = append(users.Objs, &u)
	}
	users.Page = PageBwBlacklistFromXML(soup)
	return users
}

// HasMore mirrors the has_more property.
func (u BawuBlacklistUsers) HasMore() bool { return u.Page.HasMore }
