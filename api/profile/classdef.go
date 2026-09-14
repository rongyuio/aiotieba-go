package profile

import (
	"strconv"
	"strings"

	"github.com/rongyuio/aiotieba-go/api/classdef"
	pb "github.com/rongyuio/aiotieba-go/api/profile/protobuf"
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/protobuf"
)

// Ref 为 profile 端点标识一个用户，对应 Python 模块的
// `uid_or_portrait: str | int` 参数。
type Ref struct {
	UserID   int64
	Portrait string
}

// ByUserID 通过数字 id 引用一个用户。
func ByUserID(id int64) Ref { return Ref{UserID: id} }

// ByPortrait 通过 portrait 引用一个用户。
func ByPortrait(portrait string) Ref { return Ref{Portrait: portrait} }

// UserInfoPF 用户信息。
type UserInfoPF struct {
	UserID      int64  // user_id
	Portrait    string // portrait
	UserName    string // 用户名
	NickNameNew string // 新版昵称
	TiebaUID    int64  // 用户个人主页uid

	GLevel int32        // 贴吧成长等级
	Gender enums.Gender // 性别
	Age    float64      // 吧龄 以年为单位

	// 这些计数使用 int64，因为 Python 会以任意精度上报它们，
	// 且 total_agree_num 可能超过 math.MaxInt32。
	PostNum   int64 // 发帖数
	AgreeNum  int64 // 获赞数
	FanNum    int64 // 粉丝数
	FollowNum int64 // 关注数
	ForumNum  int64 // 关注贴吧数

	Sign  string   // 个性签名
	IP    string   // ip归属地
	Icons []string // 印记信息

	IsVIP     bool // 是否超级会员
	IsGod     bool // 是否大神
	IsBlocked bool // 是否被永久封禁屏蔽

	PrivLike  enums.PrivLike  // 关注吧列表的公开状态
	PrivReply enums.PrivReply // 帖子评论权限
}

// minDaysToFree 是阈值，超过该值的受限账号会被上报为永久封禁。
const minDaysToFree = 30

// UserInfoPFFromProto 对应 UserInfo_pf.from_proto。
func UserInfoPFFromProto(p *pb.ProfileResIdl_DataRes) UserInfoPF {
	u := p.GetUser()

	portrait := u.GetPortrait()
	// portrait 携带 "?..." 查询后缀，客户端会将其去除。
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

// privLikeOf 对应 `PrivLike(v) if v else PrivLike.PUBLIC`。
func privLikeOf(v int32) enums.PrivLike {
	if v == 0 {
		return enums.PrivLikePublic
	}
	return enums.PrivLikeFrom(int(v))
}

// privReplyOf 对应 `PrivReply(v) if v else PrivReply.ALL`。
func privReplyOf(v int32) enums.PrivReply {
	if v == 0 {
		return enums.PrivReplyAll
	}
	return enums.PrivReplyFrom(int(v))
}

// NickName 用户昵称。
func (u UserInfoPF) NickName() string { return u.NickNameNew }

// ShowName 显示名称。
func (u UserInfoPF) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// String 对应 __str__。
func (u UserInfoPF) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// LogName 用于在日志中记录用户信息。
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

// Valid 对应 __bool__。
func (u UserInfoPF) Valid() bool { return u.UserID != 0 }

// FragImagePF 图像碎片。
type FragImagePF struct {
	Src        string // 大图链接 宽960px
	OriginSrc  string // 原图链接
	OriginSize int64  // 原图大小
	Width      int64  // 图像宽度
	Height     int64  // 图像高度
	Hash       string // 百度图床hash
}

// FragImagePFFromProto 对应 FragImage_pf.from_proto（输入：Media）。
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

// ContentsPF 内容碎片列表。
type ContentsPF struct {
	classdef.Containers[any]

	Texts  []classdef.Fragment  // 纯文本碎片列表
	Emojis []classdef.FragEmoji // 表情碎片列表
	Imgs   []FragImagePF        // 图像碎片列表
	Ats    []classdef.FragAt    // @碎片列表
	Links  []classdef.FragLink  // 链接碎片列表
	Video  classdef.FragVideo   // 视频碎片
	Voice  classdef.FragVoice   // 音频碎片
}

// ContentsPFFromProto 对应 Contents_pf.from_proto（输入：PostInfoList）。
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
			// 图像来自 media 列表，而非内容碎片。
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
			// 视频与语音来自各自的专用字段。
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

// Text 文本内容。
func (c ContentsPF) Text() string { return classdef.FragmentTextOf(c.Texts) }

// ThreadPF 主题帖信息。
type ThreadPF struct {
	Contents ContentsPF // 正文内容碎片列表
	Title    string     // 标题内容
	FID      int64      // 所在吧id
	FName    string     // 所在贴吧名
	TID      int64      // 主题帖tid
	PID      int64      // 首楼回复pid
	User     UserInfoPF // 发布者的用户信息

	VoteInfo classdef.VoteInfo // 投票信息

	ViewNum  int64 // 浏览量
	ReplyNum int64 // 回复数
	ShareNum int64 // 分享数
	Agree    int64 // 点赞数
	Disagree int64 // 点踩数

	CreateTime int64 // 创建时间 10位时间戳 以秒为单位
}

// ThreadPFFromProto 对应 Thread_pf.from_proto。
//
// User 由 HomepageFromProto 填充，因为它知道该页面的所有者。
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

// Text 文本内容。
func (t ThreadPF) Text() string {
	if t.Title != "" {
		return t.Title + "\n" + t.Contents.Text()
	}
	return t.Contents.Text()
}

// AuthorID 发布者的user_id。
func (t ThreadPF) AuthorID() int64 { return t.User.UserID }

// Homepage 用户个人页信息。
type Homepage struct {
	classdef.Containers[ThreadPF]

	User UserInfoPF // 用户信息
}

// HomepageFromProto 对应 Homepage.from_proto。
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
