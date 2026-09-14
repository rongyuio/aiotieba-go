// Package sendmsg implements the send_msg API of aiotieba.
//
// It mirrors the Python package aiotieba.api.send_msg. The API is only
// available over the websocket transport.
package sendmsg

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/send_msg/protobuf"
)

// CMD is the websocket command of send_msg.
const CMD = 205001

// PackProto builds the CommitPersonalMsgReqIdl request, mirroring pack_proto.
func PackProto(userID int64, content string, recordID int64) []byte {
	req := &pb.CommitPersonalMsgReqIdl{
		Data: &pb.CommitPersonalMsgReqIdl_DataReq{
			ToUid:    userID,
			Content:  content,
			MsgType:  int32(enums.MsgTypePrivateMsg),
			RecordId: recordID,
		},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil
	}
	return out
}

// ParseBody decodes a CommitPersonalMsgResIdl response and returns the msg id,
// mirroring parse_body.
//
// The endpoint reports failures through two envelopes: the usual error and a
// message-specific blockInfo.
func ParseBody(body []byte) (int64, error) {
	res := &pb.CommitPersonalMsgResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return 0, fmt.Errorf("decoding CommitPersonalMsgResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return 0, &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}

	data := res.GetData()
	if code := data.GetBlockInfo().GetBlockErrno(); code != 0 {
		return 0, &exception.TiebaServerError{Code: int(code), Msg: data.GetBlockInfo().GetBlockErrmsg()}
	}
	return data.GetMsgId(), nil
}

// Request performs the websocket request, mirroring request.
//
// The record id is taken from the message id manager of the websocket core.
func Request(wsCore *core.WsCore, userID int64, content string) (int64, error) {
	recordID := int64(wsCore.MsgIDManager().GetRecordID())

	resp, err := wsCore.Send(PackProto(userID, content, recordID), CMD)
	if err != nil {
		return 0, err
	}
	payload, err := resp.Read()
	if err != nil {
		return 0, err
	}
	return ParseBody(payload)
}
