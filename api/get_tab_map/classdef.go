// Package gettabmap implements the get_tab_map API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_tab_map.
package gettabmap

import (
	pb "github.com/rongyuio/aiotieba/api/get_tab_map/protobuf"
)

// TabMap maps a tab name to its tab id. It mirrors
// aiotieba.api.get_tab_map._classdef.TabMap.
type TabMap struct {
	Map map[string]int64
}

// TabMapFromProto mirrors TabMap.from_proto.
func TabMapFromProto(p *pb.SearchPostForumResIdl_DataRes) TabMap {
	tabs := p.GetExactMatch().GetTabInfo()
	out := make(map[string]int64, len(tabs))
	for _, tab := range tabs {
		out[tab.GetTabName()] = int64(tab.GetTabId())
	}
	return TabMap{Map: out}
}

// Get returns the id of the named tab, mirroring __getitem__.
func (t TabMap) Get(tabName string) (int64, bool) {
	id, ok := t.Map[tabName]
	return id, ok
}

// Len returns the number of tabs. It mirrors __len__.
func (t TabMap) Len() int { return len(t.Map) }

// Valid mirrors __bool__.
func (t TabMap) Valid() bool { return len(t.Map) != 0 }
