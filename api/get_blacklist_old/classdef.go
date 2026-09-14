// Package getblacklistold implements the get_blacklist_old API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_blacklist_old.
package getblacklistold

import (
	"strconv"
	"strings"

	"github.com/rongyuio/aiotieba/api/classdef"
	pb "github.com/rongyuio/aiotieba/api/get_blacklist_old/protobuf"
	"github.com/rongyuio/aiotieba/protobuf"
)

// BlacklistOldUser is one muted user of the legacy blacklist. It mirrors
// aiotieba.api.get_blacklist_old._classdef.BlacklistOldUser.
type BlacklistOldUser struct {
	UserID      int64
	Portrait    string
	UserName    string
	NickNameOld string
	UntilTime   int64
}

// BlacklistOldUserFromProto mirrors BlacklistOldUser.from_proto.
func BlacklistOldUserFromProto(p *pb.UserMuteQueryResIdl_DataRes_MuteUser) BlacklistOldUser {
	portrait := p.GetPortrait()
	// The portrait carries a "?..." query suffix that the client strips.
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

// NickName mirrors the nick_name property.
func (u BlacklistOldUser) NickName() string { return u.NickNameOld }

// String mirrors __str__.
func (u BlacklistOldUser) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// LogName mirrors the log_name property.
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

// Valid mirrors __bool__.
func (u BlacklistOldUser) Valid() bool { return u.UserID != 0 }

// PageBlacklist is the pagination information of the legacy blacklist. It
// mirrors aiotieba.api.get_blacklist_old._classdef.Page_blacklist.
type PageBlacklist struct {
	CurrentPage int64
	HasMore     bool
	HasPrev     bool
}

// PageBlacklistFromProto mirrors Page_blacklist.from_proto.
func PageBlacklistFromProto(p *protobuf.Page) PageBlacklist {
	return PageBlacklist{
		CurrentPage: int64(p.GetCurrentPage()),
		HasMore:     p.GetHasMore() != 0,
		HasPrev:     p.GetHasPrev() != 0,
	}
}

// BlacklistOldUsers is the legacy user blacklist. It mirrors
// aiotieba.api.get_blacklist_old._classdef.BlacklistOldUsers.
type BlacklistOldUsers struct {
	classdef.Containers[BlacklistOldUser]

	Page PageBlacklist
}

// BlacklistOldUsersFromProto mirrors BlacklistOldUsers.from_proto.
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

// HasMore reports whether a next page exists, mirroring the has_more property.
func (b BlacklistOldUsers) HasMore() bool { return b.Page.HasMore }
