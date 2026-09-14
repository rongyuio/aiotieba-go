// Package getusercontents implements the get_user_contents API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_user_contents and is shared by
// its get_posts / get_threads sub-packages.
package getusercontents

import (
	"net/url"

	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/logging"
	"github.com/rongyuio/aiotieba-go/protobuf"

	pb "github.com/rongyuio/aiotieba-go/api/get_user_contents/protobuf"
)

// FragVoiceUp mirrors FragVoice_up: a voice fragment carried by a post abstract.
type FragVoiceUp struct {
	MD5      string
	Duration float64
}

// FragVoiceUpFromAbstract mirrors FragVoice_up.from_proto for the Abstract
// message (during_time is a string there).
func FragVoiceUpFromAbstract(p *protobuf.PostInfoList_PostInfoContent_Abstract) FragVoiceUp {
	return FragVoiceUp{
		MD5:      p.GetVoiceMd5(),
		Duration: float64(classdef.ParseInt64OrZero(p.GetDuringTime())) / 1000,
	}
}

// ContentsUp is the body of a user post. It mirrors Contents_up.
type ContentsUp struct {
	classdef.Containers[any]

	Texts []classdef.Fragment
	Links []classdef.FragLink
	Voice FragVoiceUp
}

// ContentsUpFromProto mirrors Contents_up.from_proto. dataProto is a
// PostInfoContent, whose post_content is the list of Abstract fragments.
func ContentsUpFromProto(p *protobuf.PostInfoList_PostInfoContent) ContentsUp {
	c := ContentsUp{}
	for _, a := range p.GetPostContent() {
		switch a.GetType() {
		case 0, 4:
			frag := classdef.FragText{Text: a.GetText()}
			c.Texts = append(c.Texts, frag)
			c.Objs = append(c.Objs, frag)
		case 1:
			frag := fragLinkFromAbstract(a)
			c.Links = append(c.Links, frag)
			c.Texts = append(c.Texts, frag)
			c.Objs = append(c.Objs, frag)
		case 10: // voice
			c.Voice = FragVoiceUpFromAbstract(a)
		default:
			c.Objs = append(c.Objs, classdef.FragUnknownFromAny(a))
		}
	}
	return c
}

// Text mirrors the text cached property.
func (c ContentsUp) Text() string { return classdef.FragmentTextOf(c.Texts) }

func fragLinkFromAbstract(a *protobuf.PostInfoList_PostInfoContent_Abstract) classdef.FragLink {
	raw, err := url.Parse(a.GetLink())
	if err != nil {
		raw = &url.URL{}
	}
	return classdef.FragLink{Text: a.GetLink(), Title: a.GetText(), RawURL: raw}
}

// UserInfoU mirrors UserInfo_u.
type UserInfoU struct {
	UserID      int64
	Portrait    string
	UserName    string
	NickNameNew string
}

// UserInfoUFromProto mirrors UserInfo_u.from_proto.
func UserInfoUFromProto(p *protobuf.PostInfoList) UserInfoU {
	return UserInfoU{
		UserID:      p.GetUserId(),
		Portrait:    classdef.TrimPortrait(p.GetUserPortrait()),
		UserName:    p.GetUserName(),
		NickNameNew: p.GetNameShow(),
	}
}

// NickName mirrors the nick_name property.
func (u UserInfoU) NickName() string { return u.NickNameNew }

// ShowName mirrors the show_name property.
func (u UserInfoU) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// UserPost mirrors UserPost.
type UserPost struct {
	Contents   ContentsUp
	FID        int64
	TID        int64
	PID        int64
	User       UserInfoU
	IsComment  bool
	CreateTime int64
}

// UserPostFromProto mirrors UserPost.from_proto. dataProto is a
// PostInfoContent.
func UserPostFromProto(p *protobuf.PostInfoList_PostInfoContent) UserPost {
	return UserPost{
		Contents:   ContentsUpFromProto(p),
		PID:        int64(p.GetPostId()),
		IsComment:  p.GetPostType() != 0,
		CreateTime: int64(p.GetCreateTime()),
	}
}

// Text mirrors the text property.
func (p UserPost) Text() string { return p.Contents.Text() }

// AuthorID mirrors the author_id property.
func (p UserPost) AuthorID() int64 { return p.User.UserID }

// UserPosts mirrors UserPosts.
type UserPosts struct {
	classdef.Containers[*UserPost]

	FID int64
	TID int64
}

// UserPostsFromProto mirrors UserPosts.from_proto.
func UserPostsFromProto(p *protobuf.PostInfoList) UserPosts {
	posts := UserPosts{
		FID: int64(p.GetForumId()),
		TID: int64(p.GetThreadId()),
	}
	for _, c := range p.GetContent() {
		up := UserPostFromProto(c)
		up.FID = posts.FID
		up.TID = posts.TID
		posts.Objs = append(posts.Objs, &up)
	}
	return posts
}

// UserPostss mirrors UserPostss.
type UserPostss struct {
	classdef.Containers[*UserPosts]
	Err error
}

// UserPostssFromProto mirrors UserPostss.from_proto.
func UserPostssFromProto(p *pb.UserPostResIdl_DataRes) UserPostss {
	var result UserPostss
	list := p.GetPostList()
	for _, item := range list {
		up := UserPostsFromProto(item)
		result.Objs = append(result.Objs, &up)
	}
	if len(list) > 0 {
		user := UserInfoUFromProto(list[0])
		for _, uposts := range result.Objs {
			for _, upost := range uposts.Objs {
				upost.User = user
			}
		}
	}
	return result
}

// ContentsUt mirrors Contents_ut: the body of a user thread.
type ContentsUt struct {
	classdef.Containers[any]

	Texts  []classdef.Fragment
	Emojis []classdef.FragEmoji
	Imgs   []classdef.FragImage
	Ats    []classdef.FragAt
	Links  []classdef.FragLink
	Video  classdef.FragVideo
	Voice  classdef.FragVoice
}

// ContentsUtFromProto mirrors Contents_ut.from_proto.
func ContentsUtFromProto(p *protobuf.PostInfoList) ContentsUt {
	c := ContentsUt{}

	for _, media := range p.GetMedia() {
		if media.GetType() == 5 {
			continue
		}
		c.Imgs = append(c.Imgs, fragImageUtFromMedia(media))
	}

	for _, proto := range p.GetFirstPostContent() {
		switch t := proto.GetType(); {
		case t == 0 || t == 9 || t == 18 || t == 27:
			frag := classdef.FragTextFromProto(proto)
			c.Texts = append(c.Texts, frag)
			c.Objs = append(c.Objs, frag)
		case t == 2 || t == 11:
			frag := classdef.FragEmojiFromProto(proto)
			c.Emojis = append(c.Emojis, frag)
			c.Objs = append(c.Objs, frag)
		case t == 3 || t == 20:
			// Images are carried by the media field.
		case t == 4:
			frag := classdef.FragAtFromProto(proto)
			c.Ats = append(c.Ats, frag)
			c.Texts = append(c.Texts, frag)
			c.Objs = append(c.Objs, frag)
		case t == 1:
			frag := classdef.FragLinkFromProto(proto)
			c.Links = append(c.Links, frag)
			c.Texts = append(c.Texts, frag)
			c.Objs = append(c.Objs, frag)
		case t == 5, t == 10:
			// Video and voice are carried by dedicated fields.
		default:
			c.Objs = append(c.Objs, classdef.FragUnknownFromProto(proto))
		}
	}
	for _, img := range c.Imgs {
		c.Objs = append(c.Objs, img)
	}

	if p.GetVideoInfo().GetVideoWidth() != 0 {
		c.Video = classdef.FragVideoFromProto(p.GetVideoInfo())
		c.Objs = append(c.Objs, c.Video)
	}
	if len(p.GetVoiceInfo()) != 0 {
		c.Voice = classdef.FragVoiceFromProto(p.GetVoiceInfo()[0])
		c.Objs = append(c.Objs, c.Voice)
	}
	return c
}

// Text mirrors the text cached property.
func (c ContentsUt) Text() string { return classdef.FragmentTextOf(c.Texts) }

func fragImageUtFromMedia(m *protobuf.Media) classdef.FragImage {
	src := m.GetSmallPic()
	return classdef.FragImage{
		Src:        src,
		BigSrc:     m.GetBigPic(),
		OriginSrc:  m.GetOriginPic(),
		OriginSize: int64(m.GetOriginSize()),
		ShowWidth:  int32(m.GetWidth()),
		ShowHeight: int32(m.GetHeight()),
		Hash:       classdef.ImageHash(src),
	}
}

// UserThread mirrors UserThread.
type UserThread struct {
	Contents   ContentsUt
	Title      string
	FID        int64
	FName      string
	TID        int64
	PID        int64
	User       UserInfoU
	Type       enums.ThreadType
	VoteInfo   classdef.VoteInfo
	ViewNum    int64
	ReplyNum   int64
	ShareNum   int64
	Agree      int64
	Disagree   int64
	CreateTime int64
}

// UserThreadFromProto mirrors UserThread.from_proto.
func UserThreadFromProto(p *protobuf.PostInfoList) UserThread {
	typeValue := enums.ThreadTypeFrom(int(p.GetThreadType()))
	if typeValue == enums.ThreadTypeUnknown {
		logging.GetLogger().Debug("unknown thread type", "tid", p.GetThreadId(), "type", p.GetThreadType())
	}

	agree := p.GetAgree()
	var agreeNum, disagreeNum int64
	if agree != nil {
		agreeNum = agree.GetAgreeNum()
		disagreeNum = agree.GetDisagreeNum()
	}

	return UserThread{
		Contents:   ContentsUtFromProto(p),
		Title:      p.GetTitle(),
		FID:        int64(p.GetForumId()),
		FName:      p.GetForumName(),
		TID:        int64(p.GetThreadId()),
		PID:        int64(p.GetPostId()),
		Type:       typeValue,
		VoteInfo:   classdef.VoteInfoFromProto(p.GetPollInfo()),
		ViewNum:    int64(p.GetFreqNum()),
		ReplyNum:   int64(p.GetReplyNum()),
		ShareNum:   int64(p.GetShareNum()),
		Agree:      agreeNum,
		Disagree:   disagreeNum,
		CreateTime: int64(p.GetCreateTime()),
	}
}

// Text mirrors the text cached property.
func (t UserThread) Text() string {
	if t.Title != "" {
		return t.Title + "\n" + t.Contents.Text()
	}
	return t.Contents.Text()
}

// UserThreads mirrors UserThreads.
type UserThreads struct {
	classdef.Containers[*UserThread]
	Err error
}

// UserThreadsFromProto mirrors UserThreads.from_proto.
func UserThreadsFromProto(p *pb.UserPostResIdl_DataRes) UserThreads {
	var result UserThreads
	list := p.GetPostList()
	for _, item := range list {
		t := UserThreadFromProto(item)
		result.Objs = append(result.Objs, &t)
	}
	if len(list) > 0 {
		user := UserInfoUFromProto(list[0])
		for _, uthread := range result.Objs {
			uthread.User = user
		}
	}
	return result
}
