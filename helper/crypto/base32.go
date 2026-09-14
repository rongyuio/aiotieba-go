package crypto

// base32EncodeLen mirrors BASE32_LEN of base32.h. It is an upper bound on the
// encoded length and is exact when len is a multiple of 5.
func base32EncodeLen(n int) int {
	if n%5 != 0 {
		return (n/5)*8 + 8
	}
	return (n / 5) * 8
}

// base32Encode is a byte-for-byte port of tbc_base32_encode.
//
// It encodes src with the RFC 4648 alphabet and no padding. The C signature
// requires a non-empty src; an empty input yields an empty result.
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
