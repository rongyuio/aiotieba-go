// Package enums defines the enumerated types used across the library.
//
// It mirrors the Python module aiotieba.enums. Python's IntEnum is modelled
// with a named integer type plus one constant per member, and Python's
// `_missing_` fallback (unknown value -> UNKNOWN) is modelled with a From
// constructor.
package enums

// Gender is the user gender.
type Gender int

const (
	GenderUnknown Gender = 0
	GenderMale    Gender = 1
	GenderFemale  Gender = 2
)

// GenderFrom maps v onto a known Gender, defaulting to GenderUnknown.
//
// Python's Gender has no _missing_ fallback and therefore raises ValueError for
// an unexpected value; this port defaults to GenderUnknown instead, matching the
// other From constructors and avoiding a hard failure on an unknown server
// value.
func GenderFrom(v int) Gender {
	switch Gender(v) {
	case GenderMale, GenderFemale:
		return Gender(v)
	default:
		return GenderUnknown
	}
}

// PrivLike is the visibility of the followed-forum list.
type PrivLike int

const (
	PrivLikeUnknown PrivLike = 0
	PrivLikePublic  PrivLike = 1
	PrivLikeFriend  PrivLike = 2
	PrivLikeHide    PrivLike = 3
)

// PrivLikeFrom maps v onto a known PrivLike, defaulting to PrivLikeUnknown.
func PrivLikeFrom(v int) PrivLike {
	switch PrivLike(v) {
	case PrivLikePublic, PrivLikeFriend, PrivLikeHide:
		return PrivLike(v)
	default:
		return PrivLikeUnknown
	}
}

// PrivReply is the comment permission of a post.
type PrivReply int

const (
	PrivReplyUnknown PrivReply = 0
	PrivReplyAll     PrivReply = 1
	PrivReplyFans    PrivReply = 5
	PrivReplyFollow  PrivReply = 6
)

// PrivReplyFrom maps v onto a known PrivReply, defaulting to PrivReplyUnknown.
func PrivReplyFrom(v int) PrivReply {
	switch PrivReply(v) {
	case PrivReplyAll, PrivReplyFans, PrivReplyFollow:
		return PrivReply(v)
	default:
		return PrivReplyUnknown
	}
}

// ThreadType is the type of a thread.
type ThreadType int

const (
	ThreadTypeUnknown  ThreadType = -1
	ThreadTypeArticle  ThreadType = 0
	ThreadTypeAlbum    ThreadType = 1
	ThreadTypeExtShare ThreadType = 6
	ThreadTypeVoice    ThreadType = 11
	ThreadTypeNetdisk  ThreadType = 14
	ThreadTypeStory    ThreadType = 31
	ThreadTypeVideo    ThreadType = 40
	ThreadTypeLive     ThreadType = 50
	ThreadTypeHelp     ThreadType = 71
	ThreadTypeVote     ThreadType = 75
	ThreadTypeLottery  ThreadType = 76
)

// ThreadTypeFrom maps v onto a known ThreadType, defaulting to ThreadTypeUnknown.
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

// ReqUInfo is a bitmask selecting which user fields to fetch.
//
// BASIC = UserID | Portrait | UserName.
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

// ThreadSortType is the sort order of a thread list.
type ThreadSortType int

const (
	ThreadSortReply  ThreadSortType = 6
	ThreadSortCreate ThreadSortType = 1
	ThreadSortHot    ThreadSortType = 3
	ThreadSortFollow ThreadSortType = 2
)

// PostSortType is the sort order of a post list.
type PostSortType int

const (
	PostSortAsc  PostSortType = 0
	PostSortDesc PostSortType = 1
	PostSortHot  PostSortType = 2
)

// BawuSearchType selects the searched field of the bawu log API.
type BawuSearchType int

const (
	BawuSearchUser BawuSearchType = 0
	BawuSearchOp   BawuSearchType = 1
)

// SearchType is the in-forum search mode.
type SearchType int

const (
	SearchAll      SearchType = 0
	SearchTime     SearchType = 1
	SearchRelation SearchType = 2
)

// GlobalSearchSortType is the sort order of the global search result.
type GlobalSearchSortType int

const (
	GlobalSearchAsc      GlobalSearchSortType = 0
	GlobalSearchRelation GlobalSearchSortType = 2
	GlobalSearchDesc     GlobalSearchSortType = 5
)

// BawuType is the type of a bawu (forum moderator) role.
type BawuType string

const (
	BawuManager     BawuType = "assist"
	BawuImageEditor BawuType = "picadmin"
	BawuVoiceEditor BawuType = "voiceadmin"
)

// BawuPermType is a bitmask of the permissions granted to a bawu.
type BawuPermType int

const (
	BawuPermNull          BawuPermType = 0
	BawuPermUnblock       BawuPermType = 1
	BawuPermUnblockAppeal BawuPermType = 2
	BawuPermRecover       BawuPermType = 4
	BawuPermRecoverAppeal BawuPermType = 8

	BawuPermAll = BawuPermUnblock | BawuPermUnblockAppeal | BawuPermRecover | BawuPermRecoverAppeal
)

// RankForumType is the category of the forum sign-in ranking.
type RankForumType int

const (
	RankForumToday     RankForumType = 0
	RankForumYesterday RankForumType = 1
	RankForumWeekly    RankForumType = 2
	RankForumMonthly   RankForumType = 3
)

// BlacklistType is a bitmask of the user blacklist behaviours.
type BlacklistType int

const (
	BlacklistNull     BlacklistType = 0
	BlacklistFollow   BlacklistType = 1
	BlacklistInteract BlacklistType = 2
	BlacklistChat     BlacklistType = 4

	BlacklistAll = BlacklistFollow | BlacklistInteract | BlacklistChat
)

// WsStatus is the state of the websocket connection.
type WsStatus int

const (
	WsStatusClosed     WsStatus = 0
	WsStatusConnecting WsStatus = 1
	WsStatusOpen       WsStatus = 2
)

// GroupType is the type of a websocket message group.
type GroupType int

const (
	GroupTypePrivateMsg GroupType = 6
	GroupTypeMisc       GroupType = 8
)

// MsgType is the type of a websocket message.
type MsgType int

const (
	MsgTypePrivateMsg MsgType = 1
	MsgTypeMisc       MsgType = 10
	MsgTypeReaded     MsgType = 22
)
