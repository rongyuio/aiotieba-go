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

// CMD 是 get_blacklist_old 的 websocket 命令字。
const CMD = 303028

// PackProto 构造 UserMuteQueryReqIdl 请求，对应 pack_proto。
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

// ParseBody 解析 UserMuteQueryResIdl 响应，对应 parse_body。
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

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "https",
		Host:     consts.AppBaseHost,
		Path:     "/c/u/user/userMuteQuery",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP 执行 app HTTP 请求，对应 request_http。
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, pn, rn int32) (BlacklistOldUsers, error) {
	resp, err := httpCore.AppProto(PackProto(httpCore.Account, pn, rn)).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return BlacklistOldUsers{}, err
	}
	return ParseBody(resp.Body())
}

// RequestWS 执行 websocket 请求，对应 request_ws。
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
