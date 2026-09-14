package getforumlevel

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/get_forum_level/protobuf"
	commonpb "github.com/rongyuio/aiotieba-go/protobuf"
)

// CMD 是 get_forum_level 的 websocket 命令字。
const CMD = 301005

// PackProto 构造 GetLevelInfoReqIdl 请求，对应 pack_proto。
func PackProto(account *core.Account, fid int64) []byte {
	req := &pb.GetLevelInfoReqIdl{
		Data: &pb.GetLevelInfoReqIdl_DataReq{
			Common:  &commonpb.CommonReq{BDUSS: account.BDUSS()},
			ForumId: fid,
		},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil
	}
	return out
}

// ParseBody 解析 GetLevelInfoResIdl 响应，对应 parse_body。
func ParseBody(body []byte) (LevelInfo, error) {
	res := &pb.GetLevelInfoResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return LevelInfo{}, fmt.Errorf("decoding GetLevelInfoResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return LevelInfo{}, &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}
	return LevelInfoFromProto(res.GetData()), nil
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "https",
		Host:     consts.AppBaseHost,
		Path:     "/c/f/forum/getLevelInfo",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP 执行 app HTTP 请求，对应 request_http。
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, fid int64) (LevelInfo, error) {
	resp, err := httpCore.AppProto(PackProto(httpCore.Account, fid)).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return LevelInfo{}, err
	}
	return ParseBody(resp.Body())
}

// RequestWS 执行 websocket 请求，对应 request_ws。
func RequestWS(wsCore *core.WsCore, fid int64) (LevelInfo, error) {
	resp, err := wsCore.Send(PackProto(wsCore.Account, fid), CMD)
	if err != nil {
		return LevelInfo{}, err
	}
	payload, err := resp.Read()
	if err != nil {
		return LevelInfo{}, err
	}
	return ParseBody(payload)
}
