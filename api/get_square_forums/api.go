package getsquareforums

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/get_square_forums/protobuf"
	commonpb "github.com/rongyuio/aiotieba-go/protobuf"
)

// CMD 是 get_square_forums 的 websocket 命令字。
const CMD = 309653

// PackProto 构造 GetForumSquareReqIdl 请求，对应 pack_proto。
func PackProto(account *core.Account, cname string, pn, rn int32) []byte {
	req := &pb.GetForumSquareReqIdl{
		Data: &pb.GetForumSquareReqIdl_DataReq{
			Common: &commonpb.CommonReq{
				BDUSS:          account.BDUSS(),
				XClientVersion: consts.LatestVersion,
			},
			ClassName: cname,
			Pn:        pn,
			Rn:        rn,
		},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil
	}
	return out
}

// ParseBody 解析 GetForumSquareResIdl 响应，对应 parse_body。
func ParseBody(body []byte) (SquareForums, error) {
	res := &pb.GetForumSquareResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return SquareForums{}, fmt.Errorf("decoding GetForumSquareResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return SquareForums{}, &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}
	return SquareForumsFromProto(res.GetData()), nil
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "https",
		Host:     consts.AppBaseHost,
		Path:     "/c/f/forum/getForumSquare",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP 执行 app HTTP 请求，对应 request_http。
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, cname string, pn, rn int32) (SquareForums, error) {
	resp, err := httpCore.AppProto(PackProto(httpCore.Account, cname, pn, rn)).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return SquareForums{}, err
	}
	return ParseBody(resp.Body())
}

// RequestWS 执行 websocket 请求，对应 request_ws。
func RequestWS(wsCore *core.WsCore, cname string, pn, rn int32) (SquareForums, error) {
	resp, err := wsCore.Send(PackProto(wsCore.Account, cname, pn, rn), CMD)
	if err != nil {
		return SquareForums{}, err
	}
	payload, err := resp.Read()
	if err != nil {
		return SquareForums{}, err
	}
	return ParseBody(payload)
}
