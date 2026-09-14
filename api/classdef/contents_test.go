package classdef

import (
	"testing"

	"github.com/rongyuio/aiotieba/protobuf"
)

func TestFragImageFromProto(t *testing.T) {
	p := &protobuf.PbContent{
		CdnSrc:     "https://imgsrc.baidu.com/forum/w=580/sign=abcdef/9f8e7d6c5b4a39281706f5e4d3c2b1a0.jpg",
		BigCdnSrc:  "https://imgsrc.baidu.com/big.jpg",
		OriginSrc:  "https://imgsrc.baidu.com/origin.jpg",
		OriginSize: 12345,
		Bsize:      "580,435",
	}

	got := FragImageFromProto(p)
	if got.Src != p.CdnSrc || got.BigSrc != p.BigCdnSrc || got.OriginSrc != p.OriginSrc {
		t.Errorf("image urls = %+v", got)
	}
	if got.OriginSize != 12345 {
		t.Errorf("origin size = %d, want 12345", got.OriginSize)
	}
	if got.ShowWidth != 580 || got.ShowHeight != 435 {
		t.Errorf("show size = %dx%d, want 580x435", got.ShowWidth, got.ShowHeight)
	}
	const wantHash = "9f8e7d6c5b4a39281706f5e4d3c2b1a0"
	if got.Hash != wantHash {
		t.Errorf("hash = %q, want %q", got.Hash, wantHash)
	}
}

func TestFragImageMalformedBsize(t *testing.T) {
	got := FragImageFromProto(&protobuf.PbContent{Bsize: "not-a-size"})
	if got.ShowWidth != 0 || got.ShowHeight != 0 {
		t.Errorf("show size = %dx%d, want 0x0", got.ShowWidth, got.ShowHeight)
	}
}

func TestFragLinkExternal(t *testing.T) {
	external := FragLinkFromProto(&protobuf.PbContent{
		Link: "https://tieba.baidu.com/mo/q/checkurl?url=https%3A%2F%2Fexample.com%2Fa%3Fb%3D1",
		Text: "example",
	})
	if !external.IsExternal() {
		t.Error("IsExternal = false, want true")
	}
	if got := external.URL().String(); got != "https://example.com/a?b=1" {
		t.Errorf("URL = %q, want https://example.com/a?b=1", got)
	}

	plain := FragLinkFromProto(&protobuf.PbContent{Link: "https://tieba.baidu.com/p/1", Text: "t"})
	if plain.IsExternal() {
		t.Error("IsExternal = true, want false")
	}
	if got := plain.URL().String(); got != "https://tieba.baidu.com/p/1" {
		t.Errorf("URL = %q", got)
	}
}

func TestFragmentFromProtoVariants(t *testing.T) {
	text := FragTextFromProto(&protobuf.PbContent{Text: "hello"})
	if text.Text != "hello" {
		t.Errorf("text = %q", text.Text)
	}

	emoji := FragEmojiFromProto(&protobuf.PbContent{Text: "image_emoticon1", C: "微笑"})
	if emoji.ID != "image_emoticon1" || emoji.Desc != "微笑" {
		t.Errorf("emoji = %+v", emoji)
	}

	at := FragAtFromProto(&protobuf.PbContent{Text: "@某人 ", Uid: 42})
	if at.Text != "@某人 " || at.UserID != 42 {
		t.Errorf("at = %+v", at)
	}

	item := FragItemFromProto(&protobuf.PbContent{Item: &protobuf.PbContent_Item{ItemName: "某商品"}})
	if item.Text != "某商品" {
		t.Errorf("item = %+v", item)
	}

	plus := FragTiebaPlusFromProto(&protobuf.PbContent{
		TiebaplusInfo: &protobuf.PbContent_TiebaPlusInfo{Desc: "desc", JumpUrl: "https://tb.cn/x"},
	})
	if plus.Text != "desc" || plus.URL.String() != "https://tb.cn/x" {
		t.Errorf("tiebaplus = %+v", plus)
	}

	video := FragVideoFromProto(&protobuf.VideoInfo{
		VideoUrl: "v", ThumbnailUrl: "c", VideoDuration: 10, VideoWidth: 640, VideoHeight: 360, PlayCount: 7,
	})
	if !video.Valid() {
		t.Error("video.Valid() = false, want true")
	}
	if video.Src != "v" || video.CoverSrc != "c" || video.Duration != 10 || video.Width != 640 ||
		video.Height != 360 || video.ViewNum != 7 {
		t.Errorf("video = %+v", video)
	}

	voice := FragVoiceFromProto(&protobuf.Voice{VoiceMd5: "md5", DuringTime: 1500})
	if !voice.Valid() {
		t.Error("voice.Valid() = false, want true")
	}
	if voice.MD5 != "md5" || voice.Duration != 1.5 {
		t.Errorf("voice = %+v", voice)
	}
	if (FragVoice{}).Valid() {
		t.Error("empty voice.Valid() = true, want false")
	}
}

func TestFragmentTextHelpers(t *testing.T) {
	frags := []Fragment{
		FragText{Text: "a"},
		FragAt{Text: "@b "},
		FragImage{Src: "x"},
		FragLink{Text: "c"},
		FragTiebaPlus{Text: "d"},
		FragUnknown{Data: nil},
	}
	if got := FragmentTextOf(frags); got != "a@b cd" {
		t.Errorf("FragmentTextOf = %q, want %q", got, "a@b cd")
	}
}

func TestFragUnknownKeepsRawData(t *testing.T) {
	p := &protobuf.PbContent{Type: 999}
	u := FragUnknownFromProto(p)
	if u.Data != p {
		t.Error("FragUnknownFromProto did not keep the raw proto")
	}
	m := map[string]any{"k": "v"}
	if FragUnknownFromJSON(m).Data == nil {
		t.Error("FragUnknownFromJSON did not keep the raw map")
	}
}

func TestContainers(t *testing.T) {
	c := Containers[int]{Objs: []int{1, 2, 3}}
	if c.Len() != 3 {
		t.Errorf("Len = %d, want 3", c.Len())
	}
	if c.Empty() {
		t.Error("Empty = true, want false")
	}
	if !(Containers[int]{}).Empty() {
		t.Error("empty Containers.Empty() = false, want true")
	}
}
