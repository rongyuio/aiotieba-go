package classdef

import (
	"strings"

	"github.com/rongyuio/aiotieba/enums"
	"github.com/rongyuio/aiotieba/protobuf"
)

// TrimPortrait strips the trailing 13 characters of a portrait that carries a
// query string, mirroring the Python expression
// `portrait[:-13] if "?" in portrait else portrait`.
func TrimPortrait(portrait string) string {
	if !strings.Contains(portrait, "?") {
		return portrait
	}
	if len(portrait) > 13 {
		return portrait[:len(portrait)-13]
	}
	return ""
}

// UserIcons returns the non-empty icon names of a user.
func UserIcons(p *protobuf.User) []string {
	icons := make([]string, 0, len(p.GetIconinfo()))
	for _, icon := range p.GetIconinfo() {
		if name := icon.GetName(); name != "" {
			icons = append(icons, name)
		}
	}
	return icons
}

// UserPrivLike returns the followed-forum visibility of a user, defaulting to
// PUBLIC like the Python API modules.
func UserPrivLike(p *protobuf.User) enums.PrivLike {
	if v := p.GetPrivSets().GetLike(); v != 0 {
		return enums.PrivLikeFrom(int(v))
	}
	return enums.PrivLikePublic
}

// UserPrivReply returns the comment permission of a user, defaulting to ALL like
// the Python API modules.
func UserPrivReply(p *protobuf.User) enums.PrivReply {
	if v := p.GetPrivSets().GetReply(); v != 0 {
		return enums.PrivReplyFrom(int(v))
	}
	return enums.PrivReplyAll
}

// SplitSize splits a "width,height" bsize string.
//
// The Python original uses str.partition(",") followed by int(); a malformed
// value raises there, while this port yields zeroes.
func SplitSize(bsize string) (int32, int32) {
	widthStr, heightStr, _ := strings.Cut(bsize, ",")
	return int32(ParseInt64OrZero(widthStr)), int32(ParseInt64OrZero(heightStr))
}

// IsUserThreadAuthor reports whether userID is the thread author. It mirrors the
// `thread.author_id == obj.author_id` comparisons of the Python API modules.
func IsUserThreadAuthor(threadAuthorID, userID int64) bool {
	return threadAuthorID != 0 && threadAuthorID == userID
}
