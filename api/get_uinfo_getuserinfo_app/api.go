package getuserinfoapp

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/get_uinfo_getuserinfo_app/protobuf"
)

// CMD 是 get_uinfo_getuserinfo_app 的 websocket 命令字。
const CMD = 303024

// PackProto 构造 GetUserInfoReqIdl 请求，对应 pack_proto。
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

// ParseBody 解析 GetUserInfoResIdl 响应，对应 parse_body。
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

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "http",
		Host:     consts.AppBaseHost,
		Path:     "/c/u/user/getuserinfo",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP 执行 app HTTP 请求，对应 request_http。
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, userID int64) (UserInfoGuinfoApp, error) {
	resp, err := httpCore.AppProto(PackProto(userID)).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return UserInfoGuinfoApp{}, err
	}
	return ParseBody(resp.Body())
}

// RequestWS 执行 websocket 请求，对应 request_ws。
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
