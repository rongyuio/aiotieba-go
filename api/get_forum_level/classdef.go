// Package getforumlevel implements the get_forum_level API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_forum_level.
package getforumlevel

import (
	pb "github.com/rongyuio/aiotieba/api/get_forum_level/protobuf"
)

// LevelInfo is the level of the logged in account in one forum. It mirrors
// aiotieba.api.get_forum_level._classdef.LevelInfo.
type LevelInfo struct {
	LevelName string
	UserLevel int64
	IsLike    int64
}

// LevelInfoFromProto mirrors LevelInfo.from_proto.
func LevelInfoFromProto(p *pb.GetLevelInfoResIdl_DataRes) LevelInfo {
	return LevelInfo{
		UserLevel: int64(p.GetUserLevel()),
		LevelName: p.GetLevelName(),
		IsLike:    int64(p.GetIsLike()),
	}
}
