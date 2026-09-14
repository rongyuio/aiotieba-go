package classdef

import (
	"strconv"

	"github.com/rongyuio/aiotieba-go/enums"
)

// UserInfo 用户信息，对应 aiotieba.api._classdef.user.UserInfo。
type UserInfo struct {
	UserID      int64  // user_id
	Portrait    string // portrait
	UserName    string // 用户名
	NickNameOld string // 旧版昵称
	NickNameNew string // 新版昵称
	TiebaUID    int64  // 用户个人主页uid

	GLevel int32        // 贴吧成长等级
	Gender enums.Gender // 性别
	Age    float64      // 吧龄 以年为单位

	// 计数类字段使用 int64，因为服务端以 64 位值上报（仅 total_agree_num 就可能超过 math.MaxInt32）。
	PostNum   int64 // 发帖数
	AgreeNum  int64 // 获赞数
	FanNum    int64 // 粉丝数
	FollowNum int64 // 关注数
	ForumNum  int64 // 关注贴吧数

	Sign  string   // 个性签名
	IP    string   // ip归属地
	Icons []string // 印记信息

	IsVIP     bool // 是否超级会员
	IsGod     bool // 是否大神
	IsBlocked bool // 是否被永久封禁屏蔽
	UK        int64
	BDUK      string
	TriggerID int64
	PrivLike  enums.PrivLike  // 关注吧列表的公开状态
	PrivReply enums.PrivReply // 帖子评论权限
}

// String 对应 __str__。
func (u UserInfo) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// NickName 用户昵称，对应 nick_name 属性。
func (u UserInfo) NickName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.NickNameOld
}

// ShowName 显示名称，对应 show_name 属性。
func (u UserInfo) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	if u.NickNameOld != "" {
		return u.NickNameOld
	}
	return u.UserName
}

// LogName 用于在日志中记录用户信息，对应 log_name 属性。
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

// Equal 对应 __eq__：两个用户 user_id 相同时视为相等。
func (u UserInfo) Equal(other UserInfo) bool {
	return u.UserID == other.UserID
}

// Valid 对应 __bool__。
func (u UserInfo) Valid() bool { return u.UserID != 0 }

// MergeFrom 用 other 的值覆盖所有字段，对应 __ior__。
func (u *UserInfo) MergeFrom(other UserInfo) {
	*u = other
}
