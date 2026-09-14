// Package getforumdetail implements the get_forum_detail API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_forum_detail.
package getforumdetail

import (
	pb "github.com/rongyuio/aiotieba-go/api/get_forum_detail/protobuf"
)

// ForumDetail is the information of a forum. It mirrors
// aiotieba.api.get_forum_detail._classdef.Forum_detail.
type ForumDetail struct {
	FID   int64
	FName string

	Category string

	SmallAvatar  string
	OriginAvatar string
	Slogan       string
	MemberNum    int64
	PostNum      int64

	HasBawu bool

	Err error
}

// ForumDetailFromProto mirrors Forum_detail.from_proto.
func ForumDetailFromProto(p *pb.GetForumDetailResIdl_DataRes) ForumDetail {
	forum := p.GetForumInfo()
	return ForumDetail{
		FID:          int64(forum.GetForumId()),
		FName:        forum.GetForumName(),
		Category:     forum.GetLv1Name(),
		SmallAvatar:  forum.GetAvatar(),
		OriginAvatar: forum.GetAvatarOrigin(),
		Slogan:       forum.GetSlogan(),
		MemberNum:    int64(forum.GetMemberCount()),
		// post_num is derived from thread_count, mirroring the Python classdef.
		PostNum: int64(forum.GetThreadCount()),
		HasBawu: p.GetElectionTab().GetNewStrategyText() == "已有吧主",
	}
}
