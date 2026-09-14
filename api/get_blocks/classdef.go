// Package getblocks implements the get_blocks API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_blocks.
package getblocks

import (
	"github.com/rongyuio/aiotieba/api/classdef"
	"github.com/rongyuio/aiotieba/helper"
	"github.com/rongyuio/aiotieba/helper/htmlutil"
)

// Block mirrors Block.
type Block struct {
	UserID      int64
	UserName    string
	NickNameOld string
	Day         int64
}

// BlockFromXML mirrors Block.from_xml.
func BlockFromXML(li *htmlutil.Node) Block {
	a := htmlutil.FirstChildTag(li, "a")
	return Block{
		UserID:      htmlutil.Atoi(htmlutil.Attr(a, "attr-uid")),
		UserName:    htmlutil.Attr(a, "attr-un"),
		NickNameOld: htmlutil.Attr(a, "attr-nn"),
		Day:         htmlutil.Atoi(htmlutil.Attr(a, "attr-blockday")),
	}
}

// PageBlock mirrors Page_block.
type PageBlock struct {
	PageSize    int64
	CurrentPage int64
	TotalPage   int64
	TotalCount  int64
	HasMore     bool
	HasPrev     bool
}

// PageBlockFromJSON mirrors Page_block.from_json.
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

// Blocks mirrors Blocks.
type Blocks struct {
	classdef.Containers[*Block]

	Page PageBlock
	Err  error
}

// BlocksFromJSON mirrors Blocks.from_json: the "content" field is an HTML
// fragment, the "page" field is JSON.
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

// HasMore mirrors the has_more property.
func (b Blocks) HasMore() bool { return b.Page.HasMore }
