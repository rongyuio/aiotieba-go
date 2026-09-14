// Package getthreads implements the get_threads API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_threads.
package getthreads

import (
	"strconv"

	"github.com/rongyuio/aiotieba/api/classdef"
	"github.com/rongyuio/aiotieba/enums"
	"github.com/rongyuio/aiotieba/logging"
	"github.com/rongyuio/aiotieba/protobuf"

	pb "github.com/rongyuio/aiotieba/api/get_threads/protobuf"
)

// FragImageFeed is an image fragment of a feed card.
type FragImageFeed struct {
	Src       string
	BigSrc    string
	OriginSrc string
	Width     int64
	Height    int64
	Hash      string
}

// FragImageFeedFromProto mirrors FragImage_feed.from_proto.
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

// FragEmojiFeed is an emoji fragment of a feed card.
type FragEmojiFeed struct {
	ID   string
	Desc string
}

// FragEmojiFeedFromProto mirrors FragEmoji_feed.from_proto.
func FragEmojiFeedFromProto(p *pb.PageData_LayoutFactory_FeedLayout_ComponentFactory_FeedContentResource_FeedContentEmoji) FragEmojiFeed {
	return FragEmojiFeed{ID: p.GetName(), Desc: p.GetC()}
}

// PageT is the page information of a thread list.
type PageT struct {
	PageSize    int32
	CurrentPage int32
	TotalPage   int32
	TotalCount  int32
	HasMore     bool
	HasPrev     bool
}

// PageTFromProto mirrors Page_t.from_proto.
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

// UserInfoT is the user information attached to a thread.
type UserInfoT struct {
	UserID      int64
	Portrait    string
	UserName    string
	NickNameNew string

	Level  int32
	GLevel int32
	Gender enums.Gender
	Icons  []string

	IsBawu    bool
	IsVIP     bool
	IsGod     bool
	PrivLike  enums.PrivLike
	PrivReply enums.PrivReply
}

// UserInfoTFromProto mirrors UserInfo_t.from_proto.
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

// String mirrors __str__.
func (u UserInfoT) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// NickName mirrors the nick_name property.
func (u UserInfoT) NickName() string { return u.NickNameNew }

// ShowName mirrors the show_name property.
func (u UserInfoT) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// LogName mirrors the log_name property.
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

// Valid mirrors __bool__.
func (u UserInfoT) Valid() bool { return u.UserID != 0 }

// Equal mirrors __eq__.
func (u UserInfoT) Equal(other UserInfoT) bool { return u.UserID == other.UserID }

// FragImageSt is an image fragment of the thread body.
type FragImageSt struct {
	Src        string
	BigSrc     string
	OriginSrc  string
	ShowWidth  int32
	ShowHeight int32
	Hash       string
}

// FragImageStFromProto mirrors FragImage_st.from_proto.
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

// ContentsT is the body of a thread.
type ContentsT struct {
	classdef.Containers[any]

	Texts       []classdef.Fragment
	Emojis      []any
	Imgs        []any
	Ats         []classdef.FragAt
	Links       []classdef.FragLink
	TiebaPluses []classdef.FragTiebaPlus
	Video       classdef.FragVideo
	Voice       classdef.FragVoice
}

// ContentsTFromProto mirrors Contents_t.from_proto.
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
			// Voice and video are carried by dedicated proto fields.
		case t == 35 || t == 36 || t == 37:
			frag := classdef.FragTiebaPlusFromProto(proto)
			c.TiebaPluses = append(c.TiebaPluses, frag)
			c.Texts = append(c.Texts, frag)
			c.Objs = append(c.Objs, frag)
		case t == 34:
			// Outdated tiebaplus.
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

// ContentsTFromFeed mirrors Contents_t.from_feed.
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

// Text mirrors the `text` cached property.
func (c ContentsT) Text() string { return classdef.FragmentTextOf(c.Texts) }

// ContentsSt is the body of a shared thread.
type ContentsSt struct {
	classdef.Containers[any]

	Texts       []classdef.Fragment
	Emojis      []classdef.FragEmoji
	Imgs        []FragImageSt
	Ats         []classdef.FragAt
	Links       []classdef.FragLink
	TiebaPluses []classdef.FragTiebaPlus
	Video       classdef.FragVideo
	Voice       classdef.FragVoice
}

// ContentsStFromProto mirrors Contents_st.from_proto.
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
			// Video is carried by a dedicated proto field.
		case t == 35 || t == 36 || t == 37:
			frag := classdef.FragTiebaPlusFromProto(proto)
			c.TiebaPluses = append(c.TiebaPluses, frag)
			c.Texts = append(c.Texts, frag)
			c.Objs = append(c.Objs, frag)
		case t == 34:
			// Outdated tiebaplus.
		default:
			c.Objs = append(c.Objs, classdef.FragUnknownFromProto(proto))
		}
	}

	// The first @fragment is dropped and the media images are appended.
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

// Text mirrors the `text` cached property.
func (c ContentsSt) Text() string { return classdef.FragmentTextOf(c.Texts) }

// ShareThread is the origin of a shared thread.
type ShareThread struct {
	Contents ContentsSt
	Title    string

	AuthorID int64

	FID   int64
	FName string
	TID   int64
	PID   int64

	VoteInfo classdef.VoteInfo
}

// ShareThreadFromProto mirrors ShareThread.from_proto.
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

// Text mirrors the `text` cached property.
func (s ShareThread) Text() string {
	if s.Title != "" {
		return s.Title + "\n" + s.Contents.Text()
	}
	return s.Contents.Text()
}

// Thread is one thread of the list.
type Thread struct {
	Contents ContentsT
	Title    string

	FID      int64
	FName    string
	TID      int64
	PID      int64
	User     UserInfoT
	AuthorID int64

	Type       enums.ThreadType
	TabID      int32
	IsGood     bool
	IsTop      bool
	IsShare    bool
	IsHide     bool
	IsLivepost bool

	VoteInfo    classdef.VoteInfo
	ShareOrigin ShareThread
	ViewNum     int64
	ReplyNum    int64
	ShareNum    int64
	Agree       int64
	Disagree    int64
	CreateTime  int64
	LastTime    int64
}

// ThreadFromProto mirrors Thread.from_proto.
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

// ThreadFromFeed mirrors Thread.from_feed.
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

// Text mirrors the `text` cached property.
func (t *Thread) Text() string {
	if t.Title != "" {
		return t.Title + "\n" + t.Contents.Text()
	}
	return t.Contents.Text()
}

// Equal mirrors __eq__.
func (t *Thread) Equal(other *Thread) bool {
	return other != nil && t.PID == other.PID
}

// ForumT is the forum of a thread list.
type ForumT struct {
	FID   int64
	FName string

	Category    string
	Subcategory string

	MemberNum int64
	PostNum   int64
	ThreadNum int64

	HasBawu bool
	HasRule bool
}

// ForumTFromProto mirrors Forum_t.from_proto.
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

// Threads is the result of get_threads.
type Threads struct {
	classdef.Containers[*Thread]

	Page   PageT
	Forum  ForumT
	TabMap map[string]int32
	Err    error
}

// ThreadsFromProto mirrors Threads.from_proto.
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

// ThreadsFromFeed mirrors Threads.from_feed.
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

// HasMore mirrors the has_more property.
func (t Threads) HasMore() bool { return t.Page.HasMore }
