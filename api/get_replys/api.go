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

// CMD 是 get_replys 的 websocket 命令字。
const CMD = 303007

// PackProto 构造 ReplyMeReqIdl 请求，对应 pack_proto。
func PackProto(account *core.Account, pn int32) []byte {
	req := &pb.ReplyMeReqIdl{
		Data: &pb.ReplyMeReqIdl_DataReq{
			Common: &commonpb.CommonReq{
				BDUSS:          account.BDUSS(),
				XClientVersion: consts.LatestVersion,
			},
			// pn 以字符串形式传输，对应 Python 模块。
			Pn: strconv.FormatInt(int64(pn), 10),
		},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil
	}
	return out
}

// ParseBody 解析 ReplyMeResIdl 响应，对应 parse_body。
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

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "https",
		Host:     consts.AppBaseHost,
		Path:     "/c/u/feed/replyme",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP 执行 app HTTP 请求，对应 request_http。
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, pn int32) (Replys, error) {
	resp, err := httpCore.AppProto(PackProto(httpCore.Account, pn)).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return Replys{}, err
	}
	return ParseBody(resp.Body())
}

// RequestWS 执行 websocket 请求，对应 request_ws。
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
