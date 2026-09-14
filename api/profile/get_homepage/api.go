// Package gethomepage implements profile.get_homepage of aiotieba.
//
// It mirrors the Python package aiotieba.api.profile.get_homepage.
package gethomepage

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba/api/profile"
	pb "github.com/rongyuio/aiotieba/api/profile/protobuf"
	"github.com/rongyuio/aiotieba/consts"
	"github.com/rongyuio/aiotieba/core"
	"github.com/rongyuio/aiotieba/exception"
	commonpb "github.com/rongyuio/aiotieba/protobuf"
)

// clientType is the client type the Python module sends.
const clientType = 2

// PackProto builds the ProfileReqIdl request, mirroring pack_proto.
func PackProto(userID int64, pn int32) []byte {
	req := &pb.ProfileReqIdl{
		Data: &pb.ProfileReqIdl_DataReq{
			Common: &commonpb.CommonReq{
				XClientVersion: consts.LatestVersion,
				XClientType:    clientType,
			},
			Uid:           userID,
			NeedPostCount: 1,
			Pn:            uint32(pn),
		},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil
	}
	return out
}

// ParseBody decodes a ProfileResIdl response, mirroring parse_body.
func ParseBody(body []byte) (profile.Homepage, error) {
	res := &pb.ProfileResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return profile.Homepage{}, fmt.Errorf("decoding ProfileResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return profile.Homepage{}, &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}
	return profile.HomepageFromProto(res.GetData()), nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "http",
		Host:     consts.AppBaseHost,
		Path:     "/c/u/user/profile",
		RawQuery: "cmd=" + strconv.Itoa(profile.CMD),
	}
}

// RequestHTTP performs the app HTTP request, mirroring request_http.
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, userID int64, pn int32) (profile.Homepage, error) {
	req, err := httpCore.PackProtoRequest(ctx, RequestURL(), PackProto(userID, pn))
	if err != nil {
		return profile.Homepage{}, err
	}
	body, err := httpCore.SendProto(req)
	if err != nil {
		return profile.Homepage{}, err
	}
	return ParseBody(body)
}

// RequestWS performs the websocket request, mirroring request_ws.
func RequestWS(wsCore *core.WsCore, userID int64, pn int32) (profile.Homepage, error) {
	resp, err := wsCore.Send(PackProto(userID, pn), profile.CMD)
	if err != nil {
		return profile.Homepage{}, err
	}
	payload, err := resp.Read()
	if err != nil {
		return profile.Homepage{}, err
	}
	return ParseBody(payload)
}
