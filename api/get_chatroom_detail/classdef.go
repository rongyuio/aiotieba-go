// Package getchatroomdetail 实现 aiotieba 的 get_chatroom_detail API。
//
// 对应 Python 包 aiotieba.api.get_chatroom_detail。
package getchatroomdetail

import (
	"github.com/rongyuio/aiotieba-go/helper"
	tbcrypto "github.com/rongyuio/aiotieba-go/helper/crypto"
)

// RobotSkill 吧群机器人的技能。
type RobotSkill struct {
	Type   int64            // 技能类型
	Name   string           // 技能名称
	Alias  string           // 技能别名
	Desc   string           // 技能描述
	IsShow bool             // 是否展示
	Items  []map[string]any // 技能参数列表
}

// RobotSkillFromJSON 对应 RobotSkill.from_json。
func RobotSkillFromJSON(data map[string]any) RobotSkill {
	items := []map[string]any{}
	for _, item := range helper.JSONSlice(data, "items") {
		if m, ok := item.(map[string]any); ok {
			items = append(items, m)
		}
	}
	return RobotSkill{
		Type:   helper.JSONInt(data, "type"),
		Name:   helper.JSONStr(data, "name"),
		Alias:  helper.JSONStr(data, "alias"),
		Desc:   helper.JSONStr(data, "desc"),
		IsShow: helper.JSONBool(data, "is_show"),
		Items:  items,
	}
}

// Robot 吧群机器人。
type Robot struct {
	Name      string       // 用户名
	NameShow  string       // 显示名称
	Portrait  string       // 头像
	UK        string       // baidu_uk
	UserID    int64        // 用户id
	RobotRole int64        // 机器人角色
	Skills    []RobotSkill // 技能列表
}

// RobotFromJSON 对应 Robot.from_json。
//
// user_id 由 user.uk 解码得到，与 Python 版调用 bduk2user_id 一致；解码失败即返回错误，
// 不静默落下 0。
func RobotFromJSON(data map[string]any) (Robot, error) {
	user := helper.JSONMap(data, "user")
	uk := helper.JSONStr(user, "uk")
	userID, err := tbcrypto.BDUKToUserID(uk)
	if err != nil {
		return Robot{}, err
	}

	skills := []RobotSkill{}
	for _, item := range helper.JSONSlice(data, "skill") {
		if m, ok := item.(map[string]any); ok {
			skills = append(skills, RobotSkillFromJSON(m))
		}
	}

	return Robot{
		Name:      helper.JSONStr(user, "name"),
		NameShow:  helper.JSONStr(user, "name_show"),
		Portrait:  helper.JSONStr(user, "portrait"),
		UK:        uk,
		UserID:    userID,
		RobotRole: helper.JSONInt(user, "robot_role"),
		Skills:    skills,
	}, nil
}

// ChatroomDetail 吧群聊天室详情。
type ChatroomDetail struct {
	RoomID             int64   // 聊天室id
	ForumID            int64   // 吧id
	ForumName          string  // 吧名
	Name               string  // 群名
	Avatar             string  // 群头像
	DiscussNum         int64   // 群成员数
	Announce           string  // 群公告
	TalkThresholdLevel int64   // 发言等级门槛
	IsTalkBanned       bool    // 是否被禁言
	Robots             []Robot // 机器人列表

	Err error // 捕获的异常
}

// ChatroomDetailFromJSON 对应 ChatroomDetail.from_json。
func ChatroomDetailFromJSON(data map[string]any) (ChatroomDetail, error) {
	basicInfo := helper.JSONMap(data, "basic_info")
	extraInfo := helper.JSONMap(data, "extra_info")
	bots := helper.JSONMap(data, "bots")

	robots := []Robot{}
	for _, item := range helper.JSONSlice(bots, "bot_list") {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		robot, err := RobotFromJSON(m)
		if err != nil {
			return ChatroomDetail{}, err
		}
		robots = append(robots, robot)
	}

	return ChatroomDetail{
		RoomID:             helper.JSONInt(basicInfo, "id"),
		ForumID:            helper.JSONInt(basicInfo, "forum_id"),
		ForumName:          helper.JSONStr(basicInfo, "forum_name"),
		Name:               helper.JSONStr(basicInfo, "name"),
		Avatar:             helper.JSONStr(basicInfo, "avatar"),
		DiscussNum:         helper.JSONInt(basicInfo, "discuss_num"),
		Announce:           helper.JSONStr(extraInfo, "announce"),
		TalkThresholdLevel: helper.JSONInt(extraInfo, "talk_threshold_level"),
		IsTalkBanned:       helper.JSONBool(extraInfo, "is_talk_banned"),
		Robots:             robots,
	}, nil
}
