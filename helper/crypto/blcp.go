package crypto

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

// bdukKey 与 bdukIV 是 baidu_uk 编解码使用的固定密钥与 IV，
// 对应 aiotieba.helper.crypto.blcp 的 _AES_KEY 与 _AES_IV。
var (
	bdukKey = []byte("AFD311832EDEEAEF")
	bdukIV  = []byte("2011121211143000")
)

// BDUKToUserID 将群聊报文中的 baidu_uk 解密为 user_id。
//
// 参数:
//
//	bduk 群聊报文中的 baidu_uk，urlsafe-base64 且去掉了末尾的 '='
//
// 对应 aiotieba.helper.crypto.blcp.bduk2user_id：补齐 base64 填充后解码，
// 再用固定 key 与 iv 做 AES-CBC 解密并移除 PKCS7 填充，最后把明文解析为整数。
func BDUKToUserID(bduk string) (int64, error) {
	if pad := len(bduk) % 4; pad != 0 {
		bduk += strings.Repeat("=", 4-pad)
	}
	encrypted, err := base64.URLEncoding.DecodeString(bduk)
	if err != nil {
		return 0, fmt.Errorf("crypto: decoding bduk: %w", err)
	}

	plain, err := CBCDecrypt(bdukKey, bdukIV, encrypted)
	if err != nil {
		return 0, fmt.Errorf("crypto: decrypting bduk: %w", err)
	}

	userID, err := strconv.ParseInt(string(plain), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("crypto: parsing the bduk user id: %w", err)
	}
	return userID, nil
}
