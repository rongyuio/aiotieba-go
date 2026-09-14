package helper

import "sync"

// forumCacheLimit is the maximum number of entries of the forum cache, mirroring
// the Python constant of 128.
const forumCacheLimit = 128

// ForumInfoCache caches the fname <-> fid mapping. It mirrors
// aiotieba.helper.cache.ForumInfoCache.
//
// The Python cache keeps both dictionaries in insertion order and evicts the
// oldest entry (OrderedDict.popitem(last=False)), so this port keeps the
// insertion order in a slice.
type ForumInfoCache struct {
	mu        sync.Mutex
	fname2fid map[string]int64
	fid2fname map[int64]string
	order     []string
}

// DefaultForumInfoCache is the process wide cache used by the API modules,
// mirroring the class-level dictionaries of the Python original.
var DefaultForumInfoCache = NewForumInfoCache()

// NewForumInfoCache creates an empty cache.
func NewForumInfoCache() *ForumInfoCache {
	return &ForumInfoCache{
		fname2fid: map[string]int64{},
		fid2fname: map[int64]string{},
	}
}

// GetFid returns the fid of fname and whether it was cached.
func (c *ForumInfoCache) GetFid(fname string) (int64, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	fid, ok := c.fname2fid[fname]
	return fid, ok
}

// GetFname returns the fname of fid and whether it was cached.
func (c *ForumInfoCache) GetFname(fid int64) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	fname, ok := c.fid2fname[fid]
	return fname, ok
}

// AddForum caches the fname <-> fid mapping, evicting the oldest entry when the
// cache is full.
func (c *ForumInfoCache) AddForum(fname string, fid int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.fname2fid) == forumCacheLimit && len(c.order) > 0 {
		oldest := c.order[0]
		c.order = c.order[1:]
		if oldFid, ok := c.fname2fid[oldest]; ok {
			delete(c.fname2fid, oldest)
			delete(c.fid2fname, oldFid)
		}
	}

	if _, exists := c.fname2fid[fname]; !exists {
		c.order = append(c.order, fname)
	}
	c.fname2fid[fname] = fid
	c.fid2fname[fid] = fname
}

// Len returns the number of cached entries.
func (c *ForumInfoCache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.fname2fid)
}
