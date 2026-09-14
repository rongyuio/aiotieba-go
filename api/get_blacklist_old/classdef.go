// Package getblacklistold 实现 aiotieba 的 get_blacklist_old API。
//
// 对应 Python 包 aiotieba.api.get_blacklist_old。
package getblacklistold

import (
	"strconv"
	"strings"

	"github.com/rongyuio/aiotieba-go/api/classdef"
	pb "github.com/rongyuio/aiotieba-go/api/get_blacklist_old/protobuf"
	"github.com/rongyuio/aiotieba-go/protobuf"
)

// BlacklistOldUser 用户信息。
type BlacklistOldUser struct {
	UserID      int64  // user_id
	Portrait    string // portrait
	UserName    string // 用户名
	NickNameOld string // 旧版昵称
	UntilTime   int64  // 解禁时间 10位时间戳 以秒为单位
}

// BlacklistOldUserFromProto 对应 BlacklistOldUser.from_proto。
func BlacklistOldUserFromProto(p *pb.UserMuteQueryResIdl_DataRes_MuteUser) BlacklistOldUser {
	portrait := p.GetPortrait()
	// portrait 带有 "?..." 查询后缀，客户端会将其去除。
	if strings.Contains(portrait, "?") && len(portrait) > 13 {
		portrait = portrait[:len(portrait)-13]
	}

	return BlacklistOldUser{
		UserID:      p.GetUserId(),
		Portrait:    portrait,
		UserName:    p.GetUserName(),
		NickNameOld: p.GetNameShow(),
		UntilTime:   int64(p.GetMuteTime()),
	}
}

// NickName 用户昵称。
func (u BlacklistOldUser) NickName() string { return u.NickNameOld }

// String 对应 __str__。
func (u BlacklistOldUser) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// LogName 用于在日志中记录用户信息。
func (u BlacklistOldUser) LogName() string {
	switch {
	case u.UserName != "":
		return u.UserName
	case u.Portrait != "":
		return u.NickNameOld + "/" + u.Portrait
	default:
		return strconv.FormatInt(u.UserID, 10)
	}
}

// Valid 对应 __bool__。
func (u BlacklistOldUser) Valid() bool { return u.UserID != 0 }

// PageBlacklist 页信息。
type PageBlacklist struct {
	CurrentPage int64 // 当前页码
	HasMore     bool  // 是否有后继页
	HasPrev     bool  // 是否有前驱页
}

// PageBlacklistFromProto 对应 Page_blacklist.from_proto。
func PageBlacklistFromProto(p *protobuf.Page) PageBlacklist {
	return PageBlacklist{
		CurrentPage: int64(p.GetCurrentPage()),
		HasMore:     p.GetHasMore() != 0,
		HasPrev:     p.GetHasPrev() != 0,
	}
}

// BlacklistOldUsers 旧版用户黑名单列表。
type BlacklistOldUsers struct {
	classdef.Containers[BlacklistOldUser]

	Page PageBlacklist // 页信息
}

// BlacklistOldUsersFromProto 对应 BlacklistOldUsers.from_proto。
func BlacklistOldUsersFromProto(p *pb.UserMuteQueryResIdl_DataRes) BlacklistOldUsers {
	list := p.GetMuteUser()
	objs := make([]BlacklistOldUser, 0, len(list))
	for _, item := range list {
		objs = append(objs, BlacklistOldUserFromProto(item))
	}

	return BlacklistOldUsers{
		Containers: classdef.Containers[BlacklistOldUser]{Objs: objs},
		Page:       PageBlacklistFromProto(p.GetPage()),
	}
}

// HasMore 是否还有下一页。
func (b BlacklistOldUsers) HasMore() bool { return b.Page.HasMore }
