package tiebauid2userinfo

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/tieba_uid2user_info/protobuf"
	commonpb "github.com/rongyuio/aiotieba-go/protobuf"
)

// CMD is the websocket command of tieba_uid2user_info.
const CMD = 309702

// PackProto builds the GetUserByTiebaUidReqIdl request, mirroring pack_proto.
func PackProto(tiebaUID int64) []byte {
	req := &pb.GetUserByTiebaUidReqIdl{
		Data: &pb.GetUserByTiebaUidReqIdl_DataReq{
			Common:   &commonpb.CommonReq{XClientVersion: consts.LatestVersion},
			TiebaUid: strconv.FormatInt(tiebaUID, 10),
		},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil
	}
	return out
}

// ParseBody decodes a GetUserByTiebaUidResIdl response, mirroring parse_body.
func ParseBody(body []byte) (UserInfoTUid, error) {
	res := &pb.GetUserByTiebaUidResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return UserInfoTUid{}, fmt.Errorf("decoding GetUserByTiebaUidResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return UserInfoTUid{}, &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}
	return UserInfoTUidFromProto(res.GetData().GetUser()), nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "http",
		Host:     consts.AppBaseHost,
		Path:     "/c/u/user/getUserByTiebaUid",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP performs the app HTTP request, mirroring request_http.
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, tiebaUID int64) (UserInfoTUid, error) {
	req, err := httpCore.PackProtoRequest(ctx, RequestURL(), PackProto(tiebaUID))
	if err != nil {
		return UserInfoTUid{}, err
	}
	body, err := httpCore.SendProto(req)
	if err != nil {
		return UserInfoTUid{}, err
	}
	return ParseBody(body)
}

// RequestWS performs the websocket request, mirroring request_ws.
func RequestWS(wsCore *core.WsCore, tiebaUID int64) (UserInfoTUid, error) {
	resp, err := wsCore.Send(PackProto(tiebaUID), CMD)
	if err != nil {
		return UserInfoTUid{}, err
	}
	payload, err := resp.Read()
	if err != nil {
		return UserInfoTUid{}, err
	}
	return ParseBody(payload)
}
