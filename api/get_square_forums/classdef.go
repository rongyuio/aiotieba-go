// Package getsquareforums implements the get_square_forums API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_square_forums.
package getsquareforums

import (
	"github.com/rongyuio/aiotieba/api/classdef"
	pb "github.com/rongyuio/aiotieba/api/get_square_forums/protobuf"
	"github.com/rongyuio/aiotieba/protobuf"
)

// SquareForum is the information of one forum of the forum square. It mirrors
// aiotieba.api.get_square_forums._classdef.SquareForum.
type SquareForum struct {
	FID        int64
	FName      string
	MemberNum  int64
	PostNum    int64
	IsFollowed bool
}

// SquareForumFromProto mirrors SquareForum.from_proto.
func SquareForumFromProto(p *pb.GetForumSquareResIdl_DataRes_RecommendForumInfo) SquareForum {
	return SquareForum{
		FID:       int64(p.GetForumId()),
		FName:     p.GetForumName(),
		MemberNum: int64(p.GetMemberCount()),
		// post_num is derived from thread_count, mirroring the Python classdef.
		PostNum:    int64(p.GetThreadCount()),
		IsFollowed: p.GetIsLike() != 0,
	}
}

// PageSquare is the pagination information of the forum square. It mirrors
// aiotieba.api.get_square_forums._classdef.Page_square.
type PageSquare struct {
	PageSize    int64
	CurrentPage int64
	TotalPage   int64
	TotalCount  int64
	HasMore     bool
	HasPrev     bool
}

// PageSquareFromProto mirrors Page_square.from_proto.
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

// SquareForums is the forum square list. It mirrors
// aiotieba.api.get_square_forums._classdef.SquareForums.
type SquareForums struct {
	classdef.Containers[SquareForum]

	Page PageSquare
}

// SquareForumsFromProto mirrors SquareForums.from_proto.
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

// HasMore reports whether a next page exists, mirroring the has_more property.
func (s SquareForums) HasMore() bool { return s.Page.HasMore }
