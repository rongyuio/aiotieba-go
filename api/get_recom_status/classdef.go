// Package getrecomstatus 实现 aiotieba 的 get_recom_status API。
//
// 对应 Python 包 aiotieba.api.get_recom_status。
package getrecomstatus

import "github.com/rongyuio/aiotieba-go/helper"

// RecomStatus 大吧主推荐功能的月度配额状态。
type RecomStatus struct {
	TotalRecomNum int64 // 本月总推荐配额
	UsedRecomNum  int64 // 本月已使用的推荐配额
	Err           error // 捕获的异常
}

// RecomStatusFromJSON 对应 RecomStatus.from_json。
func RecomStatusFromJSON(m map[string]any) RecomStatus {
	return RecomStatus{
		TotalRecomNum: helper.JSONInt(m, "total_recommend_num"),
		UsedRecomNum:  helper.JSONInt(m, "used_recommend_num"),
	}
}
