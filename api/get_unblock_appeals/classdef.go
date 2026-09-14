// Package getunblockappeals implements the get_unblock_appeals API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_unblock_appeals.
package getunblockappeals

import (
	"github.com/rongyuio/aiotieba/api/classdef"
	"github.com/rongyuio/aiotieba/helper"
)

// Appeal mirrors Appeal.
type Appeal struct {
	UserID       int64
	Portrait     string
	UserName     string
	NickName     string
	AppealID     int64
	AppealReason string
	AppealTime   int64
	PunishReason string
	PunishTime   int64
	PunishDay    int64
	OpName       string
}

// AppealFromJSON mirrors Appeal.from_json.
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

// Appeals mirrors Appeals.
type Appeals struct {
	classdef.Containers[*Appeal]

	HasMore bool
	Err     error
}

// AppealsFromJSON mirrors Appeals.from_json.
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
