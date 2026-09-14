// Package setmsgreaded implements the set_msg_readed API of aiotieba.
//
// It mirrors the Python package aiotieba.api.set_msg_readed. The API is only
// available over the websocket transport.
package setmsgreaded

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba/api/get_group_msg"
	"github.com/rongyuio/aiotieba/core"
	"github.com/rongyuio/aiotieba/enums"
	"github.com/rongyuio/aiotieba/exception"

	pb "github.com/rongyuio/aiotieba/api/set_msg_readed/protobuf"
)

// CMD is the websocket command of set_msg_readed.
const CMD = 205006

// PackProto builds the CommitReceivedPmsgReqIdl request, mirroring pack_proto.
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

// ParseBody decodes a CommitReceivedPmsgResIdl response, mirroring parse_body.
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

// Request performs the websocket request, mirroring request.
//
// The message group is the private group of the message id manager.
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
