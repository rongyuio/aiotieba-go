// Package pushnotify implements the push_notify API of aiotieba.
//
// It mirrors the Python package aiotieba.api.push_notify. The endpoint is
// pushed by the server, so the package only parses frames; register it on the
// websocket core with WsCore.RegisterCallback.
package pushnotify

import (
	"strconv"

	pb "github.com/rongyuio/aiotieba/api/push_notify/protobuf"
)

// WsNotify is a server pushed notification. It mirrors
// aiotieba.api.push_notify._classdef.WsNotify.
type WsNotify struct {
	NoteType   int64
	GroupID    int64
	GroupType  int64
	MsgID      int64
	CreateTime int64
}

// WsNotifyFromProto mirrors WsNotify.from_proto.
//
// `et` is a string timestamp which may be empty.
func WsNotifyFromProto(data *pb.PushNotifyResIdl_PusherMsg_PusherMsgInfo) WsNotify {
	var createTime int64
	if et := data.GetEt(); et != "" {
		createTime, _ = strconv.ParseInt(et, 10, 64)
	}

	return WsNotify{
		NoteType:   int64(data.GetType()),
		GroupID:    data.GetGroupId(),
		GroupType:  int64(data.GetGroupType()),
		MsgID:      data.GetMsgId(),
		CreateTime: createTime,
	}
}

// WsNotifiesFromProto mirrors the list comprehension of parse_body.
func WsNotifiesFromProto(res *pb.PushNotifyResIdl) []WsNotify {
	msgs := res.GetMultiMsg()
	notifies := make([]WsNotify, 0, len(msgs))
	for _, msg := range msgs {
		notifies = append(notifies, WsNotifyFromProto(msg.GetData()))
	}
	return notifies
}
