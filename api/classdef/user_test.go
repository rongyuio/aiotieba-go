package classdef

import (
	"testing"

	"github.com/rongyuio/aiotieba/protobuf"
)

func TestUserInfoNameHelpers(t *testing.T) {
	cases := []struct {
		name  string
		user  UserInfo
		nick  string
		show  string
		log   string
		str   string
		valid bool
	}{
		{
			name:  "full",
			user:  UserInfo{UserID: 1, UserName: "uname", NickNameNew: "new", NickNameOld: "old", Portrait: "tb.1.x"},
			nick:  "new",
			show:  "new",
			log:   "uname",
			str:   "uname",
			valid: true,
		},
		{
			name:  "old nick only",
			user:  UserInfo{UserID: 2, UserName: "uname", NickNameOld: "old", Portrait: "tb.1.x"},
			nick:  "old",
			show:  "old",
			log:   "uname",
			str:   "uname",
			valid: true,
		},
		{
			name:  "portrait only",
			user:  UserInfo{UserID: 3, Portrait: "tb.1.x", NickNameNew: "new"},
			nick:  "new",
			show:  "new",
			log:   "new/tb.1.x",
			str:   "tb.1.x",
			valid: true,
		},
		{
			name:  "id only",
			user:  UserInfo{UserID: 4},
			nick:  "",
			show:  "",
			log:   "4",
			str:   "4",
			valid: true,
		},
		{
			name:  "zero",
			user:  UserInfo{},
			nick:  "",
			show:  "",
			log:   "0",
			str:   "0",
			valid: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.user.NickName(); got != tc.nick {
				t.Errorf("NickName = %q, want %q", got, tc.nick)
			}
			if got := tc.user.ShowName(); got != tc.show {
				t.Errorf("ShowName = %q, want %q", got, tc.show)
			}
			if got := tc.user.LogName(); got != tc.log {
				t.Errorf("LogName = %q, want %q", got, tc.log)
			}
			if got := tc.user.String(); got != tc.str {
				t.Errorf("String = %q, want %q", got, tc.str)
			}
			if got := tc.user.Valid(); got != tc.valid {
				t.Errorf("Valid = %v, want %v", got, tc.valid)
			}
		})
	}
}

func TestUserInfoEqualAndMerge(t *testing.T) {
	a := UserInfo{UserID: 1, UserName: "a"}
	b := UserInfo{UserID: 1, UserName: "b"}
	c := UserInfo{UserID: 2, UserName: "c"}

	if !a.Equal(b) {
		t.Error("users with the same id must be equal")
	}
	if a.Equal(c) {
		t.Error("users with different ids must not be equal")
	}

	a.MergeFrom(c)
	if a.UserID != c.UserID || a.UserName != c.UserName {
		t.Errorf("MergeFrom result = %+v, want %+v", a, c)
	}
}

func TestVoteInfoFromProto(t *testing.T) {
	p := &protobuf.PollInfo{
		Title:     "你选哪个",
		IsMulti:   1,
		TotalNum:  20,
		TotalPoll: 35,
		Options: []*protobuf.PollInfo_PollOption{
			{Num: 20, Text: "A"},
			{Num: 15, Text: "B"},
		},
	}

	got := VoteInfoFromProto(p)
	if got.Title != "你选哪个" || !got.IsMulti {
		t.Errorf("vote info = %+v", got)
	}
	if got.TotalVote != 35 || got.TotalUser != 20 {
		t.Errorf("totals = %d/%d, want 35/20", got.TotalVote, got.TotalUser)
	}
	if got.Len() != 2 || !got.Valid() {
		t.Errorf("Len/Valid = %d/%v", got.Len(), got.Valid())
	}
	if got.Options[0].VoteNum != 20 || got.Options[0].Text != "A" {
		t.Errorf("option 0 = %+v", got.Options[0])
	}
}

func TestVoteInfoFromNilProto(t *testing.T) {
	got := VoteInfoFromProto(nil)
	if got.Valid() || got.Len() != 0 {
		t.Errorf("VoteInfoFromProto(nil) = %+v, want an empty poll", got)
	}
}
