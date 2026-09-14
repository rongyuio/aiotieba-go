// Package getthreads 实现 aiotieba 的 get_threads API。
//
// 对应 Python 包 aiotieba.api.get_threads。
package getthreads

import (
	"strconv"

	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/logging"
	"github.com/rongyuio/aiotieba-go/protobuf"

	pb "github.com/rongyuio/aiotieba-go/api/get_threads/protobuf"
)

// FragImageFeed 图像碎片。
type FragImageFeed struct {
	Src       string // 小图链接 宽720px
	BigSrc    string // 大图链接 宽960px
	OriginSrc string // 原图链接
	Width     int64  // 图像宽度
	Height    int64  // 图像高度
	Hash      string // 百度图床hash
}

// FragImageFeedFromProto 对应 FragImage_feed.from_proto。
func FragImageFeedFromProto(p *pb.PageData_LayoutFactory_FeedLayout_ComponentFactory_PicInfo) FragImageFeed {
	src := p.GetSmallPicUrl()
	origin := p.GetOriginPicUrl()
	return FragImageFeed{
		Src:       src,
		BigSrc:    p.GetBigPicUrl(),
		OriginSrc: origin,
		Width:     int64(p.GetWidth()),
		Height:    int64(p.GetHeight()),
		Hash:      classdef.ImageHash(origin),
	}
}

// FragEmojiFeed 表情碎片。
type FragEmojiFeed struct {
	ID   string // 表情图片id
	Desc string // 表情描述
}

// FragEmojiFeedFromProto 对应 FragEmoji_feed.from_proto。
func FragEmojiFeedFromProto(p *pb.PageData_LayoutFactory_FeedLayout_ComponentFactory_FeedContentResource_FeedContentEmoji) FragEmojiFeed {
	return FragEmojiFeed{ID: p.GetName(), Desc: p.GetC()}
}

// PageT 页信息。
type PageT struct {
	PageSize    int32 // 页大小
	CurrentPage int32 // 当前页码
	TotalPage   int32 // 总页码
	TotalCount  int32 // 总计数
	HasMore     bool  // 是否有后继页
	HasPrev     bool  // 是否有前驱页
}

// PageTFromProto 对应 Page_t.from_proto。
func PageTFromProto(p *protobuf.Page) PageT {
	current := p.GetCurrentPage()
	if current == 0 && p.GetPageSize() != 0 {
		current = 1
	}
	return PageT{
		PageSize:    p.GetPageSize(),
		CurrentPage: current,
		TotalPage:   p.GetTotalPage(),
		TotalCount:  p.GetTotalCount(),
		HasMore:     p.GetHasMore() != 0,
		HasPrev:     p.GetHasPrev() != 0,
	}
}

// UserInfoT 用户信息。
type UserInfoT struct {
	UserID      int64  // user_id
	Portrait    string // portrait
	UserName    string // 用户名
	NickNameNew string // 新版昵称

	Level  int32        // 等级
	GLevel int32        // 贴吧成长等级
	Gender enums.Gender // 性别
	Icons  []string     // 印记信息

	IsBawu    bool            // 是否吧务
	IsVIP     bool            // 是否超级会员
	IsGod     bool            // 是否大神
	PrivLike  enums.PrivLike  // 关注吧列表的公开状态
	PrivReply enums.PrivReply // 帖子评论权限
}

// UserInfoTFromProto 对应 UserInfo_t.from_proto。
func UserInfoTFromProto(p *protobuf.User) UserInfoT {
	return UserInfoT{
		UserID:      p.GetId(),
		Portrait:    classdef.TrimPortrait(p.GetPortrait()),
		UserName:    p.GetName(),
		NickNameNew: p.GetNameShow(),
		Level:       p.GetLevelId(),
		GLevel:      int32(p.GetUserGrowth().GetLevelId()),
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
func (u UserInfoT) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// NickName 用户昵称。
func (u UserInfoT) NickName() string { return u.NickNameNew }

// ShowName 显示名称。
func (u UserInfoT) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// LogName 用于在日志中记录用户信息。
func (u UserInfoT) LogName() string {
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
func (u UserInfoT) Valid() bool { return u.UserID != 0 }

// Equal 对应 __eq__。
func (u UserInfoT) Equal(other UserInfoT) bool { return u.UserID == other.UserID }

// FragImageSt 图像碎片。
type FragImageSt struct {
	Src        string // 小图链接 宽580px
	BigSrc     string // 大图链接 宽720px
	OriginSrc  string // 原图链接
	ShowWidth  int32  // 图像在客户端预览显示的宽度
	ShowHeight int32  // 图像在客户端预览显示的高度
	Hash       string // 百度图床hash
}

// FragImageStFromProto 对应 FragImage_st.from_proto。
func FragImageStFromProto(p *protobuf.Media) FragImageSt {
	src := p.GetWaterPic()
	return FragImageSt{
		Src:        src,
		BigSrc:     p.GetSmallPic(),
		OriginSrc:  p.GetBigPic(),
		ShowWidth:  int32(p.GetWidth()),
		ShowHeight: int32(p.GetHeight()),
		Hash:       classdef.ImageHash(src),
	}
}

// ContentsT 内容碎片列表。
type ContentsT struct {
	classdef.Containers[any]

	Texts       []classdef.Fragment      // 纯文本碎片列表
	Emojis      []any                    // 表情碎片列表
	Imgs        []any                    // 图像碎片列表
	Ats         []classdef.FragAt        // @碎片列表
	Links       []classdef.FragLink      // 链接碎片列表
	TiebaPluses []classdef.FragTiebaPlus // 贴吧plus碎片列表
	Video       classdef.FragVideo       // 视频碎片
	Voice       classdef.FragVoice       // 音频碎片
}

// ContentsTFromProto 对应 Contents_t.from_proto。
func ContentsTFromProto(p *protobuf.ThreadInfo) ContentsT {
	c := ContentsT{}
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
			frag := classdef.FragImageFromProto(proto)
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
		case t == 10, t == 5:
			// 语音与视频由专用的 proto 字段承载。
		case t == 35 || t == 36 || t == 37:
			frag := classdef.FragTiebaPlusFromProto(proto)
			c.TiebaPluses = append(c.TiebaPluses, frag)
			c.Texts = append(c.Texts, frag)
			c.Objs = append(c.Objs, frag)
		case t == 34:
			// 过时的贴吧 plus。
		default:
			c.Objs = append(c.Objs, classdef.FragUnknownFromProto(proto))
		}
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

// ContentsTFromFeed 对应 Contents_t.from_feed。
func ContentsTFromFeed(p *pb.PageData_LayoutFactory_FeedLayout) ContentsT {
	c := ContentsT{}
	for _, component := range p.GetComponents() {
		switch component.GetComponent() {
		case "feed_abstract":
			for _, proto := range component.GetFeedAbstract().GetData() {
				switch proto.GetType() {
				case 1:
					frag := classdef.FragTextFromText(proto.GetTextInfo().GetText())
					c.Texts = append(c.Texts, frag)
					c.Objs = append(c.Objs, frag)
				case 3:
					frag := FragEmojiFeedFromProto(proto.GetEmojiInfo())
					c.Emojis = append(c.Emojis, frag)
					c.Objs = append(c.Objs, frag)
				default:
					c.Objs = append(c.Objs, classdef.FragUnknownFromAny(proto))
				}
			}
		case "feed_pic":
			for _, pic := range component.GetFeedPic().GetPics() {
				frag := FragImageFeedFromProto(pic)
				c.Imgs = append(c.Imgs, frag)
				c.Objs = append(c.Objs, frag)
			}
		case "feed_head", "feed_title", "feed_social", "feed_poll":
			continue
		default:
			logging.GetLogger().Debug("unknown component type", "type", component.GetComponent())
		}
	}
	return c
}

// Text 文本内容。
func (c ContentsT) Text() string { return classdef.FragmentTextOf(c.Texts) }

// ContentsSt 内容碎片列表。
type ContentsSt struct {
	classdef.Containers[any]

	Texts       []classdef.Fragment      // 纯文本碎片列表
	Emojis      []classdef.FragEmoji     // 表情碎片列表
	Imgs        []FragImageSt            // 图像碎片列表
	Ats         []classdef.FragAt        // @碎片列表
	Links       []classdef.FragLink      // 链接碎片列表
	TiebaPluses []classdef.FragTiebaPlus // 贴吧plus碎片列表
	Video       classdef.FragVideo       // 视频碎片
	Voice       classdef.FragVoice       // 音频碎片
}

// ContentsStFromProto 对应 Contents_st.from_proto。
func ContentsStFromProto(p *protobuf.ThreadInfo_OriginThreadInfo) ContentsSt {
	c := ContentsSt{}
	for _, media := range p.GetMedia() {
		c.Imgs = append(c.Imgs, FragImageStFromProto(media))
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
		case t == 5:
			// 视频由专用的 proto 字段承载。
		case t == 35 || t == 36 || t == 37:
			frag := classdef.FragTiebaPlusFromProto(proto)
			c.TiebaPluses = append(c.TiebaPluses, frag)
			c.Texts = append(c.Texts, frag)
			c.Objs = append(c.Objs, frag)
		case t == 34:
			// 过时的贴吧 plus。
		default:
			c.Objs = append(c.Objs, classdef.FragUnknownFromProto(proto))
		}
	}

	// 第一个 @碎片会被丢弃，随后追加媒体图像。
	if len(c.Ats) > 0 {
		c.Ats = c.Ats[1:]
		if len(c.Objs) > 0 {
			c.Objs = c.Objs[1:]
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

// Text 文本内容。
func (c ContentsSt) Text() string { return classdef.FragmentTextOf(c.Texts) }

// ShareThread 被分享的主题帖信息。
type ShareThread struct {
	Contents ContentsSt // 正文内容碎片列表
	Title    string     // 标题内容

	AuthorID int64 // 发布者的user_id

	FID   int64  // 所在吧id
	FName string // 所在贴吧名
	TID   int64  // 主题帖tid
	PID   int64  // 首楼的回复id

	VoteInfo classdef.VoteInfo // 投票内容
}

// ShareThreadFromProto 对应 ShareThread.from_proto。
func ShareThreadFromProto(p *protobuf.ThreadInfo_OriginThreadInfo) ShareThread {
	contents := ContentsStFromProto(p)
	var authorID int64
	if len(p.GetContent()) != 0 {
		authorID = p.GetContent()[0].GetUid()
	}
	return ShareThread{
		Contents: contents,
		Title:    p.GetTitle(),
		AuthorID: authorID,
		FID:      p.GetFid(),
		FName:    p.GetFname(),
		TID:      classdef.ParseInt64OrZero(p.GetTid()),
		PID:      p.GetPid(),
		VoteInfo: classdef.VoteInfoFromProto(p.GetPollInfo()),
	}
}

// Text 文本内容。
func (s ShareThread) Text() string {
	if s.Title != "" {
		return s.Title + "\n" + s.Contents.Text()
	}
	return s.Contents.Text()
}

// Thread 主题帖信息。
type Thread struct {
	Contents ContentsT // 正文内容碎片列表
	Title    string    // 标题内容

	FID      int64     // 所在吧id
	FName    string    // 所在贴吧名
	TID      int64     // 主题帖tid
	PID      int64     // 首楼回复pid
	User     UserInfoT // 发布者的用户信息
	AuthorID int64     // 发布者的user_id

	Type       enums.ThreadType // 帖子类型
	TabID      int32            // 帖子所在分区id
	IsGood     bool             // 是否精品帖
	IsTop      bool             // 是否置顶帖
	IsShare    bool             // 是否分享帖
	IsHide     bool             // 是否被屏蔽
	IsLivepost bool             // 是否为置顶话题

	VoteInfo    classdef.VoteInfo // 投票信息
	ShareOrigin ShareThread       // 转发来的原帖内容
	ViewNum     int64             // 浏览量
	ReplyNum    int64             // 回复数
	ShareNum    int64             // 分享数
	Agree       int64             // 点赞数
	Disagree    int64             // 点踩数
	CreateTime  int64             // 创建时间 10位时间戳 以秒为单位
	LastTime    int64             // 最后回复时间 10位时间戳 以秒为单位
}

// ThreadFromProto 对应 Thread.from_proto。
func ThreadFromProto(p *protobuf.ThreadInfo) *Thread {
	contents := ContentsTFromProto(p)

	typeValue := enums.ThreadTypeFrom(int(p.GetThreadType()))
	if typeValue == enums.ThreadTypeUnknown {
		logging.GetLogger().Debug("unknown thread type", "tid", p.GetId(), "type", p.GetThreadType())
	}

	isShare := p.GetIsShareThread() != 0
	shareOrigin := ShareThread{}
	if isShare {
		if p.GetOriginThreadInfo().GetPid() != 0 {
			shareOrigin = ShareThreadFromProto(p.GetOriginThreadInfo())
		} else {
			isShare = false
		}
	}

	return &Thread{
		Contents:    contents,
		Title:       p.GetTitle(),
		TID:         p.GetId(),
		PID:         p.GetFirstPostId(),
		AuthorID:    p.GetAuthorId(),
		Type:        typeValue,
		TabID:       p.GetTabId(),
		IsGood:      p.GetIsGood() != 0,
		IsTop:       p.GetIsTop() != 0,
		IsShare:     isShare,
		IsHide:      p.GetIsFrsMask() != 0,
		IsLivepost:  p.GetIsLivepost() != 0,
		VoteInfo:    classdef.VoteInfoFromProto(p.GetPollInfo()),
		ShareOrigin: shareOrigin,
		ViewNum:     int64(p.GetViewNum()),
		ReplyNum:    int64(p.GetReplyNum()),
		ShareNum:    int64(p.GetShareNum()),
		Agree:       p.GetAgree().GetAgreeNum(),
		Disagree:    p.GetAgree().GetDisagreeNum(),
		CreateTime:  int64(p.GetCreateTime()),
		LastTime:    int64(p.GetLastTimeInt()),
	}
}

// ThreadFromFeed 对应 Thread.from_feed。
func ThreadFromFeed(p *pb.PageData_LayoutFactory_FeedLayout, businessInfo map[string]string) *Thread {
	contents := ContentsTFromFeed(p)

	voteInfo := classdef.VoteInfo{}
	for _, component := range p.GetComponents() {
		if component.GetComponent() != "feed_poll" {
			continue
		}
		voteInfo = classdef.VoteInfoFromProto(component.GetFeedPoll())
	}

	return &Thread{
		Contents:    contents,
		Title:       businessInfo["title"],
		TID:         classdef.ParseInt64OrZero(businessInfo["thread_id"]),
		AuthorID:    classdef.ParseInt64OrZero(businessInfo["user_id"]),
		Type:        enums.ThreadTypeFrom(int(classdef.ParseInt64OrZero(businessInfo["thread_type"]))),
		TabID:       int32(classdef.ParseInt64OrZero(businessInfo["inner_tab_id"])),
		VoteInfo:    voteInfo,
		ShareOrigin: ShareThread{},
		ViewNum:     classdef.ParseInt64OrZero(businessInfo["view_num"]),
		CreateTime:  classdef.ParseInt64OrZero(businessInfo["create_time"]),
	}
}

// Text 文本内容。
func (t *Thread) Text() string {
	if t.Title != "" {
		return t.Title + "\n" + t.Contents.Text()
	}
	return t.Contents.Text()
}

// Equal 对应 __eq__。
func (t *Thread) Equal(other *Thread) bool {
	return other != nil && t.PID == other.PID
}

// ForumT 吧信息。
type ForumT struct {
	FID   int64  // 贴吧id
	FName string // 贴吧名

	Category    string // 一级分类
	Subcategory string // 二级分类

	MemberNum int64 // 吧会员数
	PostNum   int64 // 发帖数
	ThreadNum int64 // 主题帖数

	HasBawu bool // 是否有吧务
	HasRule bool // 是否有吧规
}

// ForumTFromProto 对应 Forum_t.from_proto。
func ForumTFromProto(p *pb.FrsPageResIdl_DataRes) ForumT {
	forum := p.GetForum()
	return ForumT{
		FID:         forum.GetId(),
		FName:       forum.GetName(),
		Category:    forum.GetFirstClass(),
		Subcategory: forum.GetSecondClass(),
		MemberNum:   int64(forum.GetMemberNum()),
		PostNum:     int64(forum.GetPostNum()),
		ThreadNum:   int64(forum.GetThreadNum()),
		HasBawu:     len(forum.GetManagers()) != 0,
		HasRule:     p.GetForumRule().GetHasForumRule() != 0,
	}
}

// Threads 主题帖列表。
//
// Threads 是 get_threads 的返回结果。
type Threads struct {
	classdef.Containers[*Thread]

	Page   PageT            // 页信息
	Forum  ForumT           // 所在吧信息
	TabMap map[string]int32 // 分区名到分区id的映射表
	Err    error            // 捕获的异常
}

// ThreadsFromProto 对应 Threads.from_proto。
func ThreadsFromProto(p *pb.FrsPageResIdl_DataRes) Threads {
	page := PageTFromProto(p.GetPage())
	forum := ForumTFromProto(p)

	tabMap := make(map[string]int32, len(p.GetNavTabInfo().GetTab()))
	for _, tab := range p.GetNavTabInfo().GetTab() {
		tabMap[tab.GetTabName()] = tab.GetTabId()
	}

	threads := make([]*Thread, 0, len(p.GetThreadList()))
	for _, proto := range p.GetThreadList() {
		threads = append(threads, ThreadFromProto(proto))
	}
	users := make(map[int64]UserInfoT, len(p.GetUserList()))
	for _, proto := range p.GetUserList() {
		users[proto.GetId()] = UserInfoTFromProto(proto)
	}
	for _, thread := range threads {
		thread.FName = forum.FName
		thread.FID = forum.FID
		if user, ok := users[thread.AuthorID]; ok {
			thread.User = user
		}
	}

	return Threads{
		Containers: classdef.Containers[*Thread]{Objs: threads},
		Page:       page,
		Forum:      forum,
		TabMap:     tabMap,
	}
}

// ThreadsFromFeed 对应 Threads.from_feed。
func ThreadsFromFeed(p *pb.FrsPageResIdl_DataRes) Threads {
	page := PageTFromProto(p.GetPage())
	forum := ForumTFromProto(p)

	tabMap := make(map[string]int32, len(p.GetNavTabInfo().GetTab()))
	for _, tab := range p.GetNavTabInfo().GetTab() {
		tabMap[tab.GetTabName()] = tab.GetTabId()
	}

	threads := make([]*Thread, 0, len(p.GetPageData().GetFeedList()))
	for _, layout := range p.GetPageData().GetFeedList() {
		if layout.GetLayout() != "feed" {
			continue
		}
		feed := layout.GetFeed()
		businessInfo := make(map[string]string, len(feed.GetBusinessInfo()))
		for _, kv := range feed.GetBusinessInfo() {
			businessInfo[kv.GetKey()] = kv.GetValue()
		}
		thread := ThreadFromFeed(feed, businessInfo)
		thread.FName = forum.FName
		thread.FID = forum.FID
		threads = append(threads, thread)
	}

	return Threads{
		Containers: classdef.Containers[*Thread]{Objs: threads},
		Page:       page,
		Forum:      forum,
		TabMap:     tabMap,
	}
}

// HasMore 是否还有下一页。
func (t Threads) HasMore() bool { return t.Page.HasMore }
