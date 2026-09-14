package getuserinfoapp

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba/consts"
	"github.com/rongyuio/aiotieba/core"
	"github.com/rongyuio/aiotieba/exception"

	pb "github.com/rongyuio/aiotieba/api/get_uinfo_getuserinfo_app/protobuf"
)

// CMD is the websocket command of get_uinfo_getuserinfo_app.
const CMD = 303024

// PackProto builds the GetUserInfoReqIdl request, mirroring pack_proto.
func PackProto(userID int64) []byte {
	req := &pb.GetUserInfoReqIdl{
		Data: &pb.GetUserInfoReqIdl_DataReq{UserId: userID},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil
	}
	return out
}

// ParseBody decodes a GetUserInfoResIdl response, mirroring parse_body.
func ParseBody(body []byte) (UserInfoGuinfoApp, error) {
	res := &pb.GetUserInfoResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return UserInfoGuinfoApp{}, fmt.Errorf("decoding GetUserInfoResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return UserInfoGuinfoApp{}, &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}
	return UserInfoGuinfoAppFromProto(res.GetData().GetUser()), nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "http",
		Host:     consts.AppBaseHost,
		Path:     "/c/u/user/getuserinfo",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP performs the app HTTP request, mirroring request_http.
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, userID int64) (UserInfoGuinfoApp, error) {
	req, err := httpCore.PackProtoRequest(ctx, RequestURL(), PackProto(userID))
	if err != nil {
		return UserInfoGuinfoApp{}, err
	}
	body, err := httpCore.SendProto(req)
	if err != nil {
		return UserInfoGuinfoApp{}, err
	}
	return ParseBody(body)
}

// RequestWS performs the websocket request, mirroring request_ws.
func RequestWS(wsCore *core.WsCore, userID int64) (UserInfoGuinfoApp, error) {
	resp, err := wsCore.Send(PackProto(userID), CMD)
	if err != nil {
		return UserInfoGuinfoApp{}, err
	}
	payload, err := resp.Read()
	if err != nil {
		return UserInfoGuinfoApp{}, err
	}
	return ParseBody(payload)
}
