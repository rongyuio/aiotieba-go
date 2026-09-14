// Package setbawuperm 实现 aiotieba 的 set_bawu_perm API。
//
// 对应 Python 包 aiotieba.api.set_bawu_perm。
package setbawuperm

import (
	"context"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// permSetting 是 `perm_setting` JSON 数组中的一个元素。
//
// 字段顺序很重要，因为 JSON 文本本身是签名载荷的一部分。
type permSetting struct {
	Switch int `json:"switch"`
	Perm   int `json:"perm"`
}

// perm2id 将权限映射到服务端期望的 id，对应 Python 模块的 perm2id 表。
var perm2id = []struct {
	perm enums.BawuPermType
	id   int
}{
	{enums.BawuPermUnblock, 4},
	{enums.BawuPermUnblockAppeal, 5},
	{enums.BawuPermRecover, 3},
	{enums.BawuPermRecoverAppeal, 2},
}

// PackPermSettings 对应 pack_perm_settings。
func PackPermSettings(perms enums.BawuPermType) []permSetting {
	settings := make([]permSetting, 0, len(perm2id))
	for _, entry := range perm2id {
		settings = append(settings, permSetting{
			Switch: helper.BoolInt(perms&entry.perm != 0),
			Perm:   entry.id,
		})
	}
	return settings
}

// ParseBody 对应 parse_body。
func ParseBody(body []byte) error {
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return err
	}
	if code := helper.JSONInt(res, "no"); code != 0 {
		return &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(res, "error")}
	}
	return nil
}

// RequestURL 返回该 API 的请求地址。
func RequestURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.WebBaseHost, Path: "/mo/q/setAuthToolPerm"}
}

// Request 对应 request。
func Request(ctx context.Context, httpCore *core.HttpCore, fid int64, portrait string, perms enums.BawuPermType) error {
	data := []crypto.Param{
		{Key: "forum_id", Value: fid},
		{Key: "auth_user_portrait", Value: portrait},
		{Key: "perm_setting", Value: helper.PackJSON(PackPermSettings(perms))},
	}

	resp, err := httpCore.WebForm(data).SetContext(ctx).Post(RequestURL().String())
	if err != nil {
		return err
	}
	return ParseBody(resp.Body())
}
