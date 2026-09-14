// Package getstatistics 实现 aiotieba 的 get_statistics API。
//
// 对应 Python 包 aiotieba.api.get_statistics。
package getstatistics

import "github.com/rongyuio/aiotieba-go/helper"

// Statistics 吧务后台统计信息。
// 时间从旧到新
type Statistics struct {
	View      []int64 // 浏览量
	Thread    []int64 // 主题帖数
	NewMember []int64 // 新增吧会员数
	Post      []int64 // 回复数
	SignRatio []int64 // 签到率
	AvgTime   []int64 // 人均浏览时长
	AvgTimes  []int64 // 人均进吧次数
	Recommend []int64 // 首页推荐数
}

// StatisticsFromJSON 对应 Statistics.from_json，接收解码后的 "data" 序列。
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
