package classdef

import (
	"strconv"

	"github.com/rongyuio/aiotieba/enums"
)

// UserInfo is the user information shared by most APIs. It mirrors
// aiotieba.api._classdef.user.UserInfo.
type UserInfo struct {
	UserID      int64
	Portrait    string
	UserName    string
	NickNameOld string
	NickNameNew string
	TiebaUID    int64

	GLevel int32
	Gender enums.Gender
	Age    float64

	// The counters are int64 because the server reports them as 64 bit values
	// (total_agree_num alone may exceed math.MaxInt32).
	PostNum   int64
	AgreeNum  int64
	FanNum    int64
	FollowNum int64
	ForumNum  int64

	Sign  string
	IP    string
	Icons []string

	IsVIP     bool
	IsGod     bool
	IsBlocked bool
	UK        int64
	BDUK      string
	TriggerID int64
	PrivLike  enums.PrivLike
	PrivReply enums.PrivReply
}

// String mirrors __str__.
func (u UserInfo) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// NickName mirrors the nick_name property.
func (u UserInfo) NickName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.NickNameOld
}

// ShowName mirrors the show_name property.
func (u UserInfo) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	if u.NickNameOld != "" {
		return u.NickNameOld
	}
	return u.UserName
}

// LogName mirrors the log_name property.
func (u UserInfo) LogName() string {
	switch {
	case u.UserName != "":
		return u.UserName
	case u.Portrait != "":
		return u.NickName() + "/" + u.Portrait
	default:
		return strconv.FormatInt(u.UserID, 10)
	}
}

// Equal mirrors __eq__: two users are equal when their user ids match.
func (u UserInfo) Equal(other UserInfo) bool {
	return u.UserID == other.UserID
}

// Valid mirrors __bool__.
func (u UserInfo) Valid() bool { return u.UserID != 0 }

// MergeFrom overwrites every field with the values of other, mirroring __ior__.
func (u *UserInfo) MergeFrom(other UserInfo) {
	*u = other
}
