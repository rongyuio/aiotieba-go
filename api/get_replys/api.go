package getreplys

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/get_replys/protobuf"
	commonpb "github.com/rongyuio/aiotieba-go/protobuf"
)

// CMD is the websocket command of get_replys.
const CMD = 303007

// PackProto builds the ReplyMeReqIdl request, mirroring pack_proto.
func PackProto(account *core.Account, pn int32) []byte {
	req := &pb.ReplyMeReqIdl{
		Data: &pb.ReplyMeReqIdl_DataReq{
			Common: &commonpb.CommonReq{
				BDUSS:          account.BDUSS(),
				XClientVersion: consts.LatestVersion,
			},
			// pn travels as a string, mirroring the Python module.
			Pn: strconv.FormatInt(int64(pn), 10),
		},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil
	}
	return out
}

// ParseBody decodes a ReplyMeResIdl response, mirroring parse_body.
func ParseBody(body []byte) (Replys, error) {
	res := &pb.ReplyMeResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return Replys{}, fmt.Errorf("decoding ReplyMeResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return Replys{}, &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}
	return ReplysFromProto(res.GetData()), nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "https",
		Host:     consts.AppBaseHost,
		Path:     "/c/u/feed/replyme",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP performs the app HTTP request, mirroring request_http.
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, pn int32) (Replys, error) {
	req, err := httpCore.PackProtoRequest(ctx, RequestURL(), PackProto(httpCore.Account, pn))
	if err != nil {
		return Replys{}, err
	}
	body, err := httpCore.SendProto(req)
	if err != nil {
		return Replys{}, err
	}
	return ParseBody(body)
}

// RequestWS performs the websocket request, mirroring request_ws.
func RequestWS(wsCore *core.WsCore, pn int32) (Replys, error) {
	resp, err := wsCore.Send(PackProto(wsCore.Account, pn), CMD)
	if err != nil {
		return Replys{}, err
	}
	payload, err := resp.Read()
	if err != nil {
		return Replys{}, err
	}
	return ParseBody(payload)
}
