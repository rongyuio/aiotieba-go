// Package getbawumemberlist 实现 aiotieba 的 get_bawu_memberlist API。
//
// 对应 Python 包 aiotieba.api.get_bawu_memberlist。
package getbawumemberlist

import (
	"strings"
	"time"

	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper/htmlutil"
)

// BawuListMemberUser 吧会员信息。
type BawuListMemberUser struct {
	UserID    int64     // user_id
	Portrait  string    // portrait
	UserName  string    // 用户名
	Exp       int64     // 经验值
	Level     int64     // 等级
	ThreadNum int64     // 主题帖数
	GoodNum   int64     // 精品帖数
	JoinTime  time.Time // 关注时间
}

// BawuListMemberUserFromXML 对应 BawuListMemberUser.from_xml。
func BawuListMemberUserFromXML(tr *htmlutil.Node) BawuListMemberUser {
	var u BawuListMemberUser

	leftCell := htmlutil.FirstChildTag(tr, "td")
	if a := htmlutil.FirstChildTag(leftCell, "a"); a != nil {
		u.UserName = strings.TrimLeft(htmlutil.Text(a), " \t\r\n")
	}

	expItem := leftCell.NextSibling
	if expItem != nil {
		expItem = expItem.NextSibling
	}
	u.Exp = htmlutil.Atoi(htmlutil.String(expItem))

	levelItem := expItem.NextSibling
	u.Level = htmlutil.Atoi(htmlutil.String(levelItem))

	threadNumItem := levelItem.NextSibling
	u.ThreadNum = htmlutil.Atoi(htmlutil.String(threadNumItem))

	goodNumItem := threadNumItem.NextSibling
	u.GoodNum = htmlutil.Atoi(htmlutil.String(goodNumItem))

	joinTimeItem := goodNumItem.NextSibling
	if t, err := time.Parse("2006-01-02 15:04", htmlutil.String(joinTimeItem)); err == nil {
		u.JoinTime = t
	}

	btnGroupItem := joinTimeItem.NextSibling
	u.UserID = htmlutil.Atoi(htmlutil.Attr(btnGroupItem, "id"))
	u.Portrait = htmlutil.Attr(btnGroupItem, "portrait")

	return u
}

// BawuListMemberUsers 吧会员列表。
type BawuListMemberUsers struct {
	classdef.Containers[*BawuListMemberUser]
	Err error // 捕获的异常
}

// BawuListMemberUsersFromXML 对应 BawuListMemberUsers.from_xml。
func BawuListMemberUsersFromXML(soup *htmlutil.Node) BawuListMemberUsers {
	var users BawuListMemberUsers
	if tbody := htmlutil.Find(soup, "tbody"); tbody != nil {
		for _, tr := range htmlutil.FindAllTag(tbody, "tr") {
			u := BawuListMemberUserFromXML(tr)
			users.Objs = append(users.Objs, &u)
		}
	}
	return users
}
