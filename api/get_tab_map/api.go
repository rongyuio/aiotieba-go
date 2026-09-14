package gettabmap

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/get_tab_map/protobuf"
	commonpb "github.com/rongyuio/aiotieba-go/protobuf"
)

// CMD is the websocket command of get_tab_map.
const CMD = 309466

// PackProto builds the SearchPostForumReqIdl request, mirroring pack_proto.
func PackProto(account *core.Account, fname string) []byte {
	req := &pb.SearchPostForumReqIdl{
		Data: &pb.SearchPostForumReqIdl_DataReq{
			Common: &commonpb.CommonReq{
				BDUSS:          account.BDUSS(),
				XClientVersion: consts.LatestVersion,
			},
			Fname: fname,
		},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil
	}
	return out
}

// ParseBody decodes a SearchPostForumResIdl response, mirroring parse_body.
func ParseBody(body []byte) (TabMap, error) {
	res := &pb.SearchPostForumResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return TabMap{}, fmt.Errorf("decoding SearchPostForumResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return TabMap{}, &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}
	return TabMapFromProto(res.GetData()), nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "https",
		Host:     consts.AppBaseHost,
		Path:     "/c/f/forum/searchPostForum",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP performs the app HTTP request, mirroring request_http.
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, fname string) (TabMap, error) {
	req, err := httpCore.PackProtoRequest(ctx, RequestURL(), PackProto(httpCore.Account, fname))
	if err != nil {
		return TabMap{}, err
	}
	body, err := httpCore.SendProto(req)
	if err != nil {
		return TabMap{}, err
	}
	return ParseBody(body)
}

// RequestWS performs the websocket request, mirroring request_ws.
func RequestWS(wsCore *core.WsCore, fname string) (TabMap, error) {
	resp, err := wsCore.Send(PackProto(wsCore.Account, fname), CMD)
	if err != nil {
		return TabMap{}, err
	}
	payload, err := resp.Read()
	if err != nil {
		return TabMap{}, err
	}
	return ParseBody(payload)
}
