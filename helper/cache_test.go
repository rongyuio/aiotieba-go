package helper

import (
	"strconv"
	"testing"
)

func TestForumInfoCacheBasics(t *testing.T) {
	c := NewForumInfoCache()

	if _, ok := c.GetFid("天堂鸡汤"); ok {
		t.Error("GetFid on an empty cache reported a hit")
	}
	if _, ok := c.GetFname(1); ok {
		t.Error("GetFname on an empty cache reported a hit")
	}

	c.AddForum("天堂鸡汤", 12345)
	if fid, ok := c.GetFid("天堂鸡汤"); !ok || fid != 12345 {
		t.Errorf("GetFid = %d (present=%v), want 12345", fid, ok)
	}
	if fname, ok := c.GetFname(12345); !ok || fname != "天堂鸡汤" {
		t.Errorf("GetFname = %q (present=%v), want 天堂鸡汤", fname, ok)
	}
	if c.Len() != 1 {
		t.Errorf("Len = %d, want 1", c.Len())
	}
}

func TestForumInfoCacheEvictsOldest(t *testing.T) {
	c := NewForumInfoCache()
	for i := range forumCacheLimit {
		c.AddForum("forum"+strconv.Itoa(i), int64(i))
	}
	if c.Len() != forumCacheLimit {
		t.Fatalf("Len = %d, want %d", c.Len(), forumCacheLimit)
	}

	// Adding one more entry evicts the oldest one (forum0).
	c.AddForum("new", 9999)
	if _, ok := c.GetFid("forum0"); ok {
		t.Error("forum0 was not evicted")
	}
	if _, ok := c.GetFid("forum1"); !ok {
		t.Error("forum1 must be kept")
	}
	if fid, ok := c.GetFid("new"); !ok || fid != 9999 {
		t.Errorf("GetFid(new) = %d (present=%v)", fid, ok)
	}
	if _, ok := c.GetFname(0); ok {
		t.Error("fid 0 was not evicted from the reverse map")
	}
}

func TestDefaultForumInfoCacheExists(t *testing.T) {
	if DefaultForumInfoCache == nil {
		t.Fatal("DefaultForumInfoCache is nil")
	}
	DefaultForumInfoCache.AddForum("默认缓存", 1)
	if fid, ok := DefaultForumInfoCache.GetFid("默认缓存"); !ok || fid != 1 {
		t.Errorf("GetFid = %d (present=%v), want 1", fid, ok)
	}
}

func TestUtilsHelpers(t *testing.T) {
	if !IsPortrait("tb.1.abc") {
		t.Error("IsPortrait(tb.1.abc) = false, want true")
	}
	if IsPortrait("someone") {
		t.Error("IsPortrait(someone) = true, want false")
	}
	if IsPortrait(42) {
		t.Error("IsPortrait(non-string) = true, want false")
	}
	if !IsUserName("someone") {
		t.Error("IsUserName(someone) = false, want true")
	}
	if IsUserName("tb.1.abc") {
		t.Error("IsUserName(tb.1.abc) = true, want false")
	}

	if got := PackJSON(map[string]any{"a": 1}); got != `{"a":1}` {
		t.Errorf("PackJSON = %q, want {\"a\":1}", got)
	}

	var out map[string]any
	if err := ParseJSON([]byte(`{"b":2}`), &out); err != nil {
		t.Fatalf("ParseJSON: %v", err)
	}
	if out["b"] != float64(2) {
		t.Errorf("ParseJSON = %v", out)
	}

	if got := DefaultDatetime(); got.Year() != 1970 || got.Month() != 1 || got.Day() != 1 {
		t.Errorf("DefaultDatetime = %v", got)
	}
}
