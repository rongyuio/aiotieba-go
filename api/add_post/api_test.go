package addpost

import (
	"errors"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	pb "github.com/rongyuio/aiotieba-go/api/add_post/protobuf"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
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

func TestPackProtoFields(t *testing.T) {
	raw, err := PackProto(newAccount(t), "fname", 123, 456, "showname", "content")
	if err != nil {
		t.Fatalf("PackProto: %v", err)
	}
	req := &pb.AddPostReqIdl{}
	if err := proto.Unmarshal(raw, req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	data := req.GetData()
	common := data.GetCommon()

	if common.GetBDUSS() != strings.Repeat("b", 192) {
		t.Errorf("BDUSS = %q", common.GetBDUSS())
	}
	if common.GetXClientType() != 2 {
		t.Errorf("client type = %d, want 2", common.GetXClientType())
	}
	if common.GetXClientVersion() != "12.35.1.0" {
		t.Errorf("client version = %q", common.GetXClientVersion())
	}
	if common.GetXPhoneImei() != "000000000000000" {
		t.Errorf("phone imei = %q", common.GetXPhoneImei())
	}
	if common.GetXFrom() != "1008621x" {
		t.Errorf("from = %q", common.GetXFrom())
	}
	if common.GetModel() != "SM-G988N" {
		t.Errorf("model = %q", common.GetModel())
	}
	if common.GetNetType() != 1 {
		t.Errorf("net type = %d, want 1", common.GetNetType())
	}
	if common.GetPversion() != "1.0.3" {
		t.Errorf("pversion = %q", common.GetPversion())
	}
	if common.GetXOsVersion() != "9" {
		t.Errorf("os version = %q", common.GetXOsVersion())
	}
	if common.GetBrand() != "samsung" {
		t.Errorf("brand = %q", common.GetBrand())
	}
	if common.GetLegoLibVersion() != "3.0.0" {
		t.Errorf("lego lib version = %q", common.GetLegoLibVersion())
	}
	if common.GetSdkVer() != "2.34.0" {
		t.Errorf("sdk ver = %q", common.GetSdkVer())
	}
	if common.GetFrameworkVer() != "3340042" {
		t.Errorf("framework ver = %q", common.GetFrameworkVer())
	}
	if common.GetNawsGameVer() != "1038000" {
		t.Errorf("naws game ver = %q", common.GetNawsGameVer())
	}
	if common.GetScrW() != 720 || common.GetScrH() != 1280 {
		t.Errorf("screen = %dx%d, want 720x1280", common.GetScrW(), common.GetScrH())
	}
	if common.GetScrDip() != 1.5 {
		t.Errorf("screen dip = %v, want 1.5", common.GetScrDip())
	}
	if common.GetUserAgent() != "aiotieba/1.0.0" {
		t.Errorf("user agent = %q", common.GetUserAgent())
	}
	if common.GetDeviceScore() != "0.4" {
		t.Errorf("device score = %q", common.GetDeviceScore())
	}
	// The Python module assigns account.cuid_galaxy2 to both the cuid and
	// cuid_galaxy2 fields, and the lazily computed value must be non-empty.
	if common.GetCuid() == "" || common.GetCuidGalaxy2() == "" || common.GetCuid() != common.GetCuidGalaxy2() {
		t.Errorf("cuid = %q, cuid_galaxy2 = %q", common.GetCuid(), common.GetCuidGalaxy2())
	}
	if common.GetC3Aid() == "" {
		t.Error("c3_aid is empty")
	}
	if common.GetAndroidId() == "" {
		t.Error("android_id is empty")
	}

	if data.GetContent() != "content" {
		t.Errorf("content = %q", data.GetContent())
	}
	if data.GetFid() != "123" {
		t.Errorf("fid = %q", data.GetFid())
	}
	if data.GetTid() != "456" {
		t.Errorf("tid = %q", data.GetTid())
	}
	if data.GetKw() != "fname" {
		t.Errorf("kw = %q", data.GetKw())
	}
	if data.GetNameShow() != "showname" {
		t.Errorf("name_show = %q", data.GetNameShow())
	}
	if data.GetPostFrom() != "3" {
		t.Errorf("post_from = %q", data.GetPostFrom())
	}
	if data.GetAnonymous() != "1" {
		t.Errorf("anonymous = %q", data.GetAnonymous())
	}
	if data.GetFromFourmId() != "123" {
		t.Errorf("from_fourm_id = %q", data.GetFromFourmId())
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
