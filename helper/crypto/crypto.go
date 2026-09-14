package crypto

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
)

// CuidGalaxy2 从 android_id 生成 cuid_galaxy2。
//
// 对应 tbc_cuid_galaxy2："com.baidu"+androidID 的 md5 十六进制摘要与 "|V" 拼接，再追加该摘要的 base32 helios 哈希。
func CuidGalaxy2(androidID string) (string, error) {
	if len(androidID) != androidIDSize {
		return "", fmt.Errorf("invalid size of android_id: want %d, got %d", androidIDSize, len(androidID))
	}

	sum := md5.Sum([]byte(cuid2Prefix + androidID))

	dst := make([]byte, 0, CuidGalaxy2Size)
	for _, b := range sum {
		dst = append(dst, hexUpper[b>>4], hexUpper[b&0x0F])
	}
	dst = append(dst, '|', 'V')

	he := heliosHash(dst[:md5StrSize])
	dst = append(dst, base32Encode(he[:])...)
	return string(dst), nil
}

// C3Aid 从 android_id 与 uuid 生成 c3_aid。
//
// 对应 tbc_c3_aid："A00-" + base32(sha1("com.helios"+androidID+uuid)) + "-" + base32(前者的 helios 哈希)。
func C3Aid(androidID, uuid string) (string, error) {
	if len(androidID) != androidIDSize {
		return "", fmt.Errorf("invalid size of android_id: want %d, got %d", androidIDSize, len(androidID))
	}
	if len(uuid) != uuidSize {
		return "", fmt.Errorf("invalid size of uuid: want %d, got %d", uuidSize, len(uuid))
	}

	sum := sha1.Sum([]byte(cuid3Prefix + androidID + uuid))

	dst := make([]byte, 0, C3AidSize)
	dst = append(dst, 'A', '0', '0', '-')
	dst = append(dst, base32Encode(sum[:])...)
	dst = append(dst, '-')

	he := heliosHash(dst)
	dst = append(dst, base32Encode(he[:])...)
	return string(dst), nil
}

// Enuid 生成 BLCP 登录流程使用的 EnUid。
//
// 注意：上游 C 辅助函数 tbc_BB64Encode 从不运行其 BB64 编码器（GC02 是死代码），只是把输入复制进零填充缓冲区。本移植保留该行为，因此结果是 cuid_galaxy2 后跟 EnuidSize-len(cuid_galaxy2) 个 NUL 字节。
func Enuid(cuidGalaxy2 string) (string, error) {
	if len(cuidGalaxy2) != CuidGalaxy2Size {
		return "", fmt.Errorf("invalid size of cuid_galaxy2: want %d, got %d", CuidGalaxy2Size, len(cuidGalaxy2))
	}
	out := bb64Encode([]byte(cuidGalaxy2))
	return string(out[:EnuidSize]), nil
}

// bb64Encode 对应 tbc_BB64Encode（仅复制并补零）。
func bb64Encode(input []byte) []byte {
	resultLen := 4*((len(input)+2)/3) + 2
	out := make([]byte, resultLen)
	copy(out, input)
	return out
}
