// Package addpost implements the add_post API of aiotieba.
//
// It mirrors the Python package aiotieba.api.add_post.
package addpost

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"google.golang.org/protobuf/proto"

	pb "github.com/rongyuio/aiotieba-go/api/add_post/protobuf"
	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	commonpb "github.com/rongyuio/aiotieba-go/protobuf"
)

// CMD is the websocket command of add_post.
const CMD = 309731

// addPostClientVersion is the client version hardcoded by the Python module
// (unlike most APIs, add_post does not use LATEST_VERSION).
const addPostClientVersion = "12.35.1.0"

// PackProto builds the AddPostReqIdl request, mirroring pack_proto.
//
// The request carries the full device fingerprint because posting is a
// high-risk operation on the Tieba platform.
func PackProto(account *core.Account, fname string, fid, tid int64, showName, content string) ([]byte, error) {
	cuidGalaxy2, err := account.CuidGalaxy2()
	if err != nil {
		return nil, err
	}
	c3Aid, err := account.C3Aid()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	currentTsms := now.UnixMilli()
	// 86400 * 30 mirrors the exact arithmetic of the Python module.
	offset := int64(86400 * 30)
	eventDay := fmt.Sprintf("%d%d%d", now.Year(), int(now.Month()), now.Day())
	fidStr := strconv.FormatInt(fid, 10)
	tidStr := strconv.FormatInt(tid, 10)

	req := &pb.AddPostReqIdl{
		Data: &pb.AddPostReqIdl_DataReq{
			Common: &commonpb.CommonReq{
				BDUSS:                 account.BDUSS(),
				XClientType:           2,
				XClientVersion:        addPostClientVersion,
				XClientId:             account.ClientID(),
				XPhoneImei:            "000000000000000",
				XFrom:                 "1008621x",
				Cuid:                  cuidGalaxy2,
				XTimestamp:            currentTsms,
				Model:                 "SM-G988N",
				Tbs:                   account.Tbs(),
				NetType:               1,
				Pversion:              "1.0.3",
				XOsVersion:            "9",
				Brand:                 "samsung",
				LegoLibVersion:        "3.0.0",
				Applist:               "",
				Stoken:                account.STOKEN(),
				ZId:                   account.ZID(),
				CuidGalaxy2:           cuidGalaxy2,
				CuidGid:               "",
				C3Aid:                 c3Aid,
				SampleId:              account.SampleID(),
				ScrW:                  720,
				ScrH:                  1280,
				ScrDip:                1.5,
				QType:                 0,
				IsTeenager:            0,
				SdkVer:                "2.34.0",
				FrameworkVer:          "3340042",
				NawsGameVer:           "1038000",
				ActiveTimestamp:       currentTsms - offset,
				FirstInstallTime:      currentTsms - offset,
				LastUpdateTime:        currentTsms - offset,
				EventDay:              eventDay,
				AndroidId:             account.AndroidID(),
				Cmode:                 1,
				StartScheme:           "",
				StartType:             1,
				Idfv:                  "0",
				Extra:                 "",
				UserAgent:             "aiotieba/" + consts.Version,
				PersonalizedRecSwitch: 1,
				DeviceScore:           "0.4",
			},
			Anonymous:        "1",
			CanNoForum:       "0",
			IsFeedback:       "0",
			TakephotoNum:     "0",
			EntranceType:     "0",
			VcodeTag:         "12",
			NewVcode:         "1",
			Content:          content,
			Fid:              fidStr,
			VFid:             "",
			VFname:           "",
			Kw:               fname,
			IsBarrage:        "0",
			FromFourmId:      fidStr,
			Tid:              tidStr,
			IsAd:             "0",
			PostFrom:         "3",
			NameShow:         showName,
			IsPictxt:         "0",
			ShowCustomFigure: 0,
			IsShowBless:      0,
		},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("encoding AddPostReqIdl: %w", err)
	}
	return out, nil
}

// ParseBody decodes an AddPostResIdl response, mirroring parse_body.
func ParseBody(body []byte) error {
	res := &pb.AddPostResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return fmt.Errorf("decoding AddPostResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}
	if info := res.GetData().GetInfo(); info != nil {
		if vcode, err := strconv.Atoi(info.GetNeedVcode()); err == nil && vcode != 0 {
			return &exception.TiebaValueError{Msg: "Need verify code"}
		}
	}
	return nil
}

// RequestURL returns the endpoint of the API.
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "https",
		Host:     consts.AppBaseHost,
		Path:     "/c/c/post/add",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP performs the app HTTP request, mirroring request_http.
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, fname string, fid, tid int64, showName, content string) error {
	data, err := PackProto(httpCore.Account, fname, fid, tid, showName, content)
	if err != nil {
		return err
	}
	resp, err := httpCore.AppProto(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return err
	}
	return ParseBody(resp.Body())
}

// RequestWS performs the websocket request, mirroring request_ws.
func RequestWS(wsCore *core.WsCore, fname string, fid, tid int64, showName, content string) error {
	data, err := PackProto(wsCore.Account, fname, fid, tid, showName, content)
	if err != nil {
		return err
	}
	resp, err := wsCore.Send(data, CMD)
	if err != nil {
		return err
	}
	payload, err := resp.Read()
	if err != nil {
		return err
	}
	return ParseBody(payload)
}
