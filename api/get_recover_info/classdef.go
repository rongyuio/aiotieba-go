// Package getrecoverinfo 实现 aiotieba 的 get_recover_info API。
//
// 对应 Python 包 aiotieba.api.get_recover_info。
package getrecoverinfo

import (
	"strings"

	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper"
)

// FragTextRI 纯文本碎片。
type FragTextRI struct {
	Text string // 文本内容
}

// FragTextRIFromJSON 对应 FragText_ri.from_json。
func FragTextRIFromJSON(m map[string]any) FragTextRI {
	return FragTextRI{Text: helper.JSONStr(m, "value")}
}

// FragImageRI 图像碎片。
type FragImageRI struct {
	Src        string // 小图链接 宽720px
	ShowWidth  int64  // 图像在客户端预览显示的宽度
	ShowHeight int64  // 图像在客户端预览显示的高度
	Hash       string // 百度图床hash
}

// FragImageRIFromJSON 对应 FragImage_ri.from_json。
func FragImageRIFromJSON(m map[string]any) FragImageRI {
	src := helper.JSONStr(m, "url")
	return FragImageRI{
		Src:        src,
		ShowWidth:  helper.JSONInt(m, "width"),
		ShowHeight: helper.JSONInt(m, "height"),
		Hash:       classdef.ImageHash(src),
	}
}

// ContentsRI 内容碎片列表。
type ContentsRI struct {
	Objs  []classdef.Fragment // 所有内容碎片的混合列表
	Texts []FragTextRI        // 纯文本碎片列表
	Imgs  []FragImageRI       // 图像碎片列表
}

// ContentsRIFromJSON 对应 Contents_ri.from_json。
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
			// 图像占位符，跳过
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

// Text 文本内容。
func (c ContentsRI) Text() string {
	var b strings.Builder
	for _, t := range c.Texts {
		b.WriteString(t.Text)
	}
	return b.String()
}

// UserInfoRI 用户信息。
type UserInfoRI struct {
	Portrait    string // portrait
	UserName    string // 用户名
	NickNameNew string // 新版昵称
}

// UserInfoRIFromJSON 对应 UserInfo_ri.from_json。
func UserInfoRIFromJSON(m map[string]any) UserInfoRI {
	return UserInfoRI{
		Portrait:    classdef.TrimPortrait(helper.JSONStr(m, "portrait")),
		UserName:    helper.JSONStr(m, "user_name"),
		NickNameNew: helper.JSONStr(m, "show_nickname"),
	}
}

// RecoverInfo 待恢复帖子信息。
type RecoverInfo struct {
	Contents ContentsRI // 正文内容碎片列表
	Title    string     // 标题内容
	TID      int64      // 所在主题帖id
	PID      int64      // 回复id
	User     UserInfoRI // 发布者的用户信息
	Err      error      // 捕获的异常
}

// RecoverInfoFromJSON 对应 RecoverInfo.from_json。
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

// Text 文本内容。
func (r RecoverInfo) Text() string {
	if r.Title != "" {
		return r.Title + "\n" + r.Contents.Text()
	}
	return r.Contents.Text()
}
