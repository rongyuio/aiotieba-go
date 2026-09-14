// Package getgroupmsg 实现 aiotieba 的 get_group_msg API。
//
// 对应 Python 包 aiotieba.api.get_group_msg。该 API 仅可通过 websocket 传输使用。
package getgroupmsg

import (
	"strconv"
	"strings"

	"github.com/rongyuio/aiotieba-go/api/classdef"
	pb "github.com/rongyuio/aiotieba-go/api/get_group_msg/protobuf"
)

// UserInfoWS 用户信息。
type UserInfoWS struct {
	UserID   int64  // user_id
	Portrait string // portrait
	UserName string // 用户名
}

// UserInfoWSFromProto 对应 UserInfo_ws.from_proto。
func UserInfoWSFromProto(p *pb.GetGroupMsgResIdl_DataRes_GroupMsg_MsgInfo_UserInfo) UserInfoWS {
	portrait := p.GetPortrait()
	// portrait 携带 "?..." 查询后缀，客户端会将其去除。
	if strings.Contains(portrait, "?") && len(portrait) > 13 {
		portrait = portrait[:len(portrait)-13]
	}

	return UserInfoWS{
		UserID:   p.GetUserId(),
		Portrait: portrait,
		UserName: p.GetUserName(),
	}
}

// String 对应 __str__。
func (u UserInfoWS) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// LogName 用于在日志中记录用户信息。
func (u UserInfoWS) LogName() string { return u.String() }

// Valid 对应 __bool__。
func (u UserInfoWS) Valid() bool { return u.UserID != 0 }

// WsMessage websocket消息。
type WsMessage struct {
	MsgID      int64      // 消息id
	MsgType    int64      // 消息类型
	Text       string     // 文本内容
	User       UserInfoWS // 用户信息
	CreateTime int64      // 发送时间 10位时间戳 以秒为单位
}

// WsMessageFromProto 对应 WsMessage.from_proto。
func WsMessageFromProto(p *pb.GetGroupMsgResIdl_DataRes_GroupMsg_MsgInfo) WsMessage {
	return WsMessage{
		MsgID:      p.GetMsgId(),
		MsgType:    int64(p.GetMsgType()),
		Text:       p.GetContent(),
		User:       UserInfoWSFromProto(p.GetUserInfo()),
		CreateTime: int64(p.GetCreateTime()),
	}
}

// WsMsgGroup websocket消息组。
type WsMsgGroup struct {
	GroupID   int64       // 消息组id
	GroupType int64       // 消息组类别
	Messages  []WsMessage // 消息列表
}

// WsMsgGroupFromProto 对应 WsMsgGroup.from_proto。
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

// WsMsgGroups websocket消息组列表。
type WsMsgGroups struct {
	classdef.Containers[WsMsgGroup]
}

// WsMsgGroupsFromProto 对应 WsMsgGroups.from_proto。
func WsMsgGroupsFromProto(p *pb.GetGroupMsgResIdl_DataRes) WsMsgGroups {
	list := p.GetGroupInfo()
	objs := make([]WsMsgGroup, 0, len(list))
	for _, item := range list {
		objs = append(objs, WsMsgGroupFromProto(item))
	}
	return WsMsgGroups{Containers: classdef.Containers[WsMsgGroup]{Objs: objs}}
}
