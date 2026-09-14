package getrankforums

import (
	"testing"

	"github.com/rongyuio/aiotieba/helper/htmlutil"
)

func TestRankForumFromXML(t *testing.T) {
	soup, err := htmlutil.Parse([]byte(`<table>
		<tr class="j_rank_row"><td>1</td><td>吧名A</td><td>100</td><td>200</td><td class="clearfix"><div class="bawu"></div></td></tr>
		<tr class="j_rank_row"><td>2</td><td>吧名B</td><td>50</td><td>80</td><td class="clearfix"><div class="no_bawu"></div></td></tr>
	</table>`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	forums := RankForumsFromXML(soup)
	if len(forums.Objs) != 2 {
		t.Fatalf("len(Objs) = %d, want 2", len(forums.Objs))
	}
	a := forums.Objs[0]
	if a.FName != "吧名A" || a.SignNum != 100 || a.MemberNum != 200 || !a.HasBawu {
		t.Errorf("Objs[0] = %+v", *a)
	}
	b := forums.Objs[1]
	if b.HasBawu {
		t.Errorf("Objs[1].HasBawu = true, want false")
	}
}

func TestPageRankForumFromXML(t *testing.T) {
	soup, _ := htmlutil.Parse([]byte(`<div class="pagination"><span>2</span><a href="/x?pn=1">1</a><a href="/x?pn=5">5</a></div>`))
	p := PageRankForumFromXML(soup)
	if p.CurrentPage != 2 || p.TotalPage != 5 || !p.HasMore || !p.HasPrev {
		t.Errorf("Page = %+v", p)
	}
}
