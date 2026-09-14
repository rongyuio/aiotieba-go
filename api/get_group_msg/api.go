package getgroupmsg

import (
	"fmt"
	"strconv"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/get_group_msg/protobuf"
)

// CMD is the websocket command of get_group_msg.
const CMD = 202003

// cuidSuffix is appended to the account cuid, mirroring the Python module.
const cuidSuffix = "|com.baidu.tieba_mini12.35.1.0"

// PackProto builds the GetGroupMsgReqIdl request, mirroring pack_proto.
//
// The group ids and message ids are zipped: extra entries of either slice are
// ignored, like Python's zip(strict=False).
func PackProto(account *core.Account, groupIDs, msgIDs []int64, getType int64) []byte {
	req := &pb.GetGroupMsgReqIdl{
		Cuid: account.Cuid() + cuidSuffix,
		Data: &pb.GetGroupMsgReqIdl_DataReq{
			Gettype: strconv.FormatInt(getType, 10),
		},
	}

	n := min(len(groupIDs), len(msgIDs))
	req.Data.GroupMids = make([]*pb.GetGroupMsgReqIdl_DataReq_GroupLastId, 0, n)
	for i := range n {
		req.Data.GroupMids = append(req.Data.GroupMids, &pb.GetGroupMsgReqIdl_DataReq_GroupLastId{
			GroupId:   groupIDs[i],
			LastMsgId: msgIDs[i],
		})
	}

	out, err := proto.Marshal(req)
	if err != nil {
		return nil
	}
	return out
}

// ParseBody decodes a GetGroupMsgResIdl response, mirroring parse_body.
func ParseBody(body []byte) (WsMsgGroups, error) {
	res := &pb.GetGroupMsgResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return WsMsgGroups{}, fmt.Errorf("decoding GetGroupMsgResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return WsMsgGroups{}, &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}
	return WsMsgGroupsFromProto(res.GetData()), nil
}

// Request performs the websocket request, mirroring request.
//
// The last known message id of every group is taken from the message id
// manager of the websocket core.
func Request(wsCore *core.WsCore, groupIDs []int64, getType int64) (WsMsgGroups, error) {
	midManager := wsCore.MsgIDManager()
	msgIDs := make([]int64, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		msgIDs = append(msgIDs, int64(midManager.GetMsgID(int(groupID))))
	}

	resp, err := wsCore.Send(PackProto(wsCore.Account, groupIDs, msgIDs, getType), CMD)
	if err != nil {
		return WsMsgGroups{}, err
	}
	payload, err := resp.Read()
	if err != nil {
		return WsMsgGroups{}, err
	}
	return ParseBody(payload)
}
