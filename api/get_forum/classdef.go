// Package getforum 实现 aiotieba 的 get_forum API。
//
// 对应 Python 包 aiotieba.api.get_forum。
package getforum

import (
	"github.com/rongyuio/aiotieba-go/helper"
)

// Forum 贴吧信息。
type Forum struct {
	FID   int64  // 贴吧id
	FName string // 贴吧名

	Category    string // 一级分类
	Subcategory string // 二级分类

	SmallAvatar string // 吧头像(小)
	Slogan      string // 吧标语
	MemberNum   int64  // 吧会员数
	PostNum     int64  // 发帖数
	ThreadNum   int64  // 主题帖数

	HasBawu bool // 是否有吧务

	Err error // 捕获的异常
}

// ForumFromJSON 对应 Forum.from_json。
func ForumFromJSON(data map[string]any) Forum {
	return Forum{
		FID:         helper.JSONInt(data, "id"),
		FName:       helper.JSONStr(data, "name"),
		Category:    helper.JSONStr(data, "first_class"),
		Subcategory: helper.JSONStr(data, "second_class"),
		SmallAvatar: helper.JSONStr(data, "avatar"),
		Slogan:      helper.JSONStr(data, "slogan"),
		MemberNum:   helper.JSONInt(data, "member_num"),
		PostNum:     helper.JSONInt(data, "post_num"),
		ThreadNum:   helper.JSONInt(data, "thread_num"),
		HasBawu:     helper.HasJSONKey(data, "managers"),
	}
}
