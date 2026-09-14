package helper

import "sync"

// forumCacheLimit 是吧信息缓存的最大条目数，对应 Python 常量 128。
const forumCacheLimit = 128

// ForumInfoCache 吧信息缓存，缓存 fname <-> fid 映射，对应 aiotieba.helper.cache.ForumInfoCache。
//
// Python 的缓存用有序字典保持插入顺序并淘汰最早条目（OrderedDict.popitem(last=False)），
// 因此本移植版用切片维护插入顺序。
type ForumInfoCache struct {
	mu        sync.Mutex
	fname2fid map[string]int64
	fid2fname map[int64]string
	order     []string
}

// DefaultForumInfoCache 是各 API 模块共用的进程级缓存，对应 Python 原版的类属性字典。
var DefaultForumInfoCache = NewForumInfoCache()

// NewForumInfoCache 创建一个空缓存。
func NewForumInfoCache() *ForumInfoCache {
	return &ForumInfoCache{
		fname2fid: map[string]int64{},
		fid2fname: map[int64]string{},
	}
}

// GetFid 通过贴吧名获取 forum_id，并报告是否命中缓存。
func (c *ForumInfoCache) GetFid(fname string) (int64, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	fid, ok := c.fname2fid[fname]
	return fid, ok
}

// GetFname 通过 forum_id 获取贴吧名，并报告是否命中缓存。
func (c *ForumInfoCache) GetFname(fid int64) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	fname, ok := c.fid2fname[fid]
	return fname, ok
}

// AddForum 将贴吧名与 forum_id 的映射关系添加到缓存，缓存满时淘汰最早的条目。
//
// 缓存满时淘汰最早的条目。
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

// Len 返回缓存的条目数。
func (c *ForumInfoCache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.fname2fid)
}
