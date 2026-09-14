package classdef

import (
	"strings"

	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/protobuf"
)

// TrimPortrait 去掉带查询字符串的 portrait 末尾 13 个字符，对应 Python 表达式
// `portrait[:-13] if "?" in portrait else portrait`。
func TrimPortrait(portrait string) string {
	if !strings.Contains(portrait, "?") {
		return portrait
	}
	if len(portrait) > 13 {
		return portrait[:len(portrait)-13]
	}
	return ""
}

// UserIcons 返回用户非空的印记名称。
func UserIcons(p *protobuf.User) []string {
	icons := make([]string, 0, len(p.GetIconinfo()))
	for _, icon := range p.GetIconinfo() {
		if name := icon.GetName(); name != "" {
			icons = append(icons, name)
		}
	}
	return icons
}

// UserPrivLike 返回用户关注吧列表的公开状态，与 Python API 模块一致默认为 PUBLIC。
func UserPrivLike(p *protobuf.User) enums.PrivLike {
	if v := p.GetPrivSets().GetLike(); v != 0 {
		return enums.PrivLikeFrom(int(v))
	}
	return enums.PrivLikePublic
}

// UserPrivReply 返回用户的帖子评论权限，与 Python API 模块一致默认为 ALL。
func UserPrivReply(p *protobuf.User) enums.PrivReply {
	if v := p.GetPrivSets().GetReply(); v != 0 {
		return enums.PrivReplyFrom(int(v))
	}
	return enums.PrivReplyAll
}

// SplitSize 拆分 "width,height" 形式的 bsize 字符串。
//
// Python 原版使用 str.partition(",") 再转 int()，格式非法时会抛错；本移植版返回 0。
func SplitSize(bsize string) (int32, int32) {
	widthStr, heightStr, _ := strings.Cut(bsize, ",")
	return int32(ParseInt64OrZero(widthStr)), int32(ParseInt64OrZero(heightStr))
}

// IsUserThreadAuthor 报告 userID 是否为主题帖作者，对应 Python API 模块中的
// `thread.author_id == obj.author_id` 比较。
func IsUserThreadAuthor(threadAuthorID, userID int64) bool {
	return threadAuthorID != 0 && threadAuthorID == userID
}
