package getblacklistold

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/get_blacklist_old/protobuf"
	commonpb "github.com/rongyuio/aiotieba-go/protobuf"
)

// CMD is the websocket command of get_blacklist_old.
const CMD = 303028

// PackProto builds the UserMuteQueryReqIdl request, mirroring pack_proto.
func PackProto(account *core.Account, pn, rn int32) []byte {
	req := &pb.UserMuteQueryReqIdl{
		Data: &pb.UserMuteQueryReqIdl_DataReq{
			Common: &commonpb.CommonReq{
				BDUSS:          account.BDUSS(),
				XClientVersion: consts.LatestVersion,
			},
			Pn: uint32(pn),
			Rn: uint32(rn),
		},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil
	}
	return out
}

// ParseBody decodes a UserMuteQueryResIdl response, mirroring parse_body.
func ParseBody(body []byte) (BlacklistOldUsers, error) {
	res := &pb.UserMuteQueryResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return BlacklistOldUsers{}, fmt.Errorf("decoding UserMuteQueryResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return BlacklistOldUsers{}, &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}
	return BlacklistOldUsersFromProto(res.GetData()), nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "https",
		Host:     consts.AppBaseHost,
		Path:     "/c/u/user/userMuteQuery",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP performs the app HTTP request, mirroring request_http.
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, pn, rn int32) (BlacklistOldUsers, error) {
	resp, err := httpCore.AppProto(PackProto(httpCore.Account, pn, rn)).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return BlacklistOldUsers{}, err
	}
	return ParseBody(resp.Body())
}

// RequestWS performs the websocket request, mirroring request_ws.
func RequestWS(wsCore *core.WsCore, pn, rn int32) (BlacklistOldUsers, error) {
	resp, err := wsCore.Send(PackProto(wsCore.Account, pn, rn), CMD)
	if err != nil {
		return BlacklistOldUsers{}, err
	}
	payload, err := resp.Read()
	if err != nil {
		return BlacklistOldUsers{}, err
	}
	return ParseBody(payload)
}
