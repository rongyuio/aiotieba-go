// Package getbawuinfo implements the get_bawu_info API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_bawu_info.
package getbawuinfo

import (
	"strconv"

	pb "github.com/rongyuio/aiotieba-go/api/get_bawu_info/protobuf"
)

// UserInfoBawu is the information of a moderator. It mirrors
// aiotieba.api.get_bawu_info._classdef.UserInfo_bawu.
//
// Note that the portrait is NOT trimmed here, unlike most other user types.
type UserInfoBawu struct {
	UserID      int64
	Portrait    string
	UserName    string
	NickNameNew string
	Level       int64
}

// UserInfoBawuFromProto mirrors UserInfo_bawu.from_proto.
func UserInfoBawuFromProto(p *pb.GetBawuInfoResIdl_DataRes_BawuTeam_BawuRoleDes_BawuRoleInfoPub) UserInfoBawu {
	return UserInfoBawu{
		UserID:      p.GetUserId(),
		Portrait:    p.GetPortrait(),
		UserName:    p.GetUserName(),
		NickNameNew: p.GetNameShow(),
		Level:       int64(p.GetUserLevel()),
	}
}

// NickName mirrors the nick_name property.
func (u UserInfoBawu) NickName() string { return u.NickNameNew }

// ShowName mirrors the show_name property.
func (u UserInfoBawu) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// String mirrors __str__.
func (u UserInfoBawu) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	if u.Portrait != "" {
		return u.Portrait
	}
	return strconv.FormatInt(u.UserID, 10)
}

// LogName mirrors the log_name property.
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

// Valid mirrors __bool__.
func (u UserInfoBawu) Valid() bool { return u.UserID != 0 }

// BawuInfo is the moderator team of a forum. It mirrors
// aiotieba.api.get_bawu_info._classdef.BawuInfo.
type BawuInfo struct {
	All []UserInfoBawu

	Admin              []UserInfoBawu
	Manager            []UserInfoBawu
	VoiceEditor        []UserInfoBawu
	ImageEditor        []UserInfoBawu
	VideoEditor        []UserInfoBawu
	BroadcastEditor    []UserInfoBawu
	JournalChiefEditor []UserInfoBawu
	JournalEditor      []UserInfoBawu
	ProfessAdmin       []UserInfoBawu
	FourthAdmin        []UserInfoBawu
}

// BawuInfoFromProto mirrors BawuInfo.from_proto.
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
	// The order of the keys below mirrors the Python extract() call order and
	// therefore the order of `all`.
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
