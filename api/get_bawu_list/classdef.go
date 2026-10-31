// Package getbawulist 实现吧务团队成员列表接口。
//
// 数据来自吧务中心的 /mo/q/bawulist，比 get_bawu_info 多出招聘余量与会员标识。
package getbawulist

import (
	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper"
)

// BawuMember 吧务成员。
type BawuMember struct {
	UserID      int64  // user_id
	UserName    string // user_name
	NickNameNew string // user_nickname
	RoleName    string // role_name 如 吧主
	Portrait    string // portrait 或 avatar_url
	LevelID     int64  // level_id
	LevelName   string // level_name
	PermCount   int64  // perm_count 已分配权限数
	IsSVIP      bool   // is_svip
}

// BawuMemberFromJSON 对应 BawuMember.from_json。
func BawuMemberFromJSON(m map[string]any) BawuMember {
	portrait := helper.JSONStr(m, "portrait")
	if portrait == "" {
		portrait = helper.JSONStr(m, "avatar_url")
	}
	return BawuMember{
		UserID:      helper.JSONInt(m, "user_id"),
		UserName:    helper.JSONStr(m, "user_name"),
		NickNameNew: helper.JSONStr(m, "user_nickname"),
		RoleName:    helper.JSONStr(m, "role_name"),
		Portrait:    classdef.TrimPortrait(portrait),
		LevelID:     helper.JSONInt(m, "level_id"),
		LevelName:   helper.JSONStr(m, "level_name"),
		PermCount:   helper.JSONInt(m, "perm_count"),
		IsSVIP:      helper.JSONInt(m, "is_svip") != 0,
	}
}

// ShowName 显示名称。
func (m BawuMember) ShowName() string {
	if m.NickNameNew != "" {
		return m.NickNameNew
	}
	return m.UserName
}

// BawuList 吧务团队列表与招聘信息。
type BawuList struct {
	classdef.Containers[*BawuMember]

	Helpers []*BawuMember // 吧务助手

	ManagerLeftNum int64 // 还能任命几个吧主
	AssistLeftNum  int64 // 还能招募几个小吧主
	IsPrivateForum bool  // 是否私密吧

	Err error // 捕获的异常
}

// BawuListFromJSON 对应 BawuList.from_json。
//
// 客户端把每个角色分桶展示，键名见 bawu-team.js 的 bawuCategoryList：
// manager / assist / picadmin / voiceadmin / videoadmin / forum_helper。
func BawuListFromJSON(data map[string]any) BawuList {
	var out BawuList
	bl := helper.JSONMap(data, "bawu_list")
	collect := func(key string, dst *[]*BawuMember) {
		for _, item := range helper.JSONSlice(bl, key) {
			if im, ok := item.(map[string]any); ok {
				m := BawuMemberFromJSON(im)
				*dst = append(*dst, &m)
			}
		}
	}
	collect("manager", &out.Objs)
	collect("assist", &out.Objs)
	collect("picadmin", &out.Objs)
	collect("voiceadmin", &out.Objs)
	collect("videoadmin", &out.Objs)
	collect("forum_helper", &out.Helpers)
	if apply := helper.JSONMap(data, "manager_apply_info"); apply != nil {
		out.ManagerLeftNum = helper.JSONInt(apply, "manager_left_num")
		out.AssistLeftNum = helper.JSONInt(apply, "assist_left_num")
	}
	out.IsPrivateForum = helper.JSONInt(data, "is_private_forum") != 0
	return out
}
