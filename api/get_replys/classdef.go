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

// UserInfoReply 用户信息。
type UserInfoReply struct {
	UserID      int64  // user_id
	Portrait    string // portrait
	UserName    string // 用户名
	NickNameNew string // 新版昵称

	PrivLike  enums.PrivLike  // 关注吧列表的公开状态
	PrivReply enums.PrivReply // 帖子评论权限
}

// UserInfoReplyFromProto 对应 UserInfo_reply.from_proto。
func UserInfoReplyFromProto(p *protobuf.User) UserInfoReply {
	portrait := p.GetPortrait()
	// portrait 携带 "?..." 查询后缀，客户端会将其去除。
	if strings.Contains(portrait, "?") && len(portrait) > 13 {
		portrait = portrait[:len(portrait)-13]
	}

	return UserInfoReply{
		UserID:      p.GetId(),
		Portrait:    portrait,
		UserName:    p.GetName(),
		NickNameNew: p.GetNameShow(),
		PrivLike:    privLikeOf(p.GetPrivSets().GetLike()),
		PrivReply:   privReplyOf(p.GetPrivSets().GetReply()),
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

// UserInfoReplyPFromProto 对应 UserInfo_reply_p.from_proto。
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

// UserInfoReplyTFromProto 对应 UserInfo_reply_t.from_proto。
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

// Reply 回复信息。
type Reply struct {
	Text  string // 文本内容
	FName string // 所在贴吧名
	TID   int64  // 所在主题帖id
	PPID  int64  // 所在楼层pid
	PID   int64  // 回复id

	User       UserInfoReply  // 发布者的用户信息
	PostUser   UserInfoReplyP // 楼层用户信息
	ThreadUser UserInfoReplyT // 楼主用户信息

	IsComment  bool  // 是否楼中楼
	CreateTime int64 // 创建时间 10位时间戳 以秒为单位
}

// ReplyFromProto 对应 Reply.from_proto。
func ReplyFromProto(p *pb.ReplyMeResIdl_DataRes_ReplyList) Reply {
	return Reply{
		Text:       p.GetContent(),
		FName:      p.GetFname(),
		TID:        int64(p.GetThreadId()),
		PPID:       int64(p.GetQuotePid()),
		PID:        int64(p.GetPostId()),
		User:       UserInfoReplyFromProto(p.GetReplyer()),
		PostUser:   UserInfoReplyPFromProto(p.GetQuoteUser()),
		ThreadUser: UserInfoReplyTFromProto(p.GetThreadAuthorUser()),
		IsComment:  p.GetIsFloor() != 0,
		CreateTime: int64(p.GetTime()),
	}
}

// AuthorID 发布者的user_id。
func (r Reply) AuthorID() int64 { return r.User.UserID }

// PageReply 页信息。
type PageReply struct {
	CurrentPage int64 // 当前页码
	HasMore     bool  // 是否有后继页
	HasPrev     bool  // 是否有前驱页
}

// PageReplyFromProto 对应 Page_reply.from_proto。
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
