// Package setblacklist implements the set_blacklist API of aiotieba.
//
// It mirrors the Python package aiotieba.api.set_blacklist.
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

// CMD is the websocket command of set_blacklist.
const CMD = 309697

// permission values reported to the server.
const (
	permissionDeny   = 1
	permissionAllow  = 2
	clientTypeAmount = 2
)

// PackProto builds the SetUserBlackReqIdl request, mirroring pack_proto.
//
// Each permission is reported as "denied" when the corresponding bit is set in
// btype and as "allowed" otherwise.
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

// ParseBody decodes a SetUserBlackResIdl response, mirroring parse_body.
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

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "https",
		Host:     consts.AppBaseHost,
		Path:     "/c/c/user/setUserBlack",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP performs the app HTTP request, mirroring request_http.
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, userID int64, btype enums.BlacklistType) error {
	resp, err := httpCore.AppProto(PackProto(httpCore.Account, userID, btype)).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return err
	}
	return ParseBody(resp.Body())
}

// RequestWS performs the websocket request, mirroring request_ws.
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
