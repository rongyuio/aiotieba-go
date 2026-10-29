// Package getusercontentpc 实现 aiotieba 的 get_user_contents_pc API。
//
// 对应 Python 包 aiotieba.api.get_user_contents_pc。
package getusercontentpc

import (
	"strconv"
	"strings"

	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/logging"
)

// FragImagePcup 图像碎片。
type FragImagePcup struct {
	Src        string // 小图链接 宽720px
	BigSrc     string // 大图链接 宽960px
	OriginSrc  string // 原图链接
	ShowWidth  int32  // 图像在客户端预览显示的宽度
	ShowHeight int32  // 图像在客户端预览显示的高度
	Hash       string // 百度图床hash
}

// FragImagePcupFromJSON 对应 FragImage_pcup.from_json。
func FragImagePcupFromJSON(m map[string]any) FragImagePcup {
	src := helper.JSONStr(m, "cdn_src")
	showWidth, showHeight := parseShowSize(helper.JSONStr(m, "bsize"))
	return FragImagePcup{
		Src:        src,
		BigSrc:     helper.JSONStr(m, "big_cdn_src"),
		OriginSrc:  helper.JSONStr(m, "origin_src"),
		ShowWidth:  showWidth,
		ShowHeight: showHeight,
		Hash:       classdef.ImageHash(src),
	}
}

// parseShowSize 拆分 "width,height" 形式的 bsize。Python 原版用 str.partition(",")，
// 格式非法时由 int() 抛错；这里退化为 0。
func parseShowSize(bsize string) (int32, int32) {
	widthStr, heightStr, _ := strings.Cut(bsize, ",")
	return int32(classdef.ParseInt64OrZero(widthStr)), int32(classdef.ParseInt64OrZero(heightStr))
}

// FragAtPcup @碎片。
//
// 嵌入 classdef.FragAt 以复用文本内容与 user_id 字段，同时满足 classdef.Fragment，
// 因而可以同时进入 Texts 与 Objs。
type FragAtPcup struct {
	classdef.FragAt        // 文本内容与被@用户的user_id
	Portrait        string // 被@用户的portrait
}

// FragAtPcupFromJSON 对应 FragAt_pcup.from_json。
func FragAtPcupFromJSON(m map[string]any) FragAtPcup {
	return FragAtPcup{
		FragAt: classdef.FragAt{
			Text:   helper.JSONStr(m, "text"),
			UserID: helper.JSONInt(m, "uid"),
		},
		Portrait: classdef.TrimPortrait(helper.JSONStr(m, "portrait")),
	}
}

// FragVideoPcup 视频碎片。
type FragVideoPcup struct {
	Src      string // 视频链接
	CoverSrc string // 封面链接
	Duration int64  // 视频长度 以秒为单位
	Width    int64  // 视频宽度
	Height   int64  // 视频高度
}

// FragVideoPcupFromJSON 对应 FragVideo_pcup.from_json。
func FragVideoPcupFromJSON(m map[string]any) FragVideoPcup {
	return FragVideoPcup{
		Src:      helper.JSONStr(m, "link"),
		CoverSrc: helper.JSONStr(m, "src"),
		Duration: helper.JSONInt(m, "during_time"),
		Width:    helper.JSONInt(m, "width"),
		Height:   helper.JSONInt(m, "height"),
	}
}

// Valid 对应 FragVideo_pcup 的 __bool__。
func (f FragVideoPcup) Valid() bool { return f.Width != 0 }

// FragVoicePcup 音频碎片。
type FragVoicePcup struct {
	MD5      string  // 音频md5
	Duration float64 // 音频长度 以秒为单位
}

// FragVoicePcupFromJSON 对应 FragVoice_pcup.from_json。
func FragVoicePcupFromJSON(m map[string]any) FragVoicePcup {
	return FragVoicePcup{
		MD5:      helper.JSONStr(m, "voice_md5"),
		Duration: float64(helper.JSONInt(m, "during_time")) / 1000,
	}
}

// Valid 对应 FragVoice_pcup 的 __bool__。
func (f FragVoicePcup) Valid() bool { return f.MD5 != "" }

// ContentsPcup 内容碎片列表。
type ContentsPcup struct {
	classdef.Containers[any]

	Texts  []classdef.Fragment  // 纯文本碎片列表
	Emojis []classdef.FragEmoji // 表情碎片列表
	Imgs   []FragImagePcup      // 图像碎片列表
	Ats    []FragAtPcup         // @碎片列表
	Links  []classdef.FragLink  // 链接碎片列表
	Video  FragVideoPcup        // 视频碎片
	Voice  FragVoicePcup        // 音频碎片
}

// ContentsPcupFromJSON 对应 Contents_pcup.from_json。参数是 content 碎片列表本身，
// 不是含 content 键的对象。
func ContentsPcupFromJSON(items []any) ContentsPcup {
	var c ContentsPcup
	for _, item := range items {
		im, ok := item.(map[string]any)
		if !ok {
			continue
		}
		switch helper.JSONInt(im, "type") {
		// 0纯文本 9电话号 18话题 27百科词条 40梗百科
		case 0, 9, 18, 27, 40:
			frag := classdef.FragTextFromJSON(im)
			c.Texts = append(c.Texts, frag)
			c.Objs = append(c.Objs, frag)
		// 表情。type 11 是带 tid 的旧表情
		case 2, 11:
			frag := classdef.FragEmoji{ID: helper.JSONStr(im, "text"), Desc: helper.JSONStr(im, "c")}
			c.Emojis = append(c.Emojis, frag)
			c.Objs = append(c.Objs, frag)
		// 图像。type 20 是带 tid 的旧图像
		case 3, 20:
			frag := FragImagePcupFromJSON(im)
			c.Imgs = append(c.Imgs, frag)
			c.Objs = append(c.Objs, frag)
		case 4:
			frag := FragAtPcupFromJSON(im)
			c.Ats = append(c.Ats, frag)
			c.Texts = append(c.Texts, frag)
			c.Objs = append(c.Objs, frag)
		case 1:
			frag := classdef.FragLinkFromJSON(im)
			c.Links = append(c.Links, frag)
			c.Texts = append(c.Texts, frag)
			c.Objs = append(c.Objs, frag)
		case 5: // 视频
			c.Video = FragVideoPcupFromJSON(im)
			c.Objs = append(c.Objs, c.Video)
		case 10: // 音频
			c.Voice = FragVoicePcupFromJSON(im)
			c.Objs = append(c.Objs, c.Voice)
		default:
			c.Objs = append(c.Objs, classdef.FragUnknownFromJSON(im))
		}
	}
	return c
}

// Text 文本内容，对应 text 缓存属性。
func (c ContentsPcup) Text() string {
	var b strings.Builder
	for _, frag := range c.Texts {
		b.WriteString(fragmentText(frag))
	}
	return b.String()
}

// fragmentText 取出类文本碎片的文本。不能用 classdef.FragmentTextOf，
// 因为它不认识本包的 FragAtPcup。
func fragmentText(frag classdef.Fragment) string {
	switch v := frag.(type) {
	case classdef.FragText:
		return v.Text
	case classdef.FragLink:
		return v.Text
	case FragAtPcup:
		return v.Text
	default:
		return ""
	}
}

// UserInfoPcu 用户信息。
type UserInfoPcu struct {
	UserID      int64  // user_id
	Portrait    string // portrait
	UserName    string // 用户名
	NickNameNew string // 新版昵称

	Gender enums.Gender // 性别

	PrivLike  enums.PrivLike  // 关注吧列表的公开状态
	PrivReply enums.PrivReply // 帖子评论权限
}

// UserInfoPcuFromJSON 对应 UserInfo_pcu.from_json。
func UserInfoPcuFromJSON(m map[string]any) UserInfoPcu {
	privSets := helper.JSONMap(m, "priv_sets")

	privLike := enums.PrivLikePublic
	if v := helper.JSONInt(privSets, "like"); v != 0 {
		privLike = enums.PrivLikeFrom(int(v))
	}
	privReply := enums.PrivReplyAll
	if v := helper.JSONInt(privSets, "reply"); v != 0 {
		privReply = enums.PrivReplyFrom(int(v))
	}

	return UserInfoPcu{
		UserID:      helper.JSONInt(m, "id"),
		Portrait:    classdef.TrimPortrait(helper.JSONStr(m, "portrait")),
		UserName:    helper.JSONStr(m, "name"),
		NickNameNew: helper.JSONStr(m, "name_show"),
		Gender:      enums.GenderFrom(int(helper.JSONInt(m, "gender"))),
		PrivLike:    privLike,
		PrivReply:   privReply,
	}
}

// String 对应 __str__。
func (u UserInfoPcu) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// Valid 对应 __bool__。
func (u UserInfoPcu) Valid() bool { return u.UserID != 0 }

// NickName 用户昵称。
func (u UserInfoPcu) NickName() string { return u.NickNameNew }

// ShowName 显示名称。
func (u UserInfoPcu) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// LogName 用于在日志中记录用户信息，对应 log_name 缓存属性。
func (u UserInfoPcu) LogName() string {
	switch {
	case u.UserName != "":
		return u.UserName
	case u.Portrait != "":
		return u.NickNameNew + "/" + u.Portrait
	default:
		return strconv.FormatInt(u.UserID, 10)
	}
}

// ForumPcup 父级贴吧信息。
type ForumPcup struct {
	FID   int64  // 贴吧id
	FName string // 贴吧名

	MemberNum int64 // 会员数
	PostNum   int64 // 帖子数

	IsLiked bool // 是否已关注
}

// ForumPcupFromJSON 对应 Forum_pcup.from_json。
func ForumPcupFromJSON(m map[string]any) ForumPcup {
	return ForumPcup{
		FID:       helper.JSONInt(m, "id"),
		FName:     helper.JSONStr(m, "name"),
		MemberNum: helper.JSONInt(m, "member_num"),
		PostNum:   helper.JSONInt(m, "post_num"),
		IsLiked:   helper.JSONBool(m, "is_liked"),
	}
}

// Valid 对应 __bool__。
func (f ForumPcup) Valid() bool { return f.FID != 0 }

// ThreadPcup 父级主题帖信息。
type ThreadPcup struct {
	Contents ContentsPcup // 首楼正文内容碎片列表
	Title    string       // 标题内容

	FID   int64       // 所在吧id
	FName string      // 所在贴吧名
	TID   int64       // 主题帖tid
	PID   int64       // 首楼回复pid
	User  UserInfoPcu // 发布者的用户信息
	Forum ForumPcup   // 所在贴吧信息

	Type       enums.ThreadType // 帖子类型
	ViewNum    int64            // 浏览量
	ReplyNum   int64            // 回复数
	ShareNum   int64            // 分享数
	Agree      int64            // 点赞数
	Disagree   int64            // 点踩数
	CreateTime int64            // 创建时间 10位时间戳 以秒为单位
	LastTime   int64            // 最后回复时间 10位时间戳 以秒为单位
	IsGood     bool             // 是否精品帖
	IsTop      bool             // 是否置顶帖
	IsDeleted  bool             // 是否已删除
}

// ThreadPcupFromJSON 对应 Thread_pcup.from_json。
func ThreadPcupFromJSON(m map[string]any) ThreadPcup {
	threadType := helper.JSONInt(m, "thread_type")
	typeValue := enums.ThreadTypeFrom(int(threadType))
	if typeValue == enums.ThreadTypeUnknown {
		logging.GetLogger().Debug().Int64("tid", helper.JSONInt(m, "tid")).Int64("type", threadType).Msg("unknown thread type")
	}

	agree := helper.JSONMap(m, "agree")

	return ThreadPcup{
		Contents:   ContentsPcupFromJSON(helper.JSONSlice(m, "first_post_content")),
		Title:      helper.JSONStr(m, "title"),
		FID:        helper.JSONInt(m, "fid"),
		FName:      helper.JSONStr(m, "fname"),
		TID:        helper.JSONInt(m, "tid"),
		PID:        helper.JSONInt(m, "first_post_id"),
		User:       UserInfoPcuFromJSON(helper.JSONMap(m, "author")),
		Forum:      ForumPcupFromJSON(helper.JSONMap(m, "forum_info")),
		Type:       typeValue,
		ViewNum:    helper.JSONInt(m, "view_num"),
		ReplyNum:   helper.JSONInt(m, "reply_num"),
		ShareNum:   helper.JSONInt(m, "share_num"),
		Agree:      helper.JSONInt(agree, "agree_num"),
		Disagree:   helper.JSONInt(agree, "disagree_num"),
		CreateTime: helper.JSONInt(m, "create_time"),
		LastTime:   helper.JSONInt(m, "last_time_int"),
		IsGood:     helper.JSONBool(m, "is_good"),
		IsTop:      helper.JSONBool(m, "is_top"),
		IsDeleted:  helper.JSONBool(m, "is_deleted"),
	}
}

// Valid 对应 __bool__。
func (t ThreadPcup) Valid() bool { return t.TID != 0 }

// Text 对应 text 缓存属性。
func (t ThreadPcup) Text() string {
	if t.Title != "" {
		return t.Title + "\n" + t.Contents.Text()
	}
	return t.Contents.Text()
}

// AuthorID 发布者的user_id。
func (t ThreadPcup) AuthorID() int64 { return t.User.UserID }

// PcUserPost 用户历史回复信息。
type PcUserPost struct {
	Contents ContentsPcup // 正文内容碎片列表

	FID    int64       // 所在吧id
	PPID   int64       // 父级回复id
	PID    int64       // 回复id
	User   UserInfoPcu // 发布者的用户信息
	Thread ThreadPcup  // 父级主题帖信息

	ContentType enums.ContentType // 帖子对象类型

	CreateTime int64 // 创建时间 10位时间戳 以秒为单位
}

// PcUserPostFromJSON 对应 PcUserPost.from_json。
//
// post_info 为空表示该条目是主题帖本身，此时退化为转发 thread_info 的首楼信息。
func PcUserPostFromJSON(m map[string]any) PcUserPost {
	thread := ThreadPcupFromJSON(helper.JSONMap(m, "thread_info"))

	var contents ContentsPcup
	var ppid, pid, createTime int64
	var user UserInfoPcu

	if postInfo := helper.JSONMap(m, "post_info"); len(postInfo) != 0 {
		contents = ContentsPcupFromJSON(helper.JSONSlice(postInfo, "content"))
		ppid = helper.JSONInt(postInfo, "quote_id")
		pid = helper.JSONInt(postInfo, "id")
		user = UserInfoPcuFromJSON(helper.JSONMap(postInfo, "author"))
		createTime = helper.JSONInt(postInfo, "time")
	} else {
		contents = thread.Contents
		ppid = 0
		pid = thread.PID
		user = thread.User
		createTime = thread.CreateTime
	}

	return PcUserPost{
		Contents:    contents,
		FID:         thread.FID,
		PPID:        ppid,
		PID:         pid,
		User:        user,
		Thread:      thread,
		ContentType: enums.ContentTypeFrom(int(helper.JSONInt(m, "type"))),
		CreateTime:  createTime,
	}
}

// Text 文本内容。
func (p PcUserPost) Text() string { return p.Contents.Text() }

// AuthorID 发布者的user_id。
func (p PcUserPost) AuthorID() int64 { return p.User.UserID }

// PcUserPosts 用户历史回复信息列表。
type PcUserPosts struct {
	classdef.Containers[*PcUserPost]
	Err error // 捕获的异常
}

// PcUserPostsFromJSON 对应 PcUserPosts.from_json。
func PcUserPostsFromJSON(m map[string]any) PcUserPosts {
	var posts PcUserPosts
	for _, item := range helper.JSONSlice(m, "list") {
		if im, ok := item.(map[string]any); ok {
			p := PcUserPostFromJSON(im)
			posts.Objs = append(posts.Objs, &p)
		}
	}
	return posts
}
