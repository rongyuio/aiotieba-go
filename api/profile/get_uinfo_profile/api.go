// Package getuinfoprofile implements profile.get_uinfo_profile of aiotieba.
//
// It mirrors the Python package aiotieba.api.profile.get_uinfo_profile.
package getuinfoprofile

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/api/profile"
	pb "github.com/rongyuio/aiotieba-go/api/profile/protobuf"
	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	commonpb "github.com/rongyuio/aiotieba-go/protobuf"
)

// PackProto builds the ProfileReqIdl request, mirroring pack_proto.
func PackProto(ref profile.Ref) []byte {
	data := &pb.ProfileReqIdl_DataReq{
		Common: &commonpb.CommonReq{
			XClientVersion: consts.LatestVersion,
			XClientType:    needPostCountClientType,
		},
		NeedPostCount: 1,
		Page:          1,
	}
	if ref.Portrait != "" {
		data.FriendUidPortrait = ref.Portrait
	} else {
		data.Uid = ref.UserID
	}

	out, err := proto.Marshal(&pb.ProfileReqIdl{Data: data})
	if err != nil {
		return nil
	}
	return out
}

// needPostCountClientType is the client type the Python module sends.
const needPostCountClientType = 2

// ParseBody decodes a ProfileResIdl response, mirroring parse_body.
func ParseBody(body []byte) (profile.UserInfoPF, error) {
	res := &pb.ProfileResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return profile.UserInfoPF{}, fmt.Errorf("decoding ProfileResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return profile.UserInfoPF{}, &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}
	return profile.UserInfoPFFromProto(res.GetData()), nil
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
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, ref profile.Ref) (profile.UserInfoPF, error) {
	req, err := httpCore.PackProtoRequest(ctx, RequestURL(), PackProto(ref))
	if err != nil {
		return profile.UserInfoPF{}, err
	}
	body, err := httpCore.SendProto(req)
	if err != nil {
		return profile.UserInfoPF{}, err
	}
	return ParseBody(body)
}

// RequestWS performs the websocket request, mirroring request_ws.
func RequestWS(wsCore *core.WsCore, ref profile.Ref) (profile.UserInfoPF, error) {
	resp, err := wsCore.Send(PackProto(ref), profile.CMD)
	if err != nil {
		return profile.UserInfoPF{}, err
	}
	payload, err := resp.Read()
	if err != nil {
		return profile.UserInfoPF{}, err
	}
	return ParseBody(payload)
}
