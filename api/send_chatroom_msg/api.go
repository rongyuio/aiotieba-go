// Package sendchatroommsg implements the send_chatroom_msg API of aiotieba.
//
// It mirrors the Python package aiotieba.api.send_chatroom_msg. The API is only
// available over the BLCP transport.
package sendchatroommsg

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strconv"
	"time"

	"github.com/rongyuio/aiotieba/core"
	"github.com/rongyuio/aiotieba/exception"
	"github.com/rongyuio/aiotieba/helper"
	"github.com/rongyuio/aiotieba/helper/crypto"
)

// appConstants mirrors the AppConstants dataclass of the Python module.
const (
	appID      = int64(10773430)
	sdkVersion = int64(11250036)
)

// serviceID and methodID of the chatroom message request.
const (
	serviceID = int64(3)
	methodID  = int64(185)
)

// bdukKey and bdukIV are the fixed key and IV used by getBDUKfromUserId.
var (
	bdukKey = []byte("AFD311832EDEEAEF")
	bdukIV  = []byte("2011121211143000")
)

// BDUKFromUserID mirrors BLCPCore.getBDUKfromUserId: the numeric user id is
// AES-CBC encrypted with a fixed key and IV and then urlsafe-base64 encoded
// without padding.
func BDUKFromUserID(userID string) string {
	enc, err := crypto.CBCEncrypt(bdukKey, bdukIV, []byte(userID))
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(enc)
}

// getMsgKey mirrors BLCPCore.getmsgkey.
func getMsgKey(bduk string) string {
	return bduk + strconv.FormatInt(time.Now().UnixMilli()*1000, 10) + strconv.FormatInt(rand.Int64(), 10)
}

// ConstructRequestData builds the Lcm request body, mirroring
// construct_request_data.
//
// It is exported for testing. The return value is handed to BLCPCore.SendLcm,
// which injects client_logid and rpc.
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
	// The `content` field is a JSON string whose "text" member is itself an
	// escaped JSON string, mirroring the triple json.dumps of the Python module.
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

// marshalJSON serializes v to compact JSON, mirroring json.dumps with
// ensure_ascii=False and no separators.
func marshalJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

// Request mirrors request.
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
