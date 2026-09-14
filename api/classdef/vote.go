package classdef

import "github.com/rongyuio/aiotieba-go/protobuf"

// VoteOption 投票选项信息。
type VoteOption struct {
	VoteNum int64  // 得票数
	Text    string // 选项描述文字
}

// VoteOptionFromProto 由 PollInfo_PollOption 构造 VoteOption。
func VoteOptionFromProto(p *protobuf.PollInfo_PollOption) VoteOption {
	return VoteOption{VoteNum: p.GetNum(), Text: p.GetText()}
}

// VoteInfo 投票信息。
type VoteInfo struct {
	Title     string       // 投票标题
	IsMulti   bool         // 是否多选
	Options   []VoteOption // 选项列表
	TotalVote int64        // 总投票数
	TotalUser int64        // 总投票人数
}

// VoteInfoFromProto 由 PollInfo 构造 VoteInfo。
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

// Len 返回选项数量，对应 __len__。
func (v VoteInfo) Len() int { return len(v.Options) }

// Valid 报告投票是否包含选项，对应 __bool__。
func (v VoteInfo) Valid() bool { return len(v.Options) != 0 }
