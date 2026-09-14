// Package getstatistics implements the get_statistics API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_statistics.
package getstatistics

import "github.com/rongyuio/aiotieba/helper"

// Statistics mirrors Statistics: the eight time series of the forum backend.
type Statistics struct {
	View      []int64
	Thread    []int64
	NewMember []int64
	Post      []int64
	SignRatio []int64
	AvgTime   []int64
	AvgTimes  []int64
	Recommend []int64
}

// StatisticsFromJSON mirrors Statistics.from_json, taking the decoded "data"
// sequence.
func StatisticsFromJSON(seq []any) Statistics {
	extract := func(i int) []int64 {
		if i < 0 || i >= len(seq) {
			return nil
		}
		entry, ok := seq[i].(map[string]any)
		if !ok {
			return nil
		}
		group, ok := entry["group"].([]any)
		if !ok || len(group) < 2 {
			return nil
		}
		g1, ok := group[1].(map[string]any)
		if !ok {
			return nil
		}
		values, ok := g1["values"].([]any)
		if !ok {
			return nil
		}
		out := make([]int64, 0, len(values))
		for _, v := range values {
			if vm, ok := v.(map[string]any); ok {
				out = append(out, helper.JSONInt(vm, "value"))
			}
		}
		return out
	}

	return Statistics{
		View:      extract(0),
		Thread:    extract(1),
		NewMember: extract(2),
		Post:      extract(3),
		SignRatio: extract(4),
		AvgTime:   extract(5),
		AvgTimes:  extract(6),
		Recommend: extract(7),
	}
}
