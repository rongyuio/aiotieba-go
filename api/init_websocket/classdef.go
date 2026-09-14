// Package initwebsocket implements the init_websocket API of aiotieba.
//
// It mirrors the Python package aiotieba.api.init_websocket.
package initwebsocket

import (
	pb "github.com/rongyuio/aiotieba-go/api/init_websocket/protobuf"
)

// WsMsgGroupInfo is the information of one websocket message group. It mirrors
// aiotieba.api.init_websocket._classdef.WsMsgGroupInfo.
type WsMsgGroupInfo struct {
	GroupID   int64
	GroupType int32
	LastMsgID int64
}

// WsMsgGroupInfoFromProto mirrors WsMsgGroupInfo.from_proto.
func WsMsgGroupInfoFromProto(p *pb.UpdateClientInfoResIdl_DataRes_GroupInfo) WsMsgGroupInfo {
	return WsMsgGroupInfo{
		GroupID:   p.GetGroupId(),
		GroupType: p.GetGroupType(),
		LastMsgID: p.GetLastMsgId(),
	}
}
