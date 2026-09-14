package crypto

// base32EncodeLen 对应 base32.h 的 BASE32_LEN。它是编码长度的上界，当 len 为 5 的倍数时精确。
func base32EncodeLen(n int) int {
	if n%5 != 0 {
		return (n/5)*8 + 8
	}
	return (n / 5) * 8
}

// base32Encode 是 tbc_base32_encode 的逐字节移植。
//
// 它使用 RFC 4648 字符表编码 src 且不填充。C 签名要求 src 非空；空输入返回空结果。
func base32Encode(src []byte) []byte {
	if len(src) == 0 {
		return nil
	}
	dst := make([]byte, 0, base32EncodeLen(len(src)))

	buffer := uint32(src[0])
	next := 1
	bitsLeft := 8
	for bitsLeft > 0 || next < len(src) {
		if bitsLeft < 5 {
			if next < len(src) {
				buffer <<= 8
				buffer |= uint32(src[next]) & 0xFF
				next++
				bitsLeft += 8
			} else {
				pad := 5 - bitsLeft
				buffer <<= uint(pad)
				bitsLeft += pad
			}
		}
		index := 0x1F & (buffer >> uint(bitsLeft-5))
		bitsLeft -= 5
		dst = append(dst, base32Alphabet[index])
	}
	return dst
}
