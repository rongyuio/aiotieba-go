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
//     差异原因尚未查清，首要嫌疑是本实现用 Android 指纹而抓包来自 iOS 客户端，
//     且 HTTP User-Agent 与 CommonReq.user_agent 都填 "aiotieba/<版本>"，
//     而真实客户端发的是完整的 iOS Safari UA
package share

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/url"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	pb "github.com/rongyuio/aiotieba-go/api/share/protobuf"
	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
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
// iosModel 同时填 model(9) 与 brand_type(27)：抓包中这两个字段取值相同，
// 共用常量可从结构上保证它们永远一致。
const (
	iosModel         = "iPhone 15 Pro"
	iosBrand         = "iPhone"
	iosOSVersion     = "18.7.2"
	iosClientVersion = consts.IOSVersion
)

// 抓包中恒定、且与具体设备无关的字段值。
const (
	// shareAppName 对应 app_name(13)，应用名，无熵。
	shareAppName = "tieba"

	// shareIDFA 对应 idfa(45)。idfa 是 iOS 概念，全 0 UUID 表示「不可用」，
	// 抓包中即为全 0，与 Android 指纹不冲突。
	shareIDFA = "00000000-0000-0000-0000-000000000000"

	// shareHybridVersion 对应 hybrid_version(88)，客户端内嵌 webview 的版本号。
	shareHybridVersion = "hybrid-main-pb_1.0.397.1"

	// shareExpFlags 对应 exp_flags(89)，是实验开关列表。
	// 实验开关按定义就是全量用户共享的灰度标记，硬编码不产生指纹关联。
	shareExpFlags = "22.8_pb_floor_reply_rebuild_b^22.8_pb_floor_reply_rebuild_a,22.8_pb_floor_reply_rebuild_origin"

	// shareCaidGenerateTime 对应 caid(46) 各渠道的 generateTime。
	// caid 在设备上是长期有效的广告标识，1、3 号样本完全相同，取 1 号样本的值固定下来。
	shareCaidGenerateTime = 1790071384
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

// newTraceHex 生成一个 32 位大写 hex，对应 x_field_23(23)。
//
// 该字段即分享链接中的 unique 参数（形如 ...&unique=63BD6FF2A2ACBB3200A3AC48F79228C7），
// 由客户端在每次分享时现生成，因此这里也按每请求随机。做成设备级常量反而比真实客户端更显眼。
func newTraceHex() string {
	return strings.ToUpper(fmt.Sprintf("%016x%016x", rand.Uint64(), rand.Uint64()))
}

// deriveHex 由设备信息派生一个稳定且设备唯一的 hex 串。
//
// caid、unique_id、x_field_59 在抓包里跨请求完全一致，属于设备级常量。
// getShareInfo 的抓包证实后两者分别是 form 里的 shoubai_cuid 与 mac。
// 这里统一由账号持久化的 android_id 派生：同一账号每次上报都相同，不同账号互不相同。
// 若改成硬编码常量，所有使用者会共享同一个设备指纹。
func deriveHex(androidID, label string, n int) string {
	sum := sha256.Sum256([]byte(label + "|" + androidID))
	return hex.EncodeToString(sum[:n])
}

// deriveUniqueID 对应 unique_id(58)。
//
// 抓包中是 51 个大写字母数字字符，并且被同步写进 user_agent 的 uniqueId/ 段。
// getShareInfo 的抓包显示它在 form 参数里叫 shoubai_cuid，取值逐字相同，
// 是客户端级的设备标识。
// 这里取 SHA-256 前 26 字节的 hex 再截到 51 位，长度对齐、字符集是其子集。
func deriveUniqueID(androidID string) string {
	return strings.ToUpper(deriveHex(androidID, "unique_id", 26))[:51]
}

// caidSubID 是 caid 中单个渠道的标识条目。
type caidSubID struct {
	Version string `json:"version"`
	Caid    string `json:"caid"`
}

// caidEntry 是 caid 中的单个渠道。
type caidEntry struct {
	Vendor       string      `json:"vendor"`
	GenerateTime string      `json:"generateTime"`
	Caid         []caidSubID `json:"caid"`
}

// buildCaid 构造 caid(46) 的值，返回 URL 编码后的 JSON。
//
// 抓包中的结构为 3 个渠道（vendor 0/1/2）的嵌套对象，末尾带 caid_valid。
// 所有 ID 都由 android_id 派生，保证同一账号恒定、不同账号互不相同。
//
// 抓包里的 factors_data 是一份包含内存容量、磁盘容量、开机时间、机型哈希等 14 项
// 设备画像的 base64 JSON。没有 Android 侧的对应真值，凭空编造比省略更容易露馅，
// 而服务端只是把它写进日志，因此这里直接省略该键。
// caidFactors 是 caid 里 factors_data 的设备画像，对应抓包中那段 base64 JSON。
//
// 抓包真值（解码后）：
//
//	{"factorsVersion":"3","factorsForCaid":{
//	  "physicalMemory":"8056045568","carrierInfo":"--","machine":"iPhone16,1",
//	  "sysFileTime":"1763470956.562973","countryCode":"CN","bootSecTime":"1785781330",
//	  "deviceName":"867E57BD062C7169995DC03CC0541C19","timeZone":"28800",
//	  "language":"zh-Hans-CN","disk":"255515516928","systemVersion":"18.7.2",
//	  "cpuNumber":"6","model":"D83AP","mobBirthTime":"1707014879.062328050"}}
//
// 其中 machine / model 是 iPhone 15 Pro 的机型码，属于客户端常量，硬编码无熵；
// deviceName、各时间戳与容量则由账号派生 —— 同一账号恒定、不同账号互不相同。
type caidFactors struct {
	Version string            `json:"factorsVersion"`
	ForCaid map[string]string `json:"factorsForCaid"`
}

// buildFactorsData 构造 factors_data，返回 base64 编码后的字符串。
//
// 旧实现直接省略该键（理由是没有 Android 侧的真值）。改用 iOS 档案后有了真值模板，
// 补齐它可以避免 caid 比真实客户端短一大截 —— 抓包中真机 diac(46) 为 1435 字节，
// 省略 factors_data 后只有 628 字节。
func buildFactorsData(account *core.Account) string {
	id := account.AndroidID()
	// 由 android_id 派生一个稳定的秒级时间戳，落在 2023-2026 之间的合理区间。
	base := int64(1672531200) // 2023-01-01
	offset := int64(0)
	for i := 0; i < 6; i++ {
		offset = offset*256 + int64(id[i])
	}
	born := base + offset%(3*365*86400)
	f := caidFactors{
		Version: "3",
		ForCaid: map[string]string{
			"physicalMemory": "8056045568",
			"carrierInfo":    "--",
			"machine":        "iPhone16,1",
			"sysFileTime":    fmt.Sprintf("%d.562973", born+86400),
			"countryCode":    "CN",
			"bootSecTime":    strconv.FormatInt(born+172800, 10),
			"deviceName":     strings.ToUpper(deriveHex(id, "device_name", 16)),
			"timeZone":       "28800",
			"language":       "zh-Hans-CN",
			"disk":           "255515516928",
			"systemVersion":  iosOSVersion,
			"cpuNumber":      "6",
			"model":          "D83AP",
			"mobBirthTime":   fmt.Sprintf("%d.062328050", born),
		},
	}
	raw, _ := json.Marshal(f)
	return base64.StdEncoding.EncodeToString(raw)
}

func buildCaid(account *core.Account) string {
	androidID := account.AndroidID()
	ts := strconv.FormatInt(shareCaidGenerateTime, 10)
	doc := struct {
		Caid        []caidEntry `json:"caid"`
		FactorsData string      `json:"factors_data"`
		CaidValid   string      `json:"caid_valid"`
	}{
		Caid: []caidEntry{
			{
				Vendor:       "0",
				GenerateTime: ts,
				Caid: []caidSubID{{
					Version: "00",
					Caid: "00_" + strings.ToUpper(deriveHex(androidID, "caid_0a", 16)) +
						"_" + strings.ToUpper(deriveHex(androidID, "caid_0b", 16)),
				}},
			},
			{
				Vendor:       "1",
				GenerateTime: ts,
				Caid: []caidSubID{{
					Version: "20250325",
					Caid:    deriveHex(androidID, "caid_1", 16),
				}},
			},
			{
				Vendor:       "2",
				GenerateTime: ts,
				Caid: []caidSubID{{
					Version: "00",
					Caid:    strings.ToUpper(account.UUID()),
				}},
			},
		},
		FactorsData: buildFactorsData(account),
		CaidValid:   "1",
	}
	// 结构体是固定形状，字段全部为字符串，json.Marshal 在这里不会失败。
	raw, _ := json.Marshal(doc)
	return url.QueryEscape(string(raw))
}

// shareIOSUserAgent 构造 iOS 客户端的 User-Agent。
//
// 抓包真值（uniqueId 段即客户端 uniqueId，与 unique_id(58) 取值相同）：
//
//	Mozilla/5.0 (iPhone; CPU iPhone OS 18_7 like Mac OS X) AppleWebKit/605.1.15
//	(KHTML, like Gecko) Mobile/15E148 tieba/22.12.1.0 uniqueId/<uniqueId> skin/default
func shareIOSUserAgent(uniqueID string) string {
	return "Mozilla/5.0 (iPhone; CPU iPhone OS 18_7 like Mac OS X) " +
		"AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 " +
		"tieba/" + iosClientVersion + " uniqueId/" + uniqueID + " skin/default"
}

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

	req := &pb.ShareReqIdl{
		Data: &pb.ShareReqIdl_DataReq{
			Common: &commonpb.CommonReq{
				BDUSS:                 account.BDUSS(),
				XClientType:           1,
				XClientVersion:        iosClientVersion,
				XClientId:             account.ClientID(),
				XFrom:                 "appstore",
				Cuid:                  deriveHex(androidID, "cuid", 16),
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
				Idfv:                  account.UUID(),
				UserAgent:             shareIOSUserAgent(deriveUniqueID(androidID)),
				PersonalizedRecSwitch: 0,
				DeviceScore:           "0.800000",

				// 以下为分享接口特有的日志字段
				AppName:       shareAppName,
				LogSource:     last.API,
				LogSourceId:   last.LogID,
				XField_18:     last.Cost,
				XField_19:     last.Result,
				XField_23:     newTraceHex(),
				DeviceModel:   iosModel,
				Idfa:          shareIDFA,
				Caid:          buildCaid(account),
				UniqueId:      deriveUniqueID(androidID),
				XField_59:     deriveHex(androidID, "x_field_59", 20),
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
