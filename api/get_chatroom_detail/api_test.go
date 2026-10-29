package getchatroomdetail

import (
	"errors"
	"testing"

	"github.com/rongyuio/aiotieba-go/exception"
)

// bdukOf123456 是 user_id 123456 的 baidu_uk，与 helper/crypto 测试中的黄金向量相同。
const bdukOf123456 = "xTBAOCyxXjl9vA8KC1jI7g"

const detailBody = `{"error_code":0,"error_msg":"","data":{` +
	`"basic_info":{"id":1001,"forum_id":42,"forum_name":"天堂鸡汤","name":"晚安群",` +
	`"avatar":"avatar.png","discuss_num":88},` +
	`"extra_info":{"announce":"群公告","talk_threshold_level":3,"is_talk_banned":"1"},` +
	`"bots":{"bot_list":[{"user":{"name":"bot","name_show":"机器人","portrait":"tb.1.abc",` +
	`"uk":"` + bdukOf123456 + `","robot_role":2},` +
	`"skill":[{"type":1,"name":"签到","alias":"sign","desc":"每日签到","is_show":1,` +
	`"items":[{"key":"a"}]}]}]}}}`

func TestParseBody(t *testing.T) {
	detail, err := ParseBody([]byte(detailBody))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if detail.RoomID != 1001 || detail.ForumID != 42 {
		t.Errorf("room/forum id = %d/%d", detail.RoomID, detail.ForumID)
	}
	if detail.ForumName != "天堂鸡汤" || detail.Name != "晚安群" || detail.Avatar != "avatar.png" {
		t.Errorf("basic info = %+v", detail)
	}
	if detail.DiscussNum != 88 {
		t.Errorf("DiscussNum = %d, want 88", detail.DiscussNum)
	}
	if detail.Announce != "群公告" || detail.TalkThresholdLevel != 3 {
		t.Errorf("extra info = %+v", detail)
	}
	if !detail.IsTalkBanned {
		t.Error("IsTalkBanned = false, want true (is_talk_banned = \"1\")")
	}
	if len(detail.Robots) != 1 {
		t.Fatalf("len(Robots) = %d, want 1", len(detail.Robots))
	}

	robot := detail.Robots[0]
	if robot.Name != "bot" || robot.NameShow != "机器人" || robot.Portrait != "tb.1.abc" {
		t.Errorf("robot identity = %+v", robot)
	}
	if robot.UK != bdukOf123456 {
		t.Errorf("robot UK = %q", robot.UK)
	}
	if robot.UserID != 123456 {
		t.Errorf("robot UserID = %d, want 123456", robot.UserID)
	}
	if robot.RobotRole != 2 {
		t.Errorf("robot RobotRole = %d, want 2", robot.RobotRole)
	}
	if len(robot.Skills) != 1 {
		t.Fatalf("len(robot.Skills) = %d, want 1", len(robot.Skills))
	}

	skill := robot.Skills[0]
	if skill.Type != 1 || skill.Name != "签到" || skill.Alias != "sign" || skill.Desc != "每日签到" {
		t.Errorf("skill = %+v", skill)
	}
	if !skill.IsShow {
		t.Error("skill IsShow = false, want true (is_show = 1)")
	}
	if len(skill.Items) != 1 || skill.Items[0]["key"] != "a" {
		t.Errorf("skill items = %+v", skill.Items)
	}
}

func TestParseBodyEmptyLists(t *testing.T) {
	body := `{"error_code":0,"data":{"basic_info":{"id":1,"forum_id":2,"forum_name":"f",` +
		`"name":"n","avatar":"a","discuss_num":0},` +
		`"extra_info":{"announce":"","talk_threshold_level":0,"is_talk_banned":0},` +
		`"bots":{"bot_list":[]}}}`

	detail, err := ParseBody([]byte(body))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if detail.IsTalkBanned {
		t.Error("IsTalkBanned = true, want false")
	}
	if len(detail.Robots) != 0 {
		t.Errorf("len(Robots) = %d, want 0", len(detail.Robots))
	}
}

func TestParseBodyInvalidBDUK(t *testing.T) {
	body := `{"error_code":0,"data":{"basic_info":{},"extra_info":{},` +
		`"bots":{"bot_list":[{"user":{"uk":"bogus"},"skill":[]}]}}}`

	if _, err := ParseBody([]byte(body)); err == nil {
		t.Error("ParseBody accepted an invalid bduk")
	}
}

func TestParseBodyServerError(t *testing.T) {
	_, err := ParseBody([]byte(`{"error_code":340006,"error_msg":"boom"}`))
	var serverErr *exception.TiebaServerError
	if !errors.As(err, &serverErr) {
		t.Fatalf("error = %v (%T), want *exception.TiebaServerError", err, err)
	}
	if serverErr.Code != 340006 || serverErr.Msg != "boom" {
		t.Errorf("server error = %+v", serverErr)
	}
}

func TestRequestURL(t *testing.T) {
	u := RequestURL()
	if u.Host != "tiebac.baidu.com" || u.Path != "/c/f/chat/getChatRoomDetail" {
		t.Errorf("url = %s", u)
	}
}
