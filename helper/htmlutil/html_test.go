package htmlutil

import "testing"

func TestFindAndClass(t *testing.T) {
	root, err := Parse([]byte(`<html><body><div class="tbui_pagination"><ul><li class="active">2</li><li>3</li></ul></div></body></html>`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	pag := FindClass(root, "div", "tbui_pagination")
	if pag == nil {
		t.Fatal("FindClass returned nil")
	}
	if li := FindClass(pag, "li", "active"); li == nil || Text(li) != "2" {
		t.Errorf("active li = %v", li)
	}
	if n := FindAllClass(pag, "li", "active"); len(n) != 1 {
		t.Errorf("FindAllClass active = %d, want 1", len(n))
	}
}

func TestStringAndText(t *testing.T) {
	root, _ := Parse([]byte(`<table><tr><td>100</td><td><a href="/x">name</a></td></tr></table>`))
	tds := FindAllTag(root, "td")
	if len(tds) != 2 {
		t.Fatalf("len(tds) = %d", len(tds))
	}
	if s := String(tds[0]); s != "100" {
		t.Errorf("String = %q, want 100", s)
	}
	if s := String(tds[1]); s != "" {
		t.Errorf("String of nested td = %q, want empty", s)
	}
	if txt := Text(tds[1]); txt != "name" {
		t.Errorf("Text = %q, want name", txt)
	}
}

func TestNextSiblingTag(t *testing.T) {
	root, _ := Parse([]byte("<table><tr><td>a</td>\n<td>b</td><td class=\"x\">c</td></tr></table>"))
	first := Find(root, "td")
	second := NextSiblingTag(first, "td")
	if second == nil || Text(second) != "b" {
		t.Fatalf("NextSiblingTag = %v", second)
	}
	third := NextSiblingTagClass(second, "td", "x")
	if third == nil || Text(third) != "c" {
		t.Fatalf("NextSiblingTagClass = %v", third)
	}
}

func TestAtoi(t *testing.T) {
	cases := map[string]int64{"100": 100, " 5 ": 5, "abc": 0, "": 0, "30天": 0}
	for in, want := range cases {
		if got := Atoi(in); got != want {
			t.Errorf("Atoi(%q) = %d, want %d", in, got, want)
		}
	}
}
