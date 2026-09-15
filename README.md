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

	threads, err := client.GetThreads(context.Background(), aiotieba.ByFName("天堂鸡汤"), aiotieba.GetThreadsArgs{Pn: 1, Rn: 30})
	if err != nil {
		panic(err)
	}
	for _, thread := range threads.Objs {
		fmt.Printf("tid=%d\n%s\n", thread.TID, thread.Text())
	}
}
```

## API 列表

方法名与 Python 版一一对齐（PascalCase），内部统一采用 `context.Context` 与显式 `error` 返回。

`Client` 共覆盖 **117 个贴吧接口**，按类别归类如下：

| 类别 | 接口数 | 覆盖范围 |
| --- | --- | --- |
| 会话与生命周期 | 13 | 账号切换、会话初始化、WebSocket 与 BLCP 容器 |
| 帖子与回复 | 29 | 帖子与回复的读取、增删、屏蔽、恢复、加精置顶、点赞点踩 |
| 用户 | 20 | 用户信息、关注与粉丝、个人主页、头像与图片、主页设置 |
| 吧（论坛） | 19 | 贴吧信息、关注、排名、广场、分区、吧内用户 |
| 吧务管理 | 23 | 吧务团队与权限、管理日志、黑名单、封禁、申诉、统计 |
| 消息与聊天 | 8 | 私信、@ 信息、吧群与聊天室 |
| 签到 | 3 | 单个签到、一键签到、成长等级签到 |
| 搜索 | 2 | 吧内搜索、全站搜索 |

**每个接口的用途说明见 [docs/api.md](docs/api.md)**，其中同时列出了包级辅助接口
（`New` 与 `With*` 选项、`By*` 引用构造、`Default*Args` 默认参数）以及 `logging` 包的公开接口。
完整类型与签名见 [Go Reference](https://pkg.go.dev/github.com/rongyuio/aiotieba-go)。

## 日志

日志基于 [zerolog](https://github.com/rs/zerolog)，控制台输出到 `stderr` 并带颜色，
消息体与 Python 版 aiotieba 逐字对齐：

```
2026-09-15 00:00:03 WRN [sign_forums] (340011, ''). args=() kwargs={}
2026-09-15 00:00:04 INF [sign_forum] Succeeded. args=('盗墓笔记',) kwargs={}
```

```go
// logging 即 github.com/rongyuio/aiotieba-go/logging
logging.SetLevel(zerolog.InfoLevel) // 默认 Debug
logging.EnableFileLog("log")        // 追加写入 log/<程序名>.log，仅首次调用生效
logging.SetLogger(&myLogger)        // 替换为自定义记录器
```

+ 只有写操作（返回 `exception.BoolResponse` 的方法）记录成功日志，读接口只记录失败
+ 失败日志的正文是异常本身：`*exception.TiebaServerError{Code: 340011}` 渲染为 `(340011, '')`
+ 文件日志无颜色、只记 INFO 及以上，按大小轮转（单文件 10MB，保留 5 份）

## 项目特色

+ 收录 **99 个贴吧接口**（`api/` 目录下每个子包对应一个接口）
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
├── docs/              # api.md：全部公开接口与用途对照表
├── protobuf/          # 通用 protobuf 生成代码
├── api/               # 99 个 API 子包（每个含 api.go / classdef.go / protobuf）
└── tools/genproto/    # protoc-gen-go 生成脚本
```
