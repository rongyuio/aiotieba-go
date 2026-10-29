// Package getreplys 实现 aiotieba 的 get_replys API。
//
// 对应 Python 包 aiotieba.api.get_replys。
package getreplys

import (
	"strconv"
	"strings"

	"github.com/rongyuio/aiotieba-go/api/classdef"
	pb "github.com/rongyuio/aiotieba-go/api/get_replys/protobuf"
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/protobuf"
)

// FragTextReply 纯文本碎片，对应上游的 `FragText_rep = FragText`。
type FragTextReply = classdef.FragText

// FragEmojiReply 表情碎片，对应上游的 `FragEmoji_rep = FragEmoji`。
type FragEmojiReply = classdef.FragEmoji

// FragAtReply @碎片，对应上游的 `FragAt_rep = FragAt`。
type FragAtReply = classdef.FragAt

// ContentsReply 内容碎片列表。
type ContentsReply struct {
	classdef.Containers[classdef.Fragment]

	Texts  []classdef.Fragment // 纯文本碎片列表
	Emojis []FragEmojiReply    // 表情碎片列表
	Ats    []FragAtReply       // @碎片列表
}

// ContentsReplyFromProto 对应 Contents_rep.from_proto，参数是 new_floor_info 的单项。
//
// 只识别文本（0）、表情（2）与 @（4）三类碎片，其余一律记为未知碎片。
// @碎片同时计入 texts，与文本碎片一起参与 text 的拼接。
func ContentsReplyFromProto(p *pb.ReplyMeResIdl_DataRes_ReplyList_NewFloorInfo) ContentsReply {
	c := ContentsReply{}
	for _, proto := range p.GetContent() {
		switch proto.GetType() {
		case 0:
			frag := classdef.FragTextFromProto(proto)
			c.Texts = append(c.Texts, frag)
			c.Objs = append(c.Objs, frag)
		case 2:
			frag := classdef.FragEmojiFromProto(proto)
			c.Emojis = append(c.Emojis, frag)
			c.Objs = append(c.Objs, frag)
		case 4:
			frag := classdef.FragAtFromProto(proto)
			c.Ats = append(c.Ats, frag)
			c.Texts = append(c.Texts, frag)
			c.Objs = append(c.Objs, frag)
		default:
			c.Objs = append(c.Objs, classdef.FragUnknownFromProto(proto))
		}
	}
	return c
}

// Text 文本内容，对应 text 缓存属性。
func (c ContentsReply) Text() string { return classdef.FragmentTextOf(c.Texts) }

// stripQuoteHeader 去掉父级楼层正文里的楼中楼引述头，对应 _strip_quote_header。
//
// 正文以「@某人: 正文」开头时，@碎片与其后的 ": " 前缀一并剥掉；
// 剥完正文为空时，连那个空的文本碎片也一起丢弃。
func stripQuoteHeader(c *ContentsReply) {
	objs := c.Objs
	if len(objs) < 2 {
		return
	}
	if _, ok := objs[0].(FragAtReply); !ok {
		return
	}

	textFrag, ok := objs[1].(FragTextReply)
	if !ok || !strings.HasPrefix(textFrag.Text, ": ") {
		return
	}

	textFrag.Text = strings.TrimPrefix(textFrag.Text, ": ")
	objs[1] = textFrag
	// texts 里存的是同一个碎片的副本，值语义要求显式同步改写。
	if len(c.Texts) > 1 {
		if _, ok := c.Texts[1].(FragTextReply); ok {
			c.Texts[1] = textFrag
		}
	}

	skip := 1
	if textFrag.Text == "" {
		skip = 2
	}

	c.Objs = objs[skip:]
	c.Texts = c.Texts[skip:]
	c.Ats = c.Ats[1:]
}

// UserInfoReply 用户信息。
type UserInfoReply struct {
	UserID      int64  // user_id
	Portrait    string // portrait
	UserName    string // 用户名
	NickNameNew string // 新版昵称

	PrivLike  enums.PrivLike  // 关注吧列表的公开状态
	PrivReply enums.PrivReply // 帖子评论权限
}

// UserInfoReplyFromProto 对应 UserInfo_rep.from_proto。
func UserInfoReplyFromProto(p *protobuf.User) UserInfoReply {
	return UserInfoReply{
		UserID:      p.GetId(),
		Portrait:    classdef.TrimPortrait(p.GetPortrait()),
		UserName:    p.GetName(),
		NickNameNew: p.GetNameShow(),
		PrivLike:    classdef.UserPrivLike(p),
		PrivReply:   classdef.UserPrivReply(p),
	}
}

// NickName 用户昵称。
func (u UserInfoReply) NickName() string { return u.NickNameNew }

// ShowName 显示名称。
func (u UserInfoReply) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// String 对应 __str__。
func (u UserInfoReply) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// LogName 用于在日志中记录用户信息。
func (u UserInfoReply) LogName() string {
	switch {
	case u.UserName != "":
		return u.UserName
	case u.Portrait != "":
		return u.NickNameNew + "/" + u.Portrait
	default:
		return strconv.FormatInt(u.UserID, 10)
	}
}

// UserInfoReplyP 用户信息。
type UserInfoReplyP struct {
	UserID      int64  // user_id
	UserName    string // 用户名
	NickNameNew string // 新版昵称
}

// UserInfoReplyPFromProto 对应 UserInfo_rep_p.from_proto。
func UserInfoReplyPFromProto(p *protobuf.User) UserInfoReplyP {
	return UserInfoReplyP{
		UserID:      p.GetId(),
		UserName:    p.GetName(),
		NickNameNew: p.GetNameShow(),
	}
}

// NickName 用户昵称。
func (u UserInfoReplyP) NickName() string { return u.NickNameNew }

// ShowName 显示名称。
func (u UserInfoReplyP) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// String 对应 __str__。
func (u UserInfoReplyP) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	return strconv.FormatInt(u.UserID, 10)
}

// LogName 用于在日志中记录用户信息。
func (u UserInfoReplyP) LogName() string {
	if u.UserName != "" {
		return u.UserName
	}
	return u.NickNameNew + "/" + strconv.FormatInt(u.UserID, 10)
}

// UserInfoReplyT 用户信息。
type UserInfoReplyT struct {
	UserID      int64  // user_id
	Portrait    string // portrait
	NickNameNew string // 新版昵称
}

// UserInfoReplyTFromProto 对应 UserInfo_rep_t.from_proto。
func UserInfoReplyTFromProto(p *protobuf.User) UserInfoReplyT {
	return UserInfoReplyT{
		UserID:      p.GetId(),
		Portrait:    p.GetPortrait(),
		NickNameNew: p.GetNameShow(),
	}
}

// NickName 用户昵称。
func (u UserInfoReplyT) NickName() string { return u.NickNameNew }

// ShowName 显示名称。
func (u UserInfoReplyT) ShowName() string { return u.NickNameNew }

// String 对应 __str__。
func (u UserInfoReplyT) String() string {
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// LogName 用于在日志中记录用户信息。
func (u UserInfoReplyT) LogName() string {
	if u.Portrait == "" {
		return strconv.FormatInt(u.UserID, 10)
	}
	return u.NickNameNew + "/" + u.Portrait
}

// PostReply 父级回复信息。
type PostReply struct {
	PID      int64          // 父级回复id
	Contents ContentsReply  // 正文内容碎片列表
	User     UserInfoReplyP // 发布者的用户信息
}

// PostReplyFromProto 对应 Post_rep.from_proto。
//
// 父级回复的正文取自 new_floor_info 的倒数第二项，长度不足三项时视为无引用。
func PostReplyFromProto(p *pb.ReplyMeResIdl_DataRes_ReplyList) PostReply {
	contents := ContentsReply{}
	if infos := p.GetNewFloorInfo(); len(infos) > 2 {
		contents = ContentsReplyFromProto(infos[len(infos)-2])
		stripQuoteHeader(&contents)
	}
	return PostReply{
		PID:      int64(p.GetQuotePid()),
		Contents: contents,
		User:     UserInfoReplyPFromProto(p.GetQuoteUser()),
	}
}

// Valid 对应 Post_rep 的 __bool__。
func (p PostReply) Valid() bool { return !p.Contents.Empty() }

// ThreadReply 父级主题帖信息。
type ThreadReply struct {
	TID      int64          // 父级主题帖id
	Contents ContentsReply  // 正文内容碎片列表
	User     UserInfoReplyT // 发布者的用户信息
}

// ThreadReplyFromProto 对应 Thread_rep.from_proto。
//
// 主题帖正文取自 new_floor_info 的首项，长度不足两项时视为无正文。
func ThreadReplyFromProto(p *pb.ReplyMeResIdl_DataRes_ReplyList) ThreadReply {
	contents := ContentsReply{}
	if infos := p.GetNewFloorInfo(); len(infos) > 1 {
		contents = ContentsReplyFromProto(infos[0])
		stripQuoteHeader(&contents)
	}
	return ThreadReply{
		TID:      int64(p.GetThreadId()),
		Contents: contents,
		User:     UserInfoReplyTFromProto(p.GetThreadAuthorUser()),
	}
}

// Valid 对应 Thread_rep 的 __bool__。
func (t ThreadReply) Valid() bool { return !t.Contents.Empty() }

// Reply 回复信息。
type Reply struct {
	Contents ContentsReply // 正文内容碎片列表

	FName       string            // 所在贴吧名
	PID         int64             // 回复id
	User        UserInfoReply     // 发布者的用户信息
	Post        PostReply         // 父级回复信息
	Thread      ThreadReply       // 父级主题帖信息
	ContentType enums.ContentType // 帖子对象类型
	CreateTime  int64             // 创建时间 10位时间戳 以秒为单位
}

// ReplyFromProto 对应 Reply.from_proto。
//
// 回复正文取自 new_floor_info 的末项，该列表为空时视为无正文。
func ReplyFromProto(p *pb.ReplyMeResIdl_DataRes_ReplyList) Reply {
	contents := ContentsReply{}
	if infos := p.GetNewFloorInfo(); len(infos) > 0 {
		contents = ContentsReplyFromProto(infos[len(infos)-1])
	}

	contentType := enums.ContentTypePost
	if p.GetIsFloor() != 0 {
		contentType = enums.ContentTypeComment
	}

	return Reply{
		Contents:    contents,
		FName:       p.GetFname(),
		PID:         int64(p.GetPostId()),
		User:        UserInfoReplyFromProto(p.GetReplyer()),
		Post:        PostReplyFromProto(p),
		Thread:      ThreadReplyFromProto(p),
		ContentType: contentType,
		CreateTime:  int64(p.GetTime()),
	}
}

// Text 文本内容，对应 text 属性。
func (r Reply) Text() string { return r.Contents.Text() }

// AuthorID 发布者的user_id。
func (r Reply) AuthorID() int64 { return r.User.UserID }

// PageReply 页信息。
type PageReply struct {
	CurrentPage int64 // 当前页码
	HasMore     bool  // 是否有后继页
	HasPrev     bool  // 是否有前驱页
}

// PageReplyFromProto 对应 Page_rep.from_proto。
func PageReplyFromProto(p *protobuf.Page) PageReply {
	return PageReply{
		CurrentPage: int64(p.GetCurrentPage()),
		HasMore:     p.GetHasMore() != 0,
		HasPrev:     p.GetHasPrev() != 0,
	}
}

// Replys 收到回复列表。
type Replys struct {
	classdef.Containers[Reply]

	Page PageReply // 页信息
}

// ReplysFromProto 对应 Replys.from_proto。
func ReplysFromProto(p *pb.ReplyMeResIdl_DataRes) Replys {
	list := p.GetReplyList()
	objs := make([]Reply, 0, len(list))
	for _, item := range list {
		objs = append(objs, ReplyFromProto(item))
	}

	return Replys{
		Containers: classdef.Containers[Reply]{Objs: objs},
		Page:       PageReplyFromProto(p.GetPage()),
	}
}

// HasMore 是否还有下一页。
func (r Replys) HasMore() bool { return r.Page.HasMore }
