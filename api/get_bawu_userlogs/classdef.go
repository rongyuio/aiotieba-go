// Package getbawuuserlogs 实现 aiotieba 的 get_bawu_userlogs API。
//
// 对应 Python 包 aiotieba.api.get_bawu_userlogs。
package getbawuuserlogs

import (
	"strings"
	"time"

	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper/htmlutil"
)

// BawuUserLog 吧务用户管理日志。
type BawuUserLog struct {
	OpType       string    // 操作类型
	OpDuration   int64     // 操作作用时长
	UserPortrait string    // 被操作用户的portrait
	OpUserName   string    // 操作人用户名
	OpTime       time.Time // 操作时间
}

// BawuUserLogFromXML 对应 BawuUserLog.from_xml。
func BawuUserLogFromXML(tr *htmlutil.Node) BawuUserLog {
	var l BawuUserLog

	leftCell := htmlutil.FirstChildTag(tr, "td")
	if a := htmlutil.FirstChildTag(leftCell, "a"); a != nil {
		if href := htmlutil.Attr(a, "href"); len(href) > 17 {
			l.UserPortrait = href[14 : len(href)-17]
		}
	}

	opTypeItem := leftCell.NextSibling
	if opTypeItem != nil {
		opTypeItem = opTypeItem.NextSibling
	}
	l.OpType = htmlutil.String(opTypeItem)

	opDurationItem := opTypeItem.NextSibling
	duration := strings.ReplaceAll(htmlutil.String(opDurationItem), " ", "")
	if strings.Contains(duration, "天") {
		l.OpDuration = htmlutil.Atoi(strings.TrimSuffix(duration, "天"))
	}

	opUserNameItem := opDurationItem.NextSibling
	l.OpUserName = htmlutil.String(opUserNameItem)

	if opTimeItem := opUserNameItem.NextSibling; opTimeItem != nil {
		if t, err := time.Parse("2006-01-02 15:04", htmlutil.String(opTimeItem)); err == nil {
			l.OpTime = t
		}
	}

	return l
}

// PageUserlog 页信息。
type PageUserlog struct {
	CurrentPage int64 // 当前页码
	TotalPage   int64 // 总页码
	TotalCount  int64 // 总计数
	HasMore     bool  // 是否有后继页
	HasPrev     bool  // 是否有前驱页
}

// PageUserlogFromXML 对应 Page_userlog.from_xml。
func PageUserlogFromXML(soup *htmlutil.Node) PageUserlog {
	var p PageUserlog
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
		if pageTag.Parent != nil && pageTag.Parent.NextSibling != nil {
			t := htmlutil.Text(pageTag.Parent.NextSibling)
			if len(t) >= 2 {
				p.TotalPage = htmlutil.Atoi(t[1 : len(t)-1])
			}
		}
	}
	p.HasMore = p.CurrentPage < p.TotalPage
	p.HasPrev = p.CurrentPage > 1
	return p
}

// BawuUserLogs 吧务用户管理日志表。
type BawuUserLogs struct {
	classdef.Containers[*BawuUserLog]

	Page PageUserlog // 页信息
	Err  error       // 捕获的异常
}

// BawuUserLogsFromXML 对应 BawuUserLogs.from_xml。
func BawuUserLogsFromXML(soup *htmlutil.Node) BawuUserLogs {
	var logs BawuUserLogs
	if tbody := htmlutil.Find(soup, "tbody"); tbody != nil {
		for _, tr := range htmlutil.FindAllTag(tbody, "tr") {
			l := BawuUserLogFromXML(tr)
			logs.Objs = append(logs.Objs, &l)
		}
	}
	logs.Page = PageUserlogFromXML(soup)
	return logs
}

// HasMore 是否还有下一页。
func (l BawuUserLogs) HasMore() bool { return l.Page.HasMore }
