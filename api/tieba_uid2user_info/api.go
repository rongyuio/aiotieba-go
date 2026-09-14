package tiebauid2userinfo

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/tieba_uid2user_info/protobuf"
	commonpb "github.com/rongyuio/aiotieba-go/protobuf"
)

// CMD 是 tieba_uid2user_info 的 websocket 命令字。
const CMD = 309702

// PackProto 构造 GetUserByTiebaUidReqIdl 请求，对应 pack_proto。
func PackProto(tiebaUID int64) []byte {
	req := &pb.GetUserByTiebaUidReqIdl{
		Data: &pb.GetUserByTiebaUidReqIdl_DataReq{
			Common:   &commonpb.CommonReq{XClientVersion: consts.LatestVersion},
			TiebaUid: strconv.FormatInt(tiebaUID, 10),
		},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil
	}
	return out
}

// ParseBody 解析 GetUserByTiebaUidResIdl 响应，对应 parse_body。
func ParseBody(body []byte) (UserInfoTUid, error) {
	res := &pb.GetUserByTiebaUidResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return UserInfoTUid{}, fmt.Errorf("decoding GetUserByTiebaUidResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return UserInfoTUid{}, &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}
	return UserInfoTUidFromProto(res.GetData().GetUser()), nil
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "http",
		Host:     consts.AppBaseHost,
		Path:     "/c/u/user/getUserByTiebaUid",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP 执行 app HTTP 请求，对应 request_http。
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, tiebaUID int64) (UserInfoTUid, error) {
	resp, err := httpCore.AppProto(PackProto(tiebaUID)).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return UserInfoTUid{}, err
	}
	return ParseBody(resp.Body())
}

// RequestWS 执行 websocket 请求，对应 request_ws。
func RequestWS(wsCore *core.WsCore, tiebaUID int64) (UserInfoTUid, error) {
	resp, err := wsCore.Send(PackProto(tiebaUID), CMD)
	if err != nil {
		return UserInfoTUid{}, err
	}
	payload, err := resp.Read()
	if err != nil {
		return UserInfoTUid{}, err
	}
	return ParseBody(payload)
}
