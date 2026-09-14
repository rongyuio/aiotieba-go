// Package getrecoverinfo implements the get_recover_info API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_recover_info.
package getrecoverinfo

import (
	"strings"

	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper"
)

// FragTextRI mirrors FragText_ri.
type FragTextRI struct {
	Text string
}

// FragTextRIFromJSON mirrors FragText_ri.from_json.
func FragTextRIFromJSON(m map[string]any) FragTextRI {
	return FragTextRI{Text: helper.JSONStr(m, "value")}
}

// FragImageRI mirrors FragImage_ri.
type FragImageRI struct {
	Src        string
	ShowWidth  int64
	ShowHeight int64
	Hash       string
}

// FragImageRIFromJSON mirrors FragImage_ri.from_json.
func FragImageRIFromJSON(m map[string]any) FragImageRI {
	src := helper.JSONStr(m, "url")
	return FragImageRI{
		Src:        src,
		ShowWidth:  helper.JSONInt(m, "width"),
		ShowHeight: helper.JSONInt(m, "height"),
		Hash:       classdef.ImageHash(src),
	}
}

// ContentsRI mirrors Contents_ri.
type ContentsRI struct {
	Objs  []classdef.Fragment
	Texts []FragTextRI
	Imgs  []FragImageRI
}

// ContentsRIFromJSON mirrors Contents_ri.from_json.
func ContentsRIFromJSON(m map[string]any) ContentsRI {
	var c ContentsRI
	for _, item := range helper.JSONSlice(m, "all_pics") {
		if im, ok := item.(map[string]any); ok {
			c.Imgs = append(c.Imgs, FragImageRIFromJSON(im))
		}
	}
	for _, item := range helper.JSONSlice(m, "content_detail") {
		im, ok := item.(map[string]any)
		if !ok {
			continue
		}
		switch helper.JSONInt(im, "type") {
		case 1:
			frag := FragTextRIFromJSON(im)
			c.Texts = append(c.Texts, frag)
			c.Objs = append(c.Objs, classdef.FragText{Text: frag.Text})
		case 3:
			// image placeholder, skipped
		default:
			c.Objs = append(c.Objs, classdef.FragUnknownFromJSON(im))
		}
	}
	for _, img := range c.Imgs {
		c.Objs = append(c.Objs, classdef.FragImage{
			Src:        img.Src,
			ShowWidth:  int32(img.ShowWidth),
			ShowHeight: int32(img.ShowHeight),
			Hash:       img.Hash,
		})
	}
	return c
}

// Text mirrors the text property.
func (c ContentsRI) Text() string {
	var b strings.Builder
	for _, t := range c.Texts {
		b.WriteString(t.Text)
	}
	return b.String()
}

// UserInfoRI mirrors UserInfo_ri.
type UserInfoRI struct {
	Portrait    string
	UserName    string
	NickNameNew string
}

// UserInfoRIFromJSON mirrors UserInfo_ri.from_json.
func UserInfoRIFromJSON(m map[string]any) UserInfoRI {
	return UserInfoRI{
		Portrait:    classdef.TrimPortrait(helper.JSONStr(m, "portrait")),
		UserName:    helper.JSONStr(m, "user_name"),
		NickNameNew: helper.JSONStr(m, "show_nickname"),
	}
}

// RecoverInfo mirrors RecoverInfo.
type RecoverInfo struct {
	Contents ContentsRI
	Title    string
	TID      int64
	PID      int64
	User     UserInfoRI
	Err      error
}

// RecoverInfoFromJSON mirrors RecoverInfo.from_json.
func RecoverInfoFromJSON(m map[string]any) RecoverInfo {
	threadInfo := helper.JSONMap(m, "thread_info")
	return RecoverInfo{
		Contents: ContentsRIFromJSON(threadInfo),
		Title:    helper.JSONStr(threadInfo, "title"),
		TID:      helper.JSONInt(threadInfo, "thread_id"),
		PID:      helper.JSONInt(threadInfo, "post_id"),
		User:     UserInfoRIFromJSON(helper.JSONMap(m, "user_info")),
	}
}

// Text mirrors the text property.
func (r RecoverInfo) Text() string {
	if r.Title != "" {
		return r.Title + "\n" + r.Contents.Text()
	}
	return r.Contents.Text()
}
