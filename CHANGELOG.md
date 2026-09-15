# 变更清单

本文件记录 aiotieba-go 的所有使用者可见变更，格式遵循
[Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，版本号遵循
[SemVer](https://semver.org/lang/zh-CN/)。

统一使用下列分类，便于检索：

| 分类 | 含义 |
| ------ | ------ |
| **破坏性变更** | 需要调用方改代码才能升级的改动 |
| **新增** | 新能力 |
| **变更** | 行为或用法发生变化 |
| **修复** | 缺陷修复 |
| **移除** | 删除了的能力（仅在有内容时出现） |
| **内部** | 不影响使用者的改动，如 CI、注释、文档（仅在有内容时出现） |

> **版本策略**：项目处于 1.x 阶段，为尽快对齐 Python 版实现，次要版本（如 1.2 → 1.3）也允许包含
> 破坏性变更。所有破坏性变更都会显式列在 `破坏性变更` 分类下，随 Release 说明一起发布；
> module 导入路径不会随版本变化。

任何面向 `master` 的变动都必须先在 `[Unreleased]` 下补充条目，发版时再整段下移为正式版本。

## [Unreleased]

### 破坏性变更

- 日志改用 [zerolog](https://github.com/rs/zerolog)：`logging.GetLogger()` 的返回值由 `*slog.Logger`
  变为 `*zerolog.Logger`，`logging.SetLevel` 的参数由 `slog.Level` 变为 `zerolog.Level`。
  原先通过 `log/slog` 接管日志的调用方需要改为传入 zerolog 记录器

### 新增

- `logging` 包新增 `PyRepr` / `PyArgs` / `PyErr` / `PyKw`，把参数与异常渲染成 Python 风格文本

### 变更

- 日志消息体与 Python 版 aiotieba 对齐，形如 `[sign_forum] Succeeded. args=('盗墓笔记',) kwargs={}`；
  失败日志的正文改为异常本身，例如 `*exception.TiebaServerError{Code: 340011}` 渲染为 `(340011, '')`
- 客户端日志改为在方法内部记录：同一方法无论失败在哪一步，参数都恒为调用方实参；
  50 个写操作（返回 `exception.BoolResponse` 的方法）新增成功日志
- `logging.EnableFileLog` 的写入端改为无颜色输出并交给 lumberjack 轮转（单文件 10MB、保留 5 份）

### 修复

- `SetThreadPrivacy` 此前只记录失败日志，现补齐成功日志，与 `SetThreadPrivate` / `SetThreadPublic` 行为一致

### 内部

- 依赖新增 `github.com/rs/zerolog`、`gopkg.in/natefinch/lumberjack.v2`，移除对 `log/slog` 的使用
- 新增 `logging_args_test.go`，解析 `client.go` 语法树校验日志参数一致性
- AGENTS.md / README.md / CONTRIBUTING.md 与实际实现对齐

## [1.2.0] - 2026-09-15

### 破坏性变更

- `GetThreads` 的第一个参数由 `fname string` 改为 `ref ForumRef`，以对齐 Python 版的 `fname_or_fid`；
  调用方需改用 `ByFName("天堂鸡汤")` 或 `ByFID(123)` 传参

### 内部

- 依赖升级：`actions/checkout` 4 → 7、`actions/setup-go` 5 → 7

## [1.1.1] - 2026-09-15

### 新增

- 新增 Release workflow，推送 `v*` tag 后自动校验版本号并创建 GitHub Release

### 变更

- CI 覆盖范围扩展到 `master` 分支

### 修复

- 测试中 User-Agent 的期望值改为从 `consts.Version` 推导，避免版本升级后失效

### 内部

- 补全各包的中文文档注释（顶层包与 client 门面、api 请求层、api 结果类型、core 与 helper、genproto 工具）
- 执行 `go mod tidy` 整理依赖

## [1.1.0] - 2026-09-15

### 变更

- HTTP 层由标准库 `net/http` 迁移到 [resty](https://github.com/go-resty/resty)（#4）

### 内部

- 重写 README：补充认证说明、快速开始、API 列表与徽章

## [1.0.0] - 2026-09-14

### 新增

- 首次发布：Python 版 [aiotieba](https://github.com/lumina37/aiotieba) 的全量 Go 移植，
  覆盖 HTTP、WebSocket、BLCP 私有协议、客户端签名与 AES 等加密、protobuf 编解码
