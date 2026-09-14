// Package getuinfoprofile 实现 aiotieba 的 profile.get_uinfo_profile API。
//
// 对应 Python 包 aiotieba.api.profile.get_uinfo_profile。
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

// PackProto 构造 ProfileReqIdl 请求，对应 pack_proto。
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

// needPostCountClientType 是 Python 模块发送的客户端类型。
const needPostCountClientType = 2

// ParseBody 解析 ProfileResIdl 响应，对应 parse_body。
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
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, ref profile.Ref) (profile.UserInfoPF, error) {
	resp, err := httpCore.AppProto(PackProto(ref)).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return profile.UserInfoPF{}, err
	}
	return ParseBody(resp.Body())
}

// RequestWS 执行 websocket 请求，对应 request_ws。
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
