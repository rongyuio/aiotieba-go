package classdef

import (
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/rongyuio/aiotieba-go/protobuf"
)

// imageHashRe mirrors _IMAGEHASH_EXP.
var imageHashRe = regexp.MustCompile(`/([a-z0-9]{32,})\.`)

// ImageHash extracts the Baidu image hash from a url, mirroring the
// _IMAGEHASH_EXP regex used by the API modules. It returns an empty string when
// the url does not contain a hash.
func ImageHash(src string) string { return imageHash(src) }

// ParseInt64OrZero parses s as an int64 and returns 0 when it is not numeric.
func ParseInt64OrZero(s string) int64 { return parseIntOrZero(s) }

// Fragment is one piece of the rich content of a post.
//
// The Python code uses a Protocol union; Go models it with a sealed interface.
type Fragment interface {
	fragment()
}

// FragText is a plain text fragment.
type FragText struct {
	Text string
}

func (FragText) fragment() {}

// FragTextFromProto mirrors FragText.from_proto.
func FragTextFromProto(p *protobuf.PbContent) FragText {
	return FragText{Text: p.GetText()}
}

// FragTextFromJSON mirrors FragText.from_json.
func FragTextFromJSON(m map[string]any) FragText {
	return FragText{Text: jsonString(m, "text")}
}

// FragTextFromText builds a FragText from a raw text value. The Python
// FragText.from_proto only reads a `text` field, so it also accepts the feed
// content messages.
func FragTextFromText(text string) FragText {
	return FragText{Text: text}
}

// FragEmoji is an emoji fragment.
type FragEmoji struct {
	ID   string
	Desc string
}

func (FragEmoji) fragment() {}

// FragEmojiFromProto mirrors FragEmoji.from_proto.
func FragEmojiFromProto(p *protobuf.PbContent) FragEmoji {
	return FragEmoji{ID: p.GetText(), Desc: p.GetC()}
}

// FragImage is an image fragment.
type FragImage struct {
	Src        string
	BigSrc     string
	OriginSrc  string
	OriginSize int64
	ShowWidth  int32
	ShowHeight int32
	Hash       string
}

func (FragImage) fragment() {}

// FragImageFromProto mirrors FragImage.from_proto.
func FragImageFromProto(p *protobuf.PbContent) FragImage {
	src := p.GetCdnSrc()
	showWidth, showHeight := splitSize(p.GetBsize())
	return FragImage{
		Src:        src,
		BigSrc:     p.GetBigCdnSrc(),
		OriginSrc:  p.GetOriginSrc(),
		OriginSize: int64(p.GetOriginSize()),
		ShowWidth:  showWidth,
		ShowHeight: showHeight,
		Hash:       imageHash(src),
	}
}

// FragAt is an "@user" fragment.
type FragAt struct {
	Text   string
	UserID int64
}

func (FragAt) fragment() {}

// FragAtFromProto mirrors FragAt.from_proto.
func FragAtFromProto(p *protobuf.PbContent) FragAt {
	return FragAt{Text: p.GetText(), UserID: p.GetUid()}
}

// FragVoice is a voice fragment.
type FragVoice struct {
	MD5      string
	Duration float64
}

func (FragVoice) fragment() {}

// FragVoiceFromProto mirrors FragVoice.from_proto.
func FragVoiceFromProto(p *protobuf.Voice) FragVoice {
	return FragVoice{MD5: p.GetVoiceMd5(), Duration: float64(p.GetDuringTime()) / 1000}
}

// Valid mirrors __bool__ of FragVoice.
func (f FragVoice) Valid() bool { return f.MD5 != "" }

// FragVideo is a video fragment.
type FragVideo struct {
	Src      string
	CoverSrc string
	Duration int64
	Width    int64
	Height   int64
	ViewNum  int64
}

func (FragVideo) fragment() {}

// FragVideoFromProto mirrors FragVideo.from_proto.
func FragVideoFromProto(p *protobuf.VideoInfo) FragVideo {
	return FragVideo{
		Src:      p.GetVideoUrl(),
		CoverSrc: p.GetThumbnailUrl(),
		Duration: int64(p.GetVideoDuration()),
		Width:    int64(p.GetVideoWidth()),
		Height:   int64(p.GetVideoHeight()),
		ViewNum:  int64(p.GetPlayCount()),
	}
}

// Valid mirrors __bool__ of FragVideo.
func (f FragVideo) Valid() bool { return f.Width != 0 }

// FragLink is a link fragment.
type FragLink struct {
	Text   string
	Title  string
	RawURL *url.URL
}

func (FragLink) fragment() {}

// FragLinkFromProto mirrors FragLink.from_proto.
func FragLinkFromProto(p *protobuf.PbContent) FragLink {
	return newFragLink(p.GetLink(), p.GetText())
}

// FragLinkFromJSON mirrors FragLink.from_json.
func FragLinkFromJSON(m map[string]any) FragLink {
	return newFragLink(jsonString(m, "link"), jsonString(m, "text"))
}

func newFragLink(text, title string) FragLink {
	raw, err := url.Parse(text)
	if err != nil {
		raw = &url.URL{}
	}
	return FragLink{Text: text, Title: title, RawURL: raw}
}

// IsExternal mirrors the is_external cached property.
func (f FragLink) IsExternal() bool {
	return f.RawURL != nil && f.RawURL.Path == "/mo/q/checkurl"
}

// URL mirrors the url cached property: the unwrapped target of a redirect link.
func (f FragLink) URL() *url.URL {
	if !f.IsExternal() {
		return f.RawURL
	}
	target := f.RawURL.Query().Get("url")
	if target == "" {
		return f.RawURL
	}
	u, err := url.Parse(target)
	if err != nil {
		return f.RawURL
	}
	return u
}

// FragTiebaPlus is a tieba-plus advertisement fragment.
type FragTiebaPlus struct {
	Text string
	URL  *url.URL
}

func (FragTiebaPlus) fragment() {}

// FragTiebaPlusFromProto mirrors FragTiebaPlus.from_proto.
func FragTiebaPlusFromProto(p *protobuf.PbContent) FragTiebaPlus {
	info := p.GetTiebaplusInfo()
	raw, err := url.Parse(info.GetJumpUrl())
	if err != nil {
		raw = &url.URL{}
	}
	return FragTiebaPlus{Text: info.GetDesc(), URL: raw}
}

// FragItem is an item fragment.
type FragItem struct {
	Text string
}

func (FragItem) fragment() {}

// FragItemFromProto mirrors FragItem.from_proto.
func FragItemFromProto(p *protobuf.PbContent) FragItem {
	return FragItem{Text: p.GetItem().GetItemName()}
}

// FragUnknown is a fragment of an unknown type. It keeps the raw proto.
type FragUnknown struct {
	Data any
}

func (FragUnknown) fragment() {}

// FragUnknownFromProto mirrors FragUnknown.from_proto.
func FragUnknownFromProto(p *protobuf.PbContent) FragUnknown {
	return FragUnknown{Data: p}
}

// FragUnknownFromAny keeps an arbitrary raw message. It mirrors
// FragUnknown.from_proto for the feed content messages, which are not
// PbContent.
func FragUnknownFromAny(data any) FragUnknown {
	return FragUnknown{Data: data}
}

// FragUnknownFromJSON mirrors FragUnknown.from_json.
func FragUnknownFromJSON(m map[string]any) FragUnknown {
	return FragUnknown{Data: m}
}

// FragmentText returns the text carried by a text-like fragment, mirroring the
// TypeFragText protocol used by the Python clients.
func FragmentText(f Fragment) string {
	switch v := f.(type) {
	case FragText:
		return v.Text
	case FragAt:
		return v.Text
	case FragLink:
		return v.Text
	case FragTiebaPlus:
		return v.Text
	case FragItem:
		return v.Text
	default:
		return ""
	}
}

// FragmentTextOf joins the text of every text-like fragment, mirroring the
// `text` property of the Python contents containers.
func FragmentTextOf(frags []Fragment) string {
	var b strings.Builder
	for _, f := range frags {
		b.WriteString(FragmentText(f))
	}
	return b.String()
}

// LogUnknownFragment logs an unknown fragment type the way the Python modules
// do with the library logger.
func LogUnknownFragment(logger *slog.Logger, tid int64, fragType int64) {
	if logger != nil {
		logger.Debug("unknown fragment type", "tid", tid, "type", fragType)
	}
}

// splitSize splits a "width,height" string. The Python original uses
// str.partition(","); malformed values yield zeroes instead of raising.
func splitSize(bsize string) (int32, int32) {
	widthStr, heightStr, _ := strings.Cut(bsize, ",")
	return int32(parseIntOrZero(widthStr)), int32(parseIntOrZero(heightStr))
}

func parseIntOrZero(s string) int64 {
	v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0
	}
	return v
}

func imageHash(src string) string {
	if m := imageHashRe.FindStringSubmatch(src); len(m) == 2 {
		return m[1]
	}
	return ""
}

func jsonString(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}
