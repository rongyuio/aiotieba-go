// Package initwebsocket 实现 aiotieba 的 init_websocket API。
//
// 对应 Python 包 aiotieba.api.init_websocket。
package initwebsocket

import (
	"github.com/rongyuio/aiotieba-go/api/classdef"
	pb "github.com/rongyuio/aiotieba-go/api/init_websocket/protobuf"
)

// WsMsgGroupInfo websocket消息组的相关信息。
type WsMsgGroupInfo struct {
	GroupID   int64 // 消息组id
	GroupType int32 // 消息组类别
	LastMsgID int64 // 最新消息的id
}

// WsMsgGroupInfoFromProto 对应 WsMsgGroupInfo.from_proto。
func WsMsgGroupInfoFromProto(p *pb.UpdateClientInfoResIdl_DataRes_GroupInfo) WsMsgGroupInfo {
	return WsMsgGroupInfo{
		GroupID:   p.GetGroupId(),
		GroupType: p.GetGroupType(),
		LastMsgID: p.GetLastMsgId(),
	}
}

// WsUserInfo websocket 初始化回包里的当前账号信息。
type WsUserInfo struct {
	UserID   int64  // user_id
	Portrait string // portrait
	UserName string // 用户名
}

// WsUserInfoFromProto 对应 UserInfo_ws.from_proto。
func WsUserInfoFromProto(p *pb.UpdateClientInfoResIdl_DataRes_UserInfo) WsUserInfo {
	return WsUserInfo{
		UserID:   p.GetUserId(),
		Portrait: classdef.TrimPortrait(p.GetPortrait()),
		UserName: p.GetUserName(),
	}
}

// InitWsInfo websocket 初始化的回包信息。
type InitWsInfo struct {
	Groups []WsMsgGroupInfo // 消息组信息列表
	User   WsUserInfo       // 当前账号的基础信息
}
