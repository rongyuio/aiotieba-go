// Package htmlutil 提供论坛后端 HTML 抓取 API 所用 BeautifulSoup DOM 辅助函数的一个小子集。
//
// 它对应 aiotieba 代码库中的 Python `bs4.BeautifulSoup` 调用，对这些 API 而言行为足够接近：按 tag/class 递归查找、首个子节点访问（`.td`、`.a` 等）、经由原始节点指针实现 `next_sibling`/`previous_sibling`、`.text`/`.string` 以及属性查找。
package htmlutil

import (
	"bytes"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// Node 是 golang.org/x/net/html.Node 的别名。NextSibling/PrevSibling/Parent 字段已经携带原始兄弟/父节点指针（包括文本节点），与 BeautifulSoup 的 `next_sibling`/`previous_sibling` 一致。
type Node = html.Node

// Parse 把 data 解析为 HTML 树，对应 bs4.BeautifulSoup(body, "lxml")。
func Parse(data []byte) (*Node, error) {
	return html.Parse(bytes.NewReader(data))
}

// IsTag 报告 n 是否为具有给定标签名的元素。
func IsTag(n *Node, tag string) bool {
	return n != nil && n.Type == html.ElementNode && n.Data == tag
}

// Attr 返回属性 key 的值，不存在时返回 ""，对应 `tag["key"]`。
func Attr(n *Node, key string) string {
	if n == nil {
		return ""
	}
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

// HasClass 报告 class 属性是否包含给定的 class 标记，对应 BeautifulSoup 的 class_ 过滤器。
func HasClass(n *Node, class string) bool {
	if n == nil || class == "" {
		return false
	}
	for _, c := range strings.Fields(Attr(n, "class")) {
		if c == class {
			return true
		}
	}
	return false
}

// Find 返回第一个具有给定标签的后代，对应 find(tag)。
func Find(n *Node, tag string) *Node {
	if n == nil {
		return nil
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if r := findRec(c, tag); r != nil {
			return r
		}
	}
	return nil
}

func findRec(n *Node, tag string) *Node {
	if IsTag(n, tag) {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if r := findRec(c, tag); r != nil {
			return r
		}
	}
	return nil
}

// FindClass 返回第一个具有给定标签与 class 的后代，对应 find(tag, class_=class)。
func FindClass(n *Node, tag, class string) *Node {
	if n == nil {
		return nil
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if r := findClassRec(c, tag, class); r != nil {
			return r
		}
	}
	return nil
}

func findClassRec(n *Node, tag, class string) *Node {
	if IsTag(n, tag) && HasClass(n, class) {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if r := findClassRec(c, tag, class); r != nil {
			return r
		}
	}
	return nil
}

// FindAllTag 返回所有具有给定标签的后代，对应 find_all(tag) / soup(tag)。
func FindAllTag(n *Node, tag string) []*Node {
	if n == nil {
		return nil
	}
	var out []*Node
	collectTag(n, tag, &out)
	return out
}

func collectTag(n *Node, tag string, out *[]*Node) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if IsTag(c, tag) {
			*out = append(*out, c)
		}
		collectTag(c, tag, out)
	}
}

// FindAllClass 返回所有具有给定标签与 class 的后代，对应 find_all(tag, class_=class)。
func FindAllClass(n *Node, tag, class string) []*Node {
	if n == nil {
		return nil
	}
	var out []*Node
	collectClass(n, tag, class, &out)
	return out
}

func collectClass(n *Node, tag, class string, out *[]*Node) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if IsTag(c, tag) && HasClass(c, class) {
			*out = append(*out, c)
		}
		collectClass(c, tag, class, out)
	}
}

// FirstChildTag 返回第一个具有给定标签的子节点，对应 `.td`/`.a`/`.input`/`.span`/`.div`/`.h1`/`.em`/`.time`/`.img` 访问器。
func FirstChildTag(n *Node, tag string) *Node {
	if n == nil {
		return nil
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if IsTag(c, tag) {
			return c
		}
	}
	return nil
}

// NextSiblingTag 返回下一个标签为 tag 的元素兄弟节点，对应 find_next_sibling(tag)。
func NextSiblingTag(n *Node, tag string) *Node {
	if n == nil {
		return nil
	}
	for s := n.NextSibling; s != nil; s = s.NextSibling {
		if IsTag(s, tag) {
			return s
		}
	}
	return nil
}

// NextSiblingTagClass 返回下一个标签为 tag 且 class 匹配的元素兄弟节点，对应 find_next_sibling(tag, class_=class)。
func NextSiblingTagClass(n *Node, tag, class string) *Node {
	if n == nil {
		return nil
	}
	for s := n.NextSibling; s != nil; s = s.NextSibling {
		if IsTag(s, tag) && HasClass(s, class) {
			return s
		}
	}
	return nil
}

// Text 返回节点拼接后的文本内容，对应 `.text`。
func Text(n *Node) string {
	if n == nil {
		return ""
	}
	var b strings.Builder
	collectText(n, &b)
	return b.String()
}

func collectText(n *Node, b *strings.Builder) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			b.WriteString(c.Data)
		} else if c.Type == html.ElementNode {
			collectText(c, b)
		}
	}
}

// Atoi 把 s 解析为 int64，非数字时返回 0，对应抓取 API 中的 `int(tag.text)` 转换。
func Atoi(s string) int64 {
	v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0
	}
	return v
}

// String 返回仅含单个文本子节点的节点的直接文本，对应 `.string`；否则返回 ""（对应 Python API 视为空/缺失值的 None）。
func String(n *Node) string {
	if n == nil || n.FirstChild == nil {
		return ""
	}
	if n.FirstChild.Type == html.TextNode && n.FirstChild.NextSibling == nil {
		return n.FirstChild.Data
	}
	return ""
}
