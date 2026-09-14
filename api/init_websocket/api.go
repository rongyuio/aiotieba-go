package initwebsocket

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

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"

	pb "github.com/rongyuio/aiotieba-go/api/init_websocket/protobuf"
)

// CMD 是 init_websocket 的 websocket 命令字。
const CMD = 1001

// publicKeyBase64 是用于加密 websocket 密钥的 DER（base64 编码）RSA 公钥，对应 Python 模块的 PUBLIC_KEY。
const publicKeyBase64 = "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwQpwBZxXJV/JVRF/uNfyMSdu7YWwRNLM8+2xbniGp2iIQHOikPpTYQjlQgMi1uvq1kZpJ32rHo3hkwjy2l0lFwr3u4Hk2Wk7vnsqYQjAlYlK0TCzjpmiI+OiPOUNVtbWHQiLiVqFtzvpvi4AU7C1iKGvc/4IS45WjHxeScHhnZZ7njS4S1UgNP/GflRIbzgbBhyZ9kEW5/OO5YfG1fy6r4KSlDJw4o/mw5XhftyIpL+5ZBVBC6E1EIiP/dd9AbK62VV1PByfPMHMixpxI3GM2qwcmFsXcCcgvUXJBa9k6zP8dDQ3csCM2QNT+CQAOxthjtp/TFWaD7MzOdsIYb3THwIDAQAB"

// deviceInfo 保留 Python device 字典的键顺序。
type deviceInfo struct {
	Cuid          string `json:"cuid"`
	ClientVersion string `json:"_client_version"`
	MsgStatus     string `json:"_msg_status"`
	CuidGalaxy2   string `json:"cuid_galaxy2"`
	ClientType    string `json:"_client_type"`
	Timestamp     string `json:"timestamp"`
}

// PackProto 构造 UpdateClientInfoReqIdl 请求，对应 pack_proto。
func PackProto(account *core.Account) ([]byte, error) {
	cuidGalaxy2, err := account.CuidGalaxy2()
	if err != nil {
		return nil, err
	}

	device, err := json.Marshal(deviceInfo{
		Cuid:          account.Cuid(),
		ClientVersion: consts.LegacyVersion,
		MsgStatus:     "1",
		CuidGalaxy2:   cuidGalaxy2,
		ClientType:    "2",
		Timestamp:     strconv.FormatInt(time.Now().UnixMilli(), 10),
	})
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
		Cuid: account.Cuid() + "|com.baidu.tieba" + consts.LegacyVersion,
		Data: &pb.UpdateClientInfoReqIdl_DataReq{
			Bduss:     account.BDUSS(),
			Device:    string(device),
			SecretKey: secretKey,
			Stoken:    account.STOKEN(),
		},
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("encoding UpdateClientInfoReqIdl: %w", err)
	}
	return out, nil
}

// ParseBody 解析 UpdateClientInfoResIdl 响应，对应 parse_body。
func ParseBody(body []byte) ([]WsMsgGroupInfo, error) {
	res := &pb.UpdateClientInfoResIdl{}
	if err := proto.Unmarshal(body, res); err != nil {
		return nil, fmt.Errorf("decoding UpdateClientInfoResIdl: %w", err)
	}
	if code := res.GetError().GetErrorno(); code != 0 {
		return nil, &exception.TiebaServerError{Code: int(code), Msg: res.GetError().GetErrmsg()}
	}

	groups := make([]WsMsgGroupInfo, 0, len(res.GetData().GetGroupInfo()))
	for _, group := range res.GetData().GetGroupInfo() {
		groups = append(groups, WsMsgGroupInfoFromProto(group))
	}
	return groups, nil
}

// Request 通过 websocket 发送初始化请求并返回消息分组，对应 request。
func Request(wsCore *core.WsCore) ([]WsMsgGroupInfo, error) {
	data, err := PackProto(wsCore.Account)
	if err != nil {
		return nil, err
	}

	// Python 客户端以未加密方式发送该帧（encrypt=False）。
	resp, err := wsCore.Send(data, CMD, core.WithoutEncrypt())
	if err != nil {
		return nil, err
	}
	payload, err := resp.Read()
	if err != nil {
		return nil, err
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
		return nil, errors.New("initwebsocket: the websocket public key is not an RSA key")
	}
	return rsaKey, nil
}
