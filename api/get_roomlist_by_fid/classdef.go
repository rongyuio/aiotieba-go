// Package getroomlistbyfid 实现 aiotieba 的 get_roomlist_by_fid API。
//
// 对应 Python 包 aiotieba.api.get_roomlist_by_fid。
package getroomlistbyfid

import "github.com/rongyuio/aiotieba-go/helper"

// RoomList 某吧的聊天室列表。
type RoomList struct {
	RoomList []map[string]any // 每个聊天室的json内容
	Err      error
}

// RoomListFromJSON 对应 RoomList.from_json。
func RoomListFromJSON(res map[string]any) RoomList {
	var rl RoomList
	data := helper.JSONMap(res, "data")
	for _, x := range helper.JSONSlice(data, "list") {
		if xm, ok := x.(map[string]any); ok {
			for _, item := range helper.JSONSlice(xm, "room_list") {
				if im, ok := item.(map[string]any); ok {
					rl.RoomList = append(rl.RoomList, im)
				}
			}
		}
	}
	return rl
}
