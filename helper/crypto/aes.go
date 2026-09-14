package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/sha1"
	"errors"
	"fmt"
)

// wsECBSalt 是用于派生 websocket AES-ECB 密钥的 PBKDF2 盐值。
var wsECBSalt = []byte{0xa4, 0x0b, 0xc8, 0x34, 0xd6, 0x95, 0xf3, 0x13}

const (
	ecbKeyLen  = 32
	ecbKeyIter = 5
)

// DeriveECBKey 从随机 sec key 派生 websocket AES-ECB 密钥。
//
// 对应 Account.aes_ecb_chiper：
// PBKDF2HMAC(SHA1, 32, salt=b"\xa4\x0b\xc8\x34\xd6\x95\xf3\x13", iterations=5).
func DeriveECBKey(secKey []byte) ([]byte, error) {
	return pbkdf2.Key(sha1.New, string(secKey), wsECBSalt, ecbKeyIter, ecbKeyLen)
}

// ECBEncrypt 先用 PKCS7 填充，再用 AES-ECB 加密明文。
//
// Go 标准库没有 ECB 模式，因此按块实现。
func ECBEncrypt(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: creating aes cipher: %w", err)
	}
	padded := pkcs7Pad(plaintext, block.BlockSize())
	out := make([]byte, len(padded))
	for i := 0; i < len(padded); i += block.BlockSize() {
		block.Encrypt(out[i:i+block.BlockSize()], padded[i:i+block.BlockSize()])
	}
	return out, nil
}

// ECBDecrypt 用 AES-ECB 解密并移除 PKCS7 填充。
func ECBDecrypt(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: creating aes cipher: %w", err)
	}
	bs := block.BlockSize()
	if len(ciphertext) == 0 || len(ciphertext)%bs != 0 {
		return nil, errors.New("crypto: ciphertext is not a multiple of the block size")
	}
	out := make([]byte, len(ciphertext))
	for i := 0; i < len(ciphertext); i += bs {
		block.Decrypt(out[i:i+bs], ciphertext[i:i+bs])
	}
	return pkcs7Unpad(out, bs)
}

// CBCEncrypt 先用 PKCS7 填充，再用 AES-CBC 加密明文。
func CBCEncrypt(key, iv, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: creating aes cipher: %w", err)
	}
	if len(iv) != block.BlockSize() {
		return nil, fmt.Errorf("crypto: invalid iv size %d", len(iv))
	}
	padded := pkcs7Pad(plaintext, block.BlockSize())
	out := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(out, padded)
	return out, nil
}

// CBCDecryptRaw 用 AES-CBC 解密，但不处理填充。
//
// Python 客户端在 init_z_id 中需要它：明文带有末尾的 MD5 后缀，必须在 PKCS7 去填充前剥离。
func CBCDecryptRaw(key, iv, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: creating aes cipher: %w", err)
	}
	bs := block.BlockSize()
	if len(iv) != bs {
		return nil, fmt.Errorf("crypto: invalid iv size %d", len(iv))
	}
	if len(ciphertext) == 0 || len(ciphertext)%bs != 0 {
		return nil, errors.New("crypto: ciphertext is not a multiple of the block size")
	}
	out := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(out, ciphertext)
	return out, nil
}

// CBCDecrypt 用 AES-CBC 解密并移除 PKCS7 填充。
func CBCDecrypt(key, iv, ciphertext []byte) ([]byte, error) {
	out, err := CBCDecryptRaw(key, iv, ciphertext)
	if err != nil {
		return nil, err
	}
	return PKCS7Unpad(out)
}

// PKCS7Unpad 移除 data 的 PKCS7 填充。
func PKCS7Unpad(data []byte) ([]byte, error) {
	return pkcs7Unpad(data, aes.BlockSize)
}

// PKCS7Pad 追加 data 的 PKCS7 填充。
func PKCS7Pad(data []byte) []byte {
	return pkcs7Pad(data, aes.BlockSize)
}

// CBCEncryptZeroIV 先用 PKCS7 填充，再用 AES-CBC 与零 IV 加密明文，对应 Account.aes_cbc_chiper。
func CBCEncryptZeroIV(key, plaintext []byte) ([]byte, error) {
	return CBCEncrypt(key, make([]byte, aes.BlockSize), plaintext)
}

// CBCDecryptZeroIV 用 AES-CBC 与零 IV 解密。
func CBCDecryptZeroIV(key, ciphertext []byte) ([]byte, error) {
	return CBCDecrypt(key, make([]byte, aes.BlockSize), ciphertext)
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	pad := blockSize - len(data)%blockSize
	out := make([]byte, len(data)+pad)
	copy(out, data)
	for i := len(data); i < len(out); i++ {
		out[i] = byte(pad)
	}
	return out
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, errors.New("crypto: invalid padded data")
	}
	pad := int(data[len(data)-1])
	if pad == 0 || pad > blockSize || pad > len(data) {
		return nil, errors.New("crypto: invalid padding")
	}
	for _, b := range data[len(data)-pad:] {
		if int(b) != pad {
			return nil, errors.New("crypto: invalid padding")
		}
	}
	return data[:len(data)-pad], nil
}
