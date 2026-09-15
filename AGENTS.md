# aiotieba-go 开发指南

本文档旨在帮助开发者快速了解项目结构，以便参与 aiotieba-go 的开发

## 项目概述

aiotieba-go 是一个使用 Go 编写的百度贴吧 API 库，module 路径为 `github.com/rongyuio/aiotieba-go`，包名为 `aiotieba`，是原 Python 版（`lumina37/aiotieba`）的全量移植。公开方法名与 Python 版一一对齐（PascalCase），内部采用 Go 惯用法（`context.Context`、显式 `error` 返回、结构体优先）。

## 目录结构

```text
aiotieba-go/
├── client.go                   # Client 门面（所有 api 的调用入口）
├── config/                     # ProxyConfig / TimeoutConfig 配置
├── consts/                     # 版本号等常量
├── enums/                      # 所有枚举类型（含 From 回退构造）
├── exception/                  # 异常类型（TiebaServerError / BoolResponse 等）
├── logging/                    # zerolog 日志（logging.go 控制台/文件 + pyrepr.go Python 风格渲染）
├── protobuf/                   # 通用 protobuf 生成代码与 .proto
├── api/                        # 多个 API 模块，每个自成一个子包
│   ├── classdef/               # 通用数据类型（UserInfo / Fragment / VoteInfo 等）
│   └── <api_name>/             # 各 API 子包
│       ├── api.go              # PackProto / ParseBody / RequestURL / RequestHTTP / RequestWS
│       ├── classdef.go         # XxxFromProto / XxxFromJSON / XxxFromXML 解析构造
│       └── protobuf/           # 该 API 的 .proto 与生成的 .pb.go
├── core/                       # 网络会话层
│   ├── account.go              # Account（用户身份相关不变量、PBKDF2 派生 AES 密钥）
│   ├── net.go                  # NetCore（连接池、代理与超时）
│   ├── http.go                 # HttpCore（三类 http 会话 + 签名 + multipart 打包）
│   ├── websocket.go            # WsCore（自定义握手 / 9 字节帧 / AES-ECB + gzip）
│   ├── blcp.go                 # BLCPCore（TLS 分帧 / 三次握手 / 心跳）
│   ├── blcp_chat.go            # BLCP 群聊（JoinChatRoom / SendLcm）
│   └── msgid.go                # 消息 id 管理器（WS 与 BLCP 共用）
├── helper/                     # 内部辅助工具
│   ├── utils.go                # JSON 取值辅助（ParseJSONMap / JSONInt ...）
│   ├── cache.go                # 吧名↔fid 双向缓存
│   ├── htmlutil/               # BeautifulSoup 最小语义的 HTML 解析
│   └── crypto/                 # 签名与 AES 等密码学（逐字节对齐 Python）
├── tools/genproto/             # protoc-gen-go 生成脚本（go run ./tools/genproto）
├── go.mod / go.sum             # Go module 定义与依赖
└── README.md                   # 项目介绍
```

## 核心概念

### 术语定义

| 术语 | 说明 |
| ------ | ------ |
| **BDUSS** | 192字符的用户身份认证 token |
| **STOKEN** | 64字符的额外 token，部分 API 需要 |
| **fid** | 吧数字 ID |
| **fname** | 吧名称 |
| **tid** | 主题帖数字 ID |
| **pid** | 回复帖数字 ID |
| **user_name** | 用户名 |
| **portrait** | 用户头像标识（如 `tb.0.xxx`） |
| **user_id** | 旧版用户数字 ID |
| **tieba_uid** | 新版用户主页数字 ID |

### 核心状态容器

1. **`Client`**：主客户端，封装所有贴吧 API 的入口（`client.go`）
2. **`Account`**：BDUSS 等用户身份相关 token 的容器
3. **`NetCore`**：管理 TCP 连接池，保存代理、超时等配置
4. **`HttpCore`**：HTTP 会话状态容器（app / app_proto / web 三类）
5. **`WsCore`**：WebSocket 会话状态容器
6. **`BLCPCore`**：百度直播聊天协议的会话状态容器

### 数据流概览

```text
用户调用 client.GetThreads(ctx, ByFName("天堂鸡汤"), args)
    │
    ├── tryInitWebsocket(ctx)（优先 WebSocket，失败降级 HTTP）
    │
    ├── WebSocket 路径:
    │   ├── PackProto() → 构造请求
    │   ├── WsCore.Send(data, cmd) → 加密 + 压缩 + 发送
    │   └── ParseBody() → 解析响应
    │
    └── HTTP 路径:
        ├── PackProto() → 构造请求
        ├── HttpCore.AppProto(data) → multipart 打包
        └── ParseBody() → 解析响应
```

## API 模块规范

每个 `api/<name>` 子包遵循统一的模板：

- **proto 类**：`PackProto`（`proto.Marshal`）→ `ParseBody`（`proto.Unmarshal` + `XxxFromProto`）→ `RequestHTTP` / `RequestWS`
- **JSON 类**：`ParseBody`（`helper.ParseJSONMap` + `XxxFromJSON`）→ `Request`（app 表单 / web GET / web 表单）
- **HTML 抓取类**：`htmlutil.Parse` + `XxxFromXML`

结果类型统一携带 `Err error` 字段并嵌入 `classdef.Containers[T]`（列表类型，内容在 `Objs` 字段），门面方法以 `(T, error)` 返回。

## 日志

日志基于 zerolog，对应 Python 模块 `aiotieba.logging`：

- 控制台输出到 `stderr`，使用 `zerolog.ConsoleWriter` 的默认配色，时间格式 `2006-01-02 15:04:05`
- 消息体与 Python 版逐字对齐：`[api名] <异常或 Succeeded>. args=(...) kwargs={...}`，由 `logging.PyRepr` / `PyArgs` / `PyErr` 渲染
- 只有写操作（返回 `exception.BoolResponse` 的方法）记录成功日志，读接口只记录失败
- `EnableFileLog` 额外写入 `log/<程序名>.log`：无颜色、只记 INFO 及以上，轮转交由 lumberjack（单文件 10MB、保留 5 份）

Go 的方法没有装饰器可用，因此日志参数靠约定维持，并由 `logging_args_test.go` 解析 `client.go` 语法树自动校验：

| 约定 | 说明 |
| ------ | ------ |
| 参数必须是调用方实参 | 同一方法内所有 `logCallError` 站点、以及成功日志，必须使用同一组实参（方法签名里除 `ctx` 外的参数），不能使用中途解析出的 `fid`、`fname` 等中间值。Python 的 `handle_exception` 每次调用只产生一条日志且参数恒为调用方实参，此约定用于对齐该行为 |
| 可选参数用 `logging.PyKw` 标记 | 未标记的值进 `args`，标记的值进 `kwargs` |
| 异常类型实现 `PyArgs()` | 让 `TiebaServerError{340011, ""}` 渲染为 Python 的 `(340011, '')` |

## 变更清单与发版

`CHANGELOG.md` 是唯一的版本变更记录，遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)：

- 任何面向 `master` 的变动都必须先在 `[Unreleased]` 段落补充条目。`.github/workflows/changelog.yml`
  会校验「改动包含 CHANGELOG.md」与「`[Unreleased]` 非空」，缺失时 PR 无法合入
- 条目按 `破坏性变更` / `新增` / `变更` / `修复` / `移除` / `内部` 分类，写清对使用者的影响，不要照抄 commit message
- 破坏性变更必须显式记录，并按 [SemVer](https://semver.org/lang/zh-CN/) 决定版本号
- 发版时把 `[Unreleased]` 改写为 `[x.y.z] - YYYY-MM-DD`、同步 `consts.Version`，再打 tag；
  `release.yml` 会用该段落作为 Release 说明，找不到段落或段落为空都会失败

完整流程见 `.github/CONTRIBUTING.md`。

## 开发规范

参阅 `.github/CONTRIBUTING.md`
