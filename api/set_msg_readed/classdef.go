package setmsgreaded

import (
	pb "github.com/rongyuio/aiotieba-go/api/set_msg_readed/protobuf"
)

// SetMsgReadedResult 设置私信已读的返回值。
type SetMsgReadedResult struct {
	ReadMsgID  int64 // 服务端记录的已读位置msg_id
	GroupID    int64 // 私信消息组id
	ToUID      int64 // 对端user_id
	ToUserType int32 // 对端用户类别
}

// SetMsgReadedResultFromProto 对应 SetMsgReadedResult.from_proto。
func SetMsgReadedResultFromProto(p *pb.CommitReceivedPmsgResIdl_DataRes) SetMsgReadedResult {
	return SetMsgReadedResult{
		ReadMsgID:  p.GetReadMsgId(),
		GroupID:    p.GetGroupId(),
		ToUID:      p.GetToUid(),
		ToUserType: p.GetToUserType(),
	}
}
