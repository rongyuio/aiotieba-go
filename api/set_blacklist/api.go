// Package setblacklist 实现 aiotieba 的 set_blacklist API。
//
// 对应 Python 包 aiotieba.api.set_blacklist。
package setblacklist

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/set_blacklist/protobuf"
	commonpb "github.com/rongyuio/aiotieba-go/protobuf"
)

// CMD 是 set_blacklist 的 websocket 命令字。
const CMD = 309697

// 上报给服务端的权限取值。
const (
	permissionDeny   = 1
	permissionAllow  = 2
	clientTypeAmount = 2
)

// PackProto 构造 SetUserBlackReqIdl 请求，对应 pack_proto。
//
// 当 btype 中对应的位被置位时，该权限上报为 "denied"，否则上报为 "allowed"。
func PackProto(account *core.Account, userID int64, btype enums.BlacklistType) []byte {
	req := &pb.SetUserBlackReqIdl{
		Data: &pb.SetUserBlackReqIdl_DataReq{
			Common: &commonpb.CommonReq{
				BDUSS:          account.BDUSS(),
				XClientType:    clientTypeAmount,
				XClientVersion: consts.LatestVersion,
			},
			BlackUid: userID,
			PermList: &pb.SetUserBlackReqIdl_DataReq_PermissionList{
				Follow:   permission(btype, enums.BlacklistFollow),
				Interact: permission(btype, enums.BlacklistInteract),
				Chat:     permission(btype, enums.BlacklistChat),
			},
		},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil
	}
	return out
}

func permission(btype, flag enums.BlacklistType) int32 {
	if btype&flag != 0 {
		return permissionDeny
	}
	return permissionAllow
}

// ParseBody 解析 SetUserBlackResIdl 响应，对应 parse_body。
func ParseBody(body []byte) error {
	res := &pb.SetUserBlackResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return fmt.Errorf("decoding SetUserBlackResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}
	return nil
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "https",
		Host:     consts.AppBaseHost,
		Path:     "/c/c/user/setUserBlack",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP 执行 app HTTP 请求，对应 request_http。
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, userID int64, btype enums.BlacklistType) error {
	resp, err := httpCore.AppProto(PackProto(httpCore.Account, userID, btype)).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return err
	}
	return ParseBody(resp.Body())
}

// RequestWS 执行 websocket 请求，对应 request_ws。
func RequestWS(wsCore *core.WsCore, userID int64, btype enums.BlacklistType) error {
	resp, err := wsCore.Send(PackProto(wsCore.Account, userID, btype), CMD)
	if err != nil {
		return err
	}
	payload, err := resp.Read()
	if err != nil {
		return err
	}
	return ParseBody(payload)
}
