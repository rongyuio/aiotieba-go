// Package addpoll implements the add_poll API of aiotieba.
//
// It mirrors the Python package aiotieba.api.add_poll.
package addpoll

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/add_poll/protobuf"
	commonpb "github.com/rongyuio/aiotieba-go/protobuf"
)

// CMD is the websocket command of add_poll.
const CMD = 309006

// forumID is hardcoded by the Python module.
const forumID = 6

// PackProto builds the AddPollReqIdl request, mirroring pack_proto.
func PackProto(account *core.Account, tid int64, options []int64) []byte {
	parts := make([]string, len(options))
	for i, opt := range options {
		parts[i] = strconv.FormatInt(opt, 10)
	}

	req := &pb.AddPollReqIdl{
		Data: &pb.AddPollReqIdl_DataReq{
			Common: &commonpb.CommonReq{
				BDUSS:          account.BDUSS(),
				XClientType:    2,
				XClientVersion: consts.LatestVersion,
			},
			ThreadId: uint64(tid),
			Options:  strings.Join(parts, ","),
			ForumId:  forumID,
		},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil
	}
	return out
}

// ParseBody decodes an AddPollResIdl response, mirroring parse_body.
func ParseBody(body []byte) error {
	res := &pb.AddPollResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return fmt.Errorf("decoding AddPollResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}
	return nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "https",
		Host:     consts.AppBaseHost,
		Path:     "/c/c/post/addPollPost",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP performs the app HTTP request, mirroring request_http.
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, tid int64, options []int64) error {
	resp, err := httpCore.AppProto(PackProto(httpCore.Account, tid, options)).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return err
	}
	return ParseBody(resp.Body())
}

// RequestWS performs the websocket request, mirroring request_ws.
func RequestWS(wsCore *core.WsCore, tid int64, options []int64) error {
	resp, err := wsCore.Send(PackProto(wsCore.Account, tid, options), CMD)
	if err != nil {
		return err
	}
	payload, err := resp.Read()
	if err != nil {
		return err
	}
	return ParseBody(payload)
}
