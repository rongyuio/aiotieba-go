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

// CMD is the websocket command of get_comments.
const CMD = 302002

// PackProto builds the PbFloorReqIdl request, mirroring pack_proto.
//
// Unlike most APIs it does not need the account.
func PackProto(tid, pid int64, pn int32, isComment bool) []byte {
	data := &pb.PbFloorReqIdl_DataReq{
		Common: &commonpb.CommonReq{
			XClientType:    2,
			XClientVersion: consts.LegacyVersion,
		},
		Kz: tid,
		Pn: pn,
	}
	// A comment id is passed as spid, a floor id as pid.
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

// ParseBody decodes a PbFloorResIdl response, mirroring parse_body.
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

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "http",
		Host:     consts.AppBaseHost,
		Path:     "/c/f/pb/floor",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP performs the app HTTP request, mirroring request_http.
func RequestHTTP(
	ctx context.Context, httpCore *core.HttpCore, tid, pid int64, pn int32, isComment bool,
) (Comments, error) {
	req, err := httpCore.PackProtoRequest(ctx, RequestURL(), PackProto(tid, pid, pn, isComment))
	if err != nil {
		return Comments{}, err
	}
	body, err := httpCore.SendProto(req)
	if err != nil {
		return Comments{}, err
	}
	return ParseBody(body)
}

// RequestWS performs the websocket request, mirroring request_ws.
func RequestWS(wsCore *core.WsCore, tid, pid int64, pn int32, isComment bool) (Comments, error) {
	resp, err := wsCore.Send(PackProto(tid, pid, pn, isComment), CMD)
	if err != nil {
		return Comments{}, err
	}
	payload, err := resp.Read()
	if err != nil {
		return Comments{}, err
	}
	return ParseBody(payload)
}
