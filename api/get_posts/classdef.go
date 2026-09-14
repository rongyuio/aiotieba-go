// Package getposts 实现 aiotieba 的 get_posts API。
//
// 对应 Python 包 aiotieba.api.get_posts。
package getposts

import (
	"strconv"
	"strings"

	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/logging"
	"github.com/rongyuio/aiotieba-go/protobuf"

	pb "github.com/rongyuio/aiotieba-go/api/get_posts/protobuf"
)

// FragImageP 图像碎片。
type FragImageP struct {
	Src        string // 小图链接
	BigSrc     string // 大图链接
	OriginSrc  string // 原图链接
	OriginSize int64  // 原图大小
	ShowWidth  int32  // 图像在客户端预览显示的宽度
	ShowHeight int32  // 图像在客户端预览显示的高度
	Hash       string // 百度图床hash
}

// FragImagePFromProto 对应 FragImage_p.from_proto（输入：PbContent）。
func FragImagePFromProto(p *protobuf.PbContent) FragImageP {
	src := p.GetCdnSrc()
	width, height := classdef.SplitSize(p.GetBsize())
	return FragImageP{
		Src:        src,
		BigSrc:     p.GetBigCdnSrc(),
		OriginSrc:  p.GetOriginSrc(),
		OriginSize: int64(p.GetOriginSize()),
		ShowWidth:  width,
		ShowHeight: height,
		Hash:       classdef.ImageHash(src),
	}
}

// FragVideoP 视频碎片。
type FragVideoP struct {
	Src      string // 视频链接
	CoverSrc string // 封面链接
	Duration int64  // 视频长度 以秒为单位
	Width    int64  // 视频宽度
	Height   int64  // 视频高度
	ViewNum  int64  // 浏览次数
}

// FragVideoPFromProto 对应 FragVideo_p.from_proto（输入：PbContent）。
func FragVideoPFromProto(p *protobuf.PbContent) FragVideoP {
	return FragVideoP{
		Src:      p.GetLink(),
		CoverSrc: p.GetSrc(),
		Duration: int64(p.GetDuringTime()),
		Width:    int64(p.GetWidth()),
		Height:   int64(p.GetHeight()),
		ViewNum:  int64(p.GetCount()),
	}
}

// Valid 对应 FragVideo_p 的 __bool__。
func (f FragVideoP) Valid() bool { return f.Width != 0 }

// ContentsP 内容碎片列表。
type ContentsP struct {
	classdef.Containers[any]

	Texts       []classdef.Fragment      // 纯文本碎片列表
	Emojis      []classdef.FragEmoji     // 表情碎片列表
	Imgs        []FragImageP             // 图像碎片列表
	Ats         []classdef.FragAt        // @碎片列表
	Links       []classdef.FragLink      // 链接碎片列表
	TiebaPluses []classdef.FragTiebaPlus // 贴吧plus碎片列表
	Video       FragVideoP               // 视频碎片
	Voice       classdef.FragVoice       // 音频碎片
}

// ContentsPFromProto 对应 Contents_p.from_proto（输入：Post）。
func ContentsPFromProto(p *protobuf.Post) ContentsP {
	c := ContentsP{}
	for _, proto := range p.GetContent() {
		switch t := proto.GetType(); {
		case t == 0 || t == 9 || t == 18 || t == 27 || t == 40:
			frag := classdef.FragTextFromProto(proto)
			c.Texts = append(c.Texts, frag)
			c.Objs = append(c.Objs, frag)
		case t == 2 || t == 11:
			frag := classdef.FragEmojiFromProto(proto)
			c.Emojis = append(c.Emojis, frag)
			c.Objs = append(c.Objs, frag)
		case t == 3 || t == 20:
			frag := FragImagePFromProto(proto)
			c.Imgs = append(c.Imgs, frag)
			c.Objs = append(c.Objs, frag)
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
		case t == 10:
			frag := classdef.FragVoiceFromProto(&protobuf.Voice{
				VoiceMd5:   proto.GetVoiceMd5(),
				DuringTime: int32(proto.GetDuringTime()),
			})
			c.Voice = frag
			c.Objs = append(c.Objs, frag)
		case t == 5:
			frag := FragVideoPFromProto(proto)
			c.Video = frag
			c.Objs = append(c.Objs, frag)
		case t == 35 || t == 36 || t == 37:
			frag := classdef.FragTiebaPlusFromProto(proto)
			c.TiebaPluses = append(c.TiebaPluses, frag)
			c.Texts = append(c.Texts, frag)
			c.Objs = append(c.Objs, frag)
		case t == 34, t == 52:
			// 过期的贴吧plus和投票碎片会被跳过。
		default:
			c.Objs = append(c.Objs, classdef.FragUnknownFromProto(proto))
		}
	}
	return c
}

// Text 文本内容。
func (c ContentsP) Text() string { return classdef.FragmentTextOf(c.Texts) }

// ContentsPc 内容碎片列表。
type ContentsPc struct {
	classdef.Containers[any]

	Texts       []classdef.Fragment      // 纯文本碎片列表
	Emojis      []classdef.FragEmoji     // 表情碎片列表
	Ats         []classdef.FragAt        // @碎片列表
	Links       []classdef.FragLink      // 链接碎片列表
	TiebaPluses []classdef.FragTiebaPlus // 贴吧plus碎片列表
	Voice       classdef.FragVoice       // 音频碎片
}

// ContentsPcFromProto 对应 Contents_pc.from_proto（输入：SubPostList）。
func ContentsPcFromProto(p *protobuf.SubPostList) ContentsPc {
	c := ContentsPc{}
	for _, proto := range p.GetContent() {
		switch t := proto.GetType(); {
		case t == 0 || t == 9 || t == 18 || t == 27:
			frag := classdef.FragTextFromProto(proto)
			c.Texts = append(c.Texts, frag)
			c.Objs = append(c.Objs, frag)
		case t == 2 || t == 11:
			frag := classdef.FragEmojiFromProto(proto)
			c.Emojis = append(c.Emojis, frag)
			c.Objs = append(c.Objs, frag)
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
		case t == 10:
			frag := classdef.FragVoiceFromProto(&protobuf.Voice{
				VoiceMd5:   proto.GetVoiceMd5(),
				DuringTime: int32(proto.GetDuringTime()),
			})
			c.Voice = frag
			c.Objs = append(c.Objs, frag)
		case t == 35 || t == 36 || t == 37:
			frag := classdef.FragTiebaPlusFromProto(proto)
			c.TiebaPluses = append(c.TiebaPluses, frag)
			c.Texts = append(c.Texts, frag)
			c.Objs = append(c.Objs, frag)
		case t == 34:
			// 过期的贴吧plus会被跳过。
		default:
			c.Objs = append(c.Objs, classdef.FragUnknownFromProto(proto))
		}
	}
	return c
}

// Text 文本内容。
func (c ContentsPc) Text() string { return classdef.FragmentTextOf(c.Texts) }

// UserInfoP 用户信息。
type UserInfoP struct {
	UserID      int64  // user_id
	Portrait    string // portrait
	UserName    string // 用户名
	NickNameNew string // 新版昵称

	Level  int32        // 等级
	GLevel int32        // 贴吧成长等级
	Gender enums.Gender // 性别
	IP     string       // ip归属地
	Icons  []string     // 印记信息

	IsBawu    bool            // 是否吧务
	IsVIP     bool            // 是否超级会员
	IsGod     bool            // 是否大神
	PrivLike  enums.PrivLike  // 关注吧列表的公开状态
	PrivReply enums.PrivReply // 帖子评论权限
}

// UserInfoPFromProto 对应 UserInfo_p.from_proto（输入：User）。
func UserInfoPFromProto(p *protobuf.User) UserInfoP {
	return UserInfoP{
		UserID:      p.GetId(),
		Portrait:    classdef.TrimPortrait(p.GetPortrait()),
		UserName:    p.GetName(),
		NickNameNew: p.GetNameShow(),
		Level:       p.GetLevelId(),
		GLevel:      int32(p.GetUserGrowth().GetLevelId()),
		Gender:      enums.Gender(p.GetGender()),
		IP:          p.GetIpAddress(),
		Icons:       classdef.UserIcons(p),
		IsBawu:      p.GetIsBawu() != 0,
		IsVIP:       len(p.GetNewTshowIcon()) != 0,
		IsGod:       p.GetNewGodData().GetStatus() != 0,
		PrivLike:    classdef.UserPrivLike(p),
		PrivReply:   classdef.UserPrivReply(p),
	}
}

// String 对应 __str__。
func (u UserInfoP) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// NickName 用户昵称。
func (u UserInfoP) NickName() string { return u.NickNameNew }

// ShowName 显示名称。
func (u UserInfoP) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// LogName 用于在日志中记录用户信息。
func (u UserInfoP) LogName() string {
	switch {
	case u.UserName != "":
		return u.UserName
	case u.Portrait != "":
		return u.NickNameNew + "/" + u.Portrait
	default:
		return strconv.FormatInt(u.UserID, 10)
	}
}

// Valid 对应 __bool__。
func (u UserInfoP) Valid() bool { return u.UserID != 0 }

// Equal 对应 __eq__。
func (u UserInfoP) Equal(other UserInfoP) bool { return u.UserID == other.UserID }

// CommentP 楼中楼信息。
type CommentP struct {
	Contents ContentsPc // 正文内容碎片列表

	FID   int64     // 所在吧id
	FName string    // 所在贴吧名
	TID   int64     // 所在主题帖id
	PPID  int64     // 所在楼层id
	PID   int64     // 楼中楼id
	User  UserInfoP // 发布者的用户信息

	AuthorID  int64 // 发布者的user_id
	ReplyToID int64 // 被回复者的user_id

	Floor          int64 // 所在楼层数
	Agree          int64 // 点赞数
	Disagree       int64 // 点踩数
	CreateTime     int64 // 创建时间 10位时间戳 以秒为单位
	IsThreadAuthor bool  // 是否楼主
}

// CommentPFromProto 对应 Comment_p.from_proto（输入：SubPostList）。
func CommentPFromProto(p *protobuf.SubPostList) *CommentP {
	contents := ContentsPcFromProto(p)

	var replyToID int64
	if len(contents.Objs) > 0 {
		if first, ok := contents.Objs[0].(classdef.FragText); ok && first.Text == "回复 " && len(p.GetContent()) > 1 {
			if uid := p.GetContent()[1].GetUid(); uid != 0 {
				replyToID = uid
				if len(contents.Objs) > 1 {
					if _, isAt := contents.Objs[1].(classdef.FragAt); isAt && len(contents.Ats) > 0 {
						contents.Ats = contents.Ats[1:]
					}
				}
				contents.Objs = dropFirstTwo(contents.Objs)
				contents.Texts = dropFirstTwo(contents.Texts)
				if len(contents.Texts) > 0 {
					if t, ok := contents.Texts[0].(classdef.FragText); ok {
						t.Text = strings.TrimPrefix(t.Text, " :")
						contents.Texts[0] = t
					}
				}
			}
		}
	}

	return &CommentP{
		Contents:   contents,
		PID:        p.GetId(),
		AuthorID:   p.GetAuthorId(),
		ReplyToID:  replyToID,
		Agree:      p.GetAgree().GetAgreeNum(),
		Disagree:   p.GetAgree().GetDisagreeNum(),
		CreateTime: int64(p.GetTime()),
	}
}

// Text 文本内容。
func (c *CommentP) Text() string { return c.Contents.Text() }

// Equal 对应 __eq__。
func (c *CommentP) Equal(other *CommentP) bool { return other != nil && c.PID == other.PID }

// Post 楼层信息。
type Post struct {
	Contents ContentsP  // 正文内容碎片列表
	Sign     string     // 小尾巴文本内容
	Comments []CommentP // 楼中楼列表
	IsAIMeme bool       // 是否是AI生成的表情包

	FID   int64     // 所在吧id
	FName string    // 所在贴吧名
	TID   int64     // 所在主题帖id
	PID   int64     // 回复id
	User  UserInfoP // 发布者的用户信息

	AuthorID       int64 // 发布者的user_id
	Floor          int64 // 楼层数
	ReplyNum       int64 // 楼中楼数
	Agree          int64 // 点赞数
	Disagree       int64 // 点踩数
	CreateTime     int64 // 创建时间
	IsThreadAuthor bool  // 是否楼主
}

// PostFromProto 对应 Post.from_proto（输入：Post）。
func PostFromProto(p *protobuf.Post) *Post {
	contents := ContentsPFromProto(p)

	var sign strings.Builder
	for _, part := range p.GetSignature().GetContent() {
		if part.GetType() == 0 {
			sign.WriteString(part.GetText())
		}
	}

	comments := make([]CommentP, 0, len(p.GetSubPostList().GetSubPostList()))
	for _, proto := range p.GetSubPostList().GetSubPostList() {
		comments = append(comments, *CommentPFromProto(proto))
	}

	return &Post{
		Contents:   contents,
		Sign:       sign.String(),
		Comments:   comments,
		IsAIMeme:   p.GetSpriteMemeInfo().GetMemeId() != 0,
		PID:        p.GetId(),
		AuthorID:   p.GetAuthorId(),
		Floor:      int64(p.GetFloor()),
		ReplyNum:   int64(p.GetSubPostNumber()),
		Agree:      p.GetAgree().GetAgreeNum(),
		Disagree:   p.GetAgree().GetDisagreeNum(),
		CreateTime: int64(p.GetTime()),
	}
}

// Text 文本内容。
func (p *Post) Text() string {
	if p.Sign != "" {
		return p.Contents.Text() + "\n" + p.Sign
	}
	return p.Contents.Text()
}

// Equal 对应 __eq__。
func (p *Post) Equal(other *Post) bool { return other != nil && p.PID == other.PID }

// PageP 页信息。
type PageP struct {
	PageSize    int32 // 页大小
	CurrentPage int32 // 当前页码
	TotalPage   int32 // 总页码
	TotalCount  int32 // 总计数
	HasMore     bool  // 是否有后继页
	HasPrev     bool  // 是否有前驱页
}

// PagePFromProto 对应 Page_p.from_proto。
func PagePFromProto(p *protobuf.Page) PageP {
	return PageP{
		PageSize:    p.GetPageSize(),
		CurrentPage: p.GetCurrentPage(),
		TotalPage:   p.GetTotalPage(),
		TotalCount:  p.GetTotalCount(),
		HasMore:     p.GetHasMore() != 0,
		HasPrev:     p.GetHasPrev() != 0,
	}
}

// ForumP 吧信息。
type ForumP struct {
	FID   int64  // 贴吧id
	FName string // 贴吧名

	Category    string // 一级分类
	Subcategory string // 二级分类

	MemberNum int64 // 吧会员数
	PostNum   int64 // 发帖数
}

// ForumPFromProto 对应 Forum_p.from_proto（输入：SimpleForum）。
func ForumPFromProto(p *protobuf.SimpleForum) ForumP {
	return ForumP{
		FID:         p.GetId(),
		FName:       p.GetName(),
		Category:    p.GetFirstClass(),
		Subcategory: p.GetSecondClass(),
		MemberNum:   int64(p.GetMemberNum()),
		PostNum:     int64(p.GetPostNum()),
	}
}

// FragImagePt 图像碎片。
type FragImagePt struct {
	Src        string // 小图链接 宽580px
	BigSrc     string // 大图链接 宽720px
	OriginSrc  string // 原图链接
	ShowWidth  int32  // 图像在客户端预览显示的宽度
	ShowHeight int32  // 图像在客户端预览显示的高度
	Hash       string // 百度图床hash
}

// FragImagePtFromProto 对应 FragImage_pt.from_proto（输入：Media）。
func FragImagePtFromProto(p *protobuf.Media) FragImagePt {
	src := p.GetWaterPic()
	return FragImagePt{
		Src:        src,
		BigSrc:     p.GetSmallPic(),
		OriginSrc:  p.GetBigPic(),
		ShowWidth:  int32(p.GetWidth()),
		ShowHeight: int32(p.GetHeight()),
		Hash:       classdef.ImageHash(src),
	}
}

// ContentsPt 内容碎片列表。
type ContentsPt struct {
	classdef.Containers[any]

	Texts       []classdef.Fragment      // 纯文本碎片列表
	Emojis      []classdef.FragEmoji     // 表情碎片列表
	Imgs        []FragImagePt            // 图像碎片列表
	Ats         []classdef.FragAt        // @碎片列表
	Links       []classdef.FragLink      // 链接碎片列表
	TiebaPluses []classdef.FragTiebaPlus // 贴吧plus碎片列表
	Video       classdef.FragVideo       // 视频碎片
	Voice       classdef.FragVoice       // 音频碎片
}

// ContentsPtFromProto 对应 Contents_pt.from_proto（输入：ThreadInfo.OriginThreadInfo）。
func ContentsPtFromProto(p *protobuf.ThreadInfo_OriginThreadInfo) ContentsPt {
	c := ContentsPt{}
	for _, media := range p.GetMedia() {
		c.Imgs = append(c.Imgs, FragImagePtFromProto(media))
	}

	for _, proto := range p.GetContent() {
		switch t := proto.GetType(); {
		case t == 0 || t == 9 || t == 18 || t == 27:
			frag := classdef.FragTextFromProto(proto)
			c.Texts = append(c.Texts, frag)
			c.Objs = append(c.Objs, frag)
		case t == 2 || t == 11:
			frag := classdef.FragEmojiFromProto(proto)
			c.Emojis = append(c.Emojis, frag)
			c.Objs = append(c.Objs, frag)
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
		case t == 35 || t == 36 || t == 37:
			frag := classdef.FragTiebaPlusFromProto(proto)
			c.TiebaPluses = append(c.TiebaPluses, frag)
			c.Texts = append(c.Texts, frag)
			c.Objs = append(c.Objs, frag)
		case t == 34:
			// 过期的贴吧plus会被跳过。
		default:
			c.Objs = append(c.Objs, classdef.FragUnknownFromProto(proto))
		}
	}

	// 第一个 @碎片会被移除，随后追加媒体图像。
	if len(c.Ats) > 0 {
		c.Ats = c.Ats[1:]
		c.Objs = dropFirst(c.Objs)
	}
	c.Objs = append(c.Objs, imgsToAny(c.Imgs)...)

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

// Text 文本内容。
func (c ContentsPt) Text() string { return classdef.FragmentTextOf(c.Texts) }

// UserInfoPt 用户信息。
//
// 与 UserInfoP 不同，它没有性别字段。
type UserInfoPt struct {
	UserID      int64  // user_id
	Portrait    string // portrait
	UserName    string // 用户名
	NickNameNew string // 新版昵称

	Level  int32    // 等级
	GLevel int32    // 贴吧成长等级
	IP     string   // ip归属地
	Icons  []string // 印记信息

	IsBawu    bool            // 是否吧务
	IsVIP     bool            // 是否超级会员
	IsGod     bool            // 是否大神
	PrivLike  enums.PrivLike  // 关注吧列表的公开状态
	PrivReply enums.PrivReply // 帖子评论权限
}

// UserInfoPtFromProto 对应 UserInfo_pt.from_proto（输入：User）。
func UserInfoPtFromProto(p *protobuf.User) UserInfoPt {
	return UserInfoPt{
		UserID:      p.GetId(),
		Portrait:    classdef.TrimPortrait(p.GetPortrait()),
		UserName:    p.GetName(),
		NickNameNew: p.GetNameShow(),
		Level:       p.GetLevelId(),
		GLevel:      int32(p.GetUserGrowth().GetLevelId()),
		IP:          p.GetIpAddress(),
		Icons:       classdef.UserIcons(p),
		IsBawu:      p.GetIsBawu() != 0,
		IsVIP:       len(p.GetNewTshowIcon()) != 0,
		IsGod:       p.GetNewGodData().GetStatus() != 0,
		PrivLike:    classdef.UserPrivLike(p),
		PrivReply:   classdef.UserPrivReply(p),
	}
}

// String 对应 __str__。
func (u UserInfoPt) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// NickName 用户昵称。
func (u UserInfoPt) NickName() string { return u.NickNameNew }

// ShowName 显示名称。
func (u UserInfoPt) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// LogName 用于在日志中记录用户信息。
func (u UserInfoPt) LogName() string {
	switch {
	case u.UserName != "":
		return u.UserName
	case u.Portrait != "":
		return u.NickNameNew + "/" + u.Portrait
	default:
		return strconv.FormatInt(u.UserID, 10)
	}
}

// Valid 对应 __bool__。
func (u UserInfoPt) Valid() bool { return u.UserID != 0 }

// ShareThreadPt 被分享的主题帖信息。
type ShareThreadPt struct {
	Contents ContentsPt // 正文内容碎片列表
	Title    string     // 标题内容

	FID      int64  // 所在吧id
	FName    string // 所在贴吧名
	TID      int64  // 主题帖tid
	AuthorID int64  // 发布者的user_id

	VoteInfo classdef.VoteInfo // 投票内容
}

// ShareThreadPtFromProto 对应 ShareThread_pt.from_proto。
func ShareThreadPtFromProto(p *protobuf.ThreadInfo_OriginThreadInfo) ShareThreadPt {
	var authorID int64
	if len(p.GetContent()) != 0 {
		authorID = p.GetContent()[0].GetUid()
	}
	return ShareThreadPt{
		Contents: ContentsPtFromProto(p),
		Title:    p.GetTitle(),
		FID:      p.GetFid(),
		FName:    p.GetFname(),
		TID:      classdef.ParseInt64OrZero(p.GetTid()),
		AuthorID: authorID,
		VoteInfo: classdef.VoteInfoFromProto(p.GetPollInfo()),
	}
}

// Text 文本内容。
func (s ShareThreadPt) Text() string {
	if s.Title != "" {
		return s.Title + "\n" + s.Contents.Text()
	}
	return s.Contents.Text()
}

// ThreadP 主题帖信息。
type ThreadP struct {
	Contents ContentsPt // 正文内容碎片列表
	Title    string     // 标题内容

	FID   int64      // 所在吧id
	FName string     // 所在贴吧名
	TID   int64      // 主题帖tid
	PID   int64      // 首楼回复pid
	User  UserInfoPt // 发布者的用户信息

	Type    enums.ThreadType // 帖子类型
	IsShare bool             // 是否分享帖

	VoteInfo    classdef.VoteInfo // 投票信息
	ShareOrigin ShareThreadPt     // 转发来的原帖内容
	ViewNum     int64             // 浏览量
	ReplyNum    int64             // 回复数
	ShareNum    int64             // 分享数
	Agree       int64             // 点赞数
	Disagree    int64             // 点踩数
	CreateTime  int64             // 创建时间 10位时间戳 以秒为单位
}

// ThreadPFromProto 对应 Thread_p.from_proto（输入：PbPageResIdl.DataRes）。
func ThreadPFromProto(p *pb.PbPageResIdl_DataRes) ThreadP {
	thread := p.GetThread()
	tid := thread.GetId()

	typeValue := enums.ThreadTypeFrom(int(thread.GetThreadType()))
	if typeValue == enums.ThreadTypeUnknown {
		logging.GetLogger().Debug("unknown thread type", "tid", tid, "type", thread.GetThreadType())
	}

	isShare := thread.GetIsShareThread() != 0

	var (
		contents    ContentsPt
		voteInfo    classdef.VoteInfo
		shareOrigin ShareThreadPt
	)
	if !isShare {
		origin := thread.GetOriginThreadInfo()
		contents = ContentsPtFromProto(origin)
		voteInfo = classdef.VoteInfoFromProto(origin.GetPollInfo())
	} else {
		shareOrigin = ShareThreadPtFromProto(thread.GetOriginThreadInfo())
	}

	return ThreadP{
		Contents:    contents,
		Title:       thread.GetTitle(),
		TID:         tid,
		PID:         thread.GetPostId(),
		User:        UserInfoPtFromProto(thread.GetAuthor()),
		Type:        typeValue,
		IsShare:     isShare,
		VoteInfo:    voteInfo,
		ShareOrigin: shareOrigin,
		ViewNum:     int64(p.GetThreadFreqNum()),
		ReplyNum:    int64(thread.GetReplyNum()),
		ShareNum:    int64(thread.GetShareNum()),
		Agree:       thread.GetAgree().GetAgreeNum(),
		Disagree:    thread.GetAgree().GetDisagreeNum(),
		CreateTime:  int64(thread.GetCreateTime()),
	}
}

// Text 文本内容。
func (t ThreadP) Text() string {
	if t.Title != "" {
		return t.Title + "\n" + t.Contents.Text()
	}
	return t.Contents.Text()
}

// AuthorID 发布者的user_id。
func (t ThreadP) AuthorID() int64 { return t.User.UserID }

// Equal 对应 __eq__。
func (t ThreadP) Equal(other ThreadP) bool { return t.PID == other.PID }

// Posts 回复列表。
//
// Posts 是 get_posts 的返回结果。
type Posts struct {
	classdef.Containers[*Post]

	Page   PageP   // 页信息
	Forum  ForumP  // 所在吧信息
	Thread ThreadP // 所在主题帖信息
	Err    error   // 捕获的异常
}

// PostsFromProto 对应 Posts.from_proto（输入：PbPageResIdl.DataRes）。
func PostsFromProto(p *pb.PbPageResIdl_DataRes) Posts {
	page := PagePFromProto(p.GetPage())
	forum := ForumPFromProto(p.GetForum())
	thread := ThreadPFromProto(p)
	thread.FID = forum.FID
	thread.FName = forum.FName

	posts := make([]*Post, 0, len(p.GetPostList()))
	for _, proto := range p.GetPostList() {
		if proto.GetChatContent().GetBotUk() != "" {
			continue
		}
		posts = append(posts, PostFromProto(proto))
	}

	users := make(map[int64]UserInfoP, len(p.GetUserList()))
	for _, proto := range p.GetUserList() {
		users[proto.GetId()] = UserInfoPFromProto(proto)
	}

	for _, post := range posts {
		post.FID = forum.FID
		post.FName = forum.FName
		post.TID = thread.TID
		post.User = users[post.AuthorID]
		post.IsThreadAuthor = thread.AuthorID() == post.AuthorID
		for i := range post.Comments {
			comment := &post.Comments[i]
			comment.FID = post.FID
			comment.FName = post.FName
			comment.TID = post.TID
			comment.PPID = post.PID
			comment.Floor = post.Floor
			comment.User = users[comment.AuthorID]
			comment.IsThreadAuthor = thread.AuthorID() == comment.AuthorID
		}
	}

	return Posts{
		Containers: classdef.Containers[*Post]{Objs: posts},
		Page:       page,
		Forum:      forum,
		Thread:     thread,
	}
}

// HasMore 是否还有下一页。
func (p Posts) HasMore() bool { return p.Page.HasMore }

func dropFirst[T any](s []T) []T {
	if len(s) == 0 {
		return nil
	}
	return s[1:]
}

func dropFirstTwo[T any](s []T) []T {
	if len(s) <= 2 {
		return nil
	}
	return s[2:]
}

func imgsToAny(imgs []FragImagePt) []any {
	out := make([]any, 0, len(imgs))
	for _, img := range imgs {
		out = append(out, img)
	}
	return out
}
