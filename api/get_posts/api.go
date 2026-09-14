package getposts

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/get_posts/protobuf"
	commonpb "github.com/rongyuio/aiotieba-go/protobuf"
)

// CMD is the websocket command of get_posts.
const CMD = 302001

// PackProto builds the PbPageReqIdl request, mirroring pack_proto.
func PackProto(
	account *core.Account,
	tid int64,
	pn, rn, sort int32,
	onlyThreadAuthor, withComments, commentSortByAgree bool,
	commentRn int32,
) []byte {
	protoRn := rn
	if rn <= 1 {
		protoRn = 2
	}

	common := &commonpb.CommonReq{
		XClientType:    2,
		XClientVersion: consts.LegacyVersion,
	}
	if withComments {
		common.BDUSS = account.BDUSS()
	}

	req := &pb.PbPageReqIdl{
		Data: &pb.PbPageReqIdl_DataReq{
			Common:        common,
			Kz:            tid,
			Pn:            pn,
			Rn:            protoRn,
			R:             sort,
			Lz:            boolToInt32(onlyThreadAuthor),
			WithFloor:     boolToInt32(withComments),
			FloorSortType: boolToInt32(commentSortByAgree),
			FloorRn:       commentRn,
		},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil
	}
	return out
}

// ParseBody decodes a PbPageResIdl response, mirroring parse_body.
func ParseBody(body []byte) (Posts, error) {
	res := &pb.PbPageResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return Posts{}, fmt.Errorf("decoding PbPageResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return Posts{}, &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}
	return PostsFromProto(res.GetData()), nil
}

// RequestURL returns the endpoint of the API. Unlike get_threads it uses HTTPS.
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "https",
		Host:     consts.AppBaseHost,
		Path:     "/c/f/pb/page",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP performs the app HTTP request, mirroring request_http.
func RequestHTTP(
	ctx context.Context, httpCore *core.HttpCore, tid int64, pn, rn, sort int32,
	onlyThreadAuthor, withComments, commentSortByAgree bool, commentRn int32,
) (Posts, error) {
	data := PackProto(httpCore.Account, tid, pn, rn, sort, onlyThreadAuthor, withComments, commentSortByAgree, commentRn)
	req, err := httpCore.PackProtoRequest(ctx, RequestURL(), data)
	if err != nil {
		return Posts{}, err
	}
	body, err := httpCore.SendProto(req)
	if err != nil {
		return Posts{}, err
	}
	return ParseBody(body)
}

// RequestWS performs the websocket request, mirroring request_ws.
func RequestWS(
	wsCore *core.WsCore, tid int64, pn, rn, sort int32,
	onlyThreadAuthor, withComments, commentSortByAgree bool, commentRn int32,
) (Posts, error) {
	data := PackProto(wsCore.Account, tid, pn, rn, sort, onlyThreadAuthor, withComments, commentSortByAgree, commentRn)
	resp, err := wsCore.Send(data, CMD)
	if err != nil {
		return Posts{}, err
	}
	payload, err := resp.Read()
	if err != nil {
		return Posts{}, err
	}
	return ParseBody(payload)
}

func boolToInt32(v bool) int32 {
	if v {
		return 1
	}
	return 0
}
