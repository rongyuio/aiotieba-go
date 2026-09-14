// Package getroomlistbyfid implements the get_roomlist_by_fid API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_roomlist_by_fid.
package getroomlistbyfid

import "github.com/rongyuio/aiotieba-go/helper"

// RoomList mirrors RoomList: the raw chatroom json of a forum.
type RoomList struct {
	RoomList []map[string]any
	Err      error
}

// RoomListFromJSON mirrors RoomList.from_json.
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
