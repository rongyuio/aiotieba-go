// Package getusercontentsthreads 实现 get_user_contents 的 get_threads 子 API。
//
// 对应 Python 包 aiotieba.api.get_user_contents.get_threads。
package getusercontentsthreads

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

// CMD 是 get_threads 的 websocket 命令字。
const CMD = 303002

// PackProto 构造 UserPostReqIdl 请求，对应 pack_proto。
func PackProto(userID int64, pn int32, publicOnly bool) []byte {
	isViewCard := int32(1)
	if publicOnly {
		isViewCard = 2
	}
	req := &pb.UserPostReqIdl{
		Data: &pb.UserPostReqIdl_DataReq{
			Common:      &commonpb.CommonReq{XClientVersion: consts.LatestVersion},
			Uid:         userID,
			IsThread:    1,
			NeedContent: 1,
			Pn:          uint32(pn),
			IsViewCard:  isViewCard,
		},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil
	}
	return out
}

// ParseBody 解析 UserPostResIdl 响应，对应 parse_body。
func ParseBody(body []byte) (getusercontents.UserThreads, error) {
	res := &pb.UserPostResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return getusercontents.UserThreads{}, fmt.Errorf("decoding UserPostResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return getusercontents.UserThreads{}, &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}
	return getusercontents.UserThreadsFromProto(res.GetData()), nil
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
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, userID int64, pn int32, publicOnly bool) (getusercontents.UserThreads, error) {
	data := PackProto(userID, pn, publicOnly)
	resp, err := httpCore.AppProto(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return getusercontents.UserThreads{}, err
	}
	return ParseBody(resp.Body())
}

// RequestWS 执行 websocket 请求，对应 request_ws。
func RequestWS(wsCore *core.WsCore, userID int64, pn int32, publicOnly bool) (getusercontents.UserThreads, error) {
	data := PackProto(userID, pn, publicOnly)
	resp, err := wsCore.Send(data, CMD)
	if err != nil {
		return getusercontents.UserThreads{}, err
	}
	payload, err := resp.Read()
	if err != nil {
		return getusercontents.UserThreads{}, err
	}
	return ParseBody(payload)
}
