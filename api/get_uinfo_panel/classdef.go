// Package getuinfopanel implements the get_uinfo_panel API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_uinfo_panel.
package getuinfopanel

import (
	"strconv"
	"strings"

	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/helper"
)

// tenThousandSuffix is the Chinese "万" (ten thousand) suffix used by the
// server for large counters.
const tenThousandSuffix = "万"

// UserInfoPanel is the user information returned by /home/get/panel. It mirrors
// aiotieba.api.get_uinfo_panel._classdef.UserInfo_panel.
type UserInfoPanel struct {
	Portrait    string
	UserName    string
	NickNameNew string
	NickNameOld string
	Gender      enums.Gender
	Age         float64
	PostNum     int64
	FanNum      int64
	IsVIP       bool
}

// UserInfoPanelFromJSON mirrors UserInfo_panel.from_json.
func UserInfoPanelFromJSON(data map[string]any) UserInfoPanel {
	var gender enums.Gender
	switch helper.JSONStr(data, "sex") {
	case "male":
		gender = enums.GenderMale
	case "female":
		gender = enums.GenderFemale
	default:
		gender = enums.GenderUnknown
	}

	var age float64
	if tbAge := helper.JSONStr(data, "tb_age"); tbAge != "-" {
		age = helper.AnyFloat64(tbAge)
	}

	isVIP := false
	if vipInfo := helper.JSONMap(data, "vipInfo"); len(vipInfo) > 0 {
		isVIP = helper.JSONInt(vipInfo, "v_status") == 3
	}

	return UserInfoPanel{
		Portrait:    helper.JSONStr(data, "portrait"),
		UserName:    helper.JSONStr(data, "name"),
		NickNameNew: helper.JSONStr(data, "show_nickname"),
		NickNameOld: helper.JSONStr(data, "name_show"),
		Gender:      gender,
		Age:         age,
		PostNum:     tbNum2Int(data, "post_num"),
		FanNum:      tbNum2Int(data, "followed_count"),
		IsVIP:       isVIP,
	}
}

// tbNum2Int mirrors _tbnum2int: a "1.2万" counter becomes 12000.
func tbNum2Int(data map[string]any, key string) int64 {
	raw, ok := helper.JSONRaw(data, key)
	if !ok {
		return 0
	}
	s, isString := raw.(string)
	if !isString {
		return helper.AnyInt64(raw)
	}
	s = strings.TrimSuffix(s, tenThousandSuffix)
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return int64(f * 1e4)
}

// NickName mirrors the nick_name property.
func (u UserInfoPanel) NickName() string { return u.NickNameNew }

// ShowName mirrors the show_name property.
func (u UserInfoPanel) ShowName() string {
	if u.NickNameNew != "" {
		return u.NickNameNew
	}
	return u.UserName
}

// String mirrors __str__.
func (u UserInfoPanel) String() string {
	if u.UserName != "" {
		return u.UserName
	}
	return u.Portrait
}

// LogName mirrors the log_name property.
func (u UserInfoPanel) LogName() string {
	if u.UserName != "" {
		return u.UserName
	}
	return u.NickNameNew + "/" + u.Portrait
}

// Valid mirrors __bool__, which is truthy while a portrait is present.
func (u UserInfoPanel) Valid() bool { return u.Portrait != "" }
