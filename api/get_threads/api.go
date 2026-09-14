package getthreads

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/get_threads/protobuf"
	commonpb "github.com/rongyuio/aiotieba-go/protobuf"
)

// CMD 是 get_threads 的 websocket 命令字。
const CMD = 301001

// PackProto 构造 FrsPageReqIdl 请求，对应 pack_proto。
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

// ParseBody 解析 FrsPageResIdl 响应，对应 parse_body。
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
func RequestHTTP(
	ctx context.Context, httpCore *core.HttpCore, fname string, pn, rn, sort int32, isGood bool, version string,
) (Threads, error) {
	resp, err := httpCore.AppProto(PackProto(fname, pn, rn, sort, isGood, version)).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return Threads{}, err
	}
	return ParseBody(resp.Body())
}

// RequestWS 执行 websocket 请求，对应 request_ws。
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
