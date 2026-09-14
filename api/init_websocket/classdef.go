// Package initwebsocket 实现 aiotieba 的 init_websocket API。
//
// 对应 Python 包 aiotieba.api.init_websocket。
package initwebsocket

import (
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
