package pushnotify

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	pb "github.com/rongyuio/aiotieba-go/api/push_notify/protobuf"
)

// CMD 是 push_notify 的 websocket 命令字。
const CMD = 202006

// ParseBody 解析 PushNotifyResIdl 帧，对应 parse_body。
//
// 与其他 websocket API 不同，该帧没有请求方：由服务器推送。
func ParseBody(body []byte) ([]WsNotify, error) {
	res := &pb.PushNotifyResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return nil, fmt.Errorf("decoding PushNotifyResIdl: %w", err)
	}
	return WsNotifiesFromProto(res), nil
}
