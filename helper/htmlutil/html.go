// Package htmlutil provides a small subset of the BeautifulSoup DOM helpers
// used by the forum-backend HTML scraping APIs.
//
// It mirrors the behaviour of the Python `bs4.BeautifulSoup` calls in the
// aiotieba codebase closely enough for those APIs: recursive find by tag/class,
// first-child access (`.td`, `.a`, ...), `next_sibling`/`previous_sibling` via
// the raw node pointers, `.text`/`.string` and attribute lookup.
package htmlutil

import (
	"bytes"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// Node aliases golang.org/x/net/html.Node. The NextSibling/PrevSibling/Parent
// fields already carry the raw sibling/parent pointers (including text nodes),
// matching BeautifulSoup's `next_sibling`/`previous_sibling`.
type Node = html.Node

// Parse parses data into an HTML tree, mirroring bs4.BeautifulSoup(body, "lxml").
func Parse(data []byte) (*Node, error) {
	return html.Parse(bytes.NewReader(data))
}

// IsTag reports whether n is an element with the given tag name.
func IsTag(n *Node, tag string) bool {
	return n != nil && n.Type == html.ElementNode && n.Data == tag
}

// Attr returns the value of the attribute key, or "" when absent. It mirrors
// `tag["key"]`.
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

// HasClass reports whether the class attribute contains the given class token.
// It mirrors the class_ filter of BeautifulSoup.
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

// Find returns the first descendant with the given tag, mirroring find(tag).
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

// FindClass returns the first descendant with the given tag and class, mirroring
// find(tag, class_=class).
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

// FindAllTag returns all descendants with the given tag, mirroring
// find_all(tag) / soup(tag).
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

// FindAllClass returns all descendants with the given tag and class, mirroring
// find_all(tag, class_=class).
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

// FirstChildTag returns the first child with the given tag, mirroring the
// `.td`/`.a`/`.input`/`.span`/`.div`/`.h1`/`.em`/`.time`/`.img` accessors.
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

// NextSiblingTag returns the next sibling that is an element with the given
// tag, mirroring find_next_sibling(tag).
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

// NextSiblingTagClass returns the next sibling that is an element with the given
// tag and class, mirroring find_next_sibling(tag, class_=class).
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

// Text returns the concatenated text content of the node, mirroring `.text`.
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

// Atoi parses s as an int64, returning 0 when it is not numeric. It mirrors the
// `int(tag.text)` conversions of the scraping APIs.
func Atoi(s string) int64 {
	v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0
	}
	return v
}

// String returns the direct text of a node that contains a single text child,
// mirroring `.string`. It returns "" otherwise (mirroring the None that the
// Python APIs treat as an empty/absent value).
func String(n *Node) string {
	if n == nil || n.FirstChild == nil {
		return ""
	}
	if n.FirstChild.Type == html.TextNode && n.FirstChild.NextSibling == nil {
		return n.FirstChild.Data
	}
	return ""
}
