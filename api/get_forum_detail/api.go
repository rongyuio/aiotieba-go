package getforumdetail

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba/consts"
	"github.com/rongyuio/aiotieba/core"
	"github.com/rongyuio/aiotieba/exception"

	pb "github.com/rongyuio/aiotieba/api/get_forum_detail/protobuf"
	commonpb "github.com/rongyuio/aiotieba/protobuf"
)

// CMD is the websocket command of get_forum_detail.
const CMD = 303021

// PackProto builds the GetForumDetailReqIdl request, mirroring pack_proto.
func PackProto(fid int64) []byte {
	req := &pb.GetForumDetailReqIdl{
		Data: &pb.GetForumDetailReqIdl_DataReq{
			ForumId: fid,
			Common:  &commonpb.CommonReq{XClientVersion: consts.LatestVersion},
		},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil
	}
	return out
}

// ParseBody decodes a GetForumDetailResIdl response, mirroring parse_body.
func ParseBody(body []byte) (ForumDetail, error) {
	res := &pb.GetForumDetailResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return ForumDetail{}, fmt.Errorf("decoding GetForumDetailResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return ForumDetail{}, &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}
	return ForumDetailFromProto(res.GetData()), nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "http",
		Host:     consts.AppBaseHost,
		Path:     "/c/f/forum/getforumdetail",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP performs the app HTTP request, mirroring request_http.
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, fid int64) (ForumDetail, error) {
	req, err := httpCore.PackProtoRequest(ctx, RequestURL(), PackProto(fid))
	if err != nil {
		return ForumDetail{}, err
	}
	body, err := httpCore.SendProto(req)
	if err != nil {
		return ForumDetail{}, err
	}
	return ParseBody(body)
}

// RequestWS performs the websocket request, mirroring request_ws.
func RequestWS(wsCore *core.WsCore, fid int64) (ForumDetail, error) {
	resp, err := wsCore.Send(PackProto(fid), CMD)
	if err != nil {
		return ForumDetail{}, err
	}
	payload, err := resp.Read()
	if err != nil {
		return ForumDetail{}, err
	}
	return ParseBody(payload)
}
