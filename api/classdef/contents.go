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

// imageHashRe 对应 _IMAGEHASH_EXP。
var imageHashRe = regexp.MustCompile(`/([a-z0-9]{32,})\.`)

// ImageHash 从 url 中提取百度图床 hash，使用各 API 模块所用的 _IMAGEHASH_EXP 正则；
// url 不含 hash 时返回空字符串。
func ImageHash(src string) string { return imageHash(src) }

// ParseInt64OrZero 把 s 解析为 int64，非数字时返回 0。
func ParseInt64OrZero(s string) int64 { return parseIntOrZero(s) }

// Fragment 是帖子富文本内容中的一片碎片。
//
// Python 使用 Protocol 联合类型，Go 用封闭接口建模。
type Fragment interface {
	fragment()
}

// FragText 纯文本碎片。
type FragText struct {
	Text string // 文本内容
}

func (FragText) fragment() {}

// FragTextFromProto 对应 FragText.from_proto。
func FragTextFromProto(p *protobuf.PbContent) FragText {
	return FragText{Text: p.GetText()}
}

// FragTextFromJSON 对应 FragText.from_json。
func FragTextFromJSON(m map[string]any) FragText {
	return FragText{Text: jsonString(m, "text")}
}

// FragTextFromText 由原始文本构造 FragText。Python 的 FragText.from_proto 只读取
// `text` 字段，因此它也能处理 feed 内容消息。
func FragTextFromText(text string) FragText {
	return FragText{Text: text}
}

// FragEmoji 表情碎片。
type FragEmoji struct {
	ID   string // 表情图片id
	Desc string // 表情描述
}

func (FragEmoji) fragment() {}

// FragEmojiFromProto 对应 FragEmoji.from_proto。
func FragEmojiFromProto(p *protobuf.PbContent) FragEmoji {
	return FragEmoji{ID: p.GetText(), Desc: p.GetC()}
}

// FragImage 图像碎片。
type FragImage struct {
	Src        string // 小图链接 宽720px
	BigSrc     string // 大图链接 宽960px
	OriginSrc  string // 原图链接
	OriginSize int64  // 原图大小
	ShowWidth  int32  // 图像在客户端预览显示的宽度
	ShowHeight int32  // 图像在客户端预览显示的高度
	Hash       string // 百度图床hash
}

func (FragImage) fragment() {}

// FragImageFromProto 对应 FragImage.from_proto。
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

// FragAt @碎片。
type FragAt struct {
	Text   string // 被@用户的昵称 含@
	UserID int64  // 被@用户的user_id
}

func (FragAt) fragment() {}

// FragAtFromProto 对应 FragAt.from_proto。
func FragAtFromProto(p *protobuf.PbContent) FragAt {
	return FragAt{Text: p.GetText(), UserID: p.GetUid()}
}

// FragVoice 音频碎片。
type FragVoice struct {
	MD5      string  // 音频md5
	Duration float64 // 音频长度 以秒为单位
}

func (FragVoice) fragment() {}

// FragVoiceFromProto 对应 FragVoice.from_proto。
func FragVoiceFromProto(p *protobuf.Voice) FragVoice {
	return FragVoice{MD5: p.GetVoiceMd5(), Duration: float64(p.GetDuringTime()) / 1000}
}

// Valid 对应 FragVoice 的 __bool__。
func (f FragVoice) Valid() bool { return f.MD5 != "" }

// FragVideo 视频碎片。
type FragVideo struct {
	Src      string // 视频链接
	CoverSrc string // 封面链接
	Duration int64  // 视频长度
	Width    int64  // 视频宽度
	Height   int64  // 视频高度
	ViewNum  int64  // 浏览次数
}

func (FragVideo) fragment() {}

// FragVideoFromProto 对应 FragVideo.from_proto。
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

// Valid 对应 FragVideo 的 __bool__。
func (f FragVideo) Valid() bool { return f.Width != 0 }

// FragLink 链接碎片。
type FragLink struct {
	Text   string   // 原链接
	Title  string   // 链接标题
	RawURL *url.URL // 解析后的原链接
}

func (FragLink) fragment() {}

// FragLinkFromProto 对应 FragLink.from_proto。
func FragLinkFromProto(p *protobuf.PbContent) FragLink {
	return newFragLink(p.GetLink(), p.GetText())
}

// FragLinkFromJSON 对应 FragLink.from_json。
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

// IsExternal 是否外部链接，对应 is_external 缓存属性。
func (f FragLink) IsExternal() bool {
	return f.RawURL != nil && f.RawURL.Path == "/mo/q/checkurl"
}

// URL 解析后的去前缀链接，对应 url 缓存属性：重定向链接展开后的目标地址。
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

// FragTiebaPlus 贴吧plus广告碎片。
type FragTiebaPlus struct {
	Text string   // 贴吧plus广告描述
	URL  *url.URL // 解析后的贴吧plus广告跳转链接
}

func (FragTiebaPlus) fragment() {}

// FragTiebaPlusFromProto 对应 FragTiebaPlus.from_proto。
func FragTiebaPlusFromProto(p *protobuf.PbContent) FragTiebaPlus {
	info := p.GetTiebaplusInfo()
	raw, err := url.Parse(info.GetJumpUrl())
	if err != nil {
		raw = &url.URL{}
	}
	return FragTiebaPlus{Text: info.GetDesc(), URL: raw}
}

// FragItem item碎片。
type FragItem struct {
	Text string // item名称
}

func (FragItem) fragment() {}

// FragItemFromProto 对应 FragItem.from_proto。
func FragItemFromProto(p *protobuf.PbContent) FragItem {
	return FragItem{Text: p.GetItem().GetItemName()}
}

// FragUnknown 未知碎片，保留原始 proto。
type FragUnknown struct {
	Data any // 原始数据
}

func (FragUnknown) fragment() {}

// FragUnknownFromProto 对应 FragUnknown.from_proto。
func FragUnknownFromProto(p *protobuf.PbContent) FragUnknown {
	return FragUnknown{Data: p}
}

// FragUnknownFromAny 保留任意原始消息，对应 feed 内容消息（并非 PbContent）场景下的
// FragUnknown.from_proto。
func FragUnknownFromAny(data any) FragUnknown {
	return FragUnknown{Data: data}
}

// FragUnknownFromJSON 对应 FragUnknown.from_json。
func FragUnknownFromJSON(m map[string]any) FragUnknown {
	return FragUnknown{Data: m}
}

// FragmentText 返回类文本碎片携带的文本，对应 Python 客户端使用的 TypeFragText 协议。
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

// FragmentTextOf 拼接所有类文本碎片的文本，对应 Python 内容容器的 `text` 属性。
func FragmentTextOf(frags []Fragment) string {
	var b strings.Builder
	for _, f := range frags {
		b.WriteString(FragmentText(f))
	}
	return b.String()
}

// LogUnknownFragment 用库日志记录器记录未知碎片类型，与 Python 模块的做法一致。
func LogUnknownFragment(logger *slog.Logger, tid int64, fragType int64) {
	if logger != nil {
		logger.Debug("unknown fragment type", "tid", tid, "type", fragType)
	}
}

// splitSize 拆分 "width,height" 字符串。Python 原版使用 str.partition(",")；
// 格式非法时返回 0 而不是抛错。
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
