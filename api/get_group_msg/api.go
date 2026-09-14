package getgroupmsg

import (
	"fmt"
	"strconv"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/get_group_msg/protobuf"
)

// CMD 是 get_group_msg 的 websocket 命令字。
const CMD = 202003

// cuidSuffix 会追加到账户 cuid 之后，对应 Python 模块。
const cuidSuffix = "|com.baidu.tieba_mini12.35.1.0"

// PackProto 构造 GetGroupMsgReqIdl 请求，对应 pack_proto。
//
// group id 与 message id 逐一配对：任一数组多出的条目会被忽略，如同 Python 的 zip(strict=False)。
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

// ParseBody 解析 GetGroupMsgResIdl 响应，对应 parse_body。
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

// Request 执行 websocket 请求，对应 request。
//
// 每个分组最后已知的消息 id 取自 websocket core 的消息 id 管理器。
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
