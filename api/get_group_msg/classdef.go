// Package getgroupmsg implements the get_group_msg API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_group_msg. The API is only
// available over the websocket transport.
package getgroupmsg

import (
	"strconv"
	"strings"

	"github.com/rongyuio/aiotieba/api/classdef"
	pb "github.com/rongyuio/aiotieba/api/get_group_msg/protobuf"
)

// UserInfoWS is the information of the sender of a websocket message. It
// mirrors aiotieba.api.get_group_msg._classdef.UserInfo_ws.
type UserInfoWS struct {
	UserID   int64
	Portrait string
	UserName string
}

// UserInfoWSFromProto mirrors UserInfo_ws.from_proto.
func UserInfoWSFromProto(p *pb.GetGroupMsgResIdl_DataRes_GroupMsg_MsgInfo_UserInfo) UserInfoWS {
	portrait := p.GetPortrait()
	// The portrait carries a "?..." query suffix that the client strips.
	if strings.Contains(portrait, "?") && len(portrait) > 13 {
		portrait = portrait[:len(portrait)-13]
	}

	return UserInfoWS{
		UserID:   p.GetUserId(),
		Portrait: portrait,
		UserName: p.GetUserName(),
	}
}

// String mirrors __str__.
func (u UserInfoWS) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// LogName mirrors the log_name property.
func (u UserInfoWS) LogName() string { return u.String() }

// Valid mirrors __bool__.
func (u UserInfoWS) Valid() bool { return u.UserID != 0 }

// WsMessage is one websocket message. It mirrors
// aiotieba.api.get_group_msg._classdef.WsMessage.
type WsMessage struct {
	MsgID      int64
	MsgType    int64
	Text       string
	User       UserInfoWS
	CreateTime int64
}

// WsMessageFromProto mirrors WsMessage.from_proto.
func WsMessageFromProto(p *pb.GetGroupMsgResIdl_DataRes_GroupMsg_MsgInfo) WsMessage {
	return WsMessage{
		MsgID:      p.GetMsgId(),
		MsgType:    int64(p.GetMsgType()),
		Text:       p.GetContent(),
		User:       UserInfoWSFromProto(p.GetUserInfo()),
		CreateTime: int64(p.GetCreateTime()),
	}
}

// WsMsgGroup is a group of websocket messages. It mirrors
// aiotieba.api.get_group_msg._classdef.WsMsgGroup.
type WsMsgGroup struct {
	GroupID   int64
	GroupType int64
	Messages  []WsMessage
}

// WsMsgGroupFromProto mirrors WsMsgGroup.from_proto.
func WsMsgGroupFromProto(p *pb.GetGroupMsgResIdl_DataRes_GroupMsg) WsMsgGroup {
	group := p.GetGroupInfo()

	list := p.GetMsgList()
	messages := make([]WsMessage, 0, len(list))
	for _, item := range list {
		messages = append(messages, WsMessageFromProto(item))
	}

	return WsMsgGroup{
		GroupID:   group.GetGroupId(),
		GroupType: int64(group.GetGroupType()),
		Messages:  messages,
	}
}

// WsMsgGroups is the list of websocket message groups. It mirrors
// aiotieba.api.get_group_msg._classdef.WsMsgGroups.
type WsMsgGroups struct {
	classdef.Containers[WsMsgGroup]
}

// WsMsgGroupsFromProto mirrors WsMsgGroups.from_proto.
func WsMsgGroupsFromProto(p *pb.GetGroupMsgResIdl_DataRes) WsMsgGroups {
	list := p.GetGroupInfo()
	objs := make([]WsMsgGroup, 0, len(list))
	for _, item := range list {
		objs = append(objs, WsMsgGroupFromProto(item))
	}
	return WsMsgGroups{Containers: classdef.Containers[WsMsgGroup]{Objs: objs}}
}
