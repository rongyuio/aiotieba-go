# aiotieba

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

## 尝试一下

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

## 项目特色

+ 收录**数十个常用 API**（`api/` 目录下每个子包对应一个贴吧接口）
+ 支持 protobuf 序列化请求参数
+ 支持 WebSocket 接口与 BLCP 群聊协议
+ 与官方版本高度一致的密码学实现（签名、`cuid_galaxy2`、`c3_aid`、`rc4_42`、AES-ECB/CBC 等，均有逐字节测试向量）
+ 全量对照原 Python 版迁移，公开方法名对齐，内部采用 Go 惯用法（`context.Context`、显式 `error` 返回、结构体）

## 目录结构

```
aiotieba/
├── client.go          # Client 门面：聚合全部 API 方法
├── config/ consts/ enums/ exception/ logging/   # 配置 / 常量 / 枚举 / 异常 / 日志
├── core/              # Account / NetCore / HttpCore / WsCore / BLCPCore
├── helper/            # utils / cache / crypto / htmlutil
├── protobuf/          # 通用 protobuf 生成代码
├── api/               # 约 100 个 API 子包（每个含 api.go / classdef.go / protobuf）
└── tools/genproto/    # protoc-gen-go 生成脚本
```
