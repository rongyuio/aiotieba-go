// Package crypto 实现百度贴吧客户端的密码学算法。
//
// 它是 Python C 扩展 aiotieba.helper.crypto 与 aiotieba.helper.crypto.sign 的逐字节移植。
package crypto

// 签名盐值，对应 Python 模块 aiotieba.helper.crypto.const 的 APP_SALT、PC_SALT 与 MISC_SALT。
const (
	AppSalt  = "tiebaclient!!!"
	PCSalt   = "36770b1f34c9bbf2e7d1a99d2b82fa9e"
	MiscSalt = "0039d79dc3cc2075129745a30237a3c4"
)

// 混入哈希缓冲区的前缀，对应 tbcrypto/cuid.c 的 CUID2_PERFIX 与 CUID3_PERFIX。
const (
	cuid2Prefix = "com.baidu"
	cuid3Prefix = "com.helios"
)

// 固定输入长度。
const (
	androidIDSize = 16
	uuidSize      = 36
	md5HashSize   = 16
	md5StrSize    = 32
	sha1HashSize  = 20
	cbcSecKeySize = 16
	rc4Size       = 16
)

// 固定输出长度，对应 const.h。
const (
	// HeliosHashSize 是 helios 哈希输出的长度。
	HeliosHashSize = 5
	// Sha1Base32Size 是 SHA1 摘要（20 字节）的 base32 长度。
	Sha1Base32Size = 32
	// CuidGalaxy2Size 是 cuid_galaxy2 的长度。
	CuidGalaxy2Size = md5StrSize + 2 + heliosBase32Size // 42
	// C3AidSize 是 c3_aid 的长度。
	C3AidSize = 4 + Sha1Base32Size + 1 + heliosBase32Size // 45
	// EnuidSize 是 enuid 的长度。
	EnuidSize = 4*((CuidGalaxy2Size+2)/3) + 1 // 57

	heliosBase32Size = 8
)

// hexUpper 对应 HEX_UPPERCASE_TABLE。
const hexUpper = "0123456789ABCDEF"

// base32Alphabet 是 tbc_base32_encode 使用的 RFC 4648 字符表。
const base32Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
