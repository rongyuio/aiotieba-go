package crypto

import (
	"encoding/hex"
	"hash/crc32"
	"strings"
	"testing"
)

// The vectors below were produced by compiling the original C sources
// (csrc/base32, csrc/crc, csrc/xxHash, csrc/tbcrypto) with gcc and printing
// deterministic outputs. They are the golden reference for this port.

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("decode hex %q: %v", s, err)
	}
	return b
}

func TestXXH32Vectors(t *testing.T) {
	cases := []struct {
		name     string
		in       string
		seed     uint32
		want     uint32
		wantCopy uint32
	}{
		{"empty", "", 0, 0x02CC5D05, 0x02CC5D05},
		{"a", "61", 0, 0x550D7456, 0x943F556E},
		{"abc", "616263", 0, 0x32D153FF, 0xAC59E17B},
		{"9bytes", "313233343536373839", 0, 0x937BAD67, 0xBC0D1E63},
		{"16bytes", "31323334353637383930313233343536", 0, 0x03BF5152, 0x51D1094E},
		{"17bytes", "3132333435363738393031323334353637", 0, 0xC6BC7CFC, 0xC6F1D17E},
		{"combaidu", "636f6d2e626169647591be894d01799c49", 0, 0xE2BFDFAE, 0x7BAF99EB},
		{"alt", "00ff00ff00ff00ff00ff00ff00ff00ff", 0, 0xA1DD87D7, 0xC8B6CAB8},
		{"abc-seed1", "616263", 1, 0xAA3DA8FF, 0x3F048927},
		{"16bytes-seed12345", "31323334353637383930313233343536", 12345, 0x48C95EDA, 0x16B5136D},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := mustHex(t, tc.in)

			if got := XXH32(buf, tc.seed); got != tc.want {
				t.Errorf("one-shot = %08X, want %08X", got, tc.want)
			}

			st := newXXH32(tc.seed)
			st.update(buf)
			if got := st.digest(); got != tc.want {
				t.Errorf("streaming = %08X, want %08X", got, tc.want)
			}

			half := len(buf) / 2
			st = newXXH32(tc.seed)
			st.update(buf[:half])
			st.update(buf[half:])
			if got := st.digest(); got != tc.want {
				t.Errorf("split = %08X, want %08X", got, tc.want)
			}

			st = newXXH32(tc.seed)
			st.update(buf)
			cpy := st.copy()
			cpy.update(buf)
			if got := cpy.digest(); got != tc.wantCopy {
				t.Errorf("copyState = %08X, want %08X", got, tc.wantCopy)
			}
		})
	}
}

func TestCRC32Vectors(t *testing.T) {
	cases := []struct {
		name string
		in   string
		prev uint32
		want uint32
	}{
		{"empty", "", 0, 0x00000000},
		{"123456789", "313233343536373839", 0, 0xCBF43926},
		{"combaidu", "636f6d2e626169647591be894d01799c49", 0, 0x49D99980},
		{"ffffffff", "ffffffff", 0, 0xFFFFFFFF},
		{"ffffffff-prev", "ffffffff", 0xCBF43926, 0x2D7AEE83},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := crc32.Update(tc.prev, crc32.IEEETable, mustHex(t, tc.in))
			if got != tc.want {
				t.Errorf("crc32 = %08X, want %08X", got, tc.want)
			}
		})
	}
}

func TestBase32Vectors(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"00", "AA"},
		{"ffffffff", "777777Y"},
		{"636f6d2e626169647591be894d01799c49", "MNXW2LTCMFUWI5MRX2EU2ALZTREQ"},
		{"0123456789abcdef0123456789abcdef01234567", "AERUKZ4JVPG66AJDIVTYTK6N54ASGRLH"},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got := string(base32Encode(mustHex(t, tc.in)))
			if got != tc.want {
				t.Errorf("base32 = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestHeliosVectors(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"00", "b223129de4"},
		{"ffffffff", "4a793e03ef"},
		{"636f6d2e626169647591be894d01799c49", "e8188dc67e"},
		{"0123456789abcdef0123456789abcdef", "a415ab168e"},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got := heliosHash(mustHex(t, tc.in))
			if hex.EncodeToString(got[:]) != tc.want {
				t.Errorf("helios = %s, want %s", hex.EncodeToString(got[:]), tc.want)
			}
		})
	}
}

func TestCuidGalaxy2Vectors(t *testing.T) {
	cases := []struct {
		android string
		want    string
	}{
		{"91be894d01799c49", "661CB8A33975DB28AD2F7D15F09E3CF0|VEDCZQMYE"},
		{"A3ED2D7B9CFC28E8", "D63C229A063928758CB25E5D314D7BF7|VVQDMRNYV"},
		{"0000000000000000", "C77D5D04D94F5F56C8A0A6DC3DBF240A|VQKEKVL4O"},
		{"ffffffffffffffff", "A1739C95D4AD1271F31F292F0EF635AF|VCFXHVF73"},
	}

	for _, tc := range cases {
		t.Run(tc.android, func(t *testing.T) {
			got, err := CuidGalaxy2(tc.android)
			if err != nil {
				t.Fatalf("CuidGalaxy2(%q) error: %v", tc.android, err)
			}
			if got != tc.want {
				t.Errorf("cuid_galaxy2 = %q, want %q", got, tc.want)
			}
		})
	}

	if _, err := CuidGalaxy2("short"); err == nil {
		t.Error("CuidGalaxy2 with a bad-sized android_id: want error, got nil")
	}
}

func TestC3AidVectors(t *testing.T) {
	cases := []struct {
		android string
		uuid    string
		want    string
	}{
		{"91be894d01799c49", "e4200716-58a8-4170-af15-ea7edeb8e513", "A00-DFXSU74JQ6DWEDQYL4ZVQJX47GLV5O33-H7AUP3GX"},
		{"A3ED2D7B9CFC28E8", "00000000-0000-0000-0000-000000000000", "A00-4X4PYHZKD4YG46DUCSD5OM6FDDIYUESC-TDWS3TMD"},
		{"0000000000000000", "12345678-1234-1234-1234-1234567890ab", "A00-AKVYF7TY3NZE5RFEOFW65YGZI4VKUEUR-IGQAYBSW"},
	}

	for _, tc := range cases {
		t.Run(tc.android, func(t *testing.T) {
			got, err := C3Aid(tc.android, tc.uuid)
			if err != nil {
				t.Fatalf("C3Aid(%q, %q) error: %v", tc.android, tc.uuid, err)
			}
			if got != tc.want {
				t.Errorf("c3_aid = %q, want %q", got, tc.want)
			}
		})
	}

	if _, err := C3Aid("short", "x"); err == nil {
		t.Error("C3Aid with bad sizes: want error, got nil")
	}
}

func TestRc442Vectors(t *testing.T) {
	cases := []struct {
		name    string
		xyus    string
		keyHex  string
		wantHex string
	}{
		{"seq", "0123456789abcdef0123456789abcdef", "000102030405060708090a0b0c0d0e0f", "ae436870d3618cb52bd57abfe6b47760"},
		{"full", "ffffffffffffffffffffffffffffffff", "ffffffffffffffffffffffffffffffff", "851dd024bc9085ef3c7c2ba64c3adab7"},
		{"zero", "00000000000000000000000000000000", "00000000000000000000000000000000", "a2558caae951fe0d1ed5c85edb4c7490"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Rc442(tc.xyus, mustHex(t, tc.keyHex))
			if err != nil {
				t.Fatalf("Rc442 error: %v", err)
			}
			if hex.EncodeToString(got) != tc.wantHex {
				t.Errorf("rc4_42 = %s, want %s", hex.EncodeToString(got), tc.wantHex)
			}
		})
	}
}

func TestEnuidVectors(t *testing.T) {
	// C reference: the result is cuid_galaxy2 followed by 15 NUL bytes.
	const cuid2 = "A3ED2D7B9CFC28E8934A3FBD3A9579C7|VZ5FKB5XS"
	const cRefPrefix = "41334544324437423943464332384538393334413346424433413935373943377c565a35464b42355853"

	got, err := Enuid(cuid2)
	if err != nil {
		t.Fatalf("Enuid error: %v", err)
	}
	if len(got) != EnuidSize {
		t.Fatalf("len(enuid) = %d, want %d", len(got), EnuidSize)
	}

	gotHex := hex.EncodeToString([]byte(got))
	want := hex.EncodeToString([]byte(cuid2)) + strings.Repeat("00", EnuidSize-len(cuid2))
	if gotHex != want {
		t.Errorf("enuid = %s, want %s", gotHex, want)
	}
	if !strings.HasPrefix(gotHex, cRefPrefix) {
		t.Errorf("enuid does not start with the C reference prefix %s: %s", cRefPrefix, gotHex)
	}

	if _, err := Enuid("short"); err == nil {
		t.Error("Enuid with a bad-sized cuid_galaxy2: want error, got nil")
	}
}

func TestComputeSignVectors(t *testing.T) {
	cases := []struct {
		name string
		data []Param
		salt string
		want string
	}{
		{
			"app",
			[]Param{{"a", "1"}, {"b", 2}},
			AppSalt,
			"42961b9881c2d7cb297e9498f9767789",
		},
		{
			"empty",
			nil,
			"",
			"d41d8cd98f00b204e9800998ecf8427e",
		},
		{
			"misc-utf8",
			[]Param{{"kw", "天堂鸡汤"}, {"pn", "1"}, {"rn", "30"}},
			MiscSalt,
			"f21526b2cd72961a6f3dc2316e32ac6e",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ComputeSign(tc.data, []byte(tc.salt)); got != tc.want {
				t.Errorf("compute_sign = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSignAppendsSignature(t *testing.T) {
	data := []Param{{"k", "v"}, {"z", 10}}
	got := Sign(data, []byte(PCSalt))

	if len(got) != 3 {
		t.Fatalf("len(sign) = %d, want 3", len(got))
	}
	if got[0] != data[0] || got[1] != data[1] {
		t.Errorf("Sign reordered the input: %v", got)
	}
	if got[2].Key != "sign" || got[2].Value != "5f00fd6cf105076d97cf23139702d64e" {
		t.Errorf("signature = %v, want sign=5f00fd6cf105076d97cf23139702d64e", got[2])
	}
}
