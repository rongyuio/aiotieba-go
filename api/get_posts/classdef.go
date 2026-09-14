// Package getposts implements the get_posts API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_posts.
package getposts

import (
	"strconv"
	"strings"

	"github.com/rongyuio/aiotieba/api/classdef"
	"github.com/rongyuio/aiotieba/enums"
	"github.com/rongyuio/aiotieba/logging"
	"github.com/rongyuio/aiotieba/protobuf"

	pb "github.com/rongyuio/aiotieba/api/get_posts/protobuf"
)

// FragImageP is an image fragment of a post body.
type FragImageP struct {
	Src        string
	BigSrc     string
	OriginSrc  string
	OriginSize int64
	ShowWidth  int32
	ShowHeight int32
	Hash       string
}

// FragImagePFromProto mirrors FragImage_p.from_proto (input: PbContent).
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

// FragVideoP is a video fragment of a post body.
type FragVideoP struct {
	Src      string
	CoverSrc string
	Duration int64
	Width    int64
	Height   int64
	ViewNum  int64
}

// FragVideoPFromProto mirrors FragVideo_p.from_proto (input: PbContent).
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

// Valid mirrors __bool__ of FragVideo_p.
func (f FragVideoP) Valid() bool { return f.Width != 0 }

// ContentsP is the body of a floor.
type ContentsP struct {
	classdef.Containers[any]

	Texts       []classdef.Fragment
	Emojis      []classdef.FragEmoji
	Imgs        []FragImageP
	Ats         []classdef.FragAt
	Links       []classdef.FragLink
	TiebaPluses []classdef.FragTiebaPlus
	Video       FragVideoP
	Voice       classdef.FragVoice
}

// ContentsPFromProto mirrors Contents_p.from_proto (input: Post).
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
			// Outdated tiebaplus and vote fragments are skipped.
		default:
			c.Objs = append(c.Objs, classdef.FragUnknownFromProto(proto))
		}
	}
	return c
}

// Text mirrors the `text` cached property.
func (c ContentsP) Text() string { return classdef.FragmentTextOf(c.Texts) }

// ContentsPc is the body of a comment.
type ContentsPc struct {
	classdef.Containers[any]

	Texts       []classdef.Fragment
	Emojis      []classdef.FragEmoji
	Ats         []classdef.FragAt
	Links       []classdef.FragLink
	TiebaPluses []classdef.FragTiebaPlus
	Voice       classdef.FragVoice
}

// ContentsPcFromProto mirrors Contents_pc.from_proto (input: SubPostList).
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
			// Outdated tiebaplus is skipped.
		default:
			c.Objs = append(c.Objs, classdef.FragUnknownFromProto(proto))
		}
	}
	return c
}

// Text mirrors the `text` cached property.
func (c ContentsPc) Text() string { return classdef.FragmentTextOf(c.Texts) }

// UserInfoP is the user information of a post.
type UserInfoP struct {
	UserID      int64
	Portrait    string
	UserName    string
	NickNameNew string

	Level  int32
	GLevel int32
	Gender enums.Gender
	IP     string
	Icons  []string

	IsBawu    bool
	IsVIP     bool
	IsGod     bool
	PrivLike  enums.PrivLike
	PrivReply enums.PrivReply
}

// UserInfoPFromProto mirrors UserInfo_p.from_proto (input: User).
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

// String mirrors __str__.
func (u UserInfoP) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// NickName mirrors the nick_name property.
func (u UserInfoP) NickName() string { return u.NickNameNew }

// ShowName mirrors the show_name property.
func (u UserInfoP) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// LogName mirrors the log_name property.
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

// Valid mirrors __bool__.
func (u UserInfoP) Valid() bool { return u.UserID != 0 }

// Equal mirrors __eq__.
func (u UserInfoP) Equal(other UserInfoP) bool { return u.UserID == other.UserID }

// CommentP is one comment of a floor.
type CommentP struct {
	Contents ContentsPc

	FID   int64
	FName string
	TID   int64
	PPID  int64
	PID   int64
	User  UserInfoP

	AuthorID  int64
	ReplyToID int64

	Floor          int64
	Agree          int64
	Disagree       int64
	CreateTime     int64
	IsThreadAuthor bool
}

// CommentPFromProto mirrors Comment_p.from_proto (input: SubPostList).
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

// Text mirrors the `text` property.
func (c *CommentP) Text() string { return c.Contents.Text() }

// Equal mirrors __eq__.
func (c *CommentP) Equal(other *CommentP) bool { return other != nil && c.PID == other.PID }

// Post is one floor of a thread.
type Post struct {
	Contents ContentsP
	Sign     string
	Comments []CommentP
	IsAIMeme bool

	FID   int64
	FName string
	TID   int64
	PID   int64
	User  UserInfoP

	AuthorID       int64
	Floor          int64
	ReplyNum       int64
	Agree          int64
	Disagree       int64
	CreateTime     int64
	IsThreadAuthor bool
}

// PostFromProto mirrors Post.from_proto (input: Post).
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

// Text mirrors the `text` cached property.
func (p *Post) Text() string {
	if p.Sign != "" {
		return p.Contents.Text() + "\n" + p.Sign
	}
	return p.Contents.Text()
}

// Equal mirrors __eq__.
func (p *Post) Equal(other *Post) bool { return other != nil && p.PID == other.PID }

// PageP is the page information of a post list.
type PageP struct {
	PageSize    int32
	CurrentPage int32
	TotalPage   int32
	TotalCount  int32
	HasMore     bool
	HasPrev     bool
}

// PagePFromProto mirrors Page_p.from_proto.
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

// ForumP is the forum of a post list.
type ForumP struct {
	FID   int64
	FName string

	Category    string
	Subcategory string

	MemberNum int64
	PostNum   int64
}

// ForumPFromProto mirrors Forum_p.from_proto (input: SimpleForum).
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

// FragImagePt is an image fragment of a shared thread.
type FragImagePt struct {
	Src        string
	BigSrc     string
	OriginSrc  string
	ShowWidth  int32
	ShowHeight int32
	Hash       string
}

// FragImagePtFromProto mirrors FragImage_pt.from_proto (input: Media).
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

// ContentsPt is the body of the thread of a post list.
type ContentsPt struct {
	classdef.Containers[any]

	Texts       []classdef.Fragment
	Emojis      []classdef.FragEmoji
	Imgs        []FragImagePt
	Ats         []classdef.FragAt
	Links       []classdef.FragLink
	TiebaPluses []classdef.FragTiebaPlus
	Video       classdef.FragVideo
	Voice       classdef.FragVoice
}

// ContentsPtFromProto mirrors Contents_pt.from_proto (input:
// ThreadInfo.OriginThreadInfo).
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
			// Outdated tiebaplus is skipped.
		default:
			c.Objs = append(c.Objs, classdef.FragUnknownFromProto(proto))
		}
	}

	// The first @fragment is dropped and the media images are appended.
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

// Text mirrors the `text` cached property.
func (c ContentsPt) Text() string { return classdef.FragmentTextOf(c.Texts) }

// UserInfoPt is the user information of the thread of a post list. Unlike
// UserInfoP it has no gender field.
type UserInfoPt struct {
	UserID      int64
	Portrait    string
	UserName    string
	NickNameNew string

	Level  int32
	GLevel int32
	IP     string
	Icons  []string

	IsBawu    bool
	IsVIP     bool
	IsGod     bool
	PrivLike  enums.PrivLike
	PrivReply enums.PrivReply
}

// UserInfoPtFromProto mirrors UserInfo_pt.from_proto (input: User).
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

// String mirrors __str__.
func (u UserInfoPt) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// NickName mirrors the nick_name property.
func (u UserInfoPt) NickName() string { return u.NickNameNew }

// ShowName mirrors the show_name property.
func (u UserInfoPt) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// LogName mirrors the log_name property.
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

// Valid mirrors __bool__.
func (u UserInfoPt) Valid() bool { return u.UserID != 0 }

// ShareThreadPt is the origin of a shared thread.
type ShareThreadPt struct {
	Contents ContentsPt
	Title    string

	FID      int64
	FName    string
	TID      int64
	AuthorID int64

	VoteInfo classdef.VoteInfo
}

// ShareThreadPtFromProto mirrors ShareThread_pt.from_proto.
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

// Text mirrors the `text` cached property.
func (s ShareThreadPt) Text() string {
	if s.Title != "" {
		return s.Title + "\n" + s.Contents.Text()
	}
	return s.Contents.Text()
}

// ThreadP is the thread of a post list.
type ThreadP struct {
	Contents ContentsPt
	Title    string

	FID   int64
	FName string
	TID   int64
	PID   int64
	User  UserInfoPt

	Type    enums.ThreadType
	IsShare bool

	VoteInfo    classdef.VoteInfo
	ShareOrigin ShareThreadPt
	ViewNum     int64
	ReplyNum    int64
	ShareNum    int64
	Agree       int64
	Disagree    int64
	CreateTime  int64
}

// ThreadPFromProto mirrors Thread_p.from_proto (input: PbPageResIdl.DataRes).
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

// Text mirrors the `text` property.
func (t ThreadP) Text() string {
	if t.Title != "" {
		return t.Title + "\n" + t.Contents.Text()
	}
	return t.Contents.Text()
}

// AuthorID mirrors the author_id property.
func (t ThreadP) AuthorID() int64 { return t.User.UserID }

// Equal mirrors __eq__.
func (t ThreadP) Equal(other ThreadP) bool { return t.PID == other.PID }

// Posts is the result of get_posts.
type Posts struct {
	classdef.Containers[*Post]

	Page   PageP
	Forum  ForumP
	Thread ThreadP
	Err    error
}

// PostsFromProto mirrors Posts.from_proto (input: PbPageResIdl.DataRes).
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

// HasMore mirrors the has_more property.
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
