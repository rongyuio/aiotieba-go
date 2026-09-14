// Package getdislikeforums implements the get_dislike_forums API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_dislike_forums.
package getdislikeforums

import (
	"github.com/rongyuio/aiotieba-go/api/classdef"
	pb "github.com/rongyuio/aiotieba-go/api/get_dislike_forums/protobuf"
	"github.com/rongyuio/aiotieba-go/protobuf"
)

// DislikeForum is one forum hidden from the home page recommendations. It
// mirrors aiotieba.api.get_dislike_forums._classdef.DislikeForum.
type DislikeForum struct {
	FID        int64
	FName      string
	MemberNum  int64
	PostNum    int64
	ThreadNum  int64
	IsFollowed bool
}

// DislikeForumFromProto mirrors DislikeForum.from_proto.
//
// is_followed is not reported by this endpoint and stays false.
func DislikeForumFromProto(p *protobuf.ForumList) DislikeForum {
	return DislikeForum{
		FID:       p.GetForumId(),
		FName:     p.GetForumName(),
		MemberNum: int64(p.GetMemberCount()),
		PostNum:   p.GetPostNum(),
		ThreadNum: p.GetThreadNum(),
	}
}

// PageDislikeF is the pagination information of the dislike list. It mirrors
// aiotieba.api.get_dislike_forums._classdef.Page_dislikef.
type PageDislikeF struct {
	CurrentPage int64
	HasMore     bool
	HasPrev     bool
}

// PageDislikeFFromProto mirrors Page_dislikef.from_proto (input: DataRes).
//
// The endpoint does not report has_prev, it is derived from the page number.
func PageDislikeFFromProto(p *pb.GetDislikeListResIdl_DataRes) PageDislikeF {
	currentPage := int64(p.GetCurPage())
	return PageDislikeF{
		CurrentPage: currentPage,
		HasMore:     p.GetHasMore() != 0,
		HasPrev:     currentPage > 1,
	}
}

// DislikeForums is the list of forums hidden from the home page
// recommendations. It mirrors
// aiotieba.api.get_dislike_forums._classdef.DislikeForums.
type DislikeForums struct {
	classdef.Containers[DislikeForum]

	Page PageDislikeF
}

// DislikeForumsFromProto mirrors DislikeForums.from_proto.
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

// HasMore reports whether a next page exists, mirroring the has_more property.
func (d DislikeForums) HasMore() bool { return d.Page.HasMore }
