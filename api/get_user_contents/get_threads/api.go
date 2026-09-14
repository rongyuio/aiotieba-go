// Package getusercontentsthreads implements the get_threads sub-API of
// get_user_contents.
//
// It mirrors aiotieba.api.get_user_contents.get_threads.
package getusercontentsthreads

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/api/get_user_contents"
	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/get_user_contents/protobuf"
	commonpb "github.com/rongyuio/aiotieba-go/protobuf"
)

// CMD is the websocket command of get_threads.
const CMD = 303002

// PackProto builds the UserPostReqIdl request, mirroring pack_proto.
func PackProto(userID int64, pn int32, publicOnly bool) []byte {
	isViewCard := int32(1)
	if publicOnly {
		isViewCard = 2
	}
	req := &pb.UserPostReqIdl{
		Data: &pb.UserPostReqIdl_DataReq{
			Common:      &commonpb.CommonReq{XClientVersion: consts.LatestVersion},
			Uid:         userID,
			IsThread:    1,
			NeedContent: 1,
			Pn:          uint32(pn),
			IsViewCard:  isViewCard,
		},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil
	}
	return out
}

// ParseBody decodes a UserPostResIdl response, mirroring parse_body.
func ParseBody(body []byte) (getusercontents.UserThreads, error) {
	res := &pb.UserPostResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return getusercontents.UserThreads{}, fmt.Errorf("decoding UserPostResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return getusercontents.UserThreads{}, &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}
	return getusercontents.UserThreadsFromProto(res.GetData()), nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "https",
		Host:     consts.AppBaseHost,
		Path:     "/c/u/feed/userpost",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP performs the app HTTP request, mirroring request_http.
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, userID int64, pn int32, publicOnly bool) (getusercontents.UserThreads, error) {
	data := PackProto(userID, pn, publicOnly)
	resp, err := httpCore.AppProto(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return getusercontents.UserThreads{}, err
	}
	return ParseBody(resp.Body())
}

// RequestWS performs the websocket request, mirroring request_ws.
func RequestWS(wsCore *core.WsCore, userID int64, pn int32, publicOnly bool) (getusercontents.UserThreads, error) {
	data := PackProto(userID, pn, publicOnly)
	resp, err := wsCore.Send(data, CMD)
	if err != nil {
		return getusercontents.UserThreads{}, err
	}
	payload, err := resp.Read()
	if err != nil {
		return getusercontents.UserThreads{}, err
	}
	return ParseBody(payload)
}
