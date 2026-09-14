// Package getrankusers implements the get_rank_users API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_rank_users.
package getrankusers

import (
	"strings"

	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/htmlutil"
)

// RankUser mirrors RankUser.
type RankUser struct {
	UserName string
	Level    int64
	Exp      int64
	IsVIP    bool
}

// RankUserFromXML mirrors RankUser.from_xml.
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

// PageRank mirrors Page_rank.
type PageRank struct {
	CurrentPage int64
	TotalPage   int64
	HasMore     bool
	HasPrev     bool
}

// PageRankFromJSON mirrors Page_rank.from_json.
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

// RankUsers mirrors RankUsers.
type RankUsers struct {
	classdef.Containers[*RankUser]

	Page PageRank
	Err  error
}

// RankUsersFromXML mirrors RankUsers.from_xml.
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

// HasMore mirrors the has_more property.
func (u RankUsers) HasMore() bool { return u.Page.HasMore }
