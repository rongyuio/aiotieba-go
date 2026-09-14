package getdislikeforums

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/get_dislike_forums/protobuf"
	commonpb "github.com/rongyuio/aiotieba-go/protobuf"
)

// CMD is the websocket command of get_dislike_forums.
const CMD = 309692

// PackProto builds the GetDislikeListReqIdl request, mirroring pack_proto.
func PackProto(account *core.Account, pn, rn int32) []byte {
	req := &pb.GetDislikeListReqIdl{
		Data: &pb.GetDislikeListReqIdl_DataReq{
			Common: &commonpb.CommonReq{
				BDUSS:          account.BDUSS(),
				XClientVersion: consts.LatestVersion,
			},
			Pn: pn,
			Rn: rn,
		},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil
	}
	return out
}

// ParseBody decodes a GetDislikeListResIdl response, mirroring parse_body.
func ParseBody(body []byte) (DislikeForums, error) {
	res := &pb.GetDislikeListResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return DislikeForums{}, fmt.Errorf("decoding GetDislikeListResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return DislikeForums{}, &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}
	return DislikeForumsFromProto(res.GetData()), nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "https",
		Host:     consts.AppBaseHost,
		Path:     "/c/u/user/getDislikeList",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP performs the app HTTP request, mirroring request_http.
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, pn, rn int32) (DislikeForums, error) {
	req, err := httpCore.PackProtoRequest(ctx, RequestURL(), PackProto(httpCore.Account, pn, rn))
	if err != nil {
		return DislikeForums{}, err
	}
	body, err := httpCore.SendProto(req)
	if err != nil {
		return DislikeForums{}, err
	}
	return ParseBody(body)
}

// RequestWS performs the websocket request, mirroring request_ws.
func RequestWS(wsCore *core.WsCore, pn, rn int32) (DislikeForums, error) {
	resp, err := wsCore.Send(PackProto(wsCore.Account, pn, rn), CMD)
	if err != nil {
		return DislikeForums{}, err
	}
	payload, err := resp.Read()
	if err != nil {
		return DislikeForums{}, err
	}
	return ParseBody(payload)
}
