// Package getusercontentpc 实现 aiotieba 的 get_user_contents_pc API。
//
// 对应 Python 包 aiotieba.api.get_user_contents_pc。
package getusercontentpc

import (
	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper"
)

// FragVoiceUp 音频碎片。
type FragVoiceUp struct {
	MD5      string  // 音频md5
	Duration float64 // 音频长度 以秒为单位
}

// FragVoiceUpFromJSON 对应 FragVoice_up.from_json。
func FragVoiceUpFromJSON(m map[string]any) FragVoiceUp {
	return FragVoiceUp{
		MD5:      helper.JSONStr(m, "voice_md5"),
		Duration: float64(helper.JSONInt(m, "during_time")) / 1000,
	}
}

// ContentsPcup 内容碎片列表。
type ContentsPcup struct {
	Objs  []classdef.Fragment // 所有内容碎片的混合列表
	Texts []classdef.Fragment // 纯文本碎片列表
	Links []classdef.FragLink // 链接碎片列表
	Voice FragVoiceUp         // 音频碎片
}

// ContentsPcupFromJSON 对应 Contents_pcup.from_json。
func ContentsPcupFromJSON(m map[string]any) ContentsPcup {
	var c ContentsPcup
	for _, item := range helper.JSONSlice(m, "content") {
		im, ok := item.(map[string]any)
		if !ok {
			continue
		}
		switch helper.JSONInt(im, "type") {
		case 0, 4:
			frag := classdef.FragTextFromJSON(im)
			c.Texts = append(c.Texts, frag)
			c.Objs = append(c.Objs, frag)
		case 1:
			frag := classdef.FragLinkFromJSON(im)
			c.Links = append(c.Links, frag)
			c.Texts = append(c.Texts, frag)
			c.Objs = append(c.Objs, frag)
		case 10: // 语音
			c.Voice = FragVoiceUpFromJSON(im)
		default:
			c.Objs = append(c.Objs, classdef.FragUnknownFromJSON(im))
		}
	}
	return c
}

// Text 文本内容。
func (c ContentsPcup) Text() string { return classdef.FragmentTextOf(c.Texts) }

// UserInfoPcu 用户信息。
type UserInfoPcu struct {
	UserID      int64  // user_id
	Portrait    string // portrait
	UserName    string // 用户名
	NickNameNew string // 新版昵称
}

// UserInfoPcuFromJSON 对应 UserInfo_pcu.from_json。
func UserInfoPcuFromJSON(m map[string]any) UserInfoPcu {
	return UserInfoPcu{
		UserID:      helper.JSONInt(m, "id"),
		Portrait:    classdef.TrimPortrait(helper.JSONStr(m, "portrait")),
		UserName:    helper.JSONStr(m, "name"),
		NickNameNew: helper.JSONStr(m, "name_show"),
	}
}

// PcUserPost 用户历史回复信息。
type PcUserPost struct {
	Contents   ContentsPcup // 正文内容碎片列表
	FID        int64        // 所在吧id
	TID        int64        // 所在主题帖id
	PID        int64        // 回复id
	User       UserInfoPcu  // 发布者的用户信息
	CreateTime int64        // 创建时间 10位时间戳 以秒为单位
}

// PcUserPostFromJSON 对应 PcUserPost.from_json。
func PcUserPostFromJSON(m map[string]any) PcUserPost {
	postInfo := helper.JSONMap(m, "post_info")
	return PcUserPost{
		Contents:   ContentsPcupFromJSON(postInfo),
		PID:        helper.JSONInt(postInfo, "id"),
		CreateTime: helper.JSONInt(postInfo, "time"),
	}
}

// Text 文本内容。
func (p PcUserPost) Text() string { return p.Contents.Text() }

// AuthorID 发布者的user_id。
func (p PcUserPost) AuthorID() int64 { return p.User.UserID }

// PcUserPosts 用户历史回复信息列表。
type PcUserPosts struct {
	classdef.Containers[*PcUserPost]
	Err error
}

// PcUserPostsFromJSON 对应 PcUserPosts.from_json。
func PcUserPostsFromJSON(m map[string]any) PcUserPosts {
	var posts PcUserPosts
	list := helper.JSONSlice(m, "list")
	if len(list) > 0 {
		if first, ok := list[0].(map[string]any); ok {
			if pi := helper.JSONMap(first, "post_info"); pi != nil {
				user := UserInfoPcuFromJSON(helper.JSONMap(pi, "author"))
				for _, item := range list {
					if im, ok := item.(map[string]any); ok {
						p := PcUserPostFromJSON(im)
						p.User = user
						posts.Objs = append(posts.Objs, &p)
					}
				}
			}
		}
	}
	return posts
}
