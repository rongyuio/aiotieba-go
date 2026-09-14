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

// CMD 是 get_last_replyers 的 websocket 命令字。
const CMD = 301001

// clientVersion 是该接口所需的旧版客户端版本，与 consts.LatestVersion 不同，
// 对应 Python 模块。
const clientVersion = "6.0.1"

// rnNeedMargin 是在向服务端请求额外条目时加到 rn 上的余量。
const rnNeedMargin = 5

// PackProto 构造 FrsPageReqIdl4lp 请求，对应 pack_proto。
//
// pn 为 1 时按 0 发送，这是该接口表示第一页的方式。
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

// ParseBody 解析 FrsPageResIdl4lp 响应，对应 parse_body。
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

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "http",
		Host:     consts.AppBaseHost,
		Path:     "/c/f/frs/page",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP 执行 app HTTP 请求，对应 request_http。
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, fname string, pn, rn int32, sort enums.ThreadSortType, isGood bool) (ThreadsLP, error) {
	resp, err := httpCore.AppProto(PackProto(fname, pn, rn, sort, isGood)).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return ThreadsLP{}, err
	}
	return ParseBody(resp.Body())
}

// RequestWS 执行 websocket 请求，对应 request_ws。
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
