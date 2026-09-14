package getbawumemberlist

import (
	"testing"

	"github.com/rongyuio/aiotieba/helper/htmlutil"
)

// The forum-backend member table puts a newline after the left_cell td only; the
// remaining cells are compact, so the raw next_sibling chain of the Python
// module resolves to the consecutive <td> elements.
func TestBawuListMemberUserFromXML(t *testing.T) {
	soup, err := htmlutil.Parse([]byte(`<table><tbody><tr><td class="left_cell"><a> user1</a></td>
<td>100</td><td>5</td><td>10</td><td>2</td><td>2020-01-02 10:30</td><td id="123" portrait="tb.1.aaa"></td></tr></tbody></table>`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	users := BawuListMemberUsersFromXML(soup)
	if len(users.Objs) != 1 {
		t.Fatalf("len(Objs) = %d, want 1", len(users.Objs))
	}
	u := users.Objs[0]
	if u.UserName != "user1" {
		t.Errorf("UserName = %q, want user1", u.UserName)
	}
	if u.Exp != 100 || u.Level != 5 || u.ThreadNum != 10 || u.GoodNum != 2 {
		t.Errorf("counters = exp %d level %d thread %d good %d", u.Exp, u.Level, u.ThreadNum, u.GoodNum)
	}
	if u.UserID != 123 || u.Portrait != "tb.1.aaa" {
		t.Errorf("UserID/Portrait = %d/%q", u.UserID, u.Portrait)
	}
	if u.JoinTime.Year() != 2020 || u.JoinTime.Month() != 1 || u.JoinTime.Day() != 2 {
		t.Errorf("JoinTime = %v", u.JoinTime)
	}
}
