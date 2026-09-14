// Package pushnotify 实现 aiotieba 的 push_notify API。
//
// 对应 Python 包 aiotieba.api.push_notify。该端点由服务端推送，因此该包只解析帧；
// 请通过 WsCore.RegisterCallback 将其注册到 websocket 核心上。
package pushnotify

import (
	"strconv"

	pb "github.com/rongyuio/aiotieba-go/api/push_notify/protobuf"
)

// WsNotify websocket主动推送消息提醒。
type WsNotify struct {
	NoteType   int64 // 提醒类别
	GroupID    int64 // 消息组id
	GroupType  int64 // 消息组类别
	MsgID      int64 // 消息id
	CreateTime int64 // 推送时间 10位时间戳 以秒为单位
}

// WsNotifyFromProto 对应 WsNotify.from_proto。
//
// `et` 是字符串形式的时间戳，可能为空。
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

// WsNotifiesFromProto 对应 parse_body 的列表推导式。
func WsNotifiesFromProto(res *pb.PushNotifyResIdl) []WsNotify {
	msgs := res.GetMultiMsg()
	notifies := make([]WsNotify, 0, len(msgs))
	for _, msg := range msgs {
		notifies = append(notifies, WsNotifyFromProto(msg.GetData()))
	}
	return notifies
}
