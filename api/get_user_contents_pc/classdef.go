// Package getusercontentpc implements the get_user_contents_pc API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_user_contents_pc.
package getusercontentpc

import (
	"github.com/rongyuio/aiotieba/api/classdef"
	"github.com/rongyuio/aiotieba/helper"
)

// FragVoiceUp mirrors FragVoice_up.
type FragVoiceUp struct {
	MD5      string
	Duration float64
}

// FragVoiceUpFromJSON mirrors FragVoice_up.from_json.
func FragVoiceUpFromJSON(m map[string]any) FragVoiceUp {
	return FragVoiceUp{
		MD5:      helper.JSONStr(m, "voice_md5"),
		Duration: float64(helper.JSONInt(m, "during_time")) / 1000,
	}
}

// ContentsPcup mirrors Contents_pcup.
type ContentsPcup struct {
	Objs  []classdef.Fragment
	Texts []classdef.Fragment
	Links []classdef.FragLink
	Voice FragVoiceUp
}

// ContentsPcupFromJSON mirrors Contents_pcup.from_json.
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
		case 10: // voice
			c.Voice = FragVoiceUpFromJSON(im)
		default:
			c.Objs = append(c.Objs, classdef.FragUnknownFromJSON(im))
		}
	}
	return c
}

// Text mirrors the text property.
func (c ContentsPcup) Text() string { return classdef.FragmentTextOf(c.Texts) }

// UserInfoPcu mirrors UserInfo_pcu.
type UserInfoPcu struct {
	UserID      int64
	Portrait    string
	UserName    string
	NickNameNew string
}

// UserInfoPcuFromJSON mirrors UserInfo_pcu.from_json.
func UserInfoPcuFromJSON(m map[string]any) UserInfoPcu {
	return UserInfoPcu{
		UserID:      helper.JSONInt(m, "id"),
		Portrait:    classdef.TrimPortrait(helper.JSONStr(m, "portrait")),
		UserName:    helper.JSONStr(m, "name"),
		NickNameNew: helper.JSONStr(m, "name_show"),
	}
}

// PcUserPost mirrors PcUserPost.
type PcUserPost struct {
	Contents   ContentsPcup
	FID        int64
	TID        int64
	PID        int64
	User       UserInfoPcu
	CreateTime int64
}

// PcUserPostFromJSON mirrors PcUserPost.from_json.
func PcUserPostFromJSON(m map[string]any) PcUserPost {
	postInfo := helper.JSONMap(m, "post_info")
	return PcUserPost{
		Contents:   ContentsPcupFromJSON(postInfo),
		PID:        helper.JSONInt(postInfo, "id"),
		CreateTime: helper.JSONInt(postInfo, "time"),
	}
}

// Text mirrors the text property.
func (p PcUserPost) Text() string { return p.Contents.Text() }

// AuthorID mirrors the author_id property.
func (p PcUserPost) AuthorID() int64 { return p.User.UserID }

// PcUserPosts mirrors PcUserPosts.
type PcUserPosts struct {
	classdef.Containers[*PcUserPost]
	Err error
}

// PcUserPostsFromJSON mirrors PcUserPosts.from_json.
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
