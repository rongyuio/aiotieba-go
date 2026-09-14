package gettabmap

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/get_tab_map/protobuf"
	commonpb "github.com/rongyuio/aiotieba-go/protobuf"
)

// CMD 是 get_tab_map 的 websocket 命令字。
const CMD = 309466

// PackProto 构造 SearchPostForumReqIdl 请求，对应 pack_proto。
func PackProto(account *core.Account, fname string) []byte {
	req := &pb.SearchPostForumReqIdl{
		Data: &pb.SearchPostForumReqIdl_DataReq{
			Common: &commonpb.CommonReq{
				BDUSS:          account.BDUSS(),
				XClientVersion: consts.LatestVersion,
			},
			Fname: fname,
		},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil
	}
	return out
}

// ParseBody 解析 SearchPostForumResIdl 响应，对应 parse_body。
func ParseBody(body []byte) (TabMap, error) {
	res := &pb.SearchPostForumResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return TabMap{}, fmt.Errorf("decoding SearchPostForumResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return TabMap{}, &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}
	return TabMapFromProto(res.GetData()), nil
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "https",
		Host:     consts.AppBaseHost,
		Path:     "/c/f/forum/searchPostForum",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP 执行 app HTTP 请求，对应 request_http。
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, fname string) (TabMap, error) {
	resp, err := httpCore.AppProto(PackProto(httpCore.Account, fname)).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return TabMap{}, err
	}
	return ParseBody(resp.Body())
}

// RequestWS 执行 websocket 请求，对应 request_ws。
func RequestWS(wsCore *core.WsCore, fname string) (TabMap, error) {
	resp, err := wsCore.Send(PackProto(wsCore.Account, fname), CMD)
	if err != nil {
		return TabMap{}, err
	}
	payload, err := resp.Read()
	if err != nil {
		return TabMap{}, err
	}
	return ParseBody(payload)
}
