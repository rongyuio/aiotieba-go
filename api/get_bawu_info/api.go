package getbawuinfo

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/get_bawu_info/protobuf"
	commonpb "github.com/rongyuio/aiotieba-go/protobuf"
)

// CMD 是 get_bawu_info 的 websocket 命令字。
const CMD = 301007

// PackProto 构造 GetBawuInfoReqIdl 请求，对应 pack_proto。
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

// ParseBody 解析 GetBawuInfoResIdl 响应，对应 parse_body。
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

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "http",
		Host:     consts.AppBaseHost,
		Path:     "/c/f/forum/getBawuInfo",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP 执行 app HTTP 请求，对应 request_http。
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, fid int64) (BawuInfo, error) {
	resp, err := httpCore.AppProto(PackProto(fid)).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return BawuInfo{}, err
	}
	return ParseBody(resp.Body())
}

// RequestWS 执行 websocket 请求，对应 request_ws。
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
