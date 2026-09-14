// Package core implements the network layer of the client: the account state
// container, the HTTP session, the websocket session and the BLCP session.
package core

import (
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// Account holds the identity related state of a Tieba user.
//
// It mirrors aiotieba.core.account.Account. The random identifiers
// (android_id, uuid, cuid, cuid_galaxy2, c3_aid, aes keys) are generated on
// first use and never change afterwards, exactly like the Python properties.
type Account struct {
	mu sync.Mutex

	bduss  string
	stoken string

	tbs         string
	androidID   string
	uuid        string
	clientID    string
	sampleID    string
	cuid        string
	cuidGalaxy2 string
	c3Aid       string
	zID         string

	aesECBSecKey []byte
	aesCBCSecKey []byte
	aesECBKey    []byte
}

// NewAccount creates an Account. BDUSS must be empty or 192 characters long and
// STOKEN must be empty or 64 characters long.
func NewAccount(bduss, stoken string) (*Account, error) {
	a := &Account{}
	if err := a.SetBDUSS(bduss); err != nil {
		return nil, err
	}
	if err := a.SetSTOKEN(stoken); err != nil {
		return nil, err
	}
	return a, nil
}

// BDUSS returns the BDUSS of the account.
func (a *Account) BDUSS() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.bduss
}

// SetBDUSS replaces the BDUSS.
func (a *Account) SetBDUSS(v string) error {
	if v != "" && len(v) != 192 {
		return fmt.Errorf("BDUSS length must be 192 characters, got %d", len(v))
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.bduss = v
	return nil
}

// STOKEN returns the STOKEN of the account.
func (a *Account) STOKEN() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.stoken
}

// SetSTOKEN replaces the STOKEN.
func (a *Account) SetSTOKEN(v string) error {
	if v != "" && len(v) != 64 {
		return fmt.Errorf("STOKEN length must be 64 characters, got %d", len(v))
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.stoken = v
	return nil
}

// AndroidID returns a random 16-character lowercase hex android_id (8 bytes).
func (a *Account) AndroidID() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.androidIDLocked()
}

func (a *Account) androidIDLocked() string {
	if a.androidID == "" {
		a.androidID = randomHex(8)
	}
	return a.androidID
}

// SetAndroidID replaces the android_id.
func (a *Account) SetAndroidID(v string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.androidID = v
}

// UUID returns a random v4 uuid.
func (a *Account) UUID() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.uuidLocked()
}

func (a *Account) uuidLocked() string {
	if a.uuid == "" {
		a.uuid = randomUUID()
	}
	return a.uuid
}

// SetUUID replaces the uuid.
func (a *Account) SetUUID(v string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.uuid = v
}

// Tbs returns the anti-CSRF token. It is empty until a login flow fills it.
func (a *Account) Tbs() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.tbs
}

// SetTbs replaces the tbs.
func (a *Account) SetTbs(v string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.tbs = v
}

// ClientID returns the client_id. It is empty until a login flow fills it.
func (a *Account) ClientID() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.clientID
}

// SetClientID replaces the client_id.
func (a *Account) SetClientID(v string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.clientID = v
}

// SampleID returns the sample_id. It is empty until a login flow fills it.
func (a *Account) SampleID() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.sampleID
}

// SetSampleID replaces the sample_id.
func (a *Account) SetSampleID(v string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.sampleID = v
}

// Cuid returns "baidutiebaapp" followed by the uuid.
func (a *Account) Cuid() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.cuidLocked()
}

func (a *Account) cuidLocked() string {
	if a.cuid == "" {
		a.cuid = "baidutiebaapp" + a.uuidLocked()
	}
	return a.cuid
}

// SetCuid replaces the cuid.
func (a *Account) SetCuid(v string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cuid = v
}

// CuidGalaxy2 returns cuid_galaxy2 derived from the android_id.
func (a *Account) CuidGalaxy2() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cuidGalaxy2 == "" {
		v, err := crypto.CuidGalaxy2(a.androidIDLocked())
		if err != nil {
			return "", fmt.Errorf("generating cuid_galaxy2: %w", err)
		}
		a.cuidGalaxy2 = v
	}
	return a.cuidGalaxy2, nil
}

// SetCuidGalaxy2 replaces cuid_galaxy2.
func (a *Account) SetCuidGalaxy2(v string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cuidGalaxy2 = v
}

// C3Aid returns c3_aid derived from the android_id and the uuid.
func (a *Account) C3Aid() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.c3Aid == "" {
		v, err := crypto.C3Aid(a.androidIDLocked(), a.uuidLocked())
		if err != nil {
			return "", fmt.Errorf("generating c3_aid: %w", err)
		}
		a.c3Aid = v
	}
	return a.c3Aid, nil
}

// SetC3Aid replaces c3_aid.
func (a *Account) SetC3Aid(v string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.c3Aid = v
}

// ZID returns the z_id. It is empty until the z_id flow fills it.
func (a *Account) ZID() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.zID
}

// SetZID replaces the z_id.
func (a *Account) SetZID(v string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.zID = v
}

// AESECBSecKey returns the random 31-byte seed of the websocket AES-ECB key.
func (a *Account) AESECBSecKey() []byte {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.aesECBSecKeyLocked()
}

func (a *Account) aesECBSecKeyLocked() []byte {
	if a.aesECBSecKey == nil {
		a.aesECBSecKey = randomBytes(31)
	}
	return a.aesECBSecKey
}

// SetAESECBSecKey replaces the AES-ECB seed.
func (a *Account) SetAESECBSecKey(v []byte) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.aesECBSecKey = v
	a.aesECBKey = nil
}

// AESECBKey returns the derived 32-byte AES-ECB key used by the websocket
// protocol, mirroring Account.aes_ecb_chiper.
func (a *Account) AESECBKey() ([]byte, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.aesECBKey == nil {
		key, err := crypto.DeriveECBKey(a.aesECBSecKeyLocked())
		if err != nil {
			return nil, fmt.Errorf("deriving aes-ecb key: %w", err)
		}
		a.aesECBKey = key
	}
	return a.aesECBKey, nil
}

// AESCBCSecKey returns the random 16-byte AES-CBC key.
func (a *Account) AESCBCSecKey() []byte {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.aesCBCSecKeyLocked()
}

func (a *Account) aesCBCSecKeyLocked() []byte {
	if a.aesCBCSecKey == nil {
		a.aesCBCSecKey = randomBytes(16)
	}
	return a.aesCBCSecKey
}

// SetAESCBCSecKey replaces the AES-CBC key.
func (a *Account) SetAESCBCSecKey(v []byte) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.aesCBCSecKey = v
}

// ToDict serializes the account, mirroring Account.to_dict. Only the fields
// that are set are included.
func (a *Account) ToDict() map[string]any {
	a.mu.Lock()
	defer a.mu.Unlock()

	out := make(map[string]any, 13)
	if a.bduss != "" {
		out["BDUSS"] = a.bduss
	}
	if a.stoken != "" {
		out["STOKEN"] = a.stoken
	}
	if a.tbs != "" {
		out["tbs"] = a.tbs
	}
	if a.androidID != "" {
		out["android_id"] = a.androidID
	}
	if a.uuid != "" {
		out["uuid"] = a.uuid
	}
	if a.clientID != "" {
		out["client_id"] = a.clientID
	}
	if a.sampleID != "" {
		out["sample_id"] = a.sampleID
	}
	if a.cuid != "" {
		out["cuid"] = a.cuid
	}
	if a.cuidGalaxy2 != "" {
		out["cuid_galaxy2"] = a.cuidGalaxy2
	}
	if a.c3Aid != "" {
		out["c3_aid"] = a.c3Aid
	}
	if a.zID != "" {
		out["z_id"] = a.zID
	}
	if a.aesECBSecKey != nil {
		out["aes_ecb_sec_key"] = a.aesECBSecKey
	}
	if a.aesCBCSecKey != nil {
		out["aes_cbc_sec_key"] = a.aesCBCSecKey
	}
	return out
}

// FromDict restores an account from a map, mirroring Account.from_dict.
func FromDict(dict map[string]any) (*Account, error) {
	a := &Account{}
	for key, raw := range dict {
		switch key {
		case "BDUSS":
			if s, ok := raw.(string); ok {
				a.bduss = s
			}
		case "STOKEN":
			if s, ok := raw.(string); ok {
				a.stoken = s
			}
		case "tbs":
			if s, ok := raw.(string); ok {
				a.tbs = s
			}
		case "android_id":
			if s, ok := raw.(string); ok {
				a.androidID = s
			}
		case "uuid":
			if s, ok := raw.(string); ok {
				a.uuid = s
			}
		case "client_id":
			if s, ok := raw.(string); ok {
				a.clientID = s
			}
		case "sample_id":
			if s, ok := raw.(string); ok {
				a.sampleID = s
			}
		case "cuid":
			if s, ok := raw.(string); ok {
				a.cuid = s
			}
		case "cuid_galaxy2":
			if s, ok := raw.(string); ok {
				a.cuidGalaxy2 = s
			}
		case "c3_aid":
			if s, ok := raw.(string); ok {
				a.c3Aid = s
			}
		case "z_id":
			if s, ok := raw.(string); ok {
				a.zID = s
			}
		case "aes_ecb_sec_key":
			if b, ok := raw.([]byte); ok {
				a.aesECBSecKey = b
			}
		case "aes_cbc_sec_key":
			if b, ok := raw.([]byte); ok {
				a.aesCBCSecKey = b
			}
		}
	}
	return a, nil
}

// Equal reports whether two accounts carry the same BDUSS, mirroring
// Account.__eq__.
func (a *Account) Equal(other *Account) bool {
	return other != nil && a.BDUSS() == other.BDUSS()
}

func randomBytes(n int) []byte {
	b := make([]byte, n)
	_, _ = crand.Read(b)
	return b
}

func randomHex(n int) string {
	return hex.EncodeToString(randomBytes(n))
}

func randomUUID() string {
	var b [16]byte
	_, _ = crand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
