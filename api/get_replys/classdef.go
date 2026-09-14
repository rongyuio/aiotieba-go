// Package getreplys implements the get_replys API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_replys.
package getreplys

import (
	"strconv"
	"strings"

	"github.com/rongyuio/aiotieba-go/api/classdef"
	pb "github.com/rongyuio/aiotieba-go/api/get_replys/protobuf"
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/protobuf"
)

// UserInfoReply is the information of the user who replied. It mirrors
// aiotieba.api.get_replys._classdef.UserInfo_reply.
type UserInfoReply struct {
	UserID      int64
	Portrait    string
	UserName    string
	NickNameNew string

	PrivLike  enums.PrivLike
	PrivReply enums.PrivReply
}

// UserInfoReplyFromProto mirrors UserInfo_reply.from_proto.
func UserInfoReplyFromProto(p *protobuf.User) UserInfoReply {
	portrait := p.GetPortrait()
	// The portrait carries a "?..." query suffix that the client strips.
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
func (u UserInfoReply) NickName() string { return u.NickNameNew }

// ShowName mirrors the show_name property.
func (u UserInfoReply) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// String mirrors __str__.
func (u UserInfoReply) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// LogName mirrors the log_name property.
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

// UserInfoReplyP is the information of the user of the quoted floor. It mirrors
// aiotieba.api.get_replys._classdef.UserInfo_reply_p.
type UserInfoReplyP struct {
	UserID      int64
	UserName    string
	NickNameNew string
}

// UserInfoReplyPFromProto mirrors UserInfo_reply_p.from_proto.
func UserInfoReplyPFromProto(p *protobuf.User) UserInfoReplyP {
	return UserInfoReplyP{
		UserID:      p.GetId(),
		UserName:    p.GetName(),
		NickNameNew: p.GetNameShow(),
	}
}

// NickName mirrors the nick_name property.
func (u UserInfoReplyP) NickName() string { return u.NickNameNew }

// ShowName mirrors the show_name property.
func (u UserInfoReplyP) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// String mirrors __str__.
func (u UserInfoReplyP) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	return strconv.FormatInt(u.UserID, 10)
}

// LogName mirrors the log_name property.
func (u UserInfoReplyP) LogName() string {
	if u.UserName != "" {
		return u.UserName
	}
	return u.NickNameNew + "/" + strconv.FormatInt(u.UserID, 10)
}

// UserInfoReplyT is the information of the thread author. It mirrors
// aiotieba.api.get_replys._classdef.UserInfo_reply_t.
type UserInfoReplyT struct {
	UserID      int64
	Portrait    string
	NickNameNew string
}

// UserInfoReplyTFromProto mirrors UserInfo_reply_t.from_proto.
func UserInfoReplyTFromProto(p *protobuf.User) UserInfoReplyT {
	return UserInfoReplyT{
		UserID:      p.GetId(),
		Portrait:    p.GetPortrait(),
		NickNameNew: p.GetNameShow(),
	}
}

// NickName mirrors the nick_name property.
func (u UserInfoReplyT) NickName() string { return u.NickNameNew }

// ShowName mirrors the show_name property.
func (u UserInfoReplyT) ShowName() string { return u.NickNameNew }

// String mirrors __str__.
func (u UserInfoReplyT) String() string {
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// LogName mirrors the log_name property.
func (u UserInfoReplyT) LogName() string {
	if u.Portrait == "" {
		return strconv.FormatInt(u.UserID, 10)
	}
	return u.NickNameNew + "/" + u.Portrait
}

// Reply is one reply received by the logged in account. It mirrors
// aiotieba.api.get_replys._classdef.Reply.
type Reply struct {
	Text  string
	FName string
	TID   int64
	PPID  int64
	PID   int64

	User       UserInfoReply
	PostUser   UserInfoReplyP
	ThreadUser UserInfoReplyT

	IsComment  bool
	CreateTime int64
}

// ReplyFromProto mirrors Reply.from_proto.
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

// AuthorID mirrors the author_id property.
func (r Reply) AuthorID() int64 { return r.User.UserID }

// PageReply is the pagination information of the reply list. It mirrors
// aiotieba.api.get_replys._classdef.Page_reply.
type PageReply struct {
	CurrentPage int64
	HasMore     bool
	HasPrev     bool
}

// PageReplyFromProto mirrors Page_reply.from_proto.
func PageReplyFromProto(p *protobuf.Page) PageReply {
	return PageReply{
		CurrentPage: int64(p.GetCurrentPage()),
		HasMore:     p.GetHasMore() != 0,
		HasPrev:     p.GetHasPrev() != 0,
	}
}

// Replys is the list of replies received by the logged in account. It mirrors
// aiotieba.api.get_replys._classdef.Replys.
type Replys struct {
	classdef.Containers[Reply]

	Page PageReply
}

// ReplysFromProto mirrors Replys.from_proto.
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

// HasMore reports whether a next page exists, mirroring the has_more property.
func (r Replys) HasMore() bool { return r.Page.HasMore }
