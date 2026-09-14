package getbawuinfo

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba/consts"
	"github.com/rongyuio/aiotieba/core"
	"github.com/rongyuio/aiotieba/exception"

	pb "github.com/rongyuio/aiotieba/api/get_bawu_info/protobuf"
	commonpb "github.com/rongyuio/aiotieba/protobuf"
)

// CMD is the websocket command of get_bawu_info.
const CMD = 301007

// PackProto builds the GetBawuInfoReqIdl request, mirroring pack_proto.
func PackProto(fid int64) []byte {
	req := &pb.GetBawuInfoReqIdl{
		Data: &pb.GetBawuInfoReqIdl_DataReq{
			Common: &commonpb.CommonReq{XClientVersion: consts.LatestVersion},
			Fid:    uint64(fid),
		},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil
	}
	return out
}

// ParseBody decodes a GetBawuInfoResIdl response, mirroring parse_body.
func ParseBody(body []byte) (BawuInfo, error) {
	res := &pb.GetBawuInfoResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return BawuInfo{}, fmt.Errorf("decoding GetBawuInfoResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return BawuInfo{}, &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}
	return BawuInfoFromProto(res.GetData()), nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "http",
		Host:     consts.AppBaseHost,
		Path:     "/c/f/forum/getBawuInfo",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP performs the app HTTP request, mirroring request_http.
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, fid int64) (BawuInfo, error) {
	req, err := httpCore.PackProtoRequest(ctx, RequestURL(), PackProto(fid))
	if err != nil {
		return BawuInfo{}, err
	}
	body, err := httpCore.SendProto(req)
	if err != nil {
		return BawuInfo{}, err
	}
	return ParseBody(body)
}

// RequestWS performs the websocket request, mirroring request_ws.
func RequestWS(wsCore *core.WsCore, fid int64) (BawuInfo, error) {
	resp, err := wsCore.Send(PackProto(fid), CMD)
	if err != nil {
		return BawuInfo{}, err
	}
	payload, err := resp.Read()
	if err != nil {
		return BawuInfo{}, err
	}
	return ParseBody(payload)
}
