// Package getunblockappeals 实现 aiotieba 的 get_unblock_appeals API。
//
// 对应 Python 包 aiotieba.api.get_unblock_appeals。
package getunblockappeals

import (
	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper"
)

// Appeal 申诉请求信息。
type Appeal struct {
	UserID       int64  // 申诉用户id
	Portrait     string // 申诉用户portrait
	UserName     string // 申诉用户名
	NickName     string // 申诉用户昵称
	AppealID     int64  // 申诉id
	AppealReason string // 申诉理由
	AppealTime   int64  // 申诉时间 10位时间戳 以秒为单位
	PunishReason string // 封禁理由
	PunishTime   int64  // 封禁开始时间 10位时间戳 以秒为单位
	PunishDay    int64  // 封禁天数
	OpName       string // 操作人用户名
}

// AppealFromJSON 对应 Appeal.from_json。
func AppealFromJSON(m map[string]any) Appeal {
	user := helper.JSONMap(m, "user")
	return Appeal{
		UserID:       helper.JSONInt(user, "id"),
		Portrait:     classdef.TrimPortrait(helper.JSONStr(user, "portrait")),
		UserName:     helper.JSONStr(user, "name"),
		NickName:     helper.JSONStr(user, "name_show"),
		AppealID:     helper.JSONInt(m, "appeal_id"),
		AppealReason: helper.JSONStr(m, "appeal_reason"),
		AppealTime:   helper.JSONInt(m, "appeal_time"),
		PunishReason: helper.JSONStr(m, "punish_reason"),
		PunishTime:   helper.JSONInt(m, "punish_start_time"),
		PunishDay:    helper.JSONInt(m, "punish_day_num"),
		OpName:       helper.JSONStr(m, "operate_man"),
	}
}

// Appeals 申诉请求列表。
type Appeals struct {
	classdef.Containers[*Appeal]

	HasMore bool  // 是否还有下一页
	Err     error // 捕获的异常
}

// AppealsFromJSON 对应 Appeals.from_json。
func AppealsFromJSON(m map[string]any) Appeals {
	var appeals Appeals
	for _, item := range helper.JSONSlice(m, "appeal_list") {
		if im, ok := item.(map[string]any); ok {
			a := AppealFromJSON(im)
			appeals.Objs = append(appeals.Objs, &a)
		}
	}
	appeals.HasMore = helper.JSONBool(m, "has_more")
	return appeals
}
