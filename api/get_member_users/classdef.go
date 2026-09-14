// Package getmemberusers implements the get_member_users API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_member_users.
package getmemberusers

import (
	"strings"

	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper/htmlutil"
)

// MemberUser mirrors MemberUser.
type MemberUser struct {
	UserName string
	Portrait string
	Level    int64
}

// MemberUserFromXML mirrors MemberUser.from_xml.
func MemberUserFromXML(div *htmlutil.Node) MemberUser {
	var u MemberUser
	if a := htmlutil.FirstChildTag(div, "a"); a != nil {
		u.UserName = htmlutil.Attr(a, "title")
		if href := htmlutil.Attr(a, "href"); len(href) > 14 {
			u.Portrait = href[14:]
		}
	}
	if span := htmlutil.FirstChildTag(div, "span"); span != nil {
		classes := strings.Fields(htmlutil.Attr(span, "class"))
		if len(classes) > 1 && len(classes[1]) > 12 {
			u.Level = htmlutil.Atoi(classes[1][12:])
		}
	}
	return u
}

// PageMember mirrors Page_member.
type PageMember struct {
	CurrentPage int64
	TotalPage   int64
	HasMore     bool
	HasPrev     bool
}

// PageMemberFromXML mirrors Page_member.from_xml.
func PageMemberFromXML(li *htmlutil.Node) PageMember {
	var p PageMember
	p.CurrentPage = htmlutil.Atoi(htmlutil.Text(li))
	if li != nil && li.Parent != nil && li.Parent.NextSibling != nil {
		t := htmlutil.Text(li.Parent.NextSibling)
		if len(t) >= 2 {
			p.TotalPage = htmlutil.Atoi(t[1 : len(t)-1])
		}
	}
	p.HasMore = p.CurrentPage < p.TotalPage
	p.HasPrev = p.CurrentPage > 1
	return p
}

// MemberUsers mirrors MemberUsers.
type MemberUsers struct {
	classdef.Containers[*MemberUser]

	Page PageMember
	Err  error
}

// MemberUsersFromXML mirrors MemberUsers.from_xml.
func MemberUsersFromXML(soup *htmlutil.Node) MemberUsers {
	var users MemberUsers
	for _, div := range htmlutil.FindAllClass(soup, "div", "name_wrap") {
		u := MemberUserFromXML(div)
		users.Objs = append(users.Objs, &u)
	}
	var pageTag *htmlutil.Node
	if pag := htmlutil.FindClass(soup, "div", "tbui_pagination"); pag != nil {
		pageTag = htmlutil.FindClass(pag, "li", "active")
	}
	users.Page = PageMemberFromXML(pageTag)
	return users
}

// HasMore mirrors the has_more property.
func (u MemberUsers) HasMore() bool { return u.Page.HasMore }
