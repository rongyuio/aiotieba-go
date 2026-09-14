// Package getlastreplyers 实现 aiotieba 的 get_last_replyers API。
//
// 对应 Python 包 aiotieba.api.get_last_replyers。
package getlastreplyers

import (
	"strconv"
	"strings"

	"github.com/rongyuio/aiotieba-go/api/classdef"
	pb "github.com/rongyuio/aiotieba-go/api/get_last_replyers/protobuf"
	"github.com/rongyuio/aiotieba-go/protobuf"
)

// PageLP 页信息。
type PageLP struct {
	PageSize    int64 // 页大小
	CurrentPage int64 // 当前页码
	TotalPage   int64 // 总页码
	TotalCount  int64 // 总计数
	HasMore     bool  // 是否有后继页
	HasPrev     bool  // 是否有前驱页
}

// PageLPFromProto 对应 Page_lp.from_proto。
//
// current_page 为 0 且 page_size 非 0 表示第一页。
func PageLPFromProto(p *protobuf.Page) PageLP {
	currentPage := int64(p.GetCurrentPage())
	if currentPage == 0 && p.GetPageSize() != 0 {
		currentPage = 1
	}

	return PageLP{
		PageSize:    int64(p.GetPageSize()),
		CurrentPage: currentPage,
		TotalPage:   int64(p.GetTotalPage()),
		TotalCount:  int64(p.GetTotalCount()),
		HasMore:     p.GetHasMore() != 0,
		HasPrev:     p.GetHasPrev() != 0,
	}
}

// UserInfoLP 用户信息。
type UserInfoLP struct {
	UserID      int64  // user_id
	Portrait    string // portrait
	UserName    string // 用户名
	NickNameOld string // 旧版昵称
}

// UserInfoLPFromProto 对应 UserInfo_lp.from_proto。
func UserInfoLPFromProto(p *protobuf.User) UserInfoLP {
	portrait := p.GetPortrait()
	// portrait 携带 "?..." 查询后缀，客户端会将其去除。
	if strings.Contains(portrait, "?") && len(portrait) > 13 {
		portrait = portrait[:len(portrait)-13]
	}

	return UserInfoLP{
		UserID:      p.GetId(),
		Portrait:    portrait,
		UserName:    p.GetName(),
		NickNameOld: p.GetNameShow(),
	}
}

// NickName 用户昵称。
func (u UserInfoLP) NickName() string { return u.NickNameOld }

// ShowName 显示名称。
func (u UserInfoLP) ShowName() string {
	if u.NickNameOld != "" {
		return u.NickNameOld
	}
	return u.UserName
}

// String 对应 __str__。
func (u UserInfoLP) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// LogName 用于在日志中记录用户信息。
func (u UserInfoLP) LogName() string {
	switch {
	case u.UserName != "":
		return u.UserName
	case u.Portrait != "":
		return u.NickNameOld + "/" + u.Portrait
	default:
		return strconv.FormatInt(u.UserID, 10)
	}
}

// LastReplyer 最后回复者的用户信息。
type LastReplyer struct {
	UserID      int64  // user_id
	UserName    string // 用户名
	NickNameOld string // 旧版昵称
}

// LastReplyerFromProto 对应 LastReplyer.from_proto。
func LastReplyerFromProto(p *protobuf.User) LastReplyer {
	return LastReplyer{
		UserID:      p.GetId(),
		UserName:    p.GetName(),
		NickNameOld: p.GetNameShow(),
	}
}

// NickName 用户昵称。
func (l LastReplyer) NickName() string { return l.NickNameOld }

// ShowName 显示名称。
func (l LastReplyer) ShowName() string {
	if l.NickNameOld != "" {
		return l.NickNameOld
	}
	return l.UserName
}

// String 对应 __str__。
func (l LastReplyer) String() string {
	if l.UserName != "" {
		return l.UserName
	}
	return strconv.FormatInt(l.UserID, 10)
}

// LogName 用于在日志中记录用户信息。
func (l LastReplyer) LogName() string {
	if l.UserName != "" {
		return l.UserName
	}
	return strconv.FormatInt(l.UserID, 10)
}

// ThreadLP 主题帖信息。
//
// FID 和 FName 由 ThreadsLPFromProto 填充，因为它知道所属吧。
type ThreadLP struct {
	Title string // 标题内容
	FID   int64  // 所在吧id
	FName string // 所在贴吧名
	TID   int64  // 主题帖tid
	PID   int64  // 首楼回复pid

	User        UserInfoLP  // 发布者的用户信息
	LastReplyer LastReplyer // 最后回复者的用户信息

	IsGood     bool  // 是否精品帖
	IsTop      bool  // 是否置顶帖
	CreateTime int64 // 创建时间 10位时间戳 以秒为单位
	LastTime   int64 // 最后回复时间 10位时间戳 以秒为单位
}

// ThreadLPFromProto 对应 Thread_lp.from_proto。
func ThreadLPFromProto(p *protobuf.ThreadInfo) ThreadLP {
	return ThreadLP{
		Title:       p.GetTitle(),
		TID:         p.GetId(),
		PID:         p.GetFirstPostId(),
		User:        UserInfoLPFromProto(p.GetAuthor()),
		LastReplyer: LastReplyerFromProto(p.GetLastReplyer()),
		IsGood:      p.GetIsGood() != 0,
		IsTop:       p.GetIsTop() != 0,
		CreateTime:  int64(p.GetCreateTime()),
		LastTime:    int64(p.GetLastTimeInt()),
	}
}

// Text 文本内容。
func (t ThreadLP) Text() string { return t.Title }

// AuthorID 发布者的user_id。
func (t ThreadLP) AuthorID() int64 { return t.User.UserID }

// ForumLP 吧信息。
type ForumLP struct {
	FID   int64  // 贴吧id
	FName string // 贴吧名
}

// ForumLPFromProto 对应 Forum_lp.from_proto。
func ForumLPFromProto(p *pb.FrsPageResIdl4Lp_DataRes) ForumLP {
	forum := p.GetForum()
	return ForumLP{
		FID:   forum.GetId(),
		FName: forum.GetName(),
	}
}

// ThreadsLP 主题帖列表。
type ThreadsLP struct {
	classdef.Containers[ThreadLP]

	Page  PageLP  // 页信息
	Forum ForumLP // 所在吧信息
}

// ThreadsLPFromProto 对应 Threads_lp.from_proto。
func ThreadsLPFromProto(p *pb.FrsPageResIdl4Lp_DataRes) ThreadsLP {
	page := PageLPFromProto(p.GetPage())
	forum := ForumLPFromProto(p)

	list := p.GetThreadList()
	objs := make([]ThreadLP, 0, len(list))
	for _, item := range list {
		thread := ThreadLPFromProto(item)
		// 主题帖列表不携带吧信息，由外层继承。
		thread.FName = forum.FName
		thread.FID = forum.FID
		objs = append(objs, thread)
	}

	return ThreadsLP{
		Containers: classdef.Containers[ThreadLP]{Objs: objs},
		Page:       page,
		Forum:      forum,
	}
}

// HasMore 是否还有下一页。
func (t ThreadsLP) HasMore() bool { return t.Page.HasMore }
