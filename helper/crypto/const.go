// Package crypto implements the Baidu Tieba client cryptography.
//
// It is a byte-for-byte port of the Python C extension
// aiotieba.helper.crypto combined with aiotieba.helper.crypto.sign.
package crypto

// Signature salts. They mirror APP_SALT, PC_SALT and MISC_SALT of the Python
// module aiotieba.helper.crypto.const.
const (
	AppSalt  = "tiebaclient!!!"
	PCSalt   = "36770b1f34c9bbf2e7d1a99d2b82fa9e"
	MiscSalt = "0039d79dc3cc2075129745a30237a3c4"
)

// Prefixes mixed into the hashed buffers, mirroring CUID2_PERFIX and
// CUID3_PERFIX of tbcrypto/cuid.c.
const (
	cuid2Prefix = "com.baidu"
	cuid3Prefix = "com.helios"
)

// Fixed input sizes.
const (
	androidIDSize = 16
	uuidSize      = 36
	md5HashSize   = 16
	md5StrSize    = 32
	sha1HashSize  = 20
	cbcSecKeySize = 16
	rc4Size       = 16
)

// Fixed output sizes, mirroring const.h.
const (
	// HeliosHashSize is the size of the helios hash output.
	HeliosHashSize = 5
	// Sha1Base32Size is the base32 length of a SHA1 digest (20 bytes).
	Sha1Base32Size = 32
	// CuidGalaxy2Size is the length of cuid_galaxy2.
	CuidGalaxy2Size = md5StrSize + 2 + heliosBase32Size // 42
	// C3AidSize is the length of c3_aid.
	C3AidSize = 4 + Sha1Base32Size + 1 + heliosBase32Size // 45
	// EnuidSize is the length of enuid.
	EnuidSize = 4*((CuidGalaxy2Size+2)/3) + 1 // 57

	heliosBase32Size = 8
)

// hexUpper mirrors HEX_UPPERCASE_TABLE.
const hexUpper = "0123456789ABCDEF"

// base32Alphabet is the RFC 4648 alphabet used by tbc_base32_encode.
const base32Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
