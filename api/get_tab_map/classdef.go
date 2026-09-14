// Package gettabmap 实现 aiotieba 的 get_tab_map API。
//
// 对应 Python 包 aiotieba.api.get_tab_map。
package gettabmap

import (
	pb "github.com/rongyuio/aiotieba-go/api/get_tab_map/protobuf"
)

// TabMap 分区名到分区id的映射。
type TabMap struct {
	Map map[string]int64 // 分区名到分区id的映射
}

// TabMapFromProto 对应 TabMap.from_proto。
func TabMapFromProto(p *pb.SearchPostForumResIdl_DataRes) TabMap {
	tabs := p.GetExactMatch().GetTabInfo()
	out := make(map[string]int64, len(tabs))
	for _, tab := range tabs {
		out[tab.GetTabName()] = int64(tab.GetTabId())
	}
	return TabMap{Map: out}
}

// Get 返回指定分区的 id，对应 __getitem__。
func (t TabMap) Get(tabName string) (int64, bool) {
	id, ok := t.Map[tabName]
	return id, ok
}

// Len 返回分区数量，对应 __len__。
func (t TabMap) Len() int { return len(t.Map) }

// Valid 对应 __bool__。
func (t TabMap) Valid() bool { return len(t.Map) != 0 }
