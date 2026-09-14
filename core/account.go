// Package core 实现客户端的网络层：账号状态容器、HTTP 会话、websocket 会话与 BLCP 会话。
package core

import (
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// Account 贴吧的用户参数容器，保存用户身份相关的状态，对应 aiotieba.core.account.Account。
//
// 随机标识（android_id、uuid、cuid、cuid_galaxy2、c3_aid、aes 密钥）在首次使用时生成，
// 之后不再变化，与 Python 的属性行为一致。
type Account struct {
	mu sync.Mutex

	bduss  string
	stoken string

	tbs         string
	androidID   string
	uuid        string
	clientID    string
	sampleID    string
	cuid        string
	cuidGalaxy2 string
	c3Aid       string
	zID         string

	aesECBSecKey []byte
	aesCBCSecKey []byte
	aesECBKey    []byte
}

// NewAccount 创建 Account。BDUSS 必须为空或 192 个字符，STOKEN 必须为空或 64 个字符。
//
// 参数:
//
//	bduss BDUSS
//	stoken 网页STOKEN
//
// BDUSS 必须为空或 192 个字符，STOKEN 必须为空或 64 个字符。
func NewAccount(bduss, stoken string) (*Account, error) {
	a := &Account{}
	if err := a.SetBDUSS(bduss); err != nil {
		return nil, err
	}
	if err := a.SetSTOKEN(stoken); err != nil {
		return nil, err
	}
	return a, nil
}

// BDUSS 当前账号的 BDUSS。
func (a *Account) BDUSS() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.bduss
}

// SetBDUSS 替换 BDUSS。
func (a *Account) SetBDUSS(v string) error {
	if v != "" && len(v) != 192 {
		return fmt.Errorf("BDUSS length must be 192 characters, got %d", len(v))
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.bduss = v
	return nil
}

// STOKEN 当前账号的 STOKEN。
func (a *Account) STOKEN() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.stoken
}

// SetSTOKEN 替换 STOKEN。
func (a *Account) SetSTOKEN(v string) error {
	if v != "" && len(v) != 64 {
		return fmt.Errorf("STOKEN length must be 64 characters, got %d", len(v))
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.stoken = v
	return nil
}

// AndroidID 返回一个随机的 android_id 长度为16的16进制字符串 包含8字节信息 字母为小写。
// 在初始化后该属性便不会再发生变化。
func (a *Account) AndroidID() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.androidIDLocked()
}

func (a *Account) androidIDLocked() string {
	if a.androidID == "" {
		a.androidID = randomHex(8)
	}
	return a.androidID
}

// SetAndroidID 替换 android_id。
func (a *Account) SetAndroidID(v string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.androidID = v
}

// UUID 使用 uuid.uuid4 生成并返回一个随机的 uuid 包含16字节信息。
// 在初始化后该属性便不会再发生变化。
func (a *Account) UUID() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.uuidLocked()
}

func (a *Account) uuidLocked() string {
	if a.uuid == "" {
		a.uuid = randomUUID()
	}
	return a.uuid
}

// SetUUID 替换 uuid。
func (a *Account) SetUUID(v string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.uuid = v
}

// Tbs 返回一个可作为请求参数的反csrf校验码tbs 长度为26的16进制字符串 字母为小写。
// 在初始化后该属性便不会再发生变化。
func (a *Account) Tbs() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.tbs
}

// SetTbs 替换 tbs。
func (a *Account) SetTbs(v string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.tbs = v
}

// ClientID 返回一个可作为请求参数的 client_id 例: wappc_1653660000000_123。
// 在初始化后该属性便不会再发生变化。
func (a *Account) ClientID() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.clientID
}

// SetClientID 替换 client_id。
func (a *Account) SetClientID(v string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.clientID = v
}

// SampleID 返回一个可作为请求参数的 sample_id 例: 104505_3-105324_2-...-107269_1。
// 在初始化后该属性便不会再发生变化。
func (a *Account) SampleID() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.sampleID
}

// SetSampleID 替换 sample_id。
func (a *Account) SetSampleID(v string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.sampleID = v
}

// Cuid 返回一个可作为请求参数的 cuid 例: baidutiebaappe4200716-58a8-4170-af15-ea7edeb8e513。
// 在初始化后该属性便不会再发生变化。此实现仅用于 9.x 等旧版本，11.x 后请使用 CuidGalaxy2 填充对应字段。
//
// Cuid 由 "baidutiebaapp" 与 uuid 拼接而成。
func (a *Account) Cuid() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.cuidLocked()
}

func (a *Account) cuidLocked() string {
	if a.cuid == "" {
		a.cuid = "baidutiebaapp" + a.uuidLocked()
	}
	return a.cuid
}

// SetCuid 替换 cuid。
func (a *Account) SetCuid(v string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cuid = v
}

// CuidGalaxy2 返回一个可作为请求参数的 cuid_galaxy2 例: A3ED2D7B9CFC28E8934A3FBD3A9579C7|VZ5FKB5XS。
// 在初始化后该属性便不会再发生变化。此实现与 12.x 版本及以前的官方实现一致。
//
// CuidGalaxy2 由 android_id 推导得到。
func (a *Account) CuidGalaxy2() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cuidGalaxy2 == "" {
		v, err := crypto.CuidGalaxy2(a.androidIDLocked())
		if err != nil {
			return "", fmt.Errorf("generating cuid_galaxy2: %w", err)
		}
		a.cuidGalaxy2 = v
	}
	return a.cuidGalaxy2, nil
}

// SetCuidGalaxy2 替换 cuid_galaxy2。
func (a *Account) SetCuidGalaxy2(v string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cuidGalaxy2 = v
}

// C3Aid 返回一个可作为请求参数的 c3_aid 例: A00-ZNU3O3EP74D727LMQY745CZSGZQJQZGP-3JXCKC7X。
// 在初始化后该属性便不会再发生变化。此实现与 12.x 版本及以前的官方实现一致。
//
// C3Aid 由 android_id 与 uuid 推导得到。
func (a *Account) C3Aid() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.c3Aid == "" {
		v, err := crypto.C3Aid(a.androidIDLocked(), a.uuidLocked())
		if err != nil {
			return "", fmt.Errorf("generating c3_aid: %w", err)
		}
		a.c3Aid = v
	}
	return a.c3Aid, nil
}

// SetC3Aid 替换 c3_aid。
func (a *Account) SetC3Aid(v string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.c3Aid = v
}

// ZID 返回一个可作为请求参数的 z_id。
// 在初始化后该属性便不会再发生变化。此实现与 12.x 版本及以前的官方实现一致。
//
// ZID 在 z_id 流程填充前为空。
func (a *Account) ZID() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.zID
}

// SetZID 替换 z_id。
func (a *Account) SetZID(v string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.zID = v
}

// AESECBSecKey 返回一个供贴吧 AES-ECB 加密使用的随机密码，长度为31字节。
// 在初始化后该属性便不会再发生变化。
func (a *Account) AESECBSecKey() []byte {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.aesECBSecKeyLocked()
}

func (a *Account) aesECBSecKeyLocked() []byte {
	if a.aesECBSecKey == nil {
		a.aesECBSecKey = randomBytes(31)
	}
	return a.aesECBSecKey
}

// SetAESECBSecKey 替换 AES-ECB 种子。
func (a *Account) SetAESECBSecKey(v []byte) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.aesECBSecKey = v
	a.aesECBKey = nil
}

// AESECBKey 获取供贴吧 websocket 使用的 AES-ECB 加密器。
//
// AESECBKey 是 websocket 协议使用的、由种子派生的 32 字节 AES-ECB 密钥，对应 Account.aes_ecb_chiper。
func (a *Account) AESECBKey() ([]byte, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.aesECBKey == nil {
		key, err := crypto.DeriveECBKey(a.aesECBSecKeyLocked())
		if err != nil {
			return nil, fmt.Errorf("deriving aes-ecb key: %w", err)
		}
		a.aesECBKey = key
	}
	return a.aesECBKey, nil
}

// AESCBCSecKey 返回一个供贴吧 AES-CBC 加密使用的随机密码，长度为16字节。
// 在初始化后该属性便不会再发生变化。
func (a *Account) AESCBCSecKey() []byte {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.aesCBCSecKeyLocked()
}

func (a *Account) aesCBCSecKeyLocked() []byte {
	if a.aesCBCSecKey == nil {
		a.aesCBCSecKey = randomBytes(16)
	}
	return a.aesCBCSecKey
}

// SetAESCBCSecKey 替换 AES-CBC 密钥。
func (a *Account) SetAESCBCSecKey(v []byte) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.aesCBCSecKey = v
}

// ToDict 将 Account 转换为字典，只包含已设置的字段。
//
// 对应 Account.to_dict，只包含已设置的字段。
func (a *Account) ToDict() map[string]any {
	a.mu.Lock()
	defer a.mu.Unlock()

	out := make(map[string]any, 13)
	if a.bduss != "" {
		out["BDUSS"] = a.bduss
	}
	if a.stoken != "" {
		out["STOKEN"] = a.stoken
	}
	if a.tbs != "" {
		out["tbs"] = a.tbs
	}
	if a.androidID != "" {
		out["android_id"] = a.androidID
	}
	if a.uuid != "" {
		out["uuid"] = a.uuid
	}
	if a.clientID != "" {
		out["client_id"] = a.clientID
	}
	if a.sampleID != "" {
		out["sample_id"] = a.sampleID
	}
	if a.cuid != "" {
		out["cuid"] = a.cuid
	}
	if a.cuidGalaxy2 != "" {
		out["cuid_galaxy2"] = a.cuidGalaxy2
	}
	if a.c3Aid != "" {
		out["c3_aid"] = a.c3Aid
	}
	if a.zID != "" {
		out["z_id"] = a.zID
	}
	if a.aesECBSecKey != nil {
		out["aes_ecb_sec_key"] = a.aesECBSecKey
	}
	if a.aesCBCSecKey != nil {
		out["aes_cbc_sec_key"] = a.aesCBCSecKey
	}
	return out
}

// FromDict 将字典转换为 Account。
//
// 参数:
//
//	dict 包含用户参数的字典
//
// 对应 Account.from_dict。
func FromDict(dict map[string]any) (*Account, error) {
	a := &Account{}
	for key, raw := range dict {
		switch key {
		case "BDUSS":
			if s, ok := raw.(string); ok {
				a.bduss = s
			}
		case "STOKEN":
			if s, ok := raw.(string); ok {
				a.stoken = s
			}
		case "tbs":
			if s, ok := raw.(string); ok {
				a.tbs = s
			}
		case "android_id":
			if s, ok := raw.(string); ok {
				a.androidID = s
			}
		case "uuid":
			if s, ok := raw.(string); ok {
				a.uuid = s
			}
		case "client_id":
			if s, ok := raw.(string); ok {
				a.clientID = s
			}
		case "sample_id":
			if s, ok := raw.(string); ok {
				a.sampleID = s
			}
		case "cuid":
			if s, ok := raw.(string); ok {
				a.cuid = s
			}
		case "cuid_galaxy2":
			if s, ok := raw.(string); ok {
				a.cuidGalaxy2 = s
			}
		case "c3_aid":
			if s, ok := raw.(string); ok {
				a.c3Aid = s
			}
		case "z_id":
			if s, ok := raw.(string); ok {
				a.zID = s
			}
		case "aes_ecb_sec_key":
			if b, ok := raw.([]byte); ok {
				a.aesECBSecKey = b
			}
		case "aes_cbc_sec_key":
			if b, ok := raw.([]byte); ok {
				a.aesCBCSecKey = b
			}
		}
	}
	return a, nil
}

// Equal 报告两个账号的 BDUSS 是否相同，对应 Account.__eq__。
func (a *Account) Equal(other *Account) bool {
	return other != nil && a.BDUSS() == other.BDUSS()
}

func randomBytes(n int) []byte {
	b := make([]byte, n)
	_, _ = crand.Read(b)
	return b
}

func randomHex(n int) string {
	return hex.EncodeToString(randomBytes(n))
}

func randomUUID() string {
	var b [16]byte
	_, _ = crand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
