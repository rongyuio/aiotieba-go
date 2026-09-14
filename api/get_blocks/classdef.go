// Package getblocks 实现 aiotieba 的 get_blocks API。
//
// 对应 Python 包 aiotieba.api.get_blocks。
package getblocks

import (
	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/htmlutil"
)

// Block 待解封用户信息。
type Block struct {
	UserID      int64  // user_id
	UserName    string // 用户名
	NickNameOld string // 旧版昵称
	Day         int64  // 封禁天数
}

// BlockFromXML 对应 Block.from_xml。
func BlockFromXML(li *htmlutil.Node) Block {
	a := htmlutil.FirstChildTag(li, "a")
	return Block{
		UserID:      htmlutil.Atoi(htmlutil.Attr(a, "attr-uid")),
		UserName:    htmlutil.Attr(a, "attr-un"),
		NickNameOld: htmlutil.Attr(a, "attr-nn"),
		Day:         htmlutil.Atoi(htmlutil.Attr(a, "attr-blockday")),
	}
}

// PageBlock 页信息。
type PageBlock struct {
	PageSize    int64 // 页大小
	CurrentPage int64 // 当前页码
	TotalPage   int64 // 总页码
	TotalCount  int64 // 总计数
	HasMore     bool  // 是否有后继页
	HasPrev     bool  // 是否有前驱页
}

// PageBlockFromJSON 对应 Page_block.from_json。
func PageBlockFromJSON(m map[string]any) PageBlock {
	currentPage := helper.JSONInt(m, "pn")
	return PageBlock{
		PageSize:    helper.JSONInt(m, "size"),
		CurrentPage: currentPage,
		TotalPage:   helper.JSONInt(m, "total_page"),
		TotalCount:  helper.JSONInt(m, "total_count"),
		HasMore:     helper.JSONBool(m, "have_next"),
		HasPrev:     currentPage > 1,
	}
}

// Blocks 待解封用户列表。
type Blocks struct {
	classdef.Containers[*Block]

	Page PageBlock // 页信息
	Err  error     // 捕获的异常
}

// BlocksFromJSON 对应 Blocks.from_json：其中 "content" 字段是一段 HTML
// 片段，"page" 字段是 JSON。
func BlocksFromJSON(m map[string]any) (Blocks, error) {
	data := helper.JSONMap(m, "data")
	soup, err := htmlutil.Parse([]byte(helper.JSONStr(data, "content")))
	if err != nil {
		return Blocks{}, err
	}
	var blocks Blocks
	for _, li := range htmlutil.FindAllTag(soup, "li") {
		b := BlockFromXML(li)
		blocks.Objs = append(blocks.Objs, &b)
	}
	blocks.Page = PageBlockFromJSON(helper.JSONMap(data, "page"))
	return blocks, nil
}

// HasMore 是否还有下一页。
func (b Blocks) HasMore() bool { return b.Page.HasMore }
