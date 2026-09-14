// Package getbawupostlogs implements the get_bawu_postlogs API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_bawu_postlogs.
package getbawupostlogs

import (
	"strings"
	"time"

	"github.com/rongyuio/aiotieba/api/classdef"
	"github.com/rongyuio/aiotieba/helper/htmlutil"
)

// MediaPostlog mirrors Media_postlog.
type MediaPostlog struct {
	Src       string
	OriginSrc string
	Hash      string
}

// MediaPostlogFromXML mirrors Media_postlog.from_xml.
func MediaPostlogFromXML(a *htmlutil.Node) MediaPostlog {
	var m MediaPostlog
	if img := htmlutil.FirstChildTag(a, "img"); img != nil {
		m.Src = htmlutil.Attr(img, "original")
		m.Hash = classdef.ImageHash(m.Src)
	}
	m.OriginSrc = htmlutil.Attr(a, "href")
	return m
}

// BawuPostLog mirrors BawuPostLog.
type BawuPostLog struct {
	Text         string
	Title        string
	Medias       []MediaPostlog
	TID          int64
	PID          int64
	OpType       string
	PostPortrait string
	PostTime     time.Time
	OpUserName   string
	OpTime       time.Time
}

// BawuPostLogFromXML mirrors BawuPostLog.from_xml.
func BawuPostLogFromXML(tr *htmlutil.Node) BawuPostLog {
	var p BawuPostLog

	leftCell := htmlutil.FirstChildTag(tr, "td")

	postMeta := htmlutil.FindClass(leftCell, "div", "post_meta")
	if postUser := htmlutil.FirstChildTag(postMeta, "div"); postUser != nil {
		if a := htmlutil.FirstChildTag(postUser, "a"); a != nil {
			if href := htmlutil.Attr(a, "href"); len(href) > 17 {
				p.PostPortrait = href[14 : len(href)-17]
			}
		}
	}
	if timeItem := htmlutil.FirstChildTag(postMeta, "time"); timeItem != nil {
		s := htmlutil.Text(timeItem)
		// The format is "MM-DD HH:MM" without a year; the year is fixed to 1904.
		if len(s) >= 10 {
			month := htmlutil.Atoi(s[0:2])
			day := htmlutil.Atoi(s[3:5])
			hour := htmlutil.Atoi(s[7:9])
			minute := htmlutil.Atoi(s[10:])
			p.PostTime = time.Date(1904, time.Month(month), int(day), int(hour), int(minute), 0, 0, time.UTC)
		}
	}

	postContent := postMeta.NextSibling
	if titleA := htmlutil.FirstChildTag(htmlutil.FirstChildTag(postContent, "h1"), "a"); titleA != nil {
		url := htmlutil.Attr(titleA, "href")
		if q := strings.Index(url, "?"); q >= 3 {
			p.TID = htmlutil.Atoi(url[3:q])
		}
		if h := strings.LastIndex(url, "#"); h >= 0 {
			p.PID = htmlutil.Atoi(url[h+1:])
		}
		p.Title = htmlutil.Attr(titleA, "title")
	}

	textItem := htmlutil.FirstChildTag(postContent, "div")
	text := dropPrefix(htmlutil.String(textItem), 12)
	if p.PID == p.TID || !strings.HasPrefix(p.Title, "回复：") {
		p.PID = 0
		p.Text = p.Title + "\n" + text
	} else {
		p.Title = strings.TrimPrefix(p.Title, "回复：")
		p.Text = text
	}

	if mediaList := textItem.NextSibling; mediaList != nil {
		for _, a := range htmlutil.FindAllTag(mediaList, "a") {
			p.Medias = append(p.Medias, MediaPostlogFromXML(a))
		}
	}

	opTypeItem := leftCell.NextSibling
	p.OpType = htmlutil.String(opTypeItem)

	opUserNameItem := opTypeItem.NextSibling
	p.OpUserName = htmlutil.String(opUserNameItem)

	if opTimeItem := opUserNameItem.NextSibling; opTimeItem != nil {
		if t, err := time.Parse("2006-01-0215:04", htmlutil.Text(opTimeItem)); err == nil {
			p.OpTime = t
		}
	}

	return p
}

// dropPrefix mirrors the `s[12:]` slice of the Python module, on rune boundaries.
func dropPrefix(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return ""
	}
	return string(r[n:])
}

// PagePostlog mirrors Page_postlog.
type PagePostlog struct {
	CurrentPage int64
	TotalPage   int64
	TotalCount  int64
	HasMore     bool
	HasPrev     bool
}

// PagePostlogFromXML mirrors Page_postlog.from_xml.
func PagePostlogFromXML(soup *htmlutil.Node) PagePostlog {
	var p PagePostlog
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

// BawuPostLogs mirrors BawuPostLogs.
type BawuPostLogs struct {
	classdef.Containers[*BawuPostLog]

	Page PagePostlog
	Err  error
}

// BawuPostLogsFromXML mirrors BawuPostLogs.from_xml.
func BawuPostLogsFromXML(soup *htmlutil.Node) BawuPostLogs {
	var logs BawuPostLogs
	if tbody := htmlutil.Find(soup, "tbody"); tbody != nil {
		for _, tr := range htmlutil.FindAllTag(tbody, "tr") {
			l := BawuPostLogFromXML(tr)
			logs.Objs = append(logs.Objs, &l)
		}
	}
	logs.Page = PagePostlogFromXML(soup)
	return logs
}

// HasMore mirrors the has_more property.
func (l BawuPostLogs) HasMore() bool { return l.Page.HasMore }
