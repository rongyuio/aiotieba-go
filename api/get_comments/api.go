package getcomments

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/get_comments/protobuf"
	commonpb "github.com/rongyuio/aiotieba-go/protobuf"
)

// CMD 是 get_comments 的 websocket 命令字。
const CMD = 302002

// PackProto 构造 PbFloorReqIdl 请求，对应 pack_proto。
//
// 与大多数 API 不同，它不需要账户。
func PackProto(tid, pid int64, pn, sort int32, isComment bool) []byte {
	data := &pb.PbFloorReqIdl_DataReq{
		Common: &commonpb.CommonReq{
			XClientType:    2,
			XClientVersion: consts.LatestVersion,
		},
		Kz:   tid,
		Pn:   pn,
		Sort: sort,
	}
	// 评论 id 通过 spid 传递，楼层 id 通过 pid 传递。
	if isComment {
		data.Spid = pid
	} else {
		data.Pid = pid
	}

	out, err := proto.Marshal(&pb.PbFloorReqIdl{Data: data})
	if err != nil {
		return nil
	}
	return out
}

// ParseBody 解析 PbFloorResIdl 响应，对应 parse_body。
func ParseBody(body []byte) (Comments, error) {
	res := &pb.PbFloorResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return Comments{}, fmt.Errorf("decoding PbFloorResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return Comments{}, &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}
	return CommentsFromProto(res.GetData()), nil
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "http",
		Host:     consts.AppBaseHost,
		Path:     "/c/f/pb/floor",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP 执行 app HTTP 请求，对应 request_http。
func RequestHTTP(
	ctx context.Context, httpCore *core.HttpCore, tid, pid int64, pn, sort int32, isComment bool,
) (Comments, error) {
	resp, err := httpCore.AppProto(PackProto(tid, pid, pn, sort, isComment)).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return Comments{}, err
	}
	return ParseBody(resp.Body())
}

// RequestWS 执行 websocket 请求，对应 request_ws。
func RequestWS(wsCore *core.WsCore, tid, pid int64, pn, sort int32, isComment bool) (Comments, error) {
	resp, err := wsCore.Send(PackProto(tid, pid, pn, sort, isComment), CMD)
	if err != nil {
		return Comments{}, err
	}
	payload, err := resp.Read()
	if err != nil {
		return Comments{}, err
	}
	return ParseBody(payload)
}
