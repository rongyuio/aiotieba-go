// Package getbawumemberlist implements the get_bawu_memberlist API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_bawu_memberlist.
package getbawumemberlist

import (
	"strings"
	"time"

	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper/htmlutil"
)

// BawuListMemberUser mirrors BawuListMemberUser.
type BawuListMemberUser struct {
	UserID    int64
	Portrait  string
	UserName  string
	Exp       int64
	Level     int64
	ThreadNum int64
	GoodNum   int64
	JoinTime  time.Time
}

// BawuListMemberUserFromXML mirrors BawuListMemberUser.from_xml.
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

// BawuListMemberUsers mirrors BawuListMemberUsers.
type BawuListMemberUsers struct {
	classdef.Containers[*BawuListMemberUser]
	Err error
}

// BawuListMemberUsersFromXML mirrors BawuListMemberUsers.from_xml.
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
