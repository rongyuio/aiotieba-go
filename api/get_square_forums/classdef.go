// Package getsquareforums 实现 aiotieba 的 get_square_forums API。
//
// 对应 Python 包 aiotieba.api.get_square_forums。
package getsquareforums

import (
	"github.com/rongyuio/aiotieba-go/api/classdef"
	pb "github.com/rongyuio/aiotieba-go/api/get_square_forums/protobuf"
	"github.com/rongyuio/aiotieba-go/protobuf"
)

// SquareForum 吧广场贴吧信息。
type SquareForum struct {
	FID        int64  // 贴吧id
	FName      string // 贴吧名
	MemberNum  int64  // 吧会员数
	PostNum    int64  // 发帖数
	IsFollowed bool   // 是否已关注
}

// SquareForumFromProto 对应 SquareForum.from_proto。
func SquareForumFromProto(p *pb.GetForumSquareResIdl_DataRes_RecommendForumInfo) SquareForum {
	return SquareForum{
		FID:       int64(p.GetForumId()),
		FName:     p.GetForumName(),
		MemberNum: int64(p.GetMemberCount()),
		// post_num 由 thread_count 推导，对应 Python classdef。
		PostNum:    int64(p.GetThreadCount()),
		IsFollowed: p.GetIsLike() != 0,
	}
}

// PageSquare 页信息。
type PageSquare struct {
	PageSize    int64 // 页大小
	CurrentPage int64 // 当前页码
	TotalPage   int64 // 总页码
	TotalCount  int64 // 总计数
	HasMore     bool  // 是否有后继页
	HasPrev     bool  // 是否有前驱页
}

// PageSquareFromProto 对应 Page_square.from_proto。
func PageSquareFromProto(p *protobuf.Page) PageSquare {
	return PageSquare{
		PageSize:    int64(p.GetPageSize()),
		CurrentPage: int64(p.GetCurrentPage()),
		TotalPage:   int64(p.GetTotalPage()),
		TotalCount:  int64(p.GetTotalCount()),
		HasMore:     p.GetHasMore() != 0,
		HasPrev:     p.GetHasPrev() != 0,
	}
}

// SquareForums 吧广场列表。
type SquareForums struct {
	classdef.Containers[SquareForum]

	Page PageSquare // 页信息
}

// SquareForumsFromProto 对应 SquareForums.from_proto。
func SquareForumsFromProto(p *pb.GetForumSquareResIdl_DataRes) SquareForums {
	infos := p.GetForumInfo()
	objs := make([]SquareForum, 0, len(infos))
	for _, info := range infos {
		objs = append(objs, SquareForumFromProto(info))
	}

	return SquareForums{
		Containers: classdef.Containers[SquareForum]{Objs: objs},
		Page:       PageSquareFromProto(p.GetPage()),
	}
}

// HasMore 是否还有下一页。
func (s SquareForums) HasMore() bool { return s.Page.HasMore }
