// Package share 实现 aiotieba 的 share API。
//
// 该接口对应 Client.Share，向服务端上报一次「分享主题帖」的行为。
// 上游 Python 版 aiotieba 没有这个 API，实现完全依据实际抓包
// （c/c/thread/share，cmd=309480）反解而来。
//
// 该接口只上报行为，不返回分享链接或文案，并且没有 WebSocket 通路，只能走 HTTP。
//
// 关于生效方式（实测结论，非推测）：
//
//   - 服务端**不校验** share_channel：越界值 999 与合法值 0/3 一样返回 errorno=0
//   - 本实现发出的请求**不改变帖子的分享数**：对一个 ShareNum=0 的冷门帖子分享后
//     仍是 0，二十多分钟后复查仍为 0；对 ShareNum=1 的帖子连发 7 次也不变。
//     errorno=0 只表示服务端收下了这条上报
//   - **但真实客户端分享会 +1**：实测客户端分享过的两个冷门帖子分别由 0 涨到 3 与 1。
//     差异原因尚未查清。早期怀疑的两点（用安卓指纹、UA 自报 aiotieba/<版本>）已经修掉：
//     设备档案改成与抓包一致的 iOS，HTTP 头与 CommonReq 的 UA 都换成真机形态。
//     若仍不记账，下一个怀疑对象是 log_id —— 见下方 buildCaid 之后那段实测记录
package share

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"google.golang.org/protobuf/proto"

	pb "github.com/rongyuio/aiotieba-go/api/share/protobuf"
	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper/deviceid"
	commonpb "github.com/rongyuio/aiotieba-go/protobuf"
)

// CMD 是 share 的命令字。
const CMD = 309480

// 设备档案对齐 iOS 客户端。
//
// 取值直接来自真机抓包（c/c/thread/share，客户端 22.12.1.0）。
// 早期实现沿用 add_post 的 Android 一套（SM-G988N / Android 9），
// 实测服务端回 errno=0 但不记账 —— 账号在服务端的历史是 iOS 设备，
// 突然出现一台安卓机本身就是异常信号。
//
// 机型、品牌、系统版本是公开的客户端取值，硬编码不产生指纹关联；
// 而 caid / unique_id / idfv 这类设备级标识一律由 helper/deviceid
// 从账号派生，绝不照抄抓包里的真机值。
//
// iosModel 同时填 model(9) 与 brand_type(27)：抓包中这两个字段取值相同，
// 共用常量可从结构上保证它们永远一致。
const (
	iosModel         = deviceid.IOSModel
	iosBrand         = deviceid.IOSBrand
	iosOSVersion     = deviceid.IOSOSVersion
	iosClientVersion = consts.IOSVersion
)

// 抓包中恒定、且与具体设备无关的字段值。
const (
	// shareAppName 对应 app_name(13)，应用名，无熵。
	shareAppName = "tieba"

	// shareIDFA 对应 idfa(45)。idfa 是 iOS 概念，全 0 UUID 表示「不可用」，
	// 抓包中即为全 0，与 Android 指纹不冲突。
	shareIDFA = deviceid.ZeroIDFA

	// shareHybridVersion 对应 hybrid_version(88)，客户端内嵌 webview 的版本号。
	shareHybridVersion = deviceid.HybridVersion

	// shareExpFlags 对应 exp_flags(89)，是实验开关列表。
	// 实验开关按定义就是全量用户共享的灰度标记，硬编码不产生指纹关联。
	shareExpFlags = deviceid.ExpFlags
)

// 分享请求的 log_id(2) 承载的是**被分享帖子的 tid**，不是随机数。
//
// 实测证据（对真机可重放的请求做单变量替换，再读目标帖的 ShareNum）：
//
//	把 log_id 改成另一个帖子的 tid -> 那个帖子 +1，原帖子不动
//	把 log_id 改成一个不存在的值   -> 谁都不涨
//
// 服务端只认这一个字段：设备字段、Cookie、HTTP 头、sign(23) 与 m_* 四元组
// 逐个替换成随机值后，计数照样 +1。旧实现生成随机 log_id，等于每次都在
// 上报一个不存在的帖子，因此服务端回 errno=0 却从不记账。

// 设备级标识与 caid 的构造集中在 helper/deviceid，add_post 与 share 共用一份实现。
// 那里的取值全部由账号持久化的 android_id 派生：同一账号每次上报都相同，
// 不同账号互不相同。若改成硬编码常量，所有使用者会共享同一个设备指纹。

// PackProto 构造 ShareReqIdl 请求，对应 pack_proto。
//
// 设备指纹沿用 add_post 的 Android 一套（SM-G988N / Android 9），
// 并补齐抓包中出现的日志类字段。
//
// 参数:
//
//	account 账号
//	tid 被分享的主题帖id，会同时写入 log_source_id
//	shareChannel 对应 share_channel(3)，语义未确认，传 0 表示不设置该字段。
//	  实测服务端不校验该字段（越界值同样被接受），抓包中出现过的取值只有 0 与 3
//	last 上一次 app 调用的记录，用于申报 m_* 四元组。
//	  必须来自 core.HttpCore.LastCall()，不能是编造的值 —— 服务端能看出链条不成立。
func PackProto(account *core.Account, tid int64, shareChannel int32, last core.CallRecord) ([]byte, error) {
	now := time.Now()
	currentTsms := now.UnixMilli()
	// 86400 * 30 对应 Python 模块的精确算法。
	offset := int64(86400 * 30)
	// event_day 用抓包里的 8 位补零格式，与 add_post 的 %d%d%d 不同。
	eventDay := now.Format("20060102")
	androidID := account.AndroidID()
	uniqueID := deviceid.UniqueID(androidID)

	req := &pb.ShareReqIdl{
		Data: &pb.ShareReqIdl_DataReq{
			Common: &commonpb.CommonReq{
				BDUSS:                 account.BDUSS(),
				XClientType:           1,
				XClientVersion:        iosClientVersion,
				XClientId:             account.ClientID(),
				XFrom:                 "appstore",
				Cuid:                  deviceid.Cuid(androidID),
				XTimestamp:            currentTsms,
				Model:                 iosModel,
				Tbs:                   account.Tbs(),
				NetType:               1,
				XOsVersion:            iosOSVersion,
				Brand:                 iosBrand,
				LegoLibVersion:        "3.0.0",
				Applist:               "",
				Stoken:                account.STOKEN(),
				ZId:                   account.ZID(),
				CuidGid:               "",
				SampleId:              account.SampleID(),
				ScrW:                  1179,
				ScrH:                  2556,
				ScrDip:                3.0,
				QType:                 100,
				IsTeenager:            0,
				SdkVer:                "3.54.0",
				FrameworkVer:          "4.520.1",
				NawsGameVer:           "1.21.0",
				ActiveTimestamp:       currentTsms - offset,
				EventDay:              eventDay,
				Cmode:                 1,
				StartType:             1,
				Idfv:                  deviceid.IDFV(account.UUID()),
				UserAgent:             deviceid.UserAgent(uniqueID, iosClientVersion, iosOSVersion),
				PersonalizedRecSwitch: 0,
				DeviceScore:           "0.800000",

				// 以下为分享接口特有的日志字段
				AppName:       shareAppName,
				LogSource:     last.API,
				LogSourceId:   last.LogID,
				XField_18:     last.Cost,
				XField_19:     last.Result,
				XField_23:     deviceid.NewTraceHex(),
				DeviceModel:   iosModel,
				Idfa:          shareIDFA,
				Caid:          deviceid.Caid(androidID, account.UUID(), iosOSVersion),
				UniqueId:      uniqueID,
				XField_59:     deviceid.DeriveHex(androidID, "x_field_59", 20),
				HybridVersion: shareHybridVersion,
				ExpFlags:      shareExpFlags,
			},
			LogId:        tid,
			ShareChannel: shareChannel,
		},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("encoding ShareReqIdl: %w", err)
	}
	return out, nil
}

// ParseBody 解析 ShareResIdl 响应，对应 parse_body。
//
// 响应里的 data.log_id 是服务端对请求的回显，这里不做校验：
// 上报是单向的，errorno 为 0 即视为成功。
func ParseBody(body []byte) error {
	res := &pb.ShareResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return fmt.Errorf("decoding ShareResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}
	return nil
}

// RequestURL 返回该 API 的请求地址。
//
// 抓包中该请求带有 format=protobuf，项目内其他 protobuf 接口都没有这个查询参数，
// 但它在 1、3 号样本里都被显式发送，保留可以最大程度贴合真实流量。
func RequestURL() *url.URL {
	return &url.URL{
		Scheme:   "https",
		Host:     consts.AppBaseHost,
		Path:     "/c/c/thread/share",
		RawQuery: "cmd=" + strconv.Itoa(CMD) + "&format=protobuf",
	}
}

// RequestHTTP 执行 app HTTP 请求，对应 request_http。
//
// 本接口没有 WebSocket 通路，因此没有对应的 RequestWS。
//
// 先调 getShareInfo 再调 thread/share，复现真实客户端的调用序列：
// thread/share 的 m_* 四元组申报的就是紧邻其前的那次 getShareInfo ——
// 路径、服务端返回的 logid、耗时与错误码。若跳过前置调用，
// 四元组只能靠编造，而服务端能看出链条不成立（回 errorno=0 但不记账）。
func RequestHTTP(ctx context.Context, httpCore *core.HttpCore, tid, fid int64, shareChannel int32) error {
	if err := requestShareInfo(ctx, httpCore, tid, fid); err != nil {
		return err
	}
	data, err := PackProto(httpCore.Account, tid, shareChannel, httpCore.LastCall())
	if err != nil {
		return err
	}
	resp, err := httpCore.AppProto(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return err
	}
	return ParseBody(resp.Body())
}
