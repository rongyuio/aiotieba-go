// Package getusercontents 实现 aiotieba 的 get_user_contents API。
//
// 对应 Python 包 aiotieba.api.get_user_contents，由 get_posts / get_threads 子包共用。
package getusercontents

import (
	"net/url"

	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/logging"
	"github.com/rongyuio/aiotieba-go/protobuf"

	pb "github.com/rongyuio/aiotieba-go/api/get_user_contents/protobuf"
)

// FragVoiceUp 音频碎片。
type FragVoiceUp struct {
	MD5      string  // 音频md5
	Duration float64 // 音频长度 以秒为单位
}

// FragVoiceUpFromAbstract 对应 Abstract 消息的 FragVoice_up.from_proto
// （此处的 during_time 是字符串）。
func FragVoiceUpFromAbstract(p *protobuf.PostInfoList_PostInfoContent_Abstract) FragVoiceUp {
	return FragVoiceUp{
		MD5:      p.GetVoiceMd5(),
		Duration: float64(classdef.ParseInt64OrZero(p.GetDuringTime())) / 1000,
	}
}

// ContentsUp 内容碎片列表。
type ContentsUp struct {
	classdef.Containers[any]

	Texts []classdef.Fragment // 纯文本碎片列表
	Links []classdef.FragLink // 链接碎片列表
	Voice FragVoiceUp         // 音频碎片
}

// ContentsUpFromProto 对应 Contents_up.from_proto。dataProto 是一个
// PostInfoContent，其 post_content 是 Abstract 碎片列表。
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
		case 10: // 语音
			c.Voice = FragVoiceUpFromAbstract(a)
		default:
			c.Objs = append(c.Objs, classdef.FragUnknownFromAny(a))
		}
	}
	return c
}

// Text 对应 text 缓存属性。
func (c ContentsUp) Text() string { return classdef.FragmentTextOf(c.Texts) }

func fragLinkFromAbstract(a *protobuf.PostInfoList_PostInfoContent_Abstract) classdef.FragLink {
	raw, err := url.Parse(a.GetLink())
	if err != nil {
		raw = &url.URL{}
	}
	return classdef.FragLink{Text: a.GetLink(), Title: a.GetText(), RawURL: raw}
}

// UserInfoU 用户信息。
type UserInfoU struct {
	UserID      int64  // user_id
	Portrait    string // portrait
	UserName    string // 用户名
	NickNameNew string // 新版昵称
}

// UserInfoUFromProto 对应 UserInfo_u.from_proto。
func UserInfoUFromProto(p *protobuf.PostInfoList) UserInfoU {
	return UserInfoU{
		UserID:      p.GetUserId(),
		Portrait:    classdef.TrimPortrait(p.GetUserPortrait()),
		UserName:    p.GetUserName(),
		NickNameNew: p.GetNameShow(),
	}
}

// NickName 用户昵称。
func (u UserInfoU) NickName() string { return u.NickNameNew }

// ShowName 显示名称。
func (u UserInfoU) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// UserPost 用户历史回复信息。
type UserPost struct {
	Contents   ContentsUp // 正文内容碎片列表
	FID        int64      // 所在吧id
	TID        int64      // 所在主题帖id
	PID        int64      // 回复id
	User       UserInfoU  // 发布者的用户信息
	IsComment  bool       // 是否为楼中楼
	CreateTime int64      // 创建时间 10位时间戳 以秒为单位
}

// UserPostFromProto 对应 UserPost.from_proto。dataProto 是一个 PostInfoContent。
func UserPostFromProto(p *protobuf.PostInfoList_PostInfoContent) UserPost {
	return UserPost{
		Contents:   ContentsUpFromProto(p),
		PID:        int64(p.GetPostId()),
		IsComment:  p.GetPostType() != 0,
		CreateTime: int64(p.GetCreateTime()),
	}
}

// Text 文本内容。
func (p UserPost) Text() string { return p.Contents.Text() }

// AuthorID 发布者的user_id。
func (p UserPost) AuthorID() int64 { return p.User.UserID }

// UserPosts 用户历史回复信息列表。
type UserPosts struct {
	classdef.Containers[*UserPost]

	FID int64 // 所在吧id
	TID int64 // 所在主题帖id
}

// UserPostsFromProto 对应 UserPosts.from_proto。
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

// UserPostss 用户历史回复信息列表的列表。
type UserPostss struct {
	classdef.Containers[*UserPosts]
	Err error // 捕获的异常
}

// UserPostssFromProto 对应 UserPostss.from_proto。
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

// ContentsUt 内容碎片列表。
type ContentsUt struct {
	classdef.Containers[any]

	Texts  []classdef.Fragment  // 纯文本碎片列表
	Emojis []classdef.FragEmoji // 表情碎片列表
	Imgs   []classdef.FragImage // 图像碎片列表
	Ats    []classdef.FragAt    // @碎片列表
	Links  []classdef.FragLink  // 链接碎片列表
	Video  classdef.FragVideo   // 视频碎片
	Voice  classdef.FragVoice   // 音频碎片
}

// ContentsUtFromProto 对应 Contents_ut.from_proto。
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
			// 图像由 media 字段承载。
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
			// 视频与语音由专用字段承载。
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

// Text 对应 text 缓存属性。
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

// UserThread 主题帖信息。
type UserThread struct {
	Contents   ContentsUt        // 正文内容碎片列表
	Title      string            // 标题内容
	FID        int64             // 所在吧id
	FName      string            // 所在贴吧名
	TID        int64             // 主题帖tid
	PID        int64             // 首楼回复pid
	User       UserInfoU         // 发布者的用户信息
	Type       enums.ThreadType  // 帖子类型
	VoteInfo   classdef.VoteInfo // 投票信息
	ViewNum    int64             // 浏览量
	ReplyNum   int64             // 回复数
	ShareNum   int64             // 分享数
	Agree      int64             // 点赞数
	Disagree   int64             // 点踩数
	CreateTime int64             // 创建时间 10位时间戳 以秒为单位
}

// UserThreadFromProto 对应 UserThread.from_proto。
func UserThreadFromProto(p *protobuf.PostInfoList) UserThread {
	typeValue := enums.ThreadTypeFrom(int(p.GetThreadType()))
	if typeValue == enums.ThreadTypeUnknown {
		logging.GetLogger().Debug().Uint64("tid", p.GetThreadId()).Uint64("type", p.GetThreadType()).Msg("unknown thread type")
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

// Text 对应 text 缓存属性。
func (t UserThread) Text() string {
	if t.Title != "" {
		return t.Title + "\n" + t.Contents.Text()
	}
	return t.Contents.Text()
}

// UserThreads 用户发布主题帖列表。
type UserThreads struct {
	classdef.Containers[*UserThread]
	Err error // 捕获的异常
}

// UserThreadsFromProto 对应 UserThreads.from_proto。
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
