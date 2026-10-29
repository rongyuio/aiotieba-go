package addpost

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

	pb "github.com/rongyuio/aiotieba-go/api/add_post/protobuf"
	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper/deviceid"
	commonpb "github.com/rongyuio/aiotieba-go/protobuf"
)

func newAccount(t *testing.T) *core.Account {
	t.Helper()
	account, err := core.NewAccount(strings.Repeat("b", 192), "")
	if err != nil {
		t.Fatalf("NewAccount: %v", err)
	}
	return account
}

// testLastCall 模拟紧邻 add_post 之前的一次 app 调用，取值与抓包一致。
var testLastCall = core.CallRecord{
	API:    "c/f/sprite/queryMeme",
	LogID:  "48236398",
	Cost:   "313.148975",
	Result: "0",
}

// unpack 打包一次请求并解回来，便于逐字段断言。
func unpack(t *testing.T, account *core.Account, tid int64) *pb.AddPostReqIdl_DataReq {
	t.Helper()
	raw, err := PackProto(account, "fname", 123, tid, "showname", "content", testLastCall)
	if err != nil {
		t.Fatalf("PackProto: %v", err)
	}
	req := &pb.AddPostReqIdl{}
	if err := proto.Unmarshal(raw, req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return req.GetData()
}

func TestPackProtoFields(t *testing.T) {
	account := newAccount(t)
	first := unpack(t, account, 456)
	common := first.GetCommon()

	if common.GetBDUSS() != strings.Repeat("b", 192) {
		t.Errorf("BDUSS = %q", common.GetBDUSS())
	}
	// 设备档案对齐 iOS 客户端，取值来自真机抓包。
	if common.GetXClientType() != 1 {
		t.Errorf("client type = %d, want 1", common.GetXClientType())
	}
	if common.GetXClientVersion() != consts.IOSVersion {
		t.Errorf("client version = %q, want %q", common.GetXClientVersion(), consts.IOSVersion)
	}
	if common.GetXFrom() != "appstore" {
		t.Errorf("from = %q", common.GetXFrom())
	}
	if common.GetModel() != "iPhone 15 Pro" || common.GetBrand() != "iPhone" {
		t.Errorf("model/brand = %q/%q", common.GetModel(), common.GetBrand())
	}
	if common.GetXOsVersion() != "18.7.2" {
		t.Errorf("os version = %q", common.GetXOsVersion())
	}
	if common.GetDeviceModel() != common.GetModel() {
		t.Errorf("device_model = %q, model = %q, want equal", common.GetDeviceModel(), common.GetModel())
	}
	if common.GetNetType() != 1 {
		t.Errorf("net type = %d, want 1", common.GetNetType())
	}
	if common.GetLegoLibVersion() != "3.0.0" {
		t.Errorf("lego lib version = %q", common.GetLegoLibVersion())
	}
	if common.GetSdkVer() != "3.54.0" {
		t.Errorf("sdk ver = %q", common.GetSdkVer())
	}
	if common.GetFrameworkVer() != "4.520.1" {
		t.Errorf("framework ver = %q", common.GetFrameworkVer())
	}
	if common.GetNawsGameVer() != "1.21.0" {
		t.Errorf("naws game ver = %q", common.GetNawsGameVer())
	}
	if common.GetScrW() != 1179 || common.GetScrH() != 2556 {
		t.Errorf("screen = %dx%d, want 1179x2556", common.GetScrW(), common.GetScrH())
	}
	if common.GetScrDip() != 3.0 {
		t.Errorf("screen dip = %v, want 3.0", common.GetScrDip())
	}
	if common.GetQType() != 100 {
		t.Errorf("q_type = %d, want 100", common.GetQType())
	}
	if common.GetDeviceScore() != "0.800000" {
		t.Errorf("device score = %q", common.GetDeviceScore())
	}
	if common.GetPersonalizedRecSwitch() != 0 {
		t.Errorf("personalized_rec_switch = %d, want 0", common.GetPersonalizedRecSwitch())
	}
	if got := common.GetEventDay(); len(got) != 8 {
		t.Errorf("event_day = %q, want YYYYMMDD", got)
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

	// 请求体字段
	if first.GetContent() != "content" {
		t.Errorf("content = %q", first.GetContent())
	}
	if first.GetFid() != "123" || first.GetTid() != "456" {
		t.Errorf("fid/tid = %q/%q", first.GetFid(), first.GetTid())
	}
	if first.GetKw() != "fname" || first.GetNameShow() != "showname" {
		t.Errorf("kw/name_show = %q/%q", first.GetKw(), first.GetNameShow())
	}
	if first.GetAnonymous() != "0" {
		t.Errorf("anonymous = %q, want 0", first.GetAnonymous())
	}
	if first.GetPostFrom() != "pb_reply" {
		t.Errorf("post_from = %q, want pb_reply", first.GetPostFrom())
	}
	// 抓包中 DataReq.tbs(4) 与 CommonReq.tbs(11) 逐字相同。
	if first.GetTbs() != common.GetTbs() {
		t.Errorf("DataReq.tbs = %q, common.tbs = %q, want equal", first.GetTbs(), common.GetTbs())
	}
}

// TestPackProtoFingerprintIsDerived 钉死设备标识由账号派生：
// 同账号恒定、不同账号互不相同，且不是抓包里的那台真机。
func TestPackProtoFingerprintIsDerived(t *testing.T) {
	account := newAccount(t)
	a := unpack(t, account, 456).GetCommon()
	b := unpack(t, account, 456).GetCommon()

	for name, pair := range map[string][2]string{
		"cuid":       {a.GetCuid(), b.GetCuid()},
		"caid":       {a.GetCaid(), b.GetCaid()},
		"unique_id":  {a.GetUniqueId(), b.GetUniqueId()},
		"x_field_59": {a.GetXField_59(), b.GetXField_59()},
	} {
		if pair[0] == "" {
			t.Errorf("%s is empty", name)
		}
		if pair[0] != pair[1] {
			t.Errorf("%s = %q then %q, want stable", name, pair[0], pair[1])
		}
	}
	// x_field_23 是请求级标识，每次不同。
	if a.GetXField_23() == b.GetXField_23() {
		t.Errorf("x_field_23 = %q both times, want a per-request value", a.GetXField_23())
	}

	other, err := core.NewAccount(strings.Repeat("c", 192), "")
	if err != nil {
		t.Fatalf("NewAccount: %v", err)
	}
	c := unpack(t, other, 456).GetCommon()
	if a.GetCuid() == c.GetCuid() {
		t.Errorf("cuid = %q for both accounts", a.GetCuid())
	}
	if a.GetCaid() == c.GetCaid() {
		t.Errorf("caid = %q for both accounts", a.GetCaid())
	}
	if a.GetUniqueId() == c.GetUniqueId() {
		t.Errorf("unique_id = %q for both accounts", a.GetUniqueId())
	}
	if a.GetXField_59() == c.GetXField_59() {
		t.Errorf("x_field_59 = %q for both accounts", a.GetXField_59())
	}
}

// TestPackProtoDerivedShape 钉死各字段形状，形状不对更容易被识别成异常流量。
func TestPackProtoDerivedShape(t *testing.T) {
	common := unpack(t, newAccount(t), 456).GetCommon()

	if got := len(common.GetCuid()); got != 32 {
		t.Errorf("cuid length = %d, want 32", got)
	}
	if _, err := hex.DecodeString(common.GetCuid()); err != nil {
		t.Errorf("cuid = %q is not hex: %v", common.GetCuid(), err)
	}
	if got := len(common.GetUniqueId()); got != 51 {
		t.Errorf("unique_id length = %d, want 51", got)
	}
	if got := len(common.GetXField_59()); got != 40 {
		t.Errorf("x_field_59 length = %d, want 40", got)
	}
	if got := len(common.GetXField_23()); got != 32 {
		t.Errorf("x_field_23 length = %d, want 32", got)
	}
	if common.GetIdfa() != "00000000-0000-0000-0000-000000000000" {
		t.Errorf("idfa = %q", common.GetIdfa())
	}
	if got := len(common.GetIdfv()); got != 36 {
		t.Errorf("idfv length = %d, want 36", got)
	}
	// UA 的 uniqueId 段必须与 unique_id(58) 同值。
	if !strings.Contains(common.GetUserAgent(), common.GetUniqueId()) {
		t.Errorf("UA 未包含 uniqueId：UA=%q unique_id=%q",
			common.GetUserAgent(), common.GetUniqueId())
	}
	if !strings.Contains(common.GetUserAgent(), "tieba/"+consts.IOSVersion) {
		t.Errorf("UA 未包含客户端版本：%q", common.GetUserAgent())
	}
	// caid 是 URL 编码的 JSON，含 3 个渠道。
	decoded, err := url.QueryUnescape(common.GetCaid())
	if err != nil {
		t.Fatalf("caid is not URL-encoded: %v", err)
	}
	var doc struct {
		Caid []struct {
			Vendor string `json:"vendor"`
		} `json:"caid"`
		CaidValid string `json:"caid_valid"`
	}
	if err := json.Unmarshal([]byte(decoded), &doc); err != nil {
		t.Fatalf("caid is not JSON: %v", err)
	}
	if len(doc.Caid) != 3 || doc.CaidValid != "1" {
		t.Errorf("caid vendors = %d, caid_valid = %q", len(doc.Caid), doc.CaidValid)
	}
}

// TestPackProtoMFieldsComeFromLastCall 钉死 m_* 四元组来自调用方传入的上一次调用记录。
func TestPackProtoMFiledsComeFromLastCall(t *testing.T) {
	common := unpack(t, newAccount(t), 456).GetCommon()
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

	// 零值记录必须产生空字段，而不是回退到编造的常量。
	raw, err := PackProto(newAccount(t), "fname", 123, 456, "showname", "content", core.CallRecord{})
	if err != nil {
		t.Fatalf("PackProto: %v", err)
	}
	req := &pb.AddPostReqIdl{}
	if err := proto.Unmarshal(raw, req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	c := req.GetData().GetCommon()
	if c.GetLogSource() != "" || c.GetLogSourceId() != "" ||
		c.GetXField_18() != "" || c.GetXField_19() != "" {
		t.Errorf("零值 CallRecord 应产生空 m_* 字段，得到 %q/%q/%q/%q",
			c.GetLogSource(), c.GetLogSourceId(), c.GetXField_18(), c.GetXField_19())
	}
}

// TestPackProtoTabTrace 钉死 x_field_91 是 base64 的页面栈 JSON，且携带当前 tid。
func TestPackProtoTabTrace(t *testing.T) {
	const tid = int64(10709435542)
	trace := unpack(t, newAccount(t), tid).GetCommon().GetXField_91()
	if trace == "" {
		t.Fatal("x_field_91 is empty")
	}
	raw, err := base64.StdEncoding.DecodeString(trace)
	if err != nil {
		t.Fatalf("x_field_91 is not base64: %v", err)
	}
	var items []struct {
		P string `json:"p"`
		V string `json:"v"`
	}
	if err := json.Unmarshal(raw, &items); err != nil {
		t.Fatalf("x_field_91 is not JSON: %v", err)
	}
	if len(items) != 4 {
		t.Fatalf("page stack has %d items, want 4", len(items))
	}
	if got := items[3].V; got != strconv.FormatInt(tid, 10) {
		t.Errorf("page stack tid = %q, want %d", got, tid)
	}
}

func TestParseBodySuccess(t *testing.T) {
	res := &pb.AddPostResIdl{
		Data: &pb.AddPostResIdl_DataRes{
			Info: &pb.AddPostResIdl_DataRes_PostAntiInfo{NeedVcode: "0"},
		},
	}
	raw, err := proto.Marshal(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := ParseBody(raw); err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
}

func TestParseBodyServerError(t *testing.T) {
	res := &pb.AddPostResIdl{Error: &commonpb.Error{Errorno: 340006, Errmsg: "boom"}}
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

func TestParseBodyNeedVcode(t *testing.T) {
	res := &pb.AddPostResIdl{
		Data: &pb.AddPostResIdl_DataRes{
			Info: &pb.AddPostResIdl_DataRes_PostAntiInfo{NeedVcode: "1"},
		},
	}
	raw, err := proto.Marshal(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	err = ParseBody(raw)
	var valueErr *exception.TiebaValueError
	if !errors.As(err, &valueErr) {
		t.Fatalf("error = %v (%T), want *exception.TiebaValueError", err, err)
	}
	if valueErr.Msg != "Need verify code" {
		t.Errorf("value error msg = %q", valueErr.Msg)
	}
}

func TestRequestURL(t *testing.T) {
	u := RequestURL()
	if u.Scheme != "https" || u.Host != "tiebac.baidu.com" || u.Path != "/c/c/post/add" {
		t.Errorf("url = %s", u)
	}
	if u.RawQuery != "cmd=309731" {
		t.Errorf("query = %q, want cmd=309731", u.RawQuery)
	}
}

// TestUserAgentMatchesHTTPHeader 钉死请求体里的 UA 与 HTTP 头里的 UA 逐字相同。
//
// core 用 consts.IOSVersion 与同一套派生逻辑生成请求头 UA，这里用同一个函数
// 复算一遍：两处一旦不同源，真机上不可能出现的「头说一台设备、body 说另一台」就回来了。
func TestUserAgentMatchesHTTPHeader(t *testing.T) {
	account := newAccount(t)
	got := unpack(t, account, 456).GetCommon().GetUserAgent()
	want := deviceid.AppUserAgent(account.AndroidID(), consts.IOSVersion)
	if got != want {
		t.Errorf("body UA = %q, header UA = %q, want equal", got, want)
	}
}
