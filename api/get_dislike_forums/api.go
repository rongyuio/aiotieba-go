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

// CMD 是 get_dislike_forums 的 websocket 命令字。
const CMD = 309692

// PackProto 构造 GetDislikeListReqIdl 请求，对应 pack_proto。
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

// ParseBody 解析 GetDislikeListResIdl 响应，对应 parse_body。
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

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "https",
		Host:     consts.AppBaseHost,
		Path:     "/c/u/user/getDislikeList",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP 执行 app HTTP 请求，对应 request_http。
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, pn, rn int32) (DislikeForums, error) {
	resp, err := httpCore.AppProto(PackProto(httpCore.Account, pn, rn)).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return DislikeForums{}, err
	}
	return ParseBody(resp.Body())
}

// RequestWS 执行 websocket 请求，对应 request_ws。
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
