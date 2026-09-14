// Package getusercontentsposts 实现 get_user_contents 的 get_posts 子 API。
//
// 对应 Python 包 aiotieba.api.get_user_contents.get_posts。
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

// CMD 是 get_posts 的 websocket 命令字。
const CMD = 303002

// PackProto 构造 UserPostReqIdl 请求，对应 pack_proto。
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

// ParseBody 解析 UserPostResIdl 响应，对应 parse_body。
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

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "https",
		Host:     consts.AppBaseHost,
		Path:     "/c/u/feed/userpost",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP 执行 app HTTP 请求，对应 request_http。
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, userID int64, pn, rn int32, version string) (getusercontents.UserPostss, error) {
	data := PackProto(httpCore.Account, userID, pn, rn, version)
	resp, err := httpCore.AppProto(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return getusercontents.UserPostss{}, err
	}
	return ParseBody(resp.Body())
}

// RequestWS 执行 websocket 请求，对应 request_ws。
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
