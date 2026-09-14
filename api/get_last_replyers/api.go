package getlastreplyers

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
	"github.com/rongyuio/aiotieba-go/helper"

	pb "github.com/rongyuio/aiotieba-go/api/get_last_replyers/protobuf"
	commonpb "github.com/rongyuio/aiotieba-go/protobuf"
)

// CMD is the websocket command of get_last_replyers.
const CMD = 301001

// clientVersion is the legacy client version this endpoint expects. It differs
// from consts.LatestVersion, mirroring the Python module.
const clientVersion = "6.0.1"

// rnNeedMargin is added to rn when asking the server for extra entries.
const rnNeedMargin = 5

// PackProto builds the FrsPageReqIdl4lp request, mirroring pack_proto.
//
// pn 1 is sent as 0, which is how the endpoint spells the first page.
func PackProto(fname string, pn, rn int32, sort enums.ThreadSortType, isGood bool) []byte {
	if pn == 1 {
		pn = 0
	}

	req := &pb.FrsPageReqIdl4Lp{
		Data: &pb.FrsPageReqIdl4Lp_DataReq{
			Common: &commonpb.CommonReq{
				XClientType:    2,
				XClientVersion: clientVersion,
			},
			Kw:       fname,
			Pn:       pn,
			Rn:       rn,
			RnNeed:   rn + rnNeedMargin,
			IsGood:   int32(helper.BoolInt(isGood)),
			SortType: int32(sort),
		},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil
	}
	return out
}

// ParseBody decodes a FrsPageResIdl4lp response, mirroring parse_body.
func ParseBody(body []byte) (ThreadsLP, error) {
	res := &pb.FrsPageResIdl4Lp{}
	if err := proto.Unmarshal(body, res); err != nil {
		return ThreadsLP{}, fmt.Errorf("decoding FrsPageResIdl4lp: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return ThreadsLP{}, &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}
	return ThreadsLPFromProto(res.GetData()), nil
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
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, fname string, pn, rn int32, sort enums.ThreadSortType, isGood bool) (ThreadsLP, error) {
	resp, err := httpCore.AppProto(PackProto(fname, pn, rn, sort, isGood)).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return ThreadsLP{}, err
	}
	return ParseBody(resp.Body())
}

// RequestWS performs the websocket request, mirroring request_ws.
func RequestWS(wsCore *core.WsCore, fname string, pn, rn int32, sort enums.ThreadSortType, isGood bool) (ThreadsLP, error) {
	resp, err := wsCore.Send(PackProto(fname, pn, rn, sort, isGood), CMD)
	if err != nil {
		return ThreadsLP{}, err
	}
	payload, err := resp.Read()
	if err != nil {
		return ThreadsLP{}, err
	}
	return ParseBody(payload)
}
