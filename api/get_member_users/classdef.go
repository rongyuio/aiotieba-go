// Package getmemberusers 实现 aiotieba 的 get_member_users API。
//
// 对应 Python 包 aiotieba.api.get_member_users。
package getmemberusers

import (
	"strings"

	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper/htmlutil"
)

// MemberUser 最新关注用户信息。
type MemberUser struct {
	UserName string // 用户名
	Portrait string // portrait
	Level    int64  // 等级
}

// MemberUserFromXML 对应 MemberUser.from_xml。
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

// PageMember 页信息。
type PageMember struct {
	CurrentPage int64 // 当前页码
	TotalPage   int64 // 总页码
	HasMore     bool  // 是否有后继页
	HasPrev     bool  // 是否有前驱页
}

// PageMemberFromXML 对应 Page_member.from_xml。
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

// MemberUsers 最新关注用户列表。
type MemberUsers struct {
	classdef.Containers[*MemberUser]

	Page PageMember // 页信息
	Err  error      // 捕获的异常
}

// MemberUsersFromXML 对应 MemberUsers.from_xml。
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

// HasMore 是否还有下一页。
func (u MemberUsers) HasMore() bool { return u.Page.HasMore }
