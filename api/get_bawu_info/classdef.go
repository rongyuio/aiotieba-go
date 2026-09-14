// Package getbawuinfo 实现 aiotieba 的 get_bawu_info API。
//
// 对应 Python 包 aiotieba.api.get_bawu_info。
package getbawuinfo

import (
	"strconv"

	pb "github.com/rongyuio/aiotieba-go/api/get_bawu_info/protobuf"
)

// UserInfoBawu 用户信息。
//
// 注意：此处的 portrait 不做裁剪，与大多数其他用户类型不同。
type UserInfoBawu struct {
	UserID      int64  // user_id
	Portrait    string // portrait
	UserName    string // 用户名
	NickNameNew string // 新版昵称
	Level       int64  // 等级
}

// UserInfoBawuFromProto 对应 UserInfo_bawu.from_proto。
func UserInfoBawuFromProto(p *pb.GetBawuInfoResIdl_DataRes_BawuTeam_BawuRoleDes_BawuRoleInfoPub) UserInfoBawu {
	return UserInfoBawu{
		UserID:      p.GetUserId(),
		Portrait:    p.GetPortrait(),
		UserName:    p.GetUserName(),
		NickNameNew: p.GetNameShow(),
		Level:       int64(p.GetUserLevel()),
	}
}

// NickName 用户昵称。
func (u UserInfoBawu) NickName() string { return u.NickNameNew }

// ShowName 显示名称。
func (u UserInfoBawu) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// String 对应 __str__。
func (u UserInfoBawu) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// LogName 用于在日志中记录用户信息。
func (u UserInfoBawu) LogName() string {
	switch {
	case u.UserName != "":
		return u.UserName
	case u.Portrait != "":
		return u.NickNameNew + "/" + u.Portrait
	default:
		return strconv.FormatInt(u.UserID, 10)
	}
}

// Valid 对应 __bool__。
func (u UserInfoBawu) Valid() bool { return u.UserID != 0 }

// BawuInfo 吧务团队信息。
type BawuInfo struct {
	All []UserInfoBawu // 所有吧务

	Admin              []UserInfoBawu // 大吧主
	Manager            []UserInfoBawu // 小吧主
	VoiceEditor        []UserInfoBawu // 语音小编
	ImageEditor        []UserInfoBawu // 图片小编
	VideoEditor        []UserInfoBawu // 视频小编
	BroadcastEditor    []UserInfoBawu // 广播小编
	JournalChiefEditor []UserInfoBawu // 吧刊主编
	JournalEditor      []UserInfoBawu // 吧刊小编
	ProfessAdmin       []UserInfoBawu // 职业吧主
	FourthAdmin        []UserInfoBawu // 第四吧主
}

// BawuInfoFromProto 对应 BawuInfo.from_proto。
func BawuInfoFromProto(p *pb.GetBawuInfoResIdl_DataRes) BawuInfo {
	var info BawuInfo
	slots := map[string]*[]UserInfoBawu{
		"吧主":   &info.Admin,
		"小吧主":  &info.Manager,
		"语音小编": &info.VoiceEditor,
		"图片小编": &info.ImageEditor,
		"视频小编": &info.VideoEditor,
		"广播小编": &info.BroadcastEditor,
		"吧刊主编": &info.JournalChiefEditor,
		"吧刊小编": &info.JournalEditor,
		"职业吧主": &info.ProfessAdmin,
		"第四吧主": &info.FourthAdmin,
	}
	// 下面各键的顺序与 Python 的 extract() 调用顺序一致，
	// 因此也就是 `all` 的顺序。
	order := []string{"吧主", "小吧主", "语音小编", "图片小编", "视频小编", "广播小编", "吧刊主编", "吧刊小编", "职业吧主", "第四吧主"}

	team := p.GetBawuTeamInfo()
	byRole := make(map[string][]UserInfoBawu, len(team.GetBawuTeamList()))
	for _, roleDes := range team.GetBawuTeamList() {
		infos := roleDes.GetRoleInfo()
		users := make([]UserInfoBawu, 0, len(infos))
		for _, item := range infos {
			users = append(users, UserInfoBawuFromProto(item))
		}
		byRole[roleDes.GetRoleName()] = users
	}

	for _, role := range order {
		users := byRole[role]
		if len(users) == 0 {
			continue
		}
		if slot, ok := slots[role]; ok {
			*slot = users
		}
		info.All = append(info.All, users...)
	}

	return info
}
