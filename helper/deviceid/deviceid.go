// Package deviceid 由账号派生出客户端上报所需的设备标识。
//
// 贴吧的 app 接口会在 CommonReq 里携带一整套设备指纹（caid、unique_id、
// x_field_59 等）。这些值在真机上是设备级常量，抓包可以拿到，但**不能照抄**：
// 抄下来等于所有使用者共用同一个设备指纹。
//
// 这里的做法是从账号持久化的 android_id 派生：同一账号每次上报都相同，
// 不同账号互不相同，且不包含任何真实设备的信息。
package deviceid

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/url"
	"strconv"
	"strings"
)

// 与具体设备无关的客户端常量。
const (
	// iOS 设备档案。机型与系统版本是公开的客户端取值，硬编码不产生指纹关联。
	IOSModel     = "iPhone 15 Pro"
	IOSBrand     = "iPhone"
	IOSOSVersion = "18.7.2"

	// ZeroIDFA 对应 idfa(45)。idfa 是 iOS 概念，全 0 UUID 表示「不可用」，
	// 抓包中即为全 0。
	ZeroIDFA = "00000000-0000-0000-0000-000000000000"

	// HybridVersion 对应 hybrid_version(88)，客户端内嵌 webview 的版本号。
	HybridVersion = "hybrid-main-pb_1.0.397.1"

	// ExpFlags 对应 exp_flags(89)，实验开关列表。
	// 实验开关按定义就是全量用户共享的灰度标记，硬编码不产生指纹关联。
	ExpFlags = "22.8_pb_floor_reply_rebuild_b^22.8_pb_floor_reply_rebuild_a,22.8_pb_floor_reply_rebuild_origin"

	// caidGenerateTime 是 caid 各渠道的 generateTime，抓包中的固定值。
	caidGenerateTime = 1790071384
)

// NewTraceHex 生成 x_field_23(23)，抓包中是每请求不同的 32 位 hex。
func NewTraceHex() string {
	return strings.ToUpper(fmt.Sprintf("%016x%016x", rand.Uint64(), rand.Uint64()))
}

// DeriveHex 由设备信息派生一个稳定且设备唯一的 hex 串。
//
// caid、unique_id、x_field_59 在抓包里跨请求完全一致，属于设备级常量；
// 这里统一由 android_id 派生，保证同一账号恒定、不同账号互不相同。
func DeriveHex(androidID, label string, n int) string {
	sum := sha256.Sum256([]byte(label + "|" + androidID))
	return hex.EncodeToString(sum[:n])
}

// Cuid 对应 cuid(7)，抓包中是 32 位大写 hex。
func Cuid(androidID string) string {
	return strings.ToUpper(DeriveHex(androidID, "cuid", 16))
}

// IDFV 对应 idfv(60)。抓包中是全大写 UUID，与账号的 uuid 同值。
//
// uuid 由 Account 在首次使用时随机生成并持久化，本身就是账号级标识，
// 这里只做大小写对齐，不引入新的熵。
func IDFV(uuid string) string {
	return strings.ToUpper(uuid)
}

// UniqueID 对应 unique_id(58)。
//
// 抓包中是 51 个大写字母数字字符，并且被同步写进 user_agent 的 uniqueId/ 段。
// getShareInfo 的抓包显示它在 form 参数里叫 shoubai_cuid，取值逐字相同。
// 这里取 SHA-256 前 26 字节的 hex 再截到 51 位，长度对齐、字符集是其子集。
func UniqueID(androidID string) string {
	return strings.ToUpper(DeriveHex(androidID, "unique_id", 26))[:51]
}

// DeriveEpoch 由 android_id 派生一个落在 [base, base+span) 内的秒级时间戳。
//
// 用于真机上固定不变的时间类字段（安装时间、首次活跃时间等）。这类字段若填
// now 减固定天数，会随每次请求一起漂移，比一个稳定的旧时间戳更像合成流量。
func DeriveEpoch(androidID, label string, base, span int64) int64 {
	sum := sha256.Sum256([]byte(label + "|" + androidID))
	var v uint64
	for i := 0; i < 8; i++ {
		v = v<<8 | uint64(sum[i])
	}
	if span <= 0 {
		return base
	}
	return base + int64(v%uint64(span))
}

// IOSOSTag 把 18.7.2 渲染成 UA 里的 18_7：iOS 的 UA 只写到主次版本号。
func IOSOSTag(osVersion string) string {
	parts := strings.Split(osVersion, ".")
	if len(parts) > 2 {
		parts = parts[:2]
	}
	return strings.Join(parts, "_")
}

// UserAgent 构造与 uniqueID 配套的 UA。
//
// 抓包真值：
//
//	Mozilla/5.0 (iPhone; CPU iPhone OS 18_7 like Mac OS X) AppleWebKit/605.1.15
//	(KHTML, like Gecko) Mobile/15E148 tieba/<版本> uniqueId/<uniqueID> skin/default
func UserAgent(uniqueID, clientVersion, osVersion string) string {
	return "Mozilla/5.0 (iPhone; CPU iPhone OS " + IOSOSTag(osVersion) + " like Mac OS X) " +
		"AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 " +
		"tieba/" + clientVersion + " uniqueId/" + uniqueID + " skin/default"
}

// AppUserAgent 是 app / app_proto 会话的默认 UA。
//
// 客户端版本取请求体里申报的 _client_version，uniqueId 与 CommonReq.unique_id(58)
// 同源 —— 真机的这两处逐字相同，UA 里少一段或多一段都是破绽。
//
// 不把库名和库版本写进 UA：那等于向服务端声明「这不是官方客户端」，
// 而且所有使用者共用同一串指纹。
func AppUserAgent(androidID, clientVersion string) string {
	return UserAgent(UniqueID(androidID), clientVersion, IOSOSVersion)
}

// WebUserAgent 是 web 会话的 UA：app 内置 webview 的形态，不带 tieba 版本段。
func WebUserAgent() string {
	return "Mozilla/5.0 (iPhone; CPU iPhone OS " + IOSOSTag(IOSOSVersion) + " like Mac OS X) " +
		"AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148"
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

// caidFactors 是 caid 里 factors_data 的设备画像，对应抓包中那段 base64 JSON。
//
// 其中 machine / model 是机型码，属于客户端常量，硬编码无熵；
// deviceName、各时间戳则由账号派生。
type caidFactors struct {
	Version string            `json:"factorsVersion"`
	ForCaid map[string]string `json:"factorsForCaid"`
}

// buildFactorsData 构造 factors_data，返回 base64 编码后的字符串。
//
// 省略该键会让 caid 比真实客户端短一大截（抓包中 1435 字节，省略后约 628），
// 所以这里按同一结构补齐，具体数值全部由 android_id 派生。
func buildFactorsData(androidID, osVersion string) string {
	base := int64(1672531200) // 2023-01-01
	offset := int64(0)
	for i := 0; i < 6 && i < len(androidID); i++ {
		offset = offset*256 + int64(androidID[i])
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
			"deviceName":     strings.ToUpper(DeriveHex(androidID, "device_name", 16)),
			"timeZone":       "28800",
			"language":       "zh-Hans-CN",
			"disk":           "255515516928",
			"systemVersion":  osVersion,
			"cpuNumber":      "6",
			"model":          "D83AP",
			"mobBirthTime":   fmt.Sprintf("%d.062328050", born),
		},
	}
	raw, _ := json.Marshal(f)
	return base64.StdEncoding.EncodeToString(raw)
}

// Caid 构造 caid(46) 的值，返回 URL 编码后的 JSON。
//
// 抓包中的结构为 3 个渠道（vendor 0/1/2）的嵌套对象，末尾带 caid_valid。
// 所有 ID 都由 android_id / uuid 派生，保证同一账号恒定、不同账号互不相同。
func Caid(androidID, uuid, osVersion string) string {
	ts := strconv.FormatInt(caidGenerateTime, 10)
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
					Caid: "00_" + strings.ToUpper(DeriveHex(androidID, "caid_0a", 16)) +
						"_" + strings.ToUpper(DeriveHex(androidID, "caid_0b", 16)),
				}},
			},
			{
				Vendor:       "1",
				GenerateTime: ts,
				Caid: []caidSubID{{
					Version: "20250325",
					Caid:    DeriveHex(androidID, "caid_1", 16),
				}},
			},
			{
				Vendor:       "2",
				GenerateTime: ts,
				Caid: []caidSubID{{
					Version: "00",
					Caid:    strings.ToUpper(uuid),
				}},
			},
		},
		FactorsData: buildFactorsData(androidID, osVersion),
		CaidValid:   "1",
	}
	// 结构体是固定形状，字段全部为字符串，json.Marshal 在这里不会失败。
	raw, _ := json.Marshal(doc)
	return url.QueryEscape(string(raw))
}
