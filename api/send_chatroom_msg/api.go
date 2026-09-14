// Package sendchatroommsg 实现 aiotieba 的 send_chatroom_msg API。
//
// 对应 Python 包 aiotieba.api.send_chatroom_msg。该 API 仅可通过 BLCP 传输使用。
package sendchatroommsg

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strconv"
	"time"

	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// appConstants 对应 Python 模块的 AppConstants dataclass。
const (
	appID      = int64(10773430)
	sdkVersion = int64(11250036)
)

// 聊天室消息请求的 serviceID 与 methodID。
const (
	serviceID = int64(3)
	methodID  = int64(185)
)

// bdukKey 与 bdukIV 是 getBDUKfromUserId 使用的固定密钥与 IV。
var (
	bdukKey = []byte("AFD311832EDEEAEF")
	bdukIV  = []byte("2011121211143000")
)

// BDUKFromUserID 对应 BLCPCore.getBDUKfromUserId：数字 user id 会使用固定密钥与 IV
// 做 AES-CBC 加密，然后以无填充的 urlsafe-base64 编码。
func BDUKFromUserID(userID string) string {
	enc, err := crypto.CBCEncrypt(bdukKey, bdukIV, []byte(userID))
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(enc)
}

// getMsgKey 对应 BLCPCore.getmsgkey。
func getMsgKey(bduk string) string {
	return bduk + strconv.FormatInt(time.Now().UnixMilli()*1000, 10) + strconv.FormatInt(rand.Int64(), 10)
}

// ConstructRequestData 构造聊天室消息的 Lcm 请求数据，包含 app_safe_ext、content
// 以及承载名字、头像、昵称颜色、大会员标志等 UI 展示信息的 main_data。
//
// 对应 construct_request_data。
//
// 它导出以便测试。返回值会交给 BLCPCore.SendLcm，由后者注入 client_logid 与 rpc。
func ConstructRequestData(
	blcpCore *core.BLCPCore,
	roomID, uk, userID, originID int64,
	name, portrait, text string,
	fid, level int64,
	vip bool,
	glevel int64,
	atdata []map[string]any,
	robot int64,
) map[string]any {
	now := time.Now()
	ts := strconv.FormatInt(now.Unix(), 10)
	portraitTS := portrait + "?t=" + ts

	zid := ""
	if blcpCore != nil {
		zid = blcpCore.Account.ZID()
	}
	appSafeExt := map[string]any{"haotianjing": map[string]any{"zid": zid}}

	textDict := map[string]any{
		"room_id":      strconv.FormatInt(roomID, 10),
		"type":         "0",
		"to_uid":       "0",
		"vip":          strconv.Itoa(helper.BoolInt(vip)),
		"name":         name,
		"portrait":     portraitTS,
		"content_type": "0",
		"content_body": marshalJSON(map[string]any{"text": text}),
		"src":          "",
		"baidu_uk":     BDUKFromUserID(strconv.FormatInt(userID, 10)),
		"ext":          map[string]any{},
	}

	mainData := []any{}
	if vip {
		mainData = append(mainData, map[string]any{
			"icon": map[string]any{
				"height":   75,
				"priority": 2,
				"schema":   "https://tieba.baidu.com/mo/q/hybrid-business-vip/tbvip?customfullscreen=1&nonavigationbar=1",
				"type":     "1",
				"url":      "https://tieba-ares.cdn.bcebos.com/mis/2023-7/1689061482682/13afea50121d.png",
				"width":    75,
			},
			"type": 2,
		})
	}

	namedata := map[string]any{
		"text": map[string]any{
			"short_enable":   1,
			"short_length":   5,
			"short_priority": 1,
			"priority":       1,
			"str":            name,
			"suffix":         "...",
			"type":           "1",
		},
		"type": 1,
	}
	if vip {
		namedata["text"].(map[string]any)["text_color"] = map[string]any{
			"day":   "CAM_X0301",
			"night": "CAM_X0301",
			"type":  2,
		}
	}
	mainData = append(mainData, namedata)

	mainData = append(mainData,
		map[string]any{
			"icon": map[string]any{
				"height":   15,
				"priority": 5,
				"schema":   "https://tieba.baidu.com/mo/q/hybrid-main-user/taskCenter?customfullscreen=1&nonavigationbar=1",
				"type":     "3",
				"url":      "local://icon/icon_mask_level_usergrouth_" + strconv.FormatInt(glevel, 10) + "?type=webp",
				"width":    24,
			},
			"type": 2,
		},
		map[string]any{
			"icon": map[string]any{
				"height":   12,
				"priority": 2,
				"schema": "https://tieba.baidu.com/mo/q/wise-bawu-core/forum-level?customfullscreen=1&forum_id=" +
					strconv.FormatInt(fid, 10) + "&nonavigationbar=1&obj_locate=5&portrait=" + portraitTS,
				"type":  "4",
				"url":   fmt.Sprintf("local://icon/icon_level_%02d?type=webp", level),
				"width": 16,
			},
			"type": 2,
		},
	)

	ext := map[string]any{
		"main_data":   mainData,
		"is_sys_msg":  0,
		"version":     "",
		"portrait":    portraitTS,
		"robot_role":  0,
		"role":        0,
		"send_status": 0,
		"from":        "android",
		"session_id":  roomID,
		"type":        1,
		"user_name":   name,
	}
	if robot == -1 {
		ext["content"] = map[string]any{}
	} else {
		ext["content"] = map[string]any{
			"robot_params": map[string]any{"scene": "tieba_group_chat", "type": robot},
		}
	}
	ext["level"] = level
	ext["forum_id"] = fid

	textDict["ext"] = ext
	if atdata != nil {
		textDict["at_data"] = atdata
	}

	textDict["ext"] = marshalJSON(ext)
	textJSON := marshalJSON(textDict)
	// `content` 字段是一个 JSON 字符串，其 "text" 成员本身又是一个转义后的
	// JSON 字符串，对应 Python 模块的三重 json.dumps。
	content := marshalJSON(map[string]any{"text": textJSON})

	bdUk := BDUKFromUserID(strconv.FormatInt(userID, 10))
	return map[string]any{
		"method":       methodID,
		"mcast_id":     roomID,
		"role":         3,
		"token":        blcpCore.Account.BDUSS(),
		"appid":        appID,
		"uk":           uk,
		"origin_id":    originID,
		"type":         81,
		"app_safe_ext": marshalJSON(appSafeExt),
		"content":      content,
		"msg_key":      getMsgKey(bdUk),
		"account_type": 1,
		"sdk_version":  sdkVersion,
		"event_list": []map[string]any{
			{"event": "CClickSendBegin", "timestamp_ms": 0},
			{"event": "CSendBegin", "timestamp_ms": 0},
			{"event": "CIMSendBegin", "timestamp_ms": now.UnixMilli()},
		},
	}
}

// marshalJSON 把 v 序列化为紧凑 JSON，对应 json.dumps(ensure_ascii=False) 且不加分隔符。
func marshalJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

// Request 发送聊天室消息。
//
// atdata 为艾特@数据，robot 为机器人指令代码（-1 表示无，如签到10005、领取福利10004等）。
func Request(
	blcpCore *core.BLCPCore,
	roomID, uk, userID, originID int64,
	name, portrait, text string,
	fid, level int64,
	vip bool,
	glevel int64,
	atdata []map[string]any,
	robot int64,
) error {
	requestData := ConstructRequestData(
		blcpCore, roomID, uk, userID, originID, name, portrait, text, fid, level, vip, glevel, atdata, robot,
	)

	reply, err := blcpCore.SendLcm(serviceID, methodID, requestData)
	if err != nil {
		return err
	}
	if code := helper.JSONInt(reply, "err_code"); code != 0 {
		return &exception.TiebaServerError{Code: int(code), Msg: helper.JSONStr(reply, "err_msg")}
	}
	return nil
}
