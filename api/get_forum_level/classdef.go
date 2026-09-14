// Package getforumlevel 实现 aiotieba 的 get_forum_level API。
//
// 对应 Python 包 aiotieba.api.get_forum_level。
package getforumlevel

import (
	pb "github.com/rongyuio/aiotieba-go/api/get_forum_level/protobuf"
)

// LevelInfo 用户于某贴吧的等级信息。
type LevelInfo struct {
	LevelName string // 等级名称
	UserLevel int64  // 等级数值
	IsLike    int64  // 是否已关注
}

// LevelInfoFromProto 对应 LevelInfo.from_proto。
func LevelInfoFromProto(p *pb.GetLevelInfoResIdl_DataRes) LevelInfo {
	return LevelInfo{
		UserLevel: int64(p.GetUserLevel()),
		LevelName: p.GetLevelName(),
		IsLike:    int64(p.GetIsLike()),
	}
}
