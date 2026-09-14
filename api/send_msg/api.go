// Package sendmsg 实现 aiotieba 的 send_msg API。
//
// 对应 Python 包 aiotieba.api.send_msg。该 API 仅在 websocket 传输下可用。
package sendmsg

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/send_msg/protobuf"
)

// CMD 是 send_msg 的 websocket 命令字。
const CMD = 205001

// PackProto 构造 CommitPersonalMsgReqIdl 请求，对应 pack_proto。
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

// ParseBody 解析 CommitPersonalMsgResIdl 响应并返回 msg id，对应 parse_body。
//
// 该接口通过两种信封上报失败：常规 error 以及消息专用的 blockInfo。
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

// Request 执行 websocket 请求，对应 request。
//
// record id 取自 websocket core 的消息 id 管理器。
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
