// Package getcomments implements the get_comments API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_comments.
package getcomments

import (
	"strconv"
	"strings"

	"github.com/rongyuio/aiotieba/api/classdef"
	"github.com/rongyuio/aiotieba/enums"
	"github.com/rongyuio/aiotieba/logging"
	"github.com/rongyuio/aiotieba/protobuf"

	pb "github.com/rongyuio/aiotieba/api/get_comments/protobuf"
)

// ContentsC is the body of a comment.
type ContentsC struct {
	classdef.Containers[any]

	Texts       []classdef.Fragment
	Emojis      []classdef.FragEmoji
	Ats         []classdef.FragAt
	Links       []classdef.FragLink
	TiebaPluses []classdef.FragTiebaPlus
	Voice       classdef.FragVoice
}

// ContentsCFromProto mirrors Contents_c.from_proto (input: SubPostList).
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
			// Outdated tiebaplus is skipped.
		default:
			c.Objs = append(c.Objs, classdef.FragUnknownFromProto(proto))
		}
	}
	return c
}

// Text mirrors the `text` cached property.
func (c ContentsC) Text() string { return classdef.FragmentTextOf(c.Texts) }

// UserInfoC is the user information of a comment.
type UserInfoC struct {
	UserID      int64
	Portrait    string
	UserName    string
	NickNameNew string

	Level  int32
	Gender enums.Gender
	Icons  []string

	IsBawu    bool
	IsVIP     bool
	IsGod     bool
	PrivLike  enums.PrivLike
	PrivReply enums.PrivReply
}

// UserInfoCFromProto mirrors UserInfo_c.from_proto (input: User).
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

// String mirrors __str__.
func (u UserInfoC) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// NickName mirrors the nick_name property.
func (u UserInfoC) NickName() string { return u.NickNameNew }

// ShowName mirrors the show_name property.
func (u UserInfoC) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// LogName mirrors the log_name property.
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

// Valid mirrors __bool__.
func (u UserInfoC) Valid() bool { return u.UserID != 0 }

// Comment is one comment of a floor.
type Comment struct {
	Contents ContentsC

	FID   int64
	FName string
	TID   int64
	PPID  int64
	PID   int64
	User  UserInfoC

	ReplyToID int64

	Floor          int64
	Agree          int64
	Disagree       int64
	CreateTime     int64
	IsThreadAuthor bool
}

// CommentFromProto mirrors Comment.from_proto (input: SubPostList).
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

// Text mirrors the `text` property.
func (c *Comment) Text() string { return c.Contents.Text() }

// AuthorID mirrors the author_id property.
func (c *Comment) AuthorID() int64 { return c.User.UserID }

// Equal mirrors __eq__.
func (c *Comment) Equal(other *Comment) bool { return other != nil && c.PID == other.PID }

// PageC is the page information of a comment list.
type PageC struct {
	PageSize    int32
	CurrentPage int32
	TotalPage   int32
	TotalCount  int32
	HasMore     bool
	HasPrev     bool
}

// PageCFromProto mirrors Page_c.from_proto.
//
// Unlike Page_p it derives has_more/has_prev from the page numbers.
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

// ForumC is the forum of a comment list.
type ForumC struct {
	FID   int64
	FName string

	Category    string
	Subcategory string
}

// ForumCFromProto mirrors Forum_c.from_proto (input: SimpleForum).
func ForumCFromProto(p *protobuf.SimpleForum) ForumC {
	return ForumC{
		FID:         p.GetId(),
		FName:       p.GetName(),
		Category:    p.GetFirstClass(),
		Subcategory: p.GetSecondClass(),
	}
}

// UserInfoCt is the author of the thread of a comment list.
type UserInfoCt struct {
	UserID      int64
	Portrait    string
	UserName    string
	NickNameNew string

	Level int32
	IsGod bool
}

// UserInfoCtFromProto mirrors UserInfo_ct.from_proto (input: User).
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

// String mirrors __str__.
func (u UserInfoCt) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// NickName mirrors the nick_name property.
func (u UserInfoCt) NickName() string { return u.NickNameNew }

// ShowName mirrors the show_name property.
func (u UserInfoCt) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// LogName mirrors the log_name property.
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

// Valid mirrors __bool__.
func (u UserInfoCt) Valid() bool { return u.UserID != 0 }

// ThreadC is the thread of a comment list.
type ThreadC struct {
	Title string

	FID   int64
	FName string
	TID   int64
	User  UserInfoCt

	Type enums.ThreadType

	ReplyNum int64
}

// ThreadCFromProto mirrors Thread_c.from_proto (input: ThreadInfo).
func ThreadCFromProto(p *protobuf.ThreadInfo) ThreadC {
	typeValue := enums.ThreadTypeFrom(int(p.GetThreadType()))
	if typeValue == enums.ThreadTypeUnknown {
		logging.GetLogger().Debug("unknown thread type", "tid", p.GetId(), "type", p.GetThreadType())
	}
	return ThreadC{
		Title:    p.GetTitle(),
		TID:      p.GetId(),
		User:     UserInfoCtFromProto(p.GetAuthor()),
		Type:     typeValue,
		ReplyNum: int64(p.GetReplyNum()),
	}
}

// AuthorID mirrors the author_id property.
func (t ThreadC) AuthorID() int64 { return t.User.UserID }

// Equal mirrors __eq__.
func (t ThreadC) Equal(other ThreadC) bool { return t.TID == other.TID }

// FragImageCp is an image fragment of a floor.
type FragImageCp struct {
	Src        string
	BigSrc     string
	OriginSrc  string
	OriginSize int64
	ShowWidth  int32
	ShowHeight int32
	Hash       string
}

// FragImageCpFromProto mirrors FragImage_cp.from_proto (input: PbContent).
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

// ContentsCp is the body of the floor that hosts the comments.
type ContentsCp struct {
	classdef.Containers[any]

	Texts       []classdef.Fragment
	Emojis      []classdef.FragEmoji
	Imgs        []FragImageCp
	Ats         []classdef.FragAt
	Links       []classdef.FragLink
	TiebaPluses []classdef.FragTiebaPlus
	Voice       classdef.FragVoice
}

// ContentsCpFromProto mirrors Contents_cp.from_proto (input: Post).
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
			// Outdated tiebaplus is skipped.
		default:
			c.Objs = append(c.Objs, classdef.FragUnknownFromProto(proto))
		}
	}
	return c
}

// Text mirrors the `text` cached property.
func (c ContentsCp) Text() string { return classdef.FragmentTextOf(c.Texts) }

// UserInfoCp is the author of the floor that hosts the comments.
type UserInfoCp struct {
	UserID      int64
	Portrait    string
	UserName    string
	NickNameNew string

	Level  int32
	Gender enums.Gender

	IsBawu    bool
	IsVIP     bool
	IsGod     bool
	PrivLike  enums.PrivLike
	PrivReply enums.PrivReply
}

// UserInfoCpFromProto mirrors UserInfo_cp.from_proto (input: User).
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

// String mirrors __str__.
func (u UserInfoCp) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// NickName mirrors the nick_name property.
func (u UserInfoCp) NickName() string { return u.NickNameNew }

// ShowName mirrors the show_name property.
func (u UserInfoCp) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// LogName mirrors the log_name property.
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

// Valid mirrors __bool__.
func (u UserInfoCp) Valid() bool { return u.UserID != 0 }

// PostC is the floor that hosts the comments.
type PostC struct {
	Contents ContentsCp
	Sign     string

	FID   int64
	FName string
	TID   int64
	PID   int64
	User  UserInfoCp

	Floor      int64
	CreateTime int64
}

// PostCFromProto mirrors Post_c.from_proto (input: Post).
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

// Text mirrors the `text` cached property.
func (p *PostC) Text() string {
	if p.Sign != "" {
		return p.Contents.Text() + "\n" + p.Sign
	}
	return p.Contents.Text()
}

// AuthorID mirrors the author_id property.
func (p *PostC) AuthorID() int64 { return p.User.UserID }

// Equal mirrors __eq__.
func (p *PostC) Equal(other *PostC) bool { return other != nil && p.PID == other.PID }

// Comments is the result of get_comments.
type Comments struct {
	classdef.Containers[*Comment]

	Page   PageC
	Forum  ForumC
	Thread ThreadC
	Post   PostC
	Err    error
}

// CommentsFromProto mirrors Comments.from_proto (input: PbFloorResIdl.DataRes).
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

// HasMore mirrors the has_more property.
func (c Comments) HasMore() bool { return c.Page.HasMore }

func dropFirstTwo[T any](s []T) []T {
	if len(s) <= 2 {
		return nil
	}
	return s[2:]
}
