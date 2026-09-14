// Package getdislikeforums 实现 aiotieba 的 get_dislike_forums API。
//
// 对应 Python 包 aiotieba.api.get_dislike_forums。
package getdislikeforums

import (
	"github.com/rongyuio/aiotieba-go/api/classdef"
	pb "github.com/rongyuio/aiotieba-go/api/get_dislike_forums/protobuf"
	"github.com/rongyuio/aiotieba-go/protobuf"
)

// DislikeForum 首页推荐屏蔽的贴吧信息。
type DislikeForum struct {
	FID        int64  // 贴吧id
	FName      string // 贴吧名
	MemberNum  int64  // 吧会员数
	PostNum    int64  // 发帖数
	ThreadNum  int64  // 主题帖数
	IsFollowed bool
}

// DislikeForumFromProto 对应 DislikeForum.from_proto。
//
// is_followed 不被该接口上报，保持为 false。
func DislikeForumFromProto(p *protobuf.ForumList) DislikeForum {
	return DislikeForum{
		FID:       p.GetForumId(),
		FName:     p.GetForumName(),
		MemberNum: int64(p.GetMemberCount()),
		PostNum:   p.GetPostNum(),
		ThreadNum: p.GetThreadNum(),
	}
}

// PageDislikeF 页信息。
type PageDislikeF struct {
	CurrentPage int64 // 当前页码
	HasMore     bool  // 是否有后继页
	HasPrev     bool  // 是否有前驱页
}

// PageDislikeFFromProto 对应 Page_dislikef.from_proto（输入：DataRes）。
//
// 该接口不上报 has_prev，它由页码推导得出。
func PageDislikeFFromProto(p *pb.GetDislikeListResIdl_DataRes) PageDislikeF {
	currentPage := int64(p.GetCurPage())
	return PageDislikeF{
		CurrentPage: currentPage,
		HasMore:     p.GetHasMore() != 0,
		HasPrev:     currentPage > 1,
	}
}

// DislikeForums 首页推荐屏蔽的贴吧列表。
type DislikeForums struct {
	classdef.Containers[DislikeForum]

	Page PageDislikeF // 页信息
}

// DislikeForumsFromProto 对应 DislikeForums.from_proto。
func DislikeForumsFromProto(p *pb.GetDislikeListResIdl_DataRes) DislikeForums {
	list := p.GetForumList()
	objs := make([]DislikeForum, 0, len(list))
	for _, forum := range list {
		objs = append(objs, DislikeForumFromProto(forum))
	}

	return DislikeForums{
		Containers: classdef.Containers[DislikeForum]{Objs: objs},
		Page:       PageDislikeFFromProto(p),
	}
}

// HasMore 是否还有下一页。
func (d DislikeForums) HasMore() bool { return d.Page.HasMore }
