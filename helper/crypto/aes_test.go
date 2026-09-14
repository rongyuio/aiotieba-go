package crypto

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// Vectors produced with Python's cryptography package following the exact
// parameters used by aiotieba.core.account.Account.

func TestDeriveECBKeyVector(t *testing.T) {
	key, err := DeriveECBKey(make([]byte, 31))
	if err != nil {
		t.Fatalf("DeriveECBKey error: %v", err)
	}
	const want = "9ac64c08e7ecb4a7dbf043b1aa3f4601b7ec2d60519bd8dcd6cae6376497e543"
	if hex.EncodeToString(key) != want {
		t.Errorf("derived key = %s, want %s", hex.EncodeToString(key), want)
	}
}

func TestECBRoundTripVector(t *testing.T) {
	key := mustHex(t, "9ac64c08e7ecb4a7dbf043b1aa3f4601b7ec2d60519bd8dcd6cae6376497e543")
	plaintext := []byte("hello world")

	ct, err := ECBEncrypt(key, plaintext)
	if err != nil {
		t.Fatalf("ECBEncrypt error: %v", err)
	}
	const want = "eccb6da01f99e5c07c6f7e576d5627e9"
	if hex.EncodeToString(ct) != want {
		t.Errorf("ecb ciphertext = %s, want %s", hex.EncodeToString(ct), want)
	}

	pt, err := ECBDecrypt(key, ct)
	if err != nil {
		t.Fatalf("ECBDecrypt error: %v", err)
	}
	if !bytes.Equal(pt, plaintext) {
		t.Errorf("ecb round trip = %q, want %q", pt, plaintext)
	}
}

func TestCBCZeroIVRoundTripVector(t *testing.T) {
	key := bytes.Repeat([]byte{0x01}, 16)
	plaintext := []byte("hello world")

	ct, err := CBCEncryptZeroIV(key, plaintext)
	if err != nil {
		t.Fatalf("CBCEncryptZeroIV error: %v", err)
	}
	const want = "5c77a4f1c277a7629f98220cd0fa26d9"
	if hex.EncodeToString(ct) != want {
		t.Errorf("cbc ciphertext = %s, want %s", hex.EncodeToString(ct), want)
	}

	pt, err := CBCDecryptZeroIV(key, ct)
	if err != nil {
		t.Fatalf("CBCDecryptZeroIV error: %v", err)
	}
	if !bytes.Equal(pt, plaintext) {
		t.Errorf("cbc round trip = %q, want %q", pt, plaintext)
	}
}

func TestCBCWithIVVector(t *testing.T) {
	// Vector produced with Python's cryptography and the parameters used by
	// BLCPCore.getBDUKfromUserId.
	key := []byte("AFD311832EDEEAEF")
	iv := []byte("2011121211143000")
	plaintext := []byte("1234567890")

	ct, err := CBCEncrypt(key, iv, plaintext)
	if err != nil {
		t.Fatalf("CBCEncrypt error: %v", err)
	}
	const want = "27994a522e616f8c03834e252d1c2302"
	if hex.EncodeToString(ct) != want {
		t.Errorf("cbc ciphertext = %s, want %s", hex.EncodeToString(ct), want)
	}

	pt, err := CBCDecrypt(key, iv, ct)
	if err != nil {
		t.Fatalf("CBCDecrypt error: %v", err)
	}
	if !bytes.Equal(pt, plaintext) {
		t.Errorf("cbc round trip = %q, want %q", pt, plaintext)
	}

	if _, err := CBCEncrypt(key, []byte("short"), plaintext); err == nil {
		t.Error("CBCEncrypt with a bad iv size: want error, got nil")
	}
}

func TestPKCS7RoundTrip(t *testing.T) {
	for _, n := range []int{0, 1, 15, 16, 17, 31, 32, 33, 64} {
		data := bytes.Repeat([]byte{0xAB}, n)
		padded := pkcs7Pad(data, 16)
		if len(padded)%16 != 0 || len(padded) <= n {
			t.Fatalf("pkcs7Pad(%d) produced an invalid length %d", n, len(padded))
		}
		got, err := pkcs7Unpad(padded, 16)
		if err != nil {
			t.Fatalf("pkcs7Unpad(%d) error: %v", n, err)
		}
		if !bytes.Equal(got, data) {
			t.Errorf("pkcs7 round trip(%d) = %v, want %v", n, got, data)
		}
	}
}
