// Package getbawuinfo 实现 aiotieba 的 get_bawu_info API。
//
// 对应 Python 包 aiotieba.api.get_bawu_info。
package getbawuinfo

import (
	"strconv"

	pb "github.com/rongyuio/aiotieba-go/api/get_bawu_info/protobuf"
	"github.com/rongyuio/aiotieba-go/enums"
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

	Err error // 捕获的异常
}

// BawuInfoFromProto 对应 BawuInfo.from_proto。
//
// 分桶依据是每条 role_info 内层的 role_name（英文枚举名），不是外层
// BawuRoleDes.role_name —— 后者是给人看的中文名，上游已经不再用它分桶。
// 无法识别的角色直接跳过，因此 All 里也不会有它。
func BawuInfoFromProto(p *pb.GetBawuInfoResIdl_DataRes) BawuInfo {
	var info BawuInfo
	slots := map[enums.BawuType]*[]UserInfoBawu{
		enums.BawuAdmin:              &info.Admin,
		enums.BawuManager:            &info.Manager,
		enums.BawuVoiceEditor:        &info.VoiceEditor,
		enums.BawuImageEditor:        &info.ImageEditor,
		enums.BawuVideoEditor:        &info.VideoEditor,
		enums.BawuBroadcastEditor:    &info.BroadcastEditor,
		enums.BawuJournalChiefEditor: &info.JournalChiefEditor,
		enums.BawuJournalEditor:      &info.JournalEditor,
		enums.BawuProfessAdmin:       &info.ProfessAdmin,
		enums.BawuFourthAdmin:        &info.FourthAdmin,
	}

	for _, roleDes := range p.GetBawuTeamInfo().GetBawuTeamList() {
		for _, item := range roleDes.GetRoleInfo() {
			bawuType := enums.BawuTypeFrom(item.GetRoleName())
			if bawuType == enums.BawuUnknown {
				continue
			}
			user := UserInfoBawuFromProto(item)
			info.All = append(info.All, user)
			if slot, ok := slots[bawuType]; ok {
				*slot = append(*slot, user)
			}
		}
	}

	return info
}
