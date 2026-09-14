// Package getlastreplyers implements the get_last_replyers API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_last_replyers.
package getlastreplyers

import (
	"strconv"
	"strings"

	"github.com/rongyuio/aiotieba/api/classdef"
	pb "github.com/rongyuio/aiotieba/api/get_last_replyers/protobuf"
	"github.com/rongyuio/aiotieba/protobuf"
)

// PageLP is the pagination information of the thread list. It mirrors
// aiotieba.api.get_last_replyers._classdef.Page_lp.
type PageLP struct {
	PageSize    int64
	CurrentPage int64
	TotalPage   int64
	TotalCount  int64
	HasMore     bool
	HasPrev     bool
}

// PageLPFromProto mirrors Page_lp.from_proto.
//
// A zero current_page with a non-zero page_size means the first page.
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

// UserInfoLP is the information of the thread author. It mirrors
// aiotieba.api.get_last_replyers._classdef.UserInfo_lp.
type UserInfoLP struct {
	UserID      int64
	Portrait    string
	UserName    string
	NickNameOld string
}

// UserInfoLPFromProto mirrors UserInfo_lp.from_proto.
func UserInfoLPFromProto(p *protobuf.User) UserInfoLP {
	portrait := p.GetPortrait()
	// The portrait carries a "?..." query suffix that the client strips.
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

// NickName mirrors the nick_name property.
func (u UserInfoLP) NickName() string { return u.NickNameOld }

// ShowName mirrors the show_name property.
func (u UserInfoLP) ShowName() string {
	if u.NickNameOld != "" {
		return u.NickNameOld
	}
	return u.UserName
}

// String mirrors __str__.
func (u UserInfoLP) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// LogName mirrors the log_name property.
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

// LastReplyer is the information of the user who replied last. It mirrors
// aiotieba.api.get_last_replyers._classdef.LastReplyer.
type LastReplyer struct {
	UserID      int64
	UserName    string
	NickNameOld string
}

// LastReplyerFromProto mirrors LastReplyer.from_proto.
func LastReplyerFromProto(p *protobuf.User) LastReplyer {
	return LastReplyer{
		UserID:      p.GetId(),
		UserName:    p.GetName(),
		NickNameOld: p.GetNameShow(),
	}
}

// NickName mirrors the nick_name property.
func (l LastReplyer) NickName() string { return l.NickNameOld }

// ShowName mirrors the show_name property.
func (l LastReplyer) ShowName() string {
	if l.NickNameOld != "" {
		return l.NickNameOld
	}
	return l.UserName
}

// String mirrors __str__.
func (l LastReplyer) String() string {
	if l.UserName != "" {
		return l.UserName
	}
	return strconv.FormatInt(l.UserID, 10)
}

// LogName mirrors the log_name property.
func (l LastReplyer) LogName() string {
	if l.UserName != "" {
		return l.UserName
	}
	return strconv.FormatInt(l.UserID, 10)
}

// ThreadLP is one thread of the list. It mirrors
// aiotieba.api.get_last_replyers._classdef.Thread_lp.
//
// FID and FName are filled in by ThreadsLPFromProto, which knows the forum.
type ThreadLP struct {
	Title string
	FID   int64
	FName string
	TID   int64
	PID   int64

	User        UserInfoLP
	LastReplyer LastReplyer

	IsGood     bool
	IsTop      bool
	CreateTime int64
	LastTime   int64
}

// ThreadLPFromProto mirrors Thread_lp.from_proto.
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

// Text mirrors the text property.
func (t ThreadLP) Text() string { return t.Title }

// AuthorID mirrors the author_id property.
func (t ThreadLP) AuthorID() int64 { return t.User.UserID }

// ForumLP is the forum the threads belong to. It mirrors
// aiotieba.api.get_last_replyers._classdef.Forum_lp.
type ForumLP struct {
	FID   int64
	FName string
}

// ForumLPFromProto mirrors Forum_lp.from_proto.
func ForumLPFromProto(p *pb.FrsPageResIdl4Lp_DataRes) ForumLP {
	forum := p.GetForum()
	return ForumLP{
		FID:   forum.GetId(),
		FName: forum.GetName(),
	}
}

// ThreadsLP is the thread list ordered by the last reply time. It mirrors
// aiotieba.api.get_last_replyers._classdef.Threads_lp.
type ThreadsLP struct {
	classdef.Containers[ThreadLP]

	Page  PageLP
	Forum ForumLP
}

// ThreadsLPFromProto mirrors Threads_lp.from_proto.
func ThreadsLPFromProto(p *pb.FrsPageResIdl4Lp_DataRes) ThreadsLP {
	page := PageLPFromProto(p.GetPage())
	forum := ForumLPFromProto(p)

	list := p.GetThreadList()
	objs := make([]ThreadLP, 0, len(list))
	for _, item := range list {
		thread := ThreadLPFromProto(item)
		// The thread list does not carry the forum, it is inherited.
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

// HasMore reports whether a next page exists, mirroring the has_more property.
func (t ThreadsLP) HasMore() bool { return t.Page.HasMore }
