// Package getrankusers 实现 aiotieba 的 get_rank_users API。
//
// 对应 Python 包 aiotieba.api.get_rank_users。
package getrankusers

import (
	"strings"

	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/htmlutil"
)

// RankUser 等级排行榜用户信息。
type RankUser struct {
	UserName string // 用户名
	Level    int64  // 等级
	Exp      int64  // 经验值
	IsVIP    bool   // 是否超级会员
}

// RankUserFromXML 对应 RankUser.from_xml。
func RankUserFromXML(tr *htmlutil.Node) RankUser {
	var u RankUser

	td := htmlutil.FirstChildTag(tr, "td")
	userNameItem := td.NextSibling
	u.UserName = htmlutil.Text(userNameItem)

	if div := htmlutil.FirstChildTag(userNameItem, "div"); div != nil {
		u.IsVIP = htmlutil.HasClass(div, "drl_item_vip")
	}

	levelItem := userNameItem.NextSibling
	if levelDiv := htmlutil.FirstChildTag(levelItem, "div"); levelDiv != nil {
		classes := strings.Fields(htmlutil.Attr(levelDiv, "class"))
		if len(classes) > 0 && len(classes[0]) > 5 {
			u.Level = htmlutil.Atoi(classes[0][5:])
		}
	}

	expItem := levelItem.NextSibling
	u.Exp = htmlutil.Atoi(htmlutil.Text(expItem))

	return u
}

// PageRank 页信息。
type PageRank struct {
	CurrentPage int64 // 当前页码
	TotalPage   int64 // 总页码
	HasMore     bool  // 是否有后继页
	HasPrev     bool  // 是否有前驱页
}

// PageRankFromJSON 对应 Page_rank.from_json。
func PageRankFromJSON(m map[string]any) PageRank {
	currentPage := helper.JSONInt(m, "cur_page")
	totalPage := helper.JSONInt(m, "total_num")
	return PageRank{
		CurrentPage: currentPage,
		TotalPage:   totalPage,
		HasMore:     currentPage < totalPage,
		HasPrev:     currentPage > 1,
	}
}

// RankUsers 等级排行榜用户列表。
type RankUsers struct {
	classdef.Containers[*RankUser]

	Page PageRank // 页信息
	Err  error    // 捕获的异常
}

// RankUsersFromXML 对应 RankUsers.from_xml。
func RankUsersFromXML(soup *htmlutil.Node) RankUsers {
	var users RankUsers
	for _, tr := range htmlutil.FindAllTag(soup, "tr") {
		if htmlutil.HasClass(tr, "drl_list_item") || htmlutil.HasClass(tr, "drl_list_item_self") {
			u := RankUserFromXML(tr)
			users.Objs = append(users.Objs, &u)
		}
	}
	if pageItem := htmlutil.FindClass(soup, "ul", "p_rank_pager"); pageItem != nil {
		if m, err := helper.ParseJSONMap([]byte(htmlutil.Attr(pageItem, "data-field"))); err == nil {
			users.Page = PageRankFromJSON(m)
		}
	}
	return users
}

// HasMore 是否还有下一页。
func (u RankUsers) HasMore() bool { return u.Page.HasMore }
