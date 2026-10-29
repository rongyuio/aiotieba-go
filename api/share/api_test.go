package share

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	pb "github.com/rongyuio/aiotieba-go/api/share/protobuf"
	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper/deviceid"
	commonpb "github.com/rongyuio/aiotieba-go/protobuf"
)

// testTid 是测试用的主题帖id。
const testTid = int64(10713887321)

func newAccount(t *testing.T) *core.Account {
	t.Helper()
	account, err := core.NewAccount(strings.Repeat("b", 192), "")
	if err != nil {
		t.Fatalf("NewAccount: %v", err)
	}
	return account
}

// testLastCall 是测试用的「上一次调用」记录，取值与抓包中 getShareInfo
// 那一次调用一致，用于验证 m_* 四元组被原样写入请求。
var testLastCall = core.CallRecord{
	API:    "c/f/forum/getShareInfo",
	LogID:  "2171117265",
	Cost:   "147.596955",
	Result: "110003",
}

// unpack 打包一次请求并解回来，便于逐字段断言。
func unpack(t *testing.T, account *core.Account, tid int64, shareChannel int32) *pb.ShareReqIdl_DataReq {
	t.Helper()
	raw, err := PackProto(account, tid, shareChannel, testLastCall)
	if err != nil {
		t.Fatalf("PackProto: %v", err)
	}
	req := &pb.ShareReqIdl{}
	if err := proto.Unmarshal(raw, req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return req.GetData()
}

func TestPackProtoFields(t *testing.T) {
	account := newAccount(t)
	common := unpack(t, account, testTid, 0).GetCommon()

	if common.GetBDUSS() != strings.Repeat("b", 192) {
		t.Errorf("BDUSS = %q", common.GetBDUSS())
	}
	// 设备档案对齐 iOS 客户端，取值来自真机抓包。
	if common.GetXClientType() != 1 {
		t.Errorf("client type = %d, want 1", common.GetXClientType())
	}
	if common.GetXClientVersion() != iosClientVersion {
		t.Errorf("client version = %q, want %q", common.GetXClientVersion(), iosClientVersion)
	}
	if common.GetModel() != iosModel {
		t.Errorf("model = %q, want %q", common.GetModel(), iosModel)
	}
	if common.GetBrand() != iosBrand {
		t.Errorf("brand = %q, want %q", common.GetBrand(), iosBrand)
	}
	if common.GetXOsVersion() != iosOSVersion {
		t.Errorf("os version = %q, want %q", common.GetXOsVersion(), iosOSVersion)
	}
	// UA 由账号派生（含 uniqueId 段），不是可硬编码的字面量。
	wantUA := deviceid.UserAgent(deviceid.UniqueID(account.AndroidID()), iosClientVersion, iosOSVersion)
	if got := common.GetUserAgent(); got != wantUA {
		t.Errorf("user agent = %q, want %q", got, wantUA)
	}
	// UA 的 uniqueId 段与 shoubai_cuid(58) 必须同值 —— 抓包中逐字相同。
	if !strings.Contains(common.GetUserAgent(), common.GetUniqueId()) {
		t.Errorf("UA 未包含 shoubai_cuid：UA=%q cuid=%q",
			common.GetUserAgent(), common.GetUniqueId())
	}
	// iOS 客户端不发的安卓专有字段，必须为空。
	for name, v := range map[string]string{
		"_phone_imei":        common.GetXPhoneImei(),
		"pversion":           common.GetPversion(),
		"cuid_galaxy2":       common.GetCuidGalaxy2(),
		"c3_aid":             common.GetC3Aid(),
		"android_id":         common.GetAndroidId(),
		"first_install_time": strconv.FormatInt(common.GetFirstInstallTime(), 10),
		"last_update_time":   strconv.FormatInt(common.GetLastUpdateTime(), 10),
	} {
		if v != "" && v != "0" {
			t.Errorf("iOS 客户端不该发 %s，得到 %q", name, v)
		}
	}

	// 分享接口特有的日志字段。
	if common.GetAppName() != "tieba" {
		t.Errorf("app name = %q", common.GetAppName())
	}
	if common.GetLogSource() != testLastCall.API {
		t.Errorf("log source = %q, want %q", common.GetLogSource(), testLastCall.API)
	}
	if common.GetLogSourceId() != testLastCall.LogID {
		t.Errorf("log source id = %q, want %q", common.GetLogSourceId(), testLastCall.LogID)
	}
	if common.GetXField_18() != testLastCall.Cost {
		t.Errorf("x_field_18 = %q, want %q", common.GetXField_18(), testLastCall.Cost)
	}
	if common.GetXField_19() != testLastCall.Result {
		t.Errorf("x_field_19 = %q, want %q", common.GetXField_19(), testLastCall.Result)
	}
	if common.GetIdfa() != "00000000-0000-0000-0000-000000000000" {
		t.Errorf("idfa = %q", common.GetIdfa())
	}
	if common.GetHybridVersion() != shareHybridVersion {
		t.Errorf("hybrid version = %q", common.GetHybridVersion())
	}
	if common.GetExpFlags() == "" {
		t.Error("exp_flags is empty")
	}
}

// TestPackProtoDeviceModelMatchesModel 钉死 model(9) 与 device_model(27) 同源这条不变量。
// 抓包中 iOS 客户端这两个字段取值相同。
func TestPackProtoDeviceModelMatchesModel(t *testing.T) {
	common := unpack(t, newAccount(t), testTid, 0).GetCommon()
	if common.GetDeviceModel() != common.GetModel() {
		t.Errorf("device_model = %q, model = %q, want equal", common.GetDeviceModel(), common.GetModel())
	}
}

// TestPackProtoMFieldsComeFromLastCall 钉死 m_* 四元组来自调用方传入的
// 上一次调用记录，而不是任何硬编码常量。
//
// 背景：真实客户端的 thread/share 请求申报的 m_api / m_logid / m_cost / m_result，
// 与紧邻其前的 getShareInfo 调用的路径、logid、耗时、错误码逐字相同。
// 早期实现把这四个值硬编码（log_source_id 还错填成 tid），服务端能看出链条是伪造的。
func TestPackProtoMFieldsComeFromLastCall(t *testing.T) {
	account := newAccount(t)

	last := core.CallRecord{
		API:    "c/s/otherEndpoint?cmd=123&format=protobuf",
		LogID:  "9999999999",
		Cost:   "1.234567",
		Result: "0",
	}
	raw, err := PackProto(account, testTid, 0, last)
	if err != nil {
		t.Fatalf("PackProto: %v", err)
	}
	req := &pb.ShareReqIdl{}
	if err := proto.Unmarshal(raw, req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	common := req.GetData().GetCommon()

	if got := common.GetLogSource(); got != last.API {
		t.Errorf("log source = %q, want %q", got, last.API)
	}
	if got := common.GetLogSourceId(); got != last.LogID {
		t.Errorf("log source id = %q, want %q", got, last.LogID)
	}
	if got := common.GetXField_18(); got != last.Cost {
		t.Errorf("x_field_18 = %q, want %q", got, last.Cost)
	}
	if got := common.GetXField_19(); got != last.Result {
		t.Errorf("x_field_19 = %q, want %q", got, last.Result)
	}
	// log_source_id 必须不再是 tid —— 这是早期实现的具体缺陷。
	if common.GetLogSourceId() == strconv.FormatInt(testTid, 10) {
		t.Errorf("log source id 仍等于 tid %d，链条字段被误用", testTid)
	}
}

// TestPackProtoMFieldsEmptyWithoutLastCall 验证没有前置调用时四个字段为空，
// 而不是回退到编造的常量。
func TestPackProtoMFieldsEmptyWithoutLastCall(t *testing.T) {
	raw, err := PackProto(newAccount(t), testTid, 0, core.CallRecord{})
	if err != nil {
		t.Fatalf("PackProto: %v", err)
	}
	req := &pb.ShareReqIdl{}
	if err := proto.Unmarshal(raw, req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	common := req.GetData().GetCommon()
	if common.GetLogSource() != "" || common.GetLogSourceId() != "" ||
		common.GetXField_18() != "" || common.GetXField_19() != "" {
		t.Errorf("零值 CallRecord 应产生空 m_* 字段，得到 %q/%q/%q/%q",
			common.GetLogSource(), common.GetLogSourceId(),
			common.GetXField_18(), common.GetXField_19())
	}
}

// TestPackProtoDerivedFieldsAreStable 抓包中这些字段跨请求完全相同，属于设备级常量。
func TestPackProtoDerivedFieldsAreStable(t *testing.T) {
	account := newAccount(t)
	first := unpack(t, account, testTid, 0).GetCommon()
	second := unpack(t, account, testTid, 0).GetCommon()

	pairs := []struct {
		name          string
		first, second string
	}{
		{"x_field_59", first.GetXField_59(), second.GetXField_59()},
		{"unique_id", first.GetUniqueId(), second.GetUniqueId()},
		{"caid", first.GetCaid(), second.GetCaid()},
	}
	for _, p := range pairs {
		if p.first == "" {
			t.Errorf("%s is empty", p.name)
		}
		if p.first != p.second {
			t.Errorf("%s = %q then %q, want stable", p.name, p.first, p.second)
		}
	}
}

// TestPackProtoTraceFieldVaries 抓包证明 x_field_23 每次请求都不同，是请求级标识。
// 若将来有人把它改成设备级常量，这个测试会失败。
func TestPackProtoTraceFieldVaries(t *testing.T) {
	account := newAccount(t)
	first := unpack(t, account, testTid, 0).GetCommon().GetXField_23()
	second := unpack(t, account, testTid, 0).GetCommon().GetXField_23()
	if first == second {
		t.Errorf("x_field_23 = %q both times, want a per-request value", first)
	}
}

// TestPackProtoDistinctAccounts 派生字段必须按账号隔离，不能退回全局常量。
func TestPackProtoDistinctAccounts(t *testing.T) {
	other, err := core.NewAccount(strings.Repeat("c", 192), "")
	if err != nil {
		t.Fatalf("NewAccount: %v", err)
	}
	a := unpack(t, newAccount(t), testTid, 0).GetCommon()
	b := unpack(t, other, testTid, 0).GetCommon()

	if a.GetXField_59() == b.GetXField_59() {
		t.Errorf("x_field_59 = %q for both accounts", a.GetXField_59())
	}
	if a.GetUniqueId() == b.GetUniqueId() {
		t.Errorf("unique_id = %q for both accounts", a.GetUniqueId())
	}
	if a.GetCaid() == b.GetCaid() {
		t.Errorf("caid = %q for both accounts", a.GetCaid())
	}
}

// TestPackProtoDerivedShape 钉死各派生字段的形状，形状错了服务端更容易识别出异常。
func TestPackProtoDerivedShape(t *testing.T) {
	common := unpack(t, newAccount(t), testTid, 0).GetCommon()

	trace := common.GetXField_23()
	if len(trace) != 32 {
		t.Errorf("x_field_23 length = %d, want 32", len(trace))
	}
	if trace != strings.ToUpper(trace) {
		t.Errorf("x_field_23 = %q, want uppercase", trace)
	}
	if _, err := hex.DecodeString(trace); err != nil {
		t.Errorf("x_field_23 = %q is not hex: %v", trace, err)
	}

	f59 := common.GetXField_59()
	if len(f59) != 40 {
		t.Errorf("x_field_59 length = %d, want 40", len(f59))
	}
	if _, err := hex.DecodeString(f59); err != nil {
		t.Errorf("x_field_59 = %q is not hex: %v", f59, err)
	}

	if got := len(common.GetUniqueId()); got != 51 {
		t.Errorf("unique_id length = %d, want 51", got)
	}
	if got := len(common.GetIdfa()); got != 36 {
		t.Errorf("idfa length = %d, want 36", got)
	}

	// caid 是 URL 编码的 JSON：3 个渠道 + caid_valid=1，且不带 factors_data。
	decoded, err := url.QueryUnescape(common.GetCaid())
	if err != nil {
		t.Fatalf("caid is not URL-encoded: %v", err)
	}
	doc := struct {
		Caid []struct {
			Vendor string `json:"vendor"`
			Time   string `json:"generateTime"`
		} `json:"caid"`
		FactorsData string `json:"factors_data"`
		CaidValid   string `json:"caid_valid"`
	}{}
	if err := json.Unmarshal([]byte(decoded), &doc); err != nil {
		t.Fatalf("caid is not JSON: %v", err)
	}
	if len(doc.Caid) != 3 {
		t.Errorf("caid vendors = %d, want 3", len(doc.Caid))
	}
	if doc.CaidValid != "1" {
		t.Errorf("caid_valid = %q, want 1", doc.CaidValid)
	}
	// 真机 diac(46) 为 1435 字节，其中 factors_data 占大头。
	// 省略它会让整个字段只有 628 字节，与真实客户端差距明显。
	if doc.FactorsData == "" {
		t.Fatal("caid 缺少 factors_data")
	}
	factors, err := base64.StdEncoding.DecodeString(doc.FactorsData)
	if err != nil {
		t.Fatalf("factors_data 不是 base64: %v", err)
	}
	var f struct {
		Version string            `json:"factorsVersion"`
		ForCaid map[string]string `json:"factorsForCaid"`
	}
	if err := json.Unmarshal(factors, &f); err != nil {
		t.Fatalf("factors_data 不是 JSON: %v", err)
	}
	// 抓包真值里 factorsForCaid 有 14 项。
	if len(f.ForCaid) != 14 {
		t.Errorf("factorsForCaid 项数 = %d, want 14", len(f.ForCaid))
	}
	if f.ForCaid["machine"] != "iPhone16,1" {
		t.Errorf("machine = %q, want iPhone16,1", f.ForCaid["machine"])
	}
	// 机型相关取值必须与设备档案一致，不能一处 iOS 一处安卓。
	if f.ForCaid["systemVersion"] != iosOSVersion {
		t.Errorf("systemVersion = %q, want %q", f.ForCaid["systemVersion"], iosOSVersion)
	}
	// deviceName 是设备级标识，必须随账号变化。
	other := newAccount(t)
	if got := decodeFactorsDeviceName(t, deviceid.Caid(other.AndroidID(), other.UUID(), iosOSVersion)); got == f.ForCaid["deviceName"] {
		t.Errorf("两个账号的 deviceName 相同：%q", got)
	}
}

// decodeFactorsDeviceName 从 caid 里取出 factors_data 的 deviceName，便于跨账号比对。
func decodeFactorsDeviceName(t *testing.T, caid string) string {
	t.Helper()
	decoded, err := url.QueryUnescape(caid)
	if err != nil {
		t.Fatalf("caid 不是 URL 编码: %v", err)
	}
	var doc struct {
		FactorsData string `json:"factors_data"`
	}
	if err := json.Unmarshal([]byte(decoded), &doc); err != nil {
		t.Fatalf("caid 不是 JSON: %v", err)
	}
	raw, err := base64.StdEncoding.DecodeString(doc.FactorsData)
	if err != nil {
		t.Fatalf("factors_data 不是 base64: %v", err)
	}
	var f struct {
		ForCaid map[string]string `json:"factorsForCaid"`
	}
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("factors_data 不是 JSON: %v", err)
	}
	return f.ForCaid["deviceName"]
}

// TestPackProtoLogIDIsTid 钉死 log_id(2) 承载的是被分享帖子的 tid。
//
// 这是分享接口唯一被服务端校验的字段：实测把 log_id 换成另一个帖子的 tid，
// 计数就涨到那个帖子；换成不存在的值则谁都不涨。
// 旧实现生成随机 log_id，等于上报一个不存在的帖子，所以服务端回 errno=0 却从不记账。
func TestPackProtoLogIDIsTid(t *testing.T) {
	account := newAccount(t)
	for _, tid := range []int64{8737665232, 2129471057, 11108062649} {
		if got := unpack(t, account, tid, 0).GetLogId(); got != tid {
			t.Errorf("tid %d: log_id = %d, want %d", tid, got, tid)
		}
	}
}

func TestPackProtoShareChannel(t *testing.T) {
	account := newAccount(t)
	for _, ch := range []int32{0, 3} {
		if got := unpack(t, account, testTid, ch).GetShareChannel(); got != ch {
			t.Errorf("share channel = %d, want %d", got, ch)
		}
	}
}

// TestParseBodySuccess 用真实抓包的响应体做黄金向量。
//
// 原始响应是 Error{errorno:0, errmsg:"", f3:""} + DataRes{log_id:10514196161}。
// Error 还有一个未在 Error.proto 中定义的 f3 字段，会落进 unknownFields。
func TestParseBodySuccess(t *testing.T) {
	raw, err := hex.DecodeString("0a06080012001a00120608c1cdc79527")
	if err != nil {
		t.Fatalf("decode golden vector: %v", err)
	}
	if err := ParseBody(raw); err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
}

func TestParseBodyEmpty(t *testing.T) {
	if err := ParseBody(nil); err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
}

func TestParseBodyServerError(t *testing.T) {
	res := &pb.ShareResIdl{Error: &commonpb.Error{Errorno: 340006, Errmsg: "boom"}}
	raw, err := proto.Marshal(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	err = ParseBody(raw)
	var serverErr *exception.TiebaServerError
	if !errors.As(err, &serverErr) {
		t.Fatalf("error = %v (%T), want *exception.TiebaServerError", err, err)
	}
	if serverErr.Code != 340006 || serverErr.Msg != "boom" {
		t.Errorf("server error = %+v", serverErr)
	}
}

func TestRequestURL(t *testing.T) {
	u := RequestURL()
	if u.Scheme != "https" || u.Host != "tiebac.baidu.com" || u.Path != "/c/c/thread/share" {
		t.Errorf("url = %s", u)
	}
	if u.RawQuery != "cmd=309480&format=protobuf" {
		t.Errorf("query = %q, want cmd=309480&format=protobuf", u.RawQuery)
	}
}

// TestUserAgentMatchesHTTPHeader 钉死请求体里的 UA 与 HTTP 头里的 UA 逐字相同。
func TestUserAgentMatchesHTTPHeader(t *testing.T) {
	account := newAccount(t)
	got := unpack(t, account, testTid, 0).GetCommon().GetUserAgent()
	want := deviceid.AppUserAgent(account.AndroidID(), consts.IOSVersion)
	if got != want {
		t.Errorf("body UA = %q, header UA = %q, want equal", got, want)
	}
}
