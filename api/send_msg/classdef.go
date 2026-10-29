package sendmsg

import (
	pb "github.com/rongyuio/aiotieba-go/api/send_msg/protobuf"
)

// SendMsgResult 发送私信的返回值。
type SendMsgResult struct {
	GroupID    int64 // 私信消息组id
	LocalMsgID int64 // 本端消息id 由服务端回显
	ToUID      int64 // 接收方user_id
	ToUserType int32 // 接收方用户类别
}

// SendMsgResultFromProto 对应 SendMsgResult.from_proto。
func SendMsgResultFromProto(p *pb.CommitPersonalMsgResIdl_DataRes) SendMsgResult {
	return SendMsgResult{
		GroupID:    p.GetGroupId(),
		LocalMsgID: p.GetRecordId(),
		ToUID:      p.GetToUid(),
		ToUserType: p.GetToUserType(),
	}
}
