// Package getforumdetail 实现 aiotieba 的 get_forum_detail API。
//
// 对应 Python 包 aiotieba.api.get_forum_detail。
package getforumdetail

import (
	pb "github.com/rongyuio/aiotieba-go/api/get_forum_detail/protobuf"
)

// ForumDetail 贴吧信息。
type ForumDetail struct {
	FID   int64  // 贴吧id
	FName string // 贴吧名

	Category string // 一级分类

	SmallAvatar  string // 吧头像(小)
	OriginAvatar string // 吧头像(原图)
	Slogan       string // 吧标语
	MemberNum    int64  // 吧会员数
	PostNum      int64  // 发帖数

	HasBawu bool // 是否有吧务

	Err error // 捕获的异常
}

// ForumDetailFromProto 对应 Forum_detail.from_proto。
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
		// post_num 由 thread_count 推导，对应 Python 的 classdef。
		PostNum: int64(forum.GetThreadCount()),
		HasBawu: p.GetElectionTab().GetNewStrategyText() == "已有吧主",
	}
}
