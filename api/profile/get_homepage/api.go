// Package gethomepage 实现 aiotieba 的 profile.get_homepage API。
//
// 对应 Python 包 aiotieba.api.profile.get_homepage。
package gethomepage

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

// clientType 是 Python 模块发送的客户端类型。
const clientType = 2

// PackProto 构造 ProfileReqIdl 请求，对应 pack_proto。
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

// ParseBody 解析 ProfileResIdl 响应，对应 parse_body。
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

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "http",
		Host:     consts.AppBaseHost,
		Path:     "/c/u/user/profile",
		RawQuery: "cmd=" + strconv.Itoa(profile.CMD),
	}
}

// RequestHTTP 执行 app HTTP 请求，对应 request_http。
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, userID int64, pn int32) (profile.Homepage, error) {
	resp, err := httpCore.AppProto(PackProto(userID, pn)).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return profile.Homepage{}, err
	}
	return ParseBody(resp.Body())
}

// RequestWS 执行 websocket 请求，对应 request_ws。
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
