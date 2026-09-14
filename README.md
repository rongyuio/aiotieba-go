# aiotieba

一个用 **Go** 编写的百度贴吧 API 库，是原 Python 版 [aiotieba](https://github.com/rongyuio/aiotieba) 的全量移植。

`Client` 暴露与 Python 版同名（PascalCase）的方法，覆盖 HTTP、WebSocket、BLCP 私有协议、贴吧客户端签名与 AES 等加密、以及 protobuf 编解码。

## 安装

```shell
go get github.com/rongyuio/aiotieba
```

## 尝试一下

```go
package main

import (
	"context"
	"fmt"

	"github.com/rongyuio/aiotieba"
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

## 友情链接

+ [带UI的吧务管理器 (dog194/TiebaManager)](https://github.com/dog194/TiebaManager)
+ [第三方桌面客户端 (clb-128258/TiebaDesktop)](https://github.com/clb-128258/TiebaDesktop)
+ [VSCode贴吧摸鱼插件 (akacaijizhou/tieba-fish)](https://github.com/akacaijizhou/tieba-fish)
+ [eztb贴吧工具箱 (Dilettante258/eazy-tieba)](https://www.eztb.org)
+ [功能全面的贴吧管理QQ bot (TiebaMeow/TiebaManageBot)](https://github.com/TiebaMeow/TiebaManageBot)
+ [易于部署和使用的 Web 贴吧管理和自动化平台 (TiebaMeow/WebTiebaManager)](https://github.com/TiebaMeow/WebTiebaManager)
+ [灵活且高可靠的贴吧爬虫 (TiebaMeow/TiebaScraper)](https://github.com/TiebaMeow/TiebaScraper)
+ [第三方安卓客户端 (zzc10086/TiebaLite)](https://github.com/zzc10086/TiebaLite)
+ [C#版本的贴吧接口库 (BaWuZhuShou/AioTieba4DotNet)](https://github.com/BaWuZhuShou/AioTieba4DotNet)
+ [基于aiotieba的tieba bot (adk23333/BungleCat)](https://github.com/adk23333/BungleCat)
+ [基于aiotieba的贴吧管理器 (adk23333/tieba-admin)](https://github.com/adk23333/tieba-admin)
+ [贴吧protobuf定义文件合集 (clb-128258/tbclient.protobuf)](https://github.com/clb-128258/tbclient.protobuf)

## 特别鸣谢

<p align="center">
<a href="https://jb.gg/OpenSourceSupport">
    <img src="https://resources.jetbrains.com/storage/products/company/brand/logos/jb_beam.svg">
</a>
</p>

为本开源项目提供的免费产品授权
