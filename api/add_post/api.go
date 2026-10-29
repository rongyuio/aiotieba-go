// Package addpost 实现 aiotieba 的 add_post API。
//
// 对应 Python 包 aiotieba.api.add_post。
//
// 上游在 2ee06ee3 删除了该模块，本包是按删除前的 Python 实现移植、再按
// 贴吧 iOS 22.12.1.0 客户端的实际抓包（c/c/post/add，cmd=309731）更新而来。
// 设备档案与 api/share 共用同一套 iOS 取值与同一份 helper/deviceid 派生逻辑。
package addpost

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"google.golang.org/protobuf/proto"

	pb "github.com/rongyuio/aiotieba-go/api/add_post/protobuf"
	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper/deviceid"
	commonpb "github.com/rongyuio/aiotieba-go/protobuf"
)

// CMD 是 add_post 的 websocket 命令字。
const CMD = 309731

// 设备档案对齐 iOS 客户端，取值来自真机抓包。
//
// 机型、品牌、系统版本是公开的客户端取值，硬编码不产生指纹关联；
// 而 cuid / caid / unique_id 这类设备级标识一律由 helper/deviceid 从账号派生，
// 不照抄抓包里的真机值 —— 照抄等于所有使用者共用一个设备指纹。
//
// iosModel 同时填 model(9) 与 device_model(27)：抓包中这两个字段取值相同，
// 共用常量可从结构上保证它们永远一致。
const (
	iosModel         = deviceid.IOSModel
	iosBrand         = deviceid.IOSBrand
	iosOSVersion     = deviceid.IOSOSVersion
	iosClientVersion = consts.IOSVersion
)

// 抓包中恒定、与具体设备无关的字段值。
const (
	appName        = "tieba"    // app_name(13)
	legoLibVersion = "3.0.0"    // lego_lib_version(28)
	sdkVer         = "3.54.0"   // sdk_ver(42)
	frameworkVer   = "4.520.1"  // framework_ver(43)
	nawsGameVer    = "1.21.0"   // naws_game_ver(44)
	deviceScore    = "0.800000" // device_score(70)，抓包里是 8 字符定长
	scrW           = 1179       // scr_w(37)
	scrH           = 2556       // scr_h(38)
	scrDip         = 3.0        // scr_dip(39)
	qType          = 100        // q_type(40)

	vcodeTag = "12"       // vcode_tag(16)
	newVcode = "1"        // new_vcode(18)
	postFrom = "pb_reply" // post_from(80)
)

// active_timestamp(49) 在抓包中是一个固定的旧时间戳（2024-01-19），不随请求变化。
// 上游 Python 版用 now-30d，那个值会随每次请求一起漂移，比稳定值更像合成流量。
const (
	activeBaseEpoch = int64(1704067200)  // 2024-01-01
	activeSpanSecs  = int64(540 * 86400) // 约 18 个月
)

// buildTabTrace 构造 x_field_91(91)：base64 编码的页面栈 JSON。
//
// 抓包真值（解码后）：
//
//	[{"c":"home_rec","v":"","d":"0"},{"p":"my_tab","v":"","d":"5140"},
//	 {"p":"TBCMyPostAndReplyViewController","v":"","d":"686"},
//	 {"p":"pb","v":"10709435542","d":"2299"}]
//
// 只有最后一项的 v 随请求变化，是当前主题帖的 tid；各段 d 是页面耗时，抓包中固定。
func buildTabTrace(tid int64) string {
	trace := `[{"c":"home_rec","v":"","d":"0"},{"p":"my_tab","v":"","d":"5140"},` +
		`{"p":"TBCMyPostAndReplyViewController","v":"","d":"686"},` +
		`{"p":"pb","v":"` + strconv.FormatInt(tid, 10) + `","d":"2299"}]`
	return base64.StdEncoding.EncodeToString([]byte(trace))
}

// PackProto 构造 AddPostReqIdl 请求，对应 pack_proto。
//
// 发帖在贴吧平台属于高风险操作，请求携带完整的设备指纹与调用链。
//
// 参数:
//
//	account 账号
//	fname 吧名
//	fid 吧id
//	tid 主题帖id
//	showName 发帖人展示名
//	content 回复内容
//	last 上一次 app 调用的记录，用于申报 m_* 四元组（log_source / log_source_id /
//	  x_field_18 / x_field_19）。HTTP 通路必须传 core.HttpCore.LastCall()，
//	  不能编造 —— 服务端能看出链条不成立。WebSocket 通路没有前置 HTTP 调用，
//	  传零值即可，四个字段会随之省略。
func PackProto(account *core.Account, fname string, fid, tid int64, showName, content string, last core.CallRecord) ([]byte, error) {
	now := time.Now()
	currentTsms := now.UnixMilli()
	androidID := account.AndroidID()
	uniqueID := deviceid.UniqueID(androidID)
	fidStr := strconv.FormatInt(fid, 10)
	tidStr := strconv.FormatInt(tid, 10)

	req := &pb.AddPostReqIdl{
		Data: &pb.AddPostReqIdl_DataReq{
			Common: &commonpb.CommonReq{
				// 账号持有的身份凭据
				BDUSS:    account.BDUSS(),
				Stoken:   account.STOKEN(),
				ZId:      account.ZID(),
				Tbs:      account.Tbs(),
				SampleId: account.SampleID(),

				// 客户端身份：iOS 22.12.1.0
				XClientType:    1,
				XClientVersion: iosClientVersion,
				XClientId:      account.ClientID(),
				XFrom:          "appstore",
				AppName:        appName,

				// 设备档案：机型是公开常量，标识由账号派生
				Cuid:           deviceid.Cuid(androidID),
				Model:          iosModel,
				Brand:          iosBrand,
				DeviceModel:    iosModel,
				XOsVersion:     iosOSVersion,
				Idfv:           deviceid.IDFV(account.UUID()),
				Idfa:           deviceid.ZeroIDFA,
				Caid:           deviceid.Caid(androidID, account.UUID(), iosOSVersion),
				UniqueId:       uniqueID,
				XField_59:      deviceid.DeriveHex(androidID, "x_field_59", 20),
				UserAgent:      deviceid.UserAgent(uniqueID, iosClientVersion, iosOSVersion),
				ScrW:           scrW,
				ScrH:           scrH,
				ScrDip:         scrDip,
				QType:          qType,
				DeviceScore:    deviceScore,
				LegoLibVersion: legoLibVersion,
				SdkVer:         sdkVer,
				FrameworkVer:   frameworkVer,
				NawsGameVer:    nawsGameVer,

				// 时序与投放开关
				XTimestamp:            currentTsms,
				ActiveTimestamp:       deviceid.DeriveEpoch(androidID, "active_timestamp", activeBaseEpoch, activeSpanSecs) * 1000,
				EventDay:              now.Format("20060102"),
				NetType:               1,
				IsTeenager:            0,
				Cmode:                 1,
				StartType:             1,
				StartScheme:           "",
				Extra:                 "",
				PersonalizedRecSwitch: 0,

				// 客户端常量
				HybridVersion: deviceid.HybridVersion,
				ExpFlags:      deviceid.ExpFlags,

				// 日志类字段
				LogSource:   last.API,
				LogSourceId: last.LogID,
				XField_18:   last.Cost,
				XField_19:   last.Result,
				XField_23:   deviceid.NewTraceHex(),
				XField_91:   buildTabTrace(tid),
			},
			Tbs:              account.Tbs(),
			Anonymous:        "0",
			VcodeTag:         vcodeTag,
			NewVcode:         newVcode,
			Content:          content,
			Fid:              fidStr,
			Kw:               fname,
			Tid:              tidStr,
			NameShow:         showName,
			XField_33:        "pb",
			XField_39:        "1",
			XField_48:        "0",
			XField_52:        "0",
			XField_55:        "12",
			PostFrom:         postFrom,
			ShowCustomFigure: 0,
		},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("encoding AddPostReqIdl: %w", err)
	}
	return out, nil
}

// ParseBody 解析 AddPostResIdl 响应，对应 parse_body。
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

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "https",
		Host:     consts.AppBaseHost,
		Path:     "/c/c/post/add",
		RawQuery: "cmd=" + strconv.Itoa(CMD),
	}
}

// RequestHTTP 执行 app HTTP 请求，对应 request_http。
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, fname string, fid, tid int64, showName, content string) error {
	data, err := PackProto(httpCore.Account, fname, fid, tid, showName, content, httpCore.LastCall())
	if err != nil {
		return err
	}
	resp, err := httpCore.AppProto(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return err
	}
	return ParseBody(resp.Body())
}

// RequestWS 执行 websocket 请求，对应 request_ws。
//
// WebSocket 通路没有紧邻其前的 HTTP 调用可申报，m_* 四元组传零值。
func RequestWS(wsCore *core.WsCore, fname string, fid, tid int64, showName, content string) error {
	data, err := PackProto(wsCore.Account, fname, fid, tid, showName, content, core.CallRecord{})
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
