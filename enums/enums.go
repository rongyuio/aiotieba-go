// Package enums 定义库中使用的枚举类型。
//
// 对应 Python 模块 aiotieba.enums。Python 的 IntEnum 用「具名整数类型 + 每个成员一个常量」
// 建模，Python 的 `_missing_` 回退（未知值 -> UNKNOWN）用 From 构造函数建模。
package enums

// Gender 用户性别。
type Gender int

const (
	GenderUnknown Gender = 0 // 未知
	GenderMale    Gender = 1 // 男性
	GenderFemale  Gender = 2 // 女性
)

// GenderFrom 把 v 映射为已知的 Gender，默认返回 GenderUnknown。
//
// Python 的 Gender 没有 _missing_ 回退，遇到意外值会抛 ValueError；本移植版改为默认
// GenderUnknown，与其他 From 构造函数保持一致，避免服务端返回未知值时直接失败。
func GenderFrom(v int) Gender {
	switch Gender(v) {
	case GenderMale, GenderFemale:
		return Gender(v)
	default:
		return GenderUnknown
	}
}

// PrivLike 关注吧列表的公开状态。
type PrivLike int

const (
	PrivLikeUnknown PrivLike = 0 // 未知
	PrivLikePublic  PrivLike = 1 // 所有人可见
	PrivLikeFriend  PrivLike = 2 // 好友可见
	PrivLikeHide    PrivLike = 3 // 完全隐藏
)

// PrivLikeFrom 把 v 映射为已知的 PrivLike，默认返回 PrivLikeUnknown。
func PrivLikeFrom(v int) PrivLike {
	switch PrivLike(v) {
	case PrivLikePublic, PrivLikeFriend, PrivLikeHide:
		return PrivLike(v)
	default:
		return PrivLikeUnknown
	}
}

// PrivReply 帖子评论权限。
type PrivReply int

const (
	PrivReplyUnknown PrivReply = 0 // 未知
	PrivReplyAll     PrivReply = 1 // 允许所有人
	PrivReplyFans    PrivReply = 5 // 仅允许我的粉丝
	PrivReplyFollow  PrivReply = 6 // 仅允许我的关注
)

// PrivReplyFrom 把 v 映射为已知的 PrivReply，默认返回 PrivReplyUnknown。
func PrivReplyFrom(v int) PrivReply {
	switch PrivReply(v) {
	case PrivReplyAll, PrivReplyFans, PrivReplyFollow:
		return PrivReply(v)
	default:
		return PrivReplyUnknown
	}
}

// ThreadType 主题帖类型。
type ThreadType int

const (
	ThreadTypeUnknown  ThreadType = -1 // 未知
	ThreadTypeArticle  ThreadType = 0  // 图文帖
	ThreadTypeAlbum    ThreadType = 1  // 相册帖
	ThreadTypeExtShare ThreadType = 6  // 来自外部网站的分享帖
	ThreadTypeVoice    ThreadType = 11 // 语音帖
	ThreadTypeNetdisk  ThreadType = 14 // 网盘分享帖
	ThreadTypeStory    ThreadType = 31 // 会员小说帖
	ThreadTypeVideo    ThreadType = 40 // 视频帖
	ThreadTypeLive     ThreadType = 50 // 直播帖
	ThreadTypeHelp     ThreadType = 71 // 求助帖
	ThreadTypeVote     ThreadType = 75 // 打分帖
	ThreadTypeLottery  ThreadType = 76 // 抽奖帖
)

// ThreadTypeFrom 把 v 映射为已知的 ThreadType，默认返回 ThreadTypeUnknown。
func ThreadTypeFrom(v int) ThreadType {
	switch ThreadType(v) {
	case ThreadTypeArticle, ThreadTypeAlbum, ThreadTypeExtShare, ThreadTypeVoice,
		ThreadTypeNetdisk, ThreadTypeStory, ThreadTypeVideo, ThreadTypeLive,
		ThreadTypeHelp, ThreadTypeVote, ThreadTypeLottery:
		return ThreadType(v)
	default:
		return ThreadTypeUnknown
	}
}

// ReqUInfo 使用该枚举类指定待获取的用户信息字段。
//
// 各 bit 位的含义由高到低分别为 OTHER, TIEBA_UID, NICK_NAME, USER_NAME, PORTRAIT, USER_ID。
// 其中 BASIC = USER_ID | PORTRAIT | USER_NAME。
type ReqUInfo uint32

const (
	ReqUInfoUserID   ReqUInfo = 1 << iota // 1
	ReqUInfoPortrait                      // 2
	ReqUInfoUserName                      // 4
	ReqUInfoNickName                      // 8
	ReqUInfoTiebaUID                      // 16
	ReqUInfoOther                         // 32

	ReqUInfoBasic = ReqUInfoUserID | ReqUInfoPortrait | ReqUInfoUserName
	ReqUInfoAll   = ReqUInfoBasic | ReqUInfoNickName | ReqUInfoTiebaUID | ReqUInfoOther
)

// ThreadSortType 主题帖排序。
//
// 对于有热门分区的贴吧 0热门排序(HOT) 1按发布时间(CREATE) 2关注的人(FOLLOW) 34热门排序(HOT) >=6是按回复时间(REPLY)。
// 对于无热门分区的贴吧 0按回复时间(REPLY) 1按发布时间(CREATE) 2关注的人(FOLLOW) >=3按回复时间(REPLY)。
type ThreadSortType int

const (
	ThreadSortReply  ThreadSortType = 6
	ThreadSortCreate ThreadSortType = 1
	ThreadSortHot    ThreadSortType = 3
	ThreadSortFollow ThreadSortType = 2
)

// PostSortType 回复排序。
type PostSortType int

const (
	PostSortAsc  PostSortType = 0 // 时间顺序
	PostSortDesc PostSortType = 1 // 时间倒序
	PostSortHot  PostSortType = 2 // 热门序
)

// BawuSearchType 吧务后台搜索类型。
type BawuSearchType int

const (
	BawuSearchUser BawuSearchType = 0 // 搜索用户
	BawuSearchOp   BawuSearchType = 1 // 搜索操作者
)

// SearchType 搜索类型。
type SearchType int

const (
	SearchAll      SearchType = 0 // 搜索全部
	SearchTime     SearchType = 1 // app时间倒序
	SearchRelation SearchType = 2 // app相关性排序
)

// GlobalSearchSortType 全吧搜索结果排序。
type GlobalSearchSortType int

const (
	GlobalSearchAsc      GlobalSearchSortType = 0 // 最早发帖
	GlobalSearchRelation GlobalSearchSortType = 2 // 最相关
	GlobalSearchDesc     GlobalSearchSortType = 5 // 最新发帖
)

// BawuType 吧务类型。
type BawuType string

const (
	BawuManager     BawuType = "assist"     // 小吧
	BawuImageEditor BawuType = "picadmin"   // 图片小编
	BawuVoiceEditor BawuType = "voiceadmin" // 语音小编
)

// BawuPermType 吧务已分配的权限。
type BawuPermType int

const (
	BawuPermNull          BawuPermType = 0 // 无权限
	BawuPermUnblock       BawuPermType = 1 // 解除封禁
	BawuPermUnblockAppeal BawuPermType = 2 // 封禁申诉处理
	BawuPermRecover       BawuPermType = 4 // 恢复删帖
	BawuPermRecoverAppeal BawuPermType = 8 // 删帖申诉处理

	BawuPermAll = BawuPermUnblock | BawuPermUnblockAppeal | BawuPermRecover | BawuPermRecoverAppeal // 所有权限
)

// RankForumType 吧签到排行榜类别。
type RankForumType int

const (
	RankForumToday     RankForumType = 0 // 今日排行
	RankForumYesterday RankForumType = 1 // 昨日排行
	RankForumWeekly    RankForumType = 2 // 周排行
	RankForumMonthly   RankForumType = 3 // 月排行
)

// BlacklistType 用户黑名单类型。
type BlacklistType int

const (
	BlacklistNull     BlacklistType = 0 // 正常状态
	BlacklistFollow   BlacklistType = 1 // 禁止关注
	BlacklistInteract BlacklistType = 2 // 禁止互动
	BlacklistChat     BlacklistType = 4 // 禁止私信

	BlacklistAll = BlacklistFollow | BlacklistInteract | BlacklistChat // 全屏蔽
)

// WsStatus websocket 连接状态。
type WsStatus int

const (
	WsStatusClosed     WsStatus = 0 // 已关闭
	WsStatusConnecting WsStatus = 1 // 正在连接
	WsStatusOpen       WsStatus = 2 // 可用
)

// GroupType 消息组类型。
type GroupType int

const (
	GroupTypePrivateMsg GroupType = 6
	GroupTypeMisc       GroupType = 8
)

// MsgType 消息类型。
type MsgType int

const (
	MsgTypePrivateMsg MsgType = 1
	MsgTypeMisc       MsgType = 10
	MsgTypeReaded     MsgType = 22
)
