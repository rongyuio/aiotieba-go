// Package getrecomstatus implements the get_recom_status API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_recom_status.
package getrecomstatus

import "github.com/rongyuio/aiotieba-go/helper"

// RecomStatus mirrors RecomStatus.
type RecomStatus struct {
	TotalRecomNum int64
	UsedRecomNum  int64
	Err           error
}

// RecomStatusFromJSON mirrors RecomStatus.from_json.
func RecomStatusFromJSON(m map[string]any) RecomStatus {
	return RecomStatus{
		TotalRecomNum: helper.JSONInt(m, "total_recommend_num"),
		UsedRecomNum:  helper.JSONInt(m, "used_recommend_num"),
	}
}
