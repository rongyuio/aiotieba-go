# aiotieba-go

[![Go Reference](https://pkg.go.dev/badge/github.com/rongyuio/aiotieba-go.svg)](https://pkg.go.dev/github.com/rongyuio/aiotieba-go)
[![Release](https://img.shields.io/github/v/release/rongyuio/aiotieba-go)](https://github.com/rongyuio/aiotieba-go/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/rongyuio/aiotieba-go)](go.mod)
[![License](https://img.shields.io/github/license/rongyuio/aiotieba-go)](https://github.com/rongyuio/aiotieba-go/blob/master/LICENSE)

一个用 **Go** 编写的百度贴吧 API 库，是原 Python 版 [aiotieba](https://github.com/lumina37/aiotieba) 的全量移植。

`Client` 暴露与 Python 版同名（PascalCase）的方法，覆盖 HTTP、WebSocket、BLCP 私有协议、贴吧客户端签名与 AES 等加密、以及 protobuf 编解码。

## 安装

```shell
go get github.com/rongyuio/aiotieba-go
```

## 认证

调用前需要准备两个登录凭证：

| 凭证 | 长度 | 说明 |
| --- | --- | --- |
| **BDUSS** | 192 字符 | 用户身份认证 token，必填 |
| **STOKEN** | 64 字符 | 额外的安全 token，部分 API 需要，可留空 |

## 快速开始

```go
package main

import (
	"context"
	"fmt"

	"github.com/rongyuio/aiotieba-go"
)

func main() {
	client, err := aiotieba.New("你的BDUSS", "你的STOKEN")
	if err != nil {
		panic(err)
	}
	defer client.Close()

	threads, err := client.GetThreads(context.Background(), "天堂鸡汤", aiotieba.GetThreadsArgs{Pn: 1, Rn: 30})
	if err != nil {
		panic(err)
	}
	for _, thread := range threads.Objs {
		fmt.Printf("tid=%d\n%s\n", thread.Tid, thread.Text())
	}
}
```

## API 列表

> 方法名与 Python 版一一对齐（PascalCase），内部统一采用 `context.Context` 与显式 `error` 返回。完整签名见 [Go Reference](https://pkg.go.dev/github.com/rongyuio/aiotieba-go)。

### 帖子与回复

`GetThreads` `GetPosts` `GetComments` `GetLastReplyers` `AddPost`
`DelPost` `DelPosts` `DelThread` `DelThreads` `HideThread` `UnhideThread`
`Recover` `GetRecovers` `Good` `Ungood` `Top` `Untop` `Move`
`SetThreadPrivate` `SetThreadPublic` `Recommend`
`Disagree` `Unagree` `Undisagree` `AddPoll`

### 用户

`GetUserInfo` `GetSelfInfo` `GetHomepage` `GetFans` `GetFollows`
`FollowUser` `UnfollowUser` `RemoveFan` `GetUserPosts` `GetUserThreads`
`GetSelfPosts` `GetSelfThreads` `TiebaUID2UserInfo`
`GetPortrait` `GetImage` `Hash2Image`

### 吧（论坛）

`GetForum` `GetForumDetail` `GetFID` `GetFName` `GetFollowForums`
`GetSelfFollowForums` `FollowForum` `UnfollowForum` `GetSquareForums`
`GetRankForums` `GetRankUsers` `DislikeForum` `UndislikeForum`
`GetDislikeForums`

### 吧务管理

`GetBawuInfo` `GetBawuMemberlist` `GetBawuPerm` `SetBawuPerm`
`AddBawu` `DelBawu` `GetBawuPostlogs` `GetBawuUserlogs`
`Block` `Unblock` `GetBlocks` `GetUnblockAppeals`
`SetBlacklist` `GetBlacklist` `GetBlacklistOld`
`AddBawuBlacklist` `DelBawuBlacklist` `GetBawuBlacklist`

### 消息与聊天

`SendMsg` `GetGroupMsg` `SetMsgReaded` `SendChatroomMsg` `JoinChatroom`
`GetAts` `GetReplys`

### 签到

`SignForum` `SignForums` `SignGrowth`

### 搜索

`SearchExact` `SearchGlobal`

### 其他

`Login` `GetStatistics` `GetTabMap` `GetCID` `GetRecomStatus`
`GetUserForumInfo` `GetMemberUsers` `SetProfile` `SetNicknameOld`

## 项目特色

+ 收录**数十个常用 API**（`api/` 目录下每个子包对应一个贴吧接口）
+ 支持 protobuf 序列化请求参数
+ 支持 WebSocket 接口与 BLCP 群聊协议
+ 与官方版本高度一致的密码学实现（签名、`cuid_galaxy2`、`c3_aid`、`rc4_42`、AES-ECB/CBC 等，均有逐字节测试向量）
+ 全量对照原 Python 版迁移，公开方法名对齐，内部采用 Go 惯用法（`context.Context`、显式 `error` 返回、结构体）

## 目录结构

```
aiotieba-go/
├── client.go          # Client 门面：聚合全部 API 方法
├── config/ consts/ enums/ exception/ logging/   # 配置 / 常量 / 枚举 / 异常 / 日志
├── core/              # Account / NetCore / HttpCore / WsCore / BLCPCore
├── helper/            # utils / cache / crypto / htmlutil
├── protobuf/          # 通用 protobuf 生成代码
├── api/               # 约 100 个 API 子包（每个含 api.go / classdef.go / protobuf）
└── tools/genproto/    # protoc-gen-go 生成脚本
```
