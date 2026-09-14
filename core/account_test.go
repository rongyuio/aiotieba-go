package core

import (
	"encoding/hex"
	"regexp"
	"strings"
	"testing"
)

func TestNewAccountValidatesTokenLengths(t *testing.T) {
	if _, err := NewAccount("short", ""); err == nil {
		t.Error("NewAccount with a bad BDUSS: want error, got nil")
	}
	if _, err := NewAccount(strings.Repeat("b", 192), "short"); err == nil {
		t.Error("NewAccount with a bad STOKEN: want error, got nil")
	}
	if _, err := NewAccount(strings.Repeat("b", 192), strings.Repeat("s", 64)); err != nil {
		t.Errorf("NewAccount with valid tokens: %v", err)
	}
	if _, err := NewAccount("", ""); err != nil {
		t.Errorf("NewAccount with empty tokens: %v", err)
	}
}

func TestAndroidIDFormat(t *testing.T) {
	a, _ := NewAccount("", "")
	id := a.AndroidID()
	if len(id) != 16 {
		t.Fatalf("android_id length = %d, want 16", len(id))
	}
	if !regexp.MustCompile(`^[0-9a-f]{16}$`).MatchString(id) {
		t.Errorf("android_id = %q, want lowercase hex", id)
	}
	if a.AndroidID() != id {
		t.Error("android_id changed between calls")
	}
}

func TestUUIDFormat(t *testing.T) {
	a, _ := NewAccount("", "")
	u := a.UUID()
	pattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	if !pattern.MatchString(u) {
		t.Errorf("uuid = %q, want a v4 uuid", u)
	}
	if a.UUID() != u {
		t.Error("uuid changed between calls")
	}
}

func TestCuidUsesUUID(t *testing.T) {
	a, _ := NewAccount("", "")
	a.SetUUID("e4200716-58a8-4170-af15-ea7edeb8e513")
	if got := a.Cuid(); got != "baidutiebaapp"+a.UUID() {
		t.Errorf("cuid = %q", got)
	}
}

func TestCuidGalaxy2AndC3AidMatchVectors(t *testing.T) {
	a, _ := NewAccount("", "")
	a.SetAndroidID("91be894d01799c49")
	a.SetUUID("e4200716-58a8-4170-af15-ea7edeb8e513")

	got, err := a.CuidGalaxy2()
	if err != nil {
		t.Fatalf("CuidGalaxy2: %v", err)
	}
	const wantCuid = "661CB8A33975DB28AD2F7D15F09E3CF0|VEDCZQMYE"
	if got != wantCuid {
		t.Errorf("cuid_galaxy2 = %q, want %q", got, wantCuid)
	}

	got, err = a.C3Aid()
	if err != nil {
		t.Fatalf("C3Aid: %v", err)
	}
	const wantAid = "A00-DFXSU74JQ6DWEDQYL4ZVQJX47GLV5O33-H7AUP3GX"
	if got != wantAid {
		t.Errorf("c3_aid = %q, want %q", got, wantAid)
	}
}

func TestAESKeys(t *testing.T) {
	a, _ := NewAccount("", "")
	if n := len(a.AESCBCSecKey()); n != 16 {
		t.Errorf("aes_cbc_sec_key length = %d, want 16", n)
	}
	if n := len(a.AESECBSecKey()); n != 31 {
		t.Errorf("aes_ecb_sec_key length = %d, want 31", n)
	}

	// The derivation must match the Python reference vector.
	a.SetAESECBSecKey(make([]byte, 31))
	key, err := a.AESECBKey()
	if err != nil {
		t.Fatalf("AESECBKey: %v", err)
	}
	const want = "9ac64c08e7ecb4a7dbf043b1aa3f4601b7ec2d60519bd8dcd6cae6376497e543"
	if hex.EncodeToString(key) != want {
		t.Errorf("derived key = %s, want %s", hex.EncodeToString(key), want)
	}

	// Replacing the seed must invalidate the derived key.
	a.SetAESECBSecKey([]byte("0123456789012345678901234567890"))
	first, _ := a.AESECBKey()
	if hex.EncodeToString(first) == want {
		t.Error("derived key was not recomputed after replacing the seed")
	}
}

func TestAccountDictRoundTrip(t *testing.T) {
	a, _ := NewAccount(strings.Repeat("b", 192), strings.Repeat("s", 64))
	a.SetAndroidID("91be894d01799c49")
	a.SetUUID("e4200716-58a8-4170-af15-ea7edeb8e513")
	a.SetTbs("tbs-value")
	a.SetClientID("wappc_1_2")
	a.SetSampleID("104505_3")
	a.SetZID("zid-value")
	a.SetAESECBSecKey(make([]byte, 31))
	a.SetAESCBCSecKey(make([]byte, 16))
	if _, err := a.CuidGalaxy2(); err != nil {
		t.Fatalf("CuidGalaxy2: %v", err)
	}
	if _, err := a.C3Aid(); err != nil {
		t.Fatalf("C3Aid: %v", err)
	}

	dict := a.ToDict()
	restored, err := FromDict(dict)
	if err != nil {
		t.Fatalf("FromDict: %v", err)
	}

	checks := []struct {
		name string
		got  string
		want string
	}{
		{"BDUSS", restored.BDUSS(), a.BDUSS()},
		{"STOKEN", restored.STOKEN(), a.STOKEN()},
		{"tbs", restored.Tbs(), a.Tbs()},
		{"android_id", restored.AndroidID(), a.AndroidID()},
		{"uuid", restored.UUID(), a.UUID()},
		{"client_id", restored.ClientID(), a.ClientID()},
		{"sample_id", restored.SampleID(), a.SampleID()},
		{"cuid", restored.Cuid(), a.Cuid()},
		{"z_id", restored.ZID(), a.ZID()},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("restored %s = %q, want %q", c.name, c.got, c.want)
		}
	}
	if restoredCuid2, _ := restored.CuidGalaxy2(); restoredCuid2 != dict["cuid_galaxy2"] {
		t.Errorf("restored cuid_galaxy2 = %q, want %v", restoredCuid2, dict["cuid_galaxy2"])
	}
	if restoredAid, _ := restored.C3Aid(); restoredAid != dict["c3_aid"] {
		t.Errorf("restored c3_aid = %q, want %v", restoredAid, dict["c3_aid"])
	}
	if got, _ := restored.AESECBKey(); hex.EncodeToString(got) != hex.EncodeToString(keyOf(a)) {
		t.Error("restored AES-ECB key differs")
	}
}

func keyOf(a *Account) []byte {
	key, _ := a.AESECBKey()
	return key
}

func TestToDictOmitsUnsetFields(t *testing.T) {
	a, _ := NewAccount("", "")
	dict := a.ToDict()
	if len(dict) != 0 {
		t.Errorf("ToDict of a fresh account = %v, want an empty map", dict)
	}
}
