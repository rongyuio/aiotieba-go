// Package getusercontentsposts implements the get_posts sub-API of
// get_user_contents.
//
// It mirrors aiotieba.api.get_user_contents.get_posts.
package getusercontentsposts

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

// CMD is the websocket command of get_posts.
const CMD = 303002

// PackProto builds the UserPostReqIdl request, mirroring pack_proto.
func PackProto(account *core.Account, userID int64, pn, rn int32, version string) []byte {
	common := &commonpb.CommonReq{
		BDUSS:          account.BDUSS(),
		XClientVersion: version,
	}
	req := &pb.UserPostReqIdl{
		Data: &pb.UserPostReqIdl_DataReq{
			Common:      common,
			Uid:         userID,
			NeedContent: 1,
			Pn:          uint32(pn),
			Rn:          uint32(rn),
		},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil
	}
	return out
}

// ParseBody decodes a UserPostResIdl response, mirroring parse_body.
func ParseBody(body []byte) (getusercontents.UserPostss, error) {
	res := &pb.UserPostResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return getusercontents.UserPostss{}, fmt.Errorf("decoding UserPostResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return getusercontents.UserPostss{}, &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}
	return getusercontents.UserPostssFromProto(res.GetData()), nil
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
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, userID int64, pn, rn int32, version string) (getusercontents.UserPostss, error) {
	data := PackProto(httpCore.Account, userID, pn, rn, version)
	req, err := httpCore.PackProtoRequest(ctx, RequestURL(), data)
	if err != nil {
		return getusercontents.UserPostss{}, err
	}
	body, err := httpCore.SendProto(req)
	if err != nil {
		return getusercontents.UserPostss{}, err
	}
	return ParseBody(body)
}

// RequestWS performs the websocket request, mirroring request_ws.
func RequestWS(wsCore *core.WsCore, userID int64, pn, rn int32, version string) (getusercontents.UserPostss, error) {
	data := PackProto(wsCore.Account, userID, pn, rn, version)
	resp, err := wsCore.Send(data, CMD)
	if err != nil {
		return getusercontents.UserPostss{}, err
	}
	payload, err := resp.Read()
	if err != nil {
		return getusercontents.UserPostss{}, err
	}
	return ParseBody(payload)
}
