package getforumlevel

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba/consts"
	"github.com/rongyuio/aiotieba/core"
	"github.com/rongyuio/aiotieba/exception"

	pb "github.com/rongyuio/aiotieba/api/get_forum_level/protobuf"
	commonpb "github.com/rongyuio/aiotieba/protobuf"
)

// CMD is the websocket command of get_forum_level.
const CMD = 301005

// PackProto builds the GetLevelInfoReqIdl request, mirroring pack_proto.
func PackProto(account *core.Account, fid int64) []byte {
	req := &pb.GetLevelInfoReqIdl{
		Data: &pb.GetLevelInfoReqIdl_DataReq{
			Common:  &commonpb.CommonReq{BDUSS: account.BDUSS()},
			ForumId: fid,
		},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil
	}
	return out
}

// ParseBody decodes a GetLevelInfoResIdl response, mirroring parse_body.
func ParseBody(body []byte) (LevelInfo, error) {
	res := &pb.GetLevelInfoResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return LevelInfo{}, fmt.Errorf("decoding GetLevelInfoResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return LevelInfo{}, &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}
	return LevelInfoFromProto(res.GetData()), nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "https",
		Host:     consts.AppBaseHost,
		Path:     "/c/f/forum/getLevelInfo",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP performs the app HTTP request, mirroring request_http.
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, fid int64) (LevelInfo, error) {
	req, err := httpCore.PackProtoRequest(ctx, RequestURL(), PackProto(httpCore.Account, fid))
	if err != nil {
		return LevelInfo{}, err
	}
	body, err := httpCore.SendProto(req)
	if err != nil {
		return LevelInfo{}, err
	}
	return ParseBody(body)
}

// RequestWS performs the websocket request, mirroring request_ws.
func RequestWS(wsCore *core.WsCore, fid int64) (LevelInfo, error) {
	resp, err := wsCore.Send(PackProto(wsCore.Account, fid), CMD)
	if err != nil {
		return LevelInfo{}, err
	}
	payload, err := resp.Read()
	if err != nil {
		return LevelInfo{}, err
	}
	return ParseBody(payload)
}
