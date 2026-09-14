package getthreads

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba/consts"
	"github.com/rongyuio/aiotieba/core"
	"github.com/rongyuio/aiotieba/exception"

	pb "github.com/rongyuio/aiotieba/api/get_threads/protobuf"
	commonpb "github.com/rongyuio/aiotieba/protobuf"
)

// CMD is the websocket command of get_threads.
const CMD = 301001

// PackProto builds the FrsPageReqIdl request, mirroring pack_proto.
func PackProto(fname string, pn, rn, sort int32, isGood bool, version string) []byte {
	protoPn := pn
	if pn == 1 {
		protoPn = 0
	}
	var isGoodInt int32
	if isGood {
		isGoodInt = 1
	}

	req := &pb.FrsPageReqIdl{
		Data: &pb.FrsPageReqIdl_DataReq{
			Common:   &commonpb.CommonReq{XClientType: 2, XClientVersion: version},
			Kw:       fname,
			Pn:       protoPn,
			Rn:       rn,
			RnNeed:   rn + 5,
			IsGood:   isGoodInt,
			SortType: sort,
		},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil
	}
	return out
}

// ParseBody decodes a FrsPageResIdl response, mirroring parse_body.
func ParseBody(body []byte) (Threads, error) {
	res := &pb.FrsPageResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return Threads{}, fmt.Errorf("decoding FrsPageResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return Threads{}, &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}
	return ThreadsFromProto(res.GetData()), nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "http",
		Host:     consts.AppBaseHost,
		Path:     "/c/f/frs/page",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP performs the app HTTP request, mirroring request_http.
func RequestHTTP(
	ctx context.Context, httpCore *core.HttpCore, fname string, pn, rn, sort int32, isGood bool, version string,
) (Threads, error) {
	req, err := httpCore.PackProtoRequest(ctx, RequestURL(), PackProto(fname, pn, rn, sort, isGood, version))
	if err != nil {
		return Threads{}, err
	}
	body, err := httpCore.SendProto(req)
	if err != nil {
		return Threads{}, err
	}
	return ParseBody(body)
}

// RequestWS performs the websocket request, mirroring request_ws.
func RequestWS(wsCore *core.WsCore, fname string, pn, rn, sort int32, isGood bool, version string) (Threads, error) {
	resp, err := wsCore.Send(PackProto(fname, pn, rn, sort, isGood, version), CMD)
	if err != nil {
		return Threads{}, err
	}
	payload, err := resp.Read()
	if err != nil {
		return Threads{}, err
	}
	return ParseBody(payload)
}
