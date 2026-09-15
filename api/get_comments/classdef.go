// Package getcomments 实现 aiotieba 的 get_comments API。
//
// 对应 Python 包 aiotieba.api.get_comments。
package getcomments

import (
	"strconv"
	"strings"

	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/logging"
	"github.com/rongyuio/aiotieba-go/protobuf"

	pb "github.com/rongyuio/aiotieba-go/api/get_comments/protobuf"
)

// ContentsC 内容碎片列表。
type ContentsC struct {
	classdef.Containers[any]

	Texts       []classdef.Fragment      // 纯文本碎片列表
	Emojis      []classdef.FragEmoji     // 表情碎片列表
	Ats         []classdef.FragAt        // @碎片列表
	Links       []classdef.FragLink      // 链接碎片列表
	TiebaPluses []classdef.FragTiebaPlus // 贴吧plus碎片列表
	Voice       classdef.FragVoice       // 音频碎片
}

// ContentsCFromProto 对应 Contents_c.from_proto（输入：SubPostList）。
func ContentsCFromProto(p *protobuf.SubPostList) ContentsC {
	c := ContentsC{}
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
			// 过时的 tiebaplus 会被跳过。
		default:
			c.Objs = append(c.Objs, classdef.FragUnknownFromProto(proto))
		}
	}
	return c
}

// Text 文本内容。
func (c ContentsC) Text() string { return classdef.FragmentTextOf(c.Texts) }

// UserInfoC 用户信息。
type UserInfoC struct {
	UserID      int64  // user_id
	Portrait    string // portrait
	UserName    string // 用户名
	NickNameNew string // 新版昵称

	Level  int32        // 等级
	Gender enums.Gender // 性别
	Icons  []string     // 印记信息

	IsBawu    bool            // 是否吧务
	IsVIP     bool            // 是否超级会员
	IsGod     bool            // 是否大神
	PrivLike  enums.PrivLike  // 关注吧列表的公开状态
	PrivReply enums.PrivReply // 帖子评论权限
}

// UserInfoCFromProto 对应 UserInfo_c.from_proto（输入：User）。
func UserInfoCFromProto(p *protobuf.User) UserInfoC {
	return UserInfoC{
		UserID:      p.GetId(),
		Portrait:    classdef.TrimPortrait(p.GetPortrait()),
		UserName:    p.GetName(),
		NickNameNew: p.GetNameShow(),
		Level:       p.GetLevelId(),
		Gender:      enums.Gender(p.GetGender()),
		Icons:       classdef.UserIcons(p),
		IsBawu:      p.GetIsBawu() != 0,
		IsVIP:       len(p.GetNewTshowIcon()) != 0,
		IsGod:       p.GetNewGodData().GetStatus() != 0,
		PrivLike:    classdef.UserPrivLike(p),
		PrivReply:   classdef.UserPrivReply(p),
	}
}

// String 对应 __str__。
func (u UserInfoC) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// NickName 用户昵称。
func (u UserInfoC) NickName() string { return u.NickNameNew }

// ShowName 显示名称。
func (u UserInfoC) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// LogName 用于在日志中记录用户信息。
func (u UserInfoC) LogName() string {
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
func (u UserInfoC) Valid() bool { return u.UserID != 0 }

// Comment 楼中楼信息。
type Comment struct {
	Contents ContentsC // 正文内容碎片列表

	FID   int64     // 所在吧id
	FName string    // 所在贴吧名
	TID   int64     // 所在主题帖id
	PPID  int64     // 所在楼层id
	PID   int64     // 楼中楼id
	User  UserInfoC // 发布者的用户信息

	ReplyToID int64 // 被回复者的user_id

	Floor          int64 // 所在楼层数
	Agree          int64 // 点赞数
	Disagree       int64 // 点踩数
	CreateTime     int64 // 创建时间 10位时间戳 以秒为单位
	IsThreadAuthor bool  // 是否楼主
}

// CommentFromProto 对应 Comment.from_proto（输入：SubPostList）。
func CommentFromProto(p *protobuf.SubPostList) *Comment {
	contents := ContentsCFromProto(p)

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

	return &Comment{
		Contents:   contents,
		PID:        p.GetId(),
		User:       UserInfoCFromProto(p.GetAuthor()),
		ReplyToID:  replyToID,
		Agree:      p.GetAgree().GetAgreeNum(),
		Disagree:   p.GetAgree().GetDisagreeNum(),
		CreateTime: int64(p.GetTime()),
	}
}

// Text 文本内容。
func (c *Comment) Text() string { return c.Contents.Text() }

// AuthorID 发布者的user_id。
func (c *Comment) AuthorID() int64 { return c.User.UserID }

// Equal 对应 __eq__。
func (c *Comment) Equal(other *Comment) bool { return other != nil && c.PID == other.PID }

// PageC 页信息。
type PageC struct {
	PageSize    int32 // 页大小
	CurrentPage int32 // 当前页码
	TotalPage   int32 // 总页码
	TotalCount  int32 // 总计数
	HasMore     bool  // 是否有后继页
	HasPrev     bool  // 是否有前驱页
}

// PageCFromProto 对应 Page_c.from_proto。
//
// 与 Page_p 不同，它根据页码推导出 has_more/has_prev。
func PageCFromProto(p *protobuf.Page) PageC {
	current := p.GetCurrentPage()
	total := p.GetTotalPage()
	return PageC{
		PageSize:    p.GetPageSize(),
		CurrentPage: current,
		TotalPage:   total,
		TotalCount:  p.GetTotalCount(),
		HasMore:     current < total,
		HasPrev:     current > 1,
	}
}

// ForumC 吧信息。
type ForumC struct {
	FID   int64  // 贴吧id
	FName string // 贴吧名

	Category    string // 一级分类
	Subcategory string // 二级分类
}

// ForumCFromProto 对应 Forum_c.from_proto（输入：SimpleForum）。
func ForumCFromProto(p *protobuf.SimpleForum) ForumC {
	return ForumC{
		FID:         p.GetId(),
		FName:       p.GetName(),
		Category:    p.GetFirstClass(),
		Subcategory: p.GetSecondClass(),
	}
}

// UserInfoCt 用户信息。
type UserInfoCt struct {
	UserID      int64  // user_id
	Portrait    string // portrait
	UserName    string // 用户名
	NickNameNew string // 新版昵称

	Level int32 // 等级
	IsGod bool  // 是否大神
}

// UserInfoCtFromProto 对应 UserInfo_ct.from_proto（输入：User）。
func UserInfoCtFromProto(p *protobuf.User) UserInfoCt {
	return UserInfoCt{
		UserID:      p.GetId(),
		Portrait:    classdef.TrimPortrait(p.GetPortrait()),
		UserName:    p.GetName(),
		NickNameNew: p.GetNameShow(),
		Level:       p.GetLevelId(),
		IsGod:       p.GetNewGodData().GetStatus() != 0,
	}
}

// String 对应 __str__。
func (u UserInfoCt) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// NickName 用户昵称。
func (u UserInfoCt) NickName() string { return u.NickNameNew }

// ShowName 显示名称。
func (u UserInfoCt) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// LogName 用于在日志中记录用户信息。
func (u UserInfoCt) LogName() string {
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
func (u UserInfoCt) Valid() bool { return u.UserID != 0 }

// ThreadC 主题帖信息。
type ThreadC struct {
	Title string // 标题内容

	FID   int64      // 所在吧id
	FName string     // 所在贴吧名
	TID   int64      // 主题帖tid
	User  UserInfoCt // 发布者的用户信息

	Type enums.ThreadType // 帖子类型

	ReplyNum int64 // 回复数
}

// ThreadCFromProto 对应 Thread_c.from_proto（输入：ThreadInfo）。
func ThreadCFromProto(p *protobuf.ThreadInfo) ThreadC {
	typeValue := enums.ThreadTypeFrom(int(p.GetThreadType()))
	if typeValue == enums.ThreadTypeUnknown {
		logging.GetLogger().Debug().Int64("tid", p.GetId()).Int32("type", p.GetThreadType()).Msg("unknown thread type")
	}
	return ThreadC{
		Title:    p.GetTitle(),
		TID:      p.GetId(),
		User:     UserInfoCtFromProto(p.GetAuthor()),
		Type:     typeValue,
		ReplyNum: int64(p.GetReplyNum()),
	}
}

// AuthorID 发布者的user_id。
func (t ThreadC) AuthorID() int64 { return t.User.UserID }

// Equal 对应 __eq__。
func (t ThreadC) Equal(other ThreadC) bool { return t.TID == other.TID }

// FragImageCp 图像碎片。
type FragImageCp struct {
	Src        string // 小图链接 宽720px 一定是静态图
	BigSrc     string // 大图链接 宽960px
	OriginSrc  string // 原图链接
	OriginSize int64  // 原图大小
	ShowWidth  int32  // 图像在客户端预览显示的宽度
	ShowHeight int32  // 图像在客户端预览显示的高度
	Hash       string // 百度图床hash
}

// FragImageCpFromProto 对应 FragImage_cp.from_proto（输入：PbContent）。
func FragImageCpFromProto(p *protobuf.PbContent) FragImageCp {
	src := p.GetCdnSrc()
	width, height := classdef.SplitSize(p.GetBsize())
	return FragImageCp{
		Src:        src,
		BigSrc:     p.GetBigCdnSrc(),
		OriginSrc:  p.GetOriginSrc(),
		OriginSize: int64(p.GetOriginSize()),
		ShowWidth:  width,
		ShowHeight: height,
		Hash:       classdef.ImageHash(src),
	}
}

// ContentsCp 内容碎片列表。
type ContentsCp struct {
	classdef.Containers[any]

	Texts       []classdef.Fragment      // 纯文本碎片列表
	Emojis      []classdef.FragEmoji     // 表情碎片列表
	Imgs        []FragImageCp            // 图像碎片列表
	Ats         []classdef.FragAt        // @碎片列表
	Links       []classdef.FragLink      // 链接碎片列表
	TiebaPluses []classdef.FragTiebaPlus // 贴吧plus碎片列表
	Voice       classdef.FragVoice       // 音频碎片
}

// ContentsCpFromProto 对应 Contents_cp.from_proto（输入：Post）。
func ContentsCpFromProto(p *protobuf.Post) ContentsCp {
	c := ContentsCp{}
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
		case t == 3 || t == 20:
			frag := FragImageCpFromProto(proto)
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
		case t == 35 || t == 36 || t == 37:
			frag := classdef.FragTiebaPlusFromProto(proto)
			c.TiebaPluses = append(c.TiebaPluses, frag)
			c.Texts = append(c.Texts, frag)
			c.Objs = append(c.Objs, frag)
		case t == 34:
			// 过时的 tiebaplus 会被跳过。
		default:
			c.Objs = append(c.Objs, classdef.FragUnknownFromProto(proto))
		}
	}
	return c
}

// Text 文本内容。
func (c ContentsCp) Text() string { return classdef.FragmentTextOf(c.Texts) }

// UserInfoCp 用户信息。
type UserInfoCp struct {
	UserID      int64  // user_id
	Portrait    string // portrait
	UserName    string // 用户名
	NickNameNew string // 新版昵称

	Level  int32        // 等级
	Gender enums.Gender // 性别

	IsBawu    bool            // 是否吧务
	IsVIP     bool            // 是否超级会员
	IsGod     bool            // 是否大神
	PrivLike  enums.PrivLike  // 关注吧列表的公开状态
	PrivReply enums.PrivReply // 帖子评论权限
}

// UserInfoCpFromProto 对应 UserInfo_cp.from_proto（输入：User）。
func UserInfoCpFromProto(p *protobuf.User) UserInfoCp {
	return UserInfoCp{
		UserID:      p.GetId(),
		Portrait:    classdef.TrimPortrait(p.GetPortrait()),
		UserName:    p.GetName(),
		NickNameNew: p.GetNameShow(),
		Level:       p.GetLevelId(),
		Gender:      enums.Gender(p.GetGender()),
		IsBawu:      p.GetIsBawu() != 0,
		IsVIP:       len(p.GetNewTshowIcon()) != 0,
		IsGod:       p.GetNewGodData().GetStatus() != 0,
		PrivLike:    classdef.UserPrivLike(p),
		PrivReply:   classdef.UserPrivReply(p),
	}
}

// String 对应 __str__。
func (u UserInfoCp) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// NickName 用户昵称。
func (u UserInfoCp) NickName() string { return u.NickNameNew }

// ShowName 显示名称。
func (u UserInfoCp) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// LogName 用于在日志中记录用户信息。
func (u UserInfoCp) LogName() string {
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
func (u UserInfoCp) Valid() bool { return u.UserID != 0 }

// PostC 楼层信息。
type PostC struct {
	Contents ContentsCp // 正文内容碎片列表
	Sign     string     // 小尾巴文本内容

	FID   int64      // 所在吧id
	FName string     // 所在贴吧名
	TID   int64      // 所在主题帖id
	PID   int64      // 回复id
	User  UserInfoCp // 发布者的用户信息

	Floor      int64 // 楼层数
	CreateTime int64 // 创建时间 10位时间戳 以秒为单位
}

// PostCFromProto 对应 Post_c.from_proto（输入：Post）。
func PostCFromProto(p *protobuf.Post) PostC {
	var sign strings.Builder
	for _, part := range p.GetSignature().GetContent() {
		if part.GetType() == 0 {
			sign.WriteString(part.GetText())
		}
	}
	return PostC{
		Contents:   ContentsCpFromProto(p),
		Sign:       sign.String(),
		PID:        p.GetId(),
		User:       UserInfoCpFromProto(p.GetAuthor()),
		Floor:      int64(p.GetFloor()),
		CreateTime: int64(p.GetTime()),
	}
}

// Text 文本内容。
func (p *PostC) Text() string {
	if p.Sign != "" {
		return p.Contents.Text() + "\n" + p.Sign
	}
	return p.Contents.Text()
}

// AuthorID 发布者的user_id。
func (p *PostC) AuthorID() int64 { return p.User.UserID }

// Equal 对应 __eq__。
func (p *PostC) Equal(other *PostC) bool { return other != nil && p.PID == other.PID }

// Comments 楼中楼列表。
//
// Comments 是 get_comments 的返回结果。
type Comments struct {
	classdef.Containers[*Comment]

	Page   PageC   // 页信息
	Forum  ForumC  // 所在吧信息
	Thread ThreadC // 所在主题帖信息
	Post   PostC   // 所在楼层信息
	Err    error   // 捕获的异常
}

// CommentsFromProto 对应 Comments.from_proto（输入：PbFloorResIdl.DataRes）。
func CommentsFromProto(p *pb.PbFloorResIdl_DataRes) Comments {
	page := PageCFromProto(p.GetPage())
	forum := ForumCFromProto(p.GetForum())

	thread := ThreadCFromProto(p.GetThread())
	thread.FID = forum.FID
	thread.FName = forum.FName

	post := PostCFromProto(p.GetPost())
	post.FID = thread.FID
	post.FName = thread.FName
	post.TID = thread.TID

	comments := make([]*Comment, 0, len(p.GetSubpostList()))
	for _, proto := range p.GetSubpostList() {
		comments = append(comments, CommentFromProto(proto))
	}
	for _, comment := range comments {
		comment.FID = forum.FID
		comment.FName = forum.FName
		comment.TID = thread.TID
		comment.PPID = post.PID
		comment.Floor = post.Floor
		comment.IsThreadAuthor = thread.AuthorID() == comment.AuthorID()
	}

	return Comments{
		Containers: classdef.Containers[*Comment]{Objs: comments},
		Page:       page,
		Forum:      forum,
		Thread:     thread,
		Post:       post,
	}
}

// HasMore 是否还有下一页。
func (c Comments) HasMore() bool { return c.Page.HasMore }

func dropFirstTwo[T any](s []T) []T {
	if len(s) <= 2 {
		return nil
	}
	return s[2:]
}
