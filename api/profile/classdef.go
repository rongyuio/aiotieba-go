package profile

import (
	"strconv"
	"strings"

	"github.com/rongyuio/aiotieba/api/classdef"
	pb "github.com/rongyuio/aiotieba/api/profile/protobuf"
	"github.com/rongyuio/aiotieba/enums"
	"github.com/rongyuio/aiotieba/helper"
	"github.com/rongyuio/aiotieba/protobuf"
)

// Ref identifies a user for the profile endpoint. It mirrors the
// `uid_or_portrait: str | int` argument of the Python module.
type Ref struct {
	UserID   int64
	Portrait string
}

// ByUserID references a user by their numeric id.
func ByUserID(id int64) Ref { return Ref{UserID: id} }

// ByPortrait references a user by their portrait.
func ByPortrait(portrait string) Ref { return Ref{Portrait: portrait} }

// UserInfoPF is the full user information returned by the profile endpoint. It
// mirrors aiotieba.api.profile._classdef.UserInfo_pf.
type UserInfoPF struct {
	UserID      int64
	Portrait    string
	UserName    string
	NickNameNew string
	TiebaUID    int64

	GLevel int32
	Gender enums.Gender
	Age    float64

	// The counters are int64 because Python reports them with arbitrary
	// precision and total_agree_num may exceed math.MaxInt32.
	PostNum   int64
	AgreeNum  int64
	FanNum    int64
	FollowNum int64
	ForumNum  int64

	Sign  string
	IP    string
	Icons []string

	IsVIP     bool
	IsGod     bool
	IsBlocked bool

	PrivLike  enums.PrivLike
	PrivReply enums.PrivReply
}

// minDaysToFree is the threshold above which a restricted account is reported
// as permanently blocked.
const minDaysToFree = 30

// UserInfoPFFromProto mirrors UserInfo_pf.from_proto.
func UserInfoPFFromProto(p *pb.ProfileResIdl_DataRes) UserInfoPF {
	u := p.GetUser()

	portrait := u.GetPortrait()
	// The portrait carries a "?..." query suffix that the client strips.
	if strings.Contains(portrait, "?") && len(portrait) > 13 {
		portrait = portrait[:len(portrait)-13]
	}

	icons := make([]string, 0, len(u.GetIconinfo()))
	for _, icon := range u.GetIconinfo() {
		if name := icon.GetName(); name != "" {
			icons = append(icons, name)
		}
	}

	anti := p.GetAntiStat()
	isBlocked := anti.GetBlockStat() != 0 && anti.GetHideStat() != 0 && anti.GetDaysTofree() > minDaysToFree

	return UserInfoPF{
		UserID:      u.GetId(),
		Portrait:    portrait,
		UserName:    u.GetName(),
		NickNameNew: u.GetNameShow(),
		TiebaUID:    helper.AnyInt64(u.GetTiebaUid()),
		GLevel:      int32(u.GetUserGrowth().GetLevelId()),
		Gender:      enums.GenderFrom(int(u.GetSex())),
		Age:         helper.AnyFloat64(u.GetTbAge()),
		PostNum:     int64(u.GetPostNum()),
		AgreeNum:    p.GetUserAgreeInfo().GetTotalAgreeNum(),
		FanNum:      int64(u.GetFansNum()),
		FollowNum:   int64(u.GetConcernNum()),
		ForumNum:    int64(u.GetMyLikeNum()),
		Sign:        u.GetIntro(),
		IP:          u.GetIpAddress(),
		Icons:       icons,
		IsVIP:       len(u.GetNewTshowIcon()) != 0,
		IsGod:       u.GetNewGodData().GetStatus() != 0,
		IsBlocked:   isBlocked,
		PrivLike:    privLikeOf(u.GetPrivSets().GetLike()),
		PrivReply:   privReplyOf(u.GetPrivSets().GetReply()),
	}
}

// privLikeOf mirrors `PrivLike(v) if v else PrivLike.PUBLIC`.
func privLikeOf(v int32) enums.PrivLike {
	if v == 0 {
		return enums.PrivLikePublic
	}
	return enums.PrivLikeFrom(int(v))
}

// privReplyOf mirrors `PrivReply(v) if v else PrivReply.ALL`.
func privReplyOf(v int32) enums.PrivReply {
	if v == 0 {
		return enums.PrivReplyAll
	}
	return enums.PrivReplyFrom(int(v))
}

// NickName mirrors the nick_name property.
func (u UserInfoPF) NickName() string { return u.NickNameNew }

// ShowName mirrors the show_name property.
func (u UserInfoPF) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// String mirrors __str__.
func (u UserInfoPF) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// LogName mirrors the log_name property.
func (u UserInfoPF) LogName() string {
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
func (u UserInfoPF) Valid() bool { return u.UserID != 0 }

// FragImagePF is the image fragment of a profile post. It mirrors
// aiotieba.api.profile._classdef.FragImage_pf.
type FragImagePF struct {
	Src        string
	OriginSrc  string
	OriginSize int64
	Width      int64
	Height     int64
	Hash       string
}

// FragImagePFFromProto mirrors FragImage_pf.from_proto (input: Media).
func FragImagePFFromProto(p *protobuf.Media) FragImagePF {
	src := p.GetBigPic()
	return FragImagePF{
		Src:        src,
		OriginSrc:  p.GetOriginPic(),
		OriginSize: int64(p.GetOriginSize()),
		Width:      int64(p.GetWidth()),
		Height:     int64(p.GetHeight()),
		Hash:       classdef.ImageHash(src),
	}
}

// ContentsPF is the body of a profile post. It mirrors
// aiotieba.api.profile._classdef.Contents_pf.
type ContentsPF struct {
	classdef.Containers[any]

	Texts  []classdef.Fragment
	Emojis []classdef.FragEmoji
	Imgs   []FragImagePF
	Ats    []classdef.FragAt
	Links  []classdef.FragLink
	Video  classdef.FragVideo
	Voice  classdef.FragVoice
}

// ContentsPFFromProto mirrors Contents_pf.from_proto (input: PostInfoList).
func ContentsPFFromProto(p *protobuf.PostInfoList) ContentsPF {
	var c ContentsPF

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
			// Images come from the media list, not from the content fragments.
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
		case t == 5 || t == 10:
			// Video and voice come from their dedicated fields.
		default:
			c.Objs = append(c.Objs, classdef.FragUnknownFromProto(proto))
		}
	}

	for _, media := range p.GetMedia() {
		if media.GetType() == 5 {
			continue
		}
		frag := FragImagePFFromProto(media)
		c.Imgs = append(c.Imgs, frag)
		c.Objs = append(c.Objs, frag)
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

// Text mirrors the `text` cached property.
func (c ContentsPF) Text() string { return classdef.FragmentTextOf(c.Texts) }

// ThreadPF is one post of a user home page. It mirrors
// aiotieba.api.profile._classdef.Thread_pf.
type ThreadPF struct {
	Contents ContentsPF
	Title    string
	FID      int64
	FName    string
	TID      int64
	PID      int64
	User     UserInfoPF

	VoteInfo classdef.VoteInfo

	ViewNum  int64
	ReplyNum int64
	ShareNum int64
	Agree    int64
	Disagree int64

	CreateTime int64
}

// ThreadPFFromProto mirrors Thread_pf.from_proto.
//
// User is filled in by HomepageFromProto, which knows the owner of the page.
func ThreadPFFromProto(p *protobuf.PostInfoList) ThreadPF {
	return ThreadPF{
		Contents:   ContentsPFFromProto(p),
		Title:      p.GetTitle(),
		FID:        int64(p.GetForumId()),
		FName:      p.GetForumName(),
		TID:        int64(p.GetThreadId()),
		PID:        int64(p.GetPostId()),
		VoteInfo:   classdef.VoteInfoFromProto(p.GetPollInfo()),
		ViewNum:    int64(p.GetFreqNum()),
		ReplyNum:   int64(p.GetReplyNum()),
		ShareNum:   int64(p.GetShareNum()),
		Agree:      int64(p.GetAgree().GetAgreeNum()),
		Disagree:   int64(p.GetAgree().GetDisagreeNum()),
		CreateTime: int64(p.GetCreateTime()),
	}
}

// Text mirrors the `text` cached property.
func (t ThreadPF) Text() string {
	if t.Title != "" {
		return t.Title + "\n" + t.Contents.Text()
	}
	return t.Contents.Text()
}

// AuthorID mirrors the author_id property.
func (t ThreadPF) AuthorID() int64 { return t.User.UserID }

// Homepage is the information of a user home page. It mirrors
// aiotieba.api.profile._classdef.Homepage.
type Homepage struct {
	classdef.Containers[ThreadPF]

	User UserInfoPF
}

// HomepageFromProto mirrors Homepage.from_proto.
func HomepageFromProto(p *pb.ProfileResIdl_DataRes) Homepage {
	user := UserInfoPFFromProto(p)

	posts := p.GetPostList()
	objs := make([]ThreadPF, 0, len(posts))
	for _, post := range posts {
		thread := ThreadPFFromProto(post)
		thread.User = user
		objs = append(objs, thread)
	}

	return Homepage{
		Containers: classdef.Containers[ThreadPF]{Objs: objs},
		User:       user,
	}
}
