package block

import "testing"

func TestIsLoopBan(t *testing.T) {
	tests := []struct {
		day  int64
		want int
	}{
		{1, 0},
		{3, 0},
		{10, 0},
		{0, 1},
		{30, 1},
		{-1, 1},
	}
	for _, tc := range tests {
		if got := IsLoopBan(tc.day); got != tc.want {
			t.Errorf("IsLoopBan(%d) = %d, want %d", tc.day, got, tc.want)
		}
	}
}

func TestRequestURL(t *testing.T) {
	u := RequestURL()
	if u.Host != "tiebac.baidu.com" || u.Path != "/c/c/bawu/commitprison" {
		t.Errorf("url = %s", u)
	}
}
