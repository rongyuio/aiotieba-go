package crypto

import "testing"

// bdukOf123456 是 Python 版 user_id2bduk(123456) 的输出，
// 等价于用固定 key 与 iv 对 "123456" 做 AES-CBC + PKCS7 后再取无填充的 urlsafe-base64。
const bdukOf123456 = "xTBAOCyxXjl9vA8KC1jI7g"

func TestBDUKToUserID(t *testing.T) {
	userID, err := BDUKToUserID(bdukOf123456)
	if err != nil {
		t.Fatalf("BDUKToUserID: %v", err)
	}
	if userID != 123456 {
		t.Errorf("user id = %d, want 123456", userID)
	}
}

func TestBDUKToUserIDPadded(t *testing.T) {
	// bduk 末尾的 '=' 会被去掉，但补齐填充后解码同样成立。
	userID, err := BDUKToUserID(bdukOf123456 + "==")
	if err != nil {
		t.Fatalf("BDUKToUserID: %v", err)
	}
	if userID != 123456 {
		t.Errorf("user id = %d, want 123456", userID)
	}
}

func TestBDUKToUserIDRejectsGarbage(t *testing.T) {
	if _, err := BDUKToUserID("not-a-bduk!"); err == nil {
		t.Error("BDUKToUserID accepted garbage")
	}
}
