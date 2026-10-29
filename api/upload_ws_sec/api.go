package uploadwssec

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/upload_ws_sec/protobuf"
)

// CMD 是 init_websocket 的 websocket 命令字。
const CMD = 1001

// publicKeyBase64 是用于加密 websocket 密钥的 DER（base64 编码）RSA 公钥，对应 Python 模块的 PUBLIC_KEY。
const publicKeyBase64 = "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwQpwBZxXJV/JVRF/uNfyMSdu7YWwRNLM8+2xbniGp2iIQHOikPpTYQjlQgMi1uvq1kZpJ32rHo3hkwjy2l0lFwr3u4Hk2Wk7vnsqYQjAlYlK0TCzjpmiI+OiPOUNVtbWHQiLiVqFtzvpvi4AU7C1iKGvc/4IS45WjHxeScHhnZZ7njS4S1UgNP/GflRIbzgbBhyZ9kEW5/OO5YfG1fy6r4KSlDJw4o/mw5XhftyIpL+5ZBVBC6E1EIiP/dd9AbK62VV1PByfPMHMixpxI3GM2qwcmFsXcCcgvUXJBa9k6zP8dDQ3csCM2QNT+CQAOxthjtp/TFWaD7MzOdsIYb3THwIDAQAB"

// PackProto 构造 UpdateClientInfoReqIdl 请求，对应 pack_proto。
//
// clientVersion 取自会话本身，不同版本的 websocket 要各自上报自己的版本号。
// full 为 true 时上报完整设备指纹，否则只报必需的两个字段（上游同样默认关闭）。
func PackProto(account *core.Account, clientVersion string, full bool) ([]byte, error) {
	device, err := packDevice(account, clientVersion, full)
	if err != nil {
		return nil, err
	}
	deviceJSON, err := json.Marshal(device)
	if err != nil {
		return nil, fmt.Errorf("encoding the websocket device info: %w", err)
	}

	publicKey, err := parsePublicKey()
	if err != nil {
		return nil, err
	}
	secretKey, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, account.AESECBSecKey())
	if err != nil {
		return nil, fmt.Errorf("encrypting the websocket sec key: %w", err)
	}

	req := &pb.UpdateClientInfoReqIdl{
		Cuid: "-|com.baidu.tieba" + clientVersion,
		Data: &pb.UpdateClientInfoReqIdl_DataReq{
			Bduss:     account.BDUSS(),
			Device:    string(deviceJSON),
			SecretKey: secretKey,
		},
	}
	// stoken 与真实 cuid 只在 full 模式下上报，对应上游 pack_proto 的分支。
	if full {
		req.Cuid = account.Cuid() + "|com.baidu.tieba" + clientVersion
		req.Data.Stoken = account.STOKEN()
	}

	out, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("encoding UpdateClientInfoReqIdl: %w", err)
	}
	return out, nil
}

// packDevice 返回随请求上报的设备信息，对应 pack_min_device / pack_full_device。
func packDevice(account *core.Account, clientVersion string, full bool) (map[string]string, error) {
	if !full {
		return map[string]string{
			"_client_type":    "2",
			"_client_version": clientVersion,
		}, nil
	}
	return packFullDevice(account, clientVersion)
}

// packFullDevice 构造完整设备指纹。
//
// 仓库内没有调用方使用 full 模式，这里仅与上游保持一致，供需要时开启。
func packFullDevice(account *core.Account, clientVersion string) (map[string]string, error) {
	cuidGalaxy2, err := account.CuidGalaxy2()
	if err != nil {
		return nil, err
	}
	c3Aid, err := account.C3Aid()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	nowMS := now.UnixMilli()

	return map[string]string{
		"model":                   "PDEM10",
		"brand":                   "Xiaomi",
		"_phone_imei":             "000000000000000",
		"_os_version":             "13",
		"android_id":              account.AndroidID(),
		"mac":                     "",
		"_client_type":            "2",
		"_client_version":         clientVersion,
		"user_agent":              "Mozilla/5.0 (Linux; Android 13; PDEM10 Build/TKQ1.221114.001; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/114.0.5735.196 Mobile Safari/537.36 tieba/" + clientVersion,
		"_client_id":              account.ClientID(),
		"net_type":                "1",
		"net":                     "1",
		"cuid":                    cuidGalaxy2,
		"cuid_galaxy2":            cuidGalaxy2,
		"c3_aid":                  c3Aid,
		"cuid_gid":                "",
		"timestamp":               strconv.FormatInt(nowMS, 10),
		"sample_id":               account.SampleID(),
		"z_id":                    account.ZID(),
		"baiduid":                 account.BaiduID(),
		"sdk_ver":                 "3.36.0",
		"framework_ver":           "",
		"naws_game_ver":           "",
		"_phone_screen":           "1080,2340",
		"_msg_status":             "0",
		"_pic_quality":            "0",
		"pversion":                "1.0.3",
		"q_type":                  "0",
		"scr_h":                   "2340",
		"scr_w":                   "1080",
		"scr_dip":                 "2.75",
		"active_timestamp":        strconv.FormatInt(nowMS, 10),
		"first_install_time":      strconv.FormatInt(nowMS-86400000, 10),
		"update_time":             strconv.FormatInt(nowMS-3600000, 10),
		"event_day":               fmt.Sprintf("%d%d%d", now.Year(), int(now.Month()), now.Day()),
		"cmode":                   "1",
		"extra":                   "",
		"personalized_rec_switch": "1",
		"ad_context":              "",
		"start_type":              "99",
		"start_scheme":            "",
	}, nil
}

// ParseBody 解析 UpdateClientInfoResIdl 响应，对应 parse_body。
func ParseBody(body []byte) (InitWsInfo, error) {
	res := &pb.UpdateClientInfoResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return InitWsInfo{}, fmt.Errorf("decoding UpdateClientInfoResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return InitWsInfo{}, &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}

	data := res.GetData()
	userID := data.GetUserInfo().GetUserId()
	if userID == 0 || data.GetIsUserAvailable() != userID {
		return InitWsInfo{}, &exception.TiebaValueError{
			Msg: fmt.Sprintf("isUserAvailable(%d) userId(%d)", data.GetIsUserAvailable(), userID),
		}
	}

	info := InitWsInfo{
		Groups: make([]WsMsgGroupInfo, 0, len(data.GetGroupInfo())),
		User:   WsUserInfoFromProto(data.GetUserInfo()),
	}
	for _, group := range data.GetGroupInfo() {
		info.Groups = append(info.Groups, WsMsgGroupInfoFromProto(group))
	}
	return info, nil
}

// Request 通过 websocket 发送初始化请求并返回消息分组，对应 request。
func Request(wsCore *core.WsCore, full bool) (InitWsInfo, error) {
	data, err := PackProto(wsCore.Account, wsCore.ClientVersion(), full)
	if err != nil {
		return InitWsInfo{}, err
	}

	// Python 客户端以未加密方式发送该帧（encrypt=False）。
	resp, err := wsCore.Send(data, CMD, core.WithoutEncrypt())
	if err != nil {
		return InitWsInfo{}, err
	}
	payload, err := resp.Read()
	if err != nil {
		return InitWsInfo{}, err
	}
	return ParseBody(payload)
}

func parsePublicKey() (*rsa.PublicKey, error) {
	der, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("decoding the websocket RSA public key: %w", err)
	}
	key, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return nil, fmt.Errorf("parsing the websocket RSA public key: %w", err)
	}
	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("uploadwssec: the websocket public key is not an RSA key")
	}
	return rsaKey, nil
}
