// Package getats 实现 aiotieba 的 get_ats API。
//
// 对应 Python 包 aiotieba.api.get_ats。
package getats

import (
	"strings"

	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/helper"
)

// FragTextAt 纯文本碎片，对应上游的 `FragText_at = FragText`。
type FragTextAt = classdef.FragText

// FragLinkAt 链接碎片，对应上游的 `FragLink_at = FragLink`。
type FragLinkAt = classdef.FragLink

// FragEmojiAt 表情碎片。
type FragEmojiAt struct {
	ID   string // 表情图片id
	Desc string // 表情描述
}

// FragEmojiAtFromJSON 对应 FragEmoji_at.from_json。
func FragEmojiAtFromJSON(m map[string]any) FragEmojiAt {
	return FragEmojiAt{
		ID:   helper.JSONStr(m, "text"),
		Desc: helper.JSONStr(m, "c"),
	}
}

// FragImageAt 图像碎片。
type FragImageAt struct {
	Src        string // 小图链接
	OriginSize int64  // 原图大小
	ShowWidth  int32  // 图像在客户端预览显示的宽度
	ShowHeight int32  // 图像在客户端预览显示的高度
	Hash       string // 百度图床hash
}

// FragImageAtFromJSON 对应 FragImage_at.from_json。
func FragImageAtFromJSON(m map[string]any) FragImageAt {
	src := helper.JSONStr(m, "src")
	width, height := classdef.SplitSize(helper.JSONStr(m, "bsize"))
	return FragImageAt{
		Src:        src,
		OriginSize: helper.JSONInt(m, "size"),
		ShowWidth:  width,
		ShowHeight: height,
		Hash:       classdef.ImageHash(src),
	}
}

// FragAtAt @碎片。
type FragAtAt struct {
	Text     string // 被@用户的昵称 含@
	UserID   int64  // 被@用户的user_id
	Portrait string // 被@用户的portrait
}

// FragAtAtFromJSON 对应 FragAt_at.from_json。
func FragAtAtFromJSON(m map[string]any) FragAtAt {
	return FragAtAt{
		Text:     helper.JSONStr(m, "text"),
		UserID:   helper.JSONInt(m, "uid"),
		Portrait: classdef.TrimPortrait(helper.JSONStr(m, "portrait")),
	}
}

// ContentsAt 内容碎片列表。
type ContentsAt struct {
	classdef.Containers[any]

	Texts  []any               // 纯文本碎片列表
	Emojis []FragEmojiAt       // 表情碎片列表
	Imgs   []FragImageAt       // 图像碎片列表
	Ats    []FragAtAt          // @碎片列表
	Links  []classdef.FragLink // 链接碎片列表
}

// ContentsAtFromJSON 对应 Contents_at.from_json。
//
// 碎片类型按 JSON 键嗅探：有 src 视为图像，有 link 视为链接，
// type 为 4 视为 @，type 为 2 视为表情，其余按文本处理。
func ContentsAtFromJSON(contentMaps []any) ContentsAt {
	var c ContentsAt
	for _, item := range contentMaps {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		switch {
		case helper.HasJSONKey(m, "src"):
			frag := FragImageAtFromJSON(m)
			c.Imgs = append(c.Imgs, frag)
			c.Objs = append(c.Objs, frag)
		case helper.HasJSONKey(m, "link"):
			frag := classdef.FragLinkFromJSON(m)
			c.Links = append(c.Links, frag)
			c.Texts = append(c.Texts, frag)
			c.Objs = append(c.Objs, frag)
		case helper.JSONInt(m, "type") == 4:
			frag := FragAtAtFromJSON(m)
			c.Ats = append(c.Ats, frag)
			c.Texts = append(c.Texts, frag)
			c.Objs = append(c.Objs, frag)
		case helper.JSONInt(m, "type") == 2:
			frag := FragEmojiAtFromJSON(m)
			c.Emojis = append(c.Emojis, frag)
			c.Objs = append(c.Objs, frag)
		default:
			frag := classdef.FragTextFromJSON(m)
			c.Texts = append(c.Texts, frag)
			c.Objs = append(c.Objs, frag)
		}
	}
	return c
}

// Text 文本内容，对应 text 缓存属性。
func (c ContentsAt) Text() string {
	var b strings.Builder
	for _, frag := range c.Texts {
		switch t := frag.(type) {
		case classdef.FragText:
			b.WriteString(t.Text)
		case classdef.FragLink:
			b.WriteString(t.Text)
		case FragAtAt:
			b.WriteString(t.Text)
		}
	}
	return b.String()
}

// stripQuoteHeader 去掉引用楼层正文里的楼中楼引述头，对应 _strip_quote_header。
//
// 正片以「@某人: 正文」的形式开头时，@碎片与其后的 ": " 前缀都被剥掉；
// 剥完正文为空时连同该文本碎片一起丢弃。
func stripQuoteHeader(c *ContentsAt) {
	objs := c.Objs
	if len(objs) < 2 {
		return
	}
	if _, ok := objs[0].(FragAtAt); !ok {
		return
	}

	textFrag, ok := objs[1].(classdef.FragText)
	if !ok || !strings.HasPrefix(textFrag.Text, ": ") {
		return
	}

	textFrag.Text = strings.TrimPrefix(textFrag.Text, ": ")
	objs[1] = textFrag
	// Python 中 objs[1] 与 texts[1] 引用同一个碎片，原地改写两处都可见；
	// Go 的值语义要求显式同步 texts 中的副本。
	if len(c.Texts) > 1 {
		if _, ok := c.Texts[1].(classdef.FragText); ok {
			c.Texts[1] = textFrag
		}
	}

	skip := 1
	if textFrag.Text == "" {
		skip = 2
	}

	c.Objs = objs[skip:]
	c.Texts = c.Texts[skip:]
	c.Ats = c.Ats[1:]
}

// PageAt 页信息。
type PageAt struct {
	CurrentPage int64 // 当前页码
	HasMore     bool  // 是否有后继页
	HasPrev     bool  // 是否有前驱页
}

// PageAtFromJSON 对应 Page_at.from_json。
func PageAtFromJSON(m map[string]any) PageAt {
	return PageAt{
		CurrentPage: helper.JSONInt(m, "current_page"),
		HasMore:     helper.JSONBool(m, "has_more"),
		HasPrev:     helper.JSONBool(m, "has_prev"),
	}
}

// UserInfoAt 用户信息。
type UserInfoAt struct {
	UserID      int64           // user_id
	Portrait    string          // portrait
	UserName    string          // 用户名
	NickNameNew string          // 新版昵称
	PrivLike    enums.PrivLike  // 关注吧列表的公开状态
	PrivReply   enums.PrivReply // 帖子评论权限
}

// UserInfoAtFromJSON 对应 UserInfo_at.from_json。
func UserInfoAtFromJSON(m map[string]any) UserInfoAt {
	u := UserInfoAt{
		UserID:      helper.JSONInt(m, "id"),
		Portrait:    classdef.TrimPortrait(helper.JSONStr(m, "portrait")),
		UserName:    helper.JSONStr(m, "name"),
		NickNameNew: helper.JSONStr(m, "name_show"),
	}
	if priv := helper.JSONMap(m, "priv_sets"); priv != nil {
		u.PrivLike = enums.PrivLike(helper.JSONInt(priv, "like"))
		u.PrivReply = enums.PrivReply(helper.JSONInt(priv, "reply"))
	} else {
		u.PrivLike = enums.PrivLikePublic
		u.PrivReply = enums.PrivReplyAll
	}
	return u
}

// NickName 用户昵称。
func (u UserInfoAt) NickName() string { return u.NickNameNew }

// ShowName 显示名称。
func (u UserInfoAt) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// UserInfoAtP 用户信息。
type UserInfoAtP struct {
	UserID      int64  // user_id
	Portrait    string // portrait
	UserName    string // 用户名
	NickNameNew string // 新版昵称
}

// UserInfoAtPFromJSON 对应 UserInfo_at_p.from_json。
func UserInfoAtPFromJSON(m map[string]any) UserInfoAtP {
	return UserInfoAtP{
		UserID:      helper.JSONInt(m, "id"),
		Portrait:    classdef.TrimPortrait(helper.JSONStr(m, "portrait")),
		UserName:    helper.JSONStr(m, "name"),
		NickNameNew: helper.JSONStr(m, "name_show"),
	}
}

// NickName 用户昵称。
func (u UserInfoAtP) NickName() string { return u.NickNameNew }

// ShowName 显示名称。
func (u UserInfoAtP) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// UserInfoAtT 用户信息。
type UserInfoAtT struct {
	UserID      int64  // user_id
	Portrait    string // portrait
	UserName    string // 用户名
	NickNameNew string // 新版昵称
}

// UserInfoAtTFromJSON 对应 UserInfo_at_t.from_json。
func UserInfoAtTFromJSON(m map[string]any) UserInfoAtT {
	return UserInfoAtT{
		UserID:      helper.JSONInt(m, "id"),
		Portrait:    classdef.TrimPortrait(helper.JSONStr(m, "portrait")),
		UserName:    helper.JSONStr(m, "name"),
		NickNameNew: helper.JSONStr(m, "name_show"),
	}
}

// NickName 用户昵称。
func (u UserInfoAtT) NickName() string { return u.NickNameNew }

// ShowName 显示名称。
func (u UserInfoAtT) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// PostAt 父级回复信息。
type PostAt struct {
	PID      int64       // 父级回复id
	Contents ContentsAt  // 正文内容碎片列表
	User     UserInfoAtP // 发布者的用户信息
}

// PostAtFromJSON 对应 Post_at.from_json。
//
// 引用楼层的正文取自 new_floor_info 的倒数第二项，长度不足三项时视为无引用。
func PostAtFromJSON(m map[string]any) PostAt {
	contents := ContentsAt{}
	if floors := helper.JSONSlice(m, "new_floor_info"); len(floors) > 2 {
		if floor, ok := floors[len(floors)-2].(map[string]any); ok {
			contents = ContentsAtFromJSON(helper.JSONSlice(floor, "content"))
			stripQuoteHeader(&contents)
		}
	}
	return PostAt{
		PID:      helper.JSONInt(m, "quote_pid"),
		Contents: contents,
		User:     UserInfoAtPFromJSON(helper.JSONMap(m, "quote_user")),
	}
}

// Valid 对应 Post_at 的 __bool__。
func (p PostAt) Valid() bool { return !p.Contents.Empty() }

// ThreadAt 父级主题帖信息。
type ThreadAt struct {
	TID      int64       // 父级主题帖id
	Contents ContentsAt  // 主题帖内容碎片列表
	User     UserInfoAtT // 发布者的用户信息
}

// ThreadAtFromJSON 对应 Thread_at.from_json。
//
// 主题帖正文取自 new_floor_info 的首项，列表为空时视为无正文。
func ThreadAtFromJSON(m map[string]any) ThreadAt {
	contents := ContentsAt{}
	if floors := helper.JSONSlice(m, "new_floor_info"); len(floors) > 0 {
		if floor, ok := floors[0].(map[string]any); ok {
			contents = ContentsAtFromJSON(helper.JSONSlice(floor, "content"))
			stripQuoteHeader(&contents)
		}
	}
	return ThreadAt{
		TID:      helper.JSONInt(m, "thread_id"),
		Contents: contents,
		User:     UserInfoAtTFromJSON(helper.JSONMap(m, "thread_author_user")),
	}
}

// Valid 对应 Thread_at 的 __bool__。
func (t ThreadAt) Valid() bool { return !t.Contents.Empty() }

// At @信息。
type At struct {
	Text        string            // 文本内容
	FName       string            // 所在贴吧名
	FID         int64             // 所在贴吧id
	TID         int64             // 所在主题帖id
	PID         int64             // 回复id
	User        UserInfoAt        // 发布者的用户信息
	Post        PostAt            // 父级回复信息
	Thread      ThreadAt          // 父级主题帖信息
	ContentType enums.ContentType // 帖子对象类型
	CreateTime  int64             // 创建时间
}

// AtFromJSON 对应 At.from_json。
func AtFromJSON(m map[string]any) At {
	return At{
		Text:        helper.JSONStr(m, "content"),
		FName:       helper.JSONStr(m, "fname"),
		FID:         helper.JSONInt(m, "fid"),
		TID:         helper.JSONInt(m, "thread_id"),
		PID:         helper.JSONInt(m, "post_id"),
		User:        UserInfoAtFromJSON(helper.JSONMap(m, "replyer")),
		Post:        PostAtFromJSON(m),
		Thread:      ThreadAtFromJSON(m),
		ContentType: enums.ContentTypeFrom(int(helper.JSONInt(m, "type"))),
		CreateTime:  helper.JSONInt(m, "time"),
	}
}

// AuthorID 发布者的user_id。
func (a At) AuthorID() int64 { return a.User.UserID }

// Ats @信息列表。
type Ats struct {
	classdef.Containers[*At]

	Page PageAt // 页信息
	Err  error  // 捕获的异常
}

// AtsFromJSON 对应 Ats.from_json。
func AtsFromJSON(m map[string]any) Ats {
	var ats Ats
	for _, item := range helper.JSONSlice(m, "at_list") {
		if im, ok := item.(map[string]any); ok {
			at := AtFromJSON(im)
			ats.Objs = append(ats.Objs, &at)
		}
	}
	ats.Page = PageAtFromJSON(helper.JSONMap(m, "page"))
	return ats
}

// HasMore 是否还有下一页。
func (a Ats) HasMore() bool { return a.Page.HasMore }
