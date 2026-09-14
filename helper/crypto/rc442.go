package crypto

import "fmt"

// Rc442 使用贴吧所用的 RC4 变体加密 cbc_sec_key。
//
// 对应 tbc_rc4_42：以 32 字节 xyus_md5 字符串为密钥的 RC4，并对每个输出字节额外 XOR 42。
func Rc442(xyusMD5 string, cbcSecKey []byte) ([]byte, error) {
	if len(xyusMD5) != md5StrSize {
		return nil, fmt.Errorf("invalid size of xyus_md5: want %d, got %d", md5StrSize, len(xyusMD5))
	}
	if len(cbcSecKey) != cbcSecKeySize {
		return nil, fmt.Errorf("invalid size of cbc_sec_key: want %d, got %d", cbcSecKeySize, len(cbcSecKey))
	}
	return rc4_42([]byte(xyusMD5), cbcSecKey), nil
}

func rc4_42(key, src []byte) []byte {
	var m [256]byte
	for i := range m {
		m[i] = byte(i)
	}

	j := 0
	k := 0
	for i := range 256 {
		if k >= len(key) {
			k = 0
		}
		a := m[i]
		j = (j + int(a) + int(key[k])) & 0xFF
		m[i] = m[j]
		m[j] = a
		k++
	}

	dst := make([]byte, len(src))
	x, y := 0, 0
	for i := range src {
		x = (x + 1) & 0xFF
		a := int(m[x])
		y = (y + a) & 0xFF
		b := int(m[y])
		m[x] = byte(b)
		m[y] = byte(a)
		dst[i] = src[i] ^ m[byte(a+b)] ^ 42 // RC4 的 "+42" 变体
	}
	return dst
}
