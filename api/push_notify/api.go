package pushnotify

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	pb "github.com/rongyuio/aiotieba/api/push_notify/protobuf"
)

// CMD is the websocket command of push_notify.
const CMD = 202006

// ParseBody decodes a PushNotifyResIdl frame, mirroring parse_body.
//
// Unlike the other websocket APIs there is no request counterpart: the frame is
// pushed by the server.
func ParseBody(body []byte) ([]WsNotify, error) {
	res := &pb.PushNotifyResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return nil, fmt.Errorf("decoding PushNotifyResIdl: %w", err)
	}
	return WsNotifiesFromProto(res), nil
}
