package classdef

import "github.com/rongyuio/aiotieba-go/protobuf"

// VoteOption is one option of a poll. It mirrors
// aiotieba.api._classdef.vote.VoteOption.
type VoteOption struct {
	VoteNum int64
	Text    string
}

// VoteOptionFromProto builds a VoteOption from a PollInfo_PollOption.
func VoteOptionFromProto(p *protobuf.PollInfo_PollOption) VoteOption {
	return VoteOption{VoteNum: p.GetNum(), Text: p.GetText()}
}

// VoteInfo is the poll of a thread. It mirrors
// aiotieba.api._classdef.vote.VoteInfo.
type VoteInfo struct {
	Title     string
	IsMulti   bool
	Options   []VoteOption
	TotalVote int64
	TotalUser int64
}

// VoteInfoFromProto builds a VoteInfo from a PollInfo.
func VoteInfoFromProto(p *protobuf.PollInfo) VoteInfo {
	if p == nil {
		return VoteInfo{}
	}
	options := make([]VoteOption, 0, len(p.GetOptions()))
	for _, opt := range p.GetOptions() {
		options = append(options, VoteOptionFromProto(opt))
	}
	return VoteInfo{
		Title:     p.GetTitle(),
		IsMulti:   p.GetIsMulti() != 0,
		Options:   options,
		TotalVote: p.GetTotalPoll(),
		TotalUser: p.GetTotalNum(),
	}
}

// Len returns the number of options. It mirrors __len__.
func (v VoteInfo) Len() int { return len(v.Options) }

// Valid reports whether the poll has any option. It mirrors __bool__.
func (v VoteInfo) Valid() bool { return len(v.Options) != 0 }
