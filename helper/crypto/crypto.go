package crypto

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
)

// CuidGalaxy2 generates cuid_galaxy2 from an android_id.
//
// It mirrors tbc_cuid_galaxy2: the md5 hex digest of "com.baidu"+androidID is
// joined with "|V" and suffixed with the base32 helios hash of the digest.
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

// C3Aid generates c3_aid from an android_id and a uuid.
//
// It mirrors tbc_c3_aid: "A00-" + base32(sha1("com.helios"+androidID+uuid)) +
// "-" + base32(helios hash of the former).
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

// Enuid generates the EnUid used by the BLCP login flow.
//
// Note: the upstream C helper tbc_BB64Encode never runs its BB64 encoder
// (GC02 is dead code) and merely copies the input into a zero-filled buffer.
// This port preserves that behaviour, so the result is cuid_galaxy2 followed
// by EnuidSize-len(cuid_galaxy2) NUL bytes.
func Enuid(cuidGalaxy2 string) (string, error) {
	if len(cuidGalaxy2) != CuidGalaxy2Size {
		return "", fmt.Errorf("invalid size of cuid_galaxy2: want %d, got %d", CuidGalaxy2Size, len(cuidGalaxy2))
	}
	out := bb64Encode([]byte(cuidGalaxy2))
	return string(out[:EnuidSize]), nil
}

// bb64Encode mirrors tbc_BB64Encode (which is a copy plus zero padding).
func bb64Encode(input []byte) []byte {
	resultLen := 4*((len(input)+2)/3) + 2
	out := make([]byte, resultLen)
	copy(out, input)
	return out
}
