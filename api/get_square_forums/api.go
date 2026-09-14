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

// CMD is the websocket command of get_square_forums.
const CMD = 309653

// PackProto builds the GetForumSquareReqIdl request, mirroring pack_proto.
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

// ParseBody decodes a GetForumSquareResIdl response, mirroring parse_body.
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

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "https",
		Host:     consts.AppBaseHost,
		Path:     "/c/f/forum/getForumSquare",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP performs the app HTTP request, mirroring request_http.
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, cname string, pn, rn int32) (SquareForums, error) {
	resp, err := httpCore.AppProto(PackProto(httpCore.Account, cname, pn, rn)).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return SquareForums{}, err
	}
	return ParseBody(resp.Body())
}

// RequestWS performs the websocket request, mirroring request_ws.
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
