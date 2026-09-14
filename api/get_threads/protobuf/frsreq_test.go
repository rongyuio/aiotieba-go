package protobuf

import (
	"bytes"
	"encoding/hex"
	"testing"

	"google.golang.org/protobuf/proto"
)

// frsReqWire was serialized by the Python bindings generated from the same
// FrsPageReqIdl.proto, including a nested CommonReq.
const frsReqWire = "0ab2010a0ce5a4a9e5a082e9b8a1e6b1a4101e18232001ba0297010802120931322e36342e312e313a0a637569642d76616c75654080d095ffbc314a0b6d6f64656c2d76616c7565520b62647573732d76616c75655a097462732d76616c7565fa01097a69642d76616c756582022a41334544324437423943464332384538393334413346424433413935373943377c565A35464B423558539a020b4130302d5858582d595959a202083130343530355f33f80206"

func TestFrsPageReqIdlMatchesPython(t *testing.T) {
	wire, err := hex.DecodeString(frsReqWire)
	if err != nil {
		t.Fatalf("decoding hex: %v", err)
	}

	msg := &FrsPageReqIdl{}
	if err := proto.Unmarshal(wire, msg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	data := msg.GetData()
	if data == nil {
		t.Fatal("data is nil")
	}

	checks := []struct {
		name string
		got  any
		want any
	}{
		{"kw", data.GetKw(), "天堂鸡汤"},
		{"pn", data.GetPn(), int32(0)},
		{"rn", data.GetRn(), int32(30)},
		{"rn_need", data.GetRnNeed(), int32(35)},
		{"is_good", data.GetIsGood(), int32(1)},
		{"sort_type", data.GetSortType(), int32(6)},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("data.%s = %v, want %v", c.name, c.got, c.want)
		}
	}

	common := data.GetCommon()
	if common == nil {
		t.Fatal("data.common is nil")
	}
	commonChecks := []struct {
		name string
		got  any
		want any
	}{
		{"_client_type", common.GetXClientType(), int32(2)},
		{"_client_version", common.GetXClientVersion(), "12.64.1.1"},
		{"cuid", common.GetCuid(), "cuid-value"},
		{"BDUSS", common.GetBDUSS(), "bduss-value"},
		{"tbs", common.GetTbs(), "tbs-value"},
		{"cuid_galaxy2", common.GetCuidGalaxy2(), "A3ED2D7B9CFC28E8934A3FBD3A9579C7|VZ5FKB5XS"},
		{"c3_aid", common.GetC3Aid(), "A00-XXX-YYY"},
		{"sample_id", common.GetSampleId(), "104505_3"},
		{"z_id", common.GetZId(), "zid-value"},
		{"_timestamp", common.GetXTimestamp(), int64(1700000000000)},
	}
	for _, c := range commonChecks {
		if c.got != c.want {
			t.Errorf("data.common.%s = %v, want %v", c.name, c.got, c.want)
		}
	}

	got, err := proto.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !bytes.Equal(got, wire) {
		t.Errorf("re-encoded wire = %x, want %x", got, wire)
	}
}
