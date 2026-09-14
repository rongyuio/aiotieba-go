// Package setmsgreaded 实现 aiotieba 的 set_msg_readed API。
//
// 对应 Python 包 aiotieba.api.set_msg_readed。该 API 仅可通过 websocket 传输。
package setmsgreaded

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/api/get_group_msg"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/set_msg_readed/protobuf"
)

// CMD 是 set_msg_readed 的 websocket 命令字。
const CMD = 205006

// PackProto 构造 CommitReceivedPmsgReqIdl 请求，对应 pack_proto。
func PackProto(userID, groupID, msgID int64) []byte {
	req := &pb.CommitReceivedPmsgReqIdl{
		Data: &pb.CommitReceivedPmsgReqIdl_DataReq{
			GroupId: groupID,
			ToUid:   userID,
			MsgType: int32(enums.MsgTypeReaded),
			MsgId:   msgID,
		},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil
	}
	return out
}

// ParseBody 解析 CommitReceivedPmsgResIdl 响应，对应 parse_body。
func ParseBody(body []byte) error {
	res := &pb.CommitReceivedPmsgResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return fmt.Errorf("decoding CommitReceivedPmsgResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}
	return nil
}

// Request 执行 websocket 请求，对应 request。
//
// 消息组为消息 id 管理器的私有组。
func Request(wsCore *core.WsCore, message getgroupmsg.WsMessage) error {
	groupID := int64(wsCore.MsgIDManager().PrivGID)

	resp, err := wsCore.Send(PackProto(message.User.UserID, groupID, message.MsgID), CMD)
	if err != nil {
		return err
	}
	payload, err := resp.Read()
	if err != nil {
		return err
	}
	return ParseBody(payload)
}
