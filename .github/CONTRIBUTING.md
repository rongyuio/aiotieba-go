# 开发规范

## commit 规范

提交信息须遵循简化版的 [Conventional Commits](https://www.conventionalcommits.org/zh-hans/)，格式为 `<type>: <description>`。

| 类型 | 说明 |
| ------ | ------ |
| `feat` | 新功能 |
| `fix` | Bug 修复 |
| `refactor` | 重构（既非新功能也非修复） |
| `perf` | 性能优化 |
| `chore` | 日常维护（依赖更新、脚本改进等） |
| `docs` | 文档变更 |
| `test` | 测试相关 |
| `style` | 代码格式（不影响逻辑的空白、缩进等） |
| `ci` | CI/CD 配置变更 |

外部 PR 和 commit 须向 `develop` 分支而不是 `master` 分支提交

## 变更清单

`CHANGELOG.md` 是唯一的版本变更记录，遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)。

- 任何面向 `master` 的变动都必须先在 `[Unreleased]` 段落补充条目，再合入
- 条目要写清**对使用者意味着什么**，不要照抄 commit message
- 分类固定为 `破坏性变更` / `新增` / `变更` / `修复` / `移除` / `内部`，其中 `移除` 与 `内部` 只在有内容时出现
- 破坏性变更必须显式写在 `破坏性变更` 分类里，说明改动内容与迁移方式
- 版本策略：1.x 阶段允许在次要版本中引入破坏性变更（module 导入路径不变），前提是变更被显式记录并随 Release 说明发布
- 纯 CI、纯依赖升级等不影响使用者的改动，可给 PR 打 `skip-changelog` 标签，或直接在提交信息里写上
  `skip-changelog` 跳过检查

`.github/workflows/changelog.yml` 负责强制该约定：变动的文件里必须包含 `CHANGELOG.md`，且
`[Unreleased]` 段落至少有一条条目。校验不通过时，面向 `master` 的 PR 无法合入。

## 发版

1. 确认 `develop` 上的改动都已写入 `CHANGELOG.md` 的 `[Unreleased]`
2. 按 [SemVer](https://semver.org/lang/zh-CN/) 确定版本号，把 `[Unreleased]` 改写为
   `[x.y.z] - YYYY-MM-DD`，并在文件顶部补一个新的空 `[Unreleased]` 段落
3. 同步 `consts/consts.go` 中的 `Version`
4. 开一个标题为 `release: vX.Y.Z` 的 PR（`develop` → `master`）并合入，合并方式用 merge commit
5. 在 `master` 上打 tag 并推送。用附注 tag 并写明版本：`git tag -a vX.Y.Z -m "Release vX.Y.Z"`

发版后 `master` 的 HEAD 就是该 tag 指向的提交，下次发版再前进一次——这就是 `AGENTS.md` 里那条
约定，不要把零散改动单独合入 `master`。

`release.yml` 收到 tag 后会依次校验：tag 与 `consts.Version` 一致、`CHANGELOG.md` 中存在该版本的
非空段落、`go.mod` 整洁（`go mod tidy -diff` 无差异）、gofmt/build/vet/test 全绿，最后以该段落
作为说明创建 GitHub Release。任一校验失败都会直接中断，不会发出一个没有说明的版本。

**`v*` tag 一旦推送就不可撤销**：版本会被 `proxy.golang.org` 缓存，并在 `sum.golang.org` 留下永久
校验和。永远不要移动或重推已发布的 tag；要修已发布版本的内容，只能发新版本。详见 `AGENTS.md` 的
「已发布的 tag 不可移动」。

## 代码风格

- Go 代码风格遵循标准 `gofmt`（提交前运行 `gofmt -w .` 并确保 `gofmt -l .` 无输出）
- `go.mod` 保持整洁：改动依赖后运行 `go mod tidy`，CI 会用 `go mod tidy -diff` 校验
- CI 会用 `go.mod` 声明的最低 Go 版本与最新版各跑一遍，两者都必须通过
- 导入分组：标准库、第三方、本项目，按路径字典序排列

## 注释规范

全部手写 Go 源码的注释都是中文，首行以符号名开头，其后依次是「参数」段（可选）与补充说明；
溯源信息写成「对应 Python 的 `Client.get_forum`」这样的中文表述，不保留英文原文。

「参数」段写成 `// 参数:` + 一行空 `//` + 制表符缩进的条目，条目用 **Go 侧的实际参数名**，
不照抄 Python 的 `id_` / `fname_or_fid`；结构体参数写成一行指向结构体，不展开其字段：

```go
// GetThreads 获取首页帖子。
//
// 参数:
//
//	ref 目标贴吧名或fid 优先贴吧名
//	args 可选参数，详见 GetThreadsArgs
```

**不要手工对抗 gofmt**：Go 1.19 起的 `gofmt` 会解析并重排文档注释——以空格或制表符起始的连续
注释行会被识别为 preformatted 代码块，归一化为 `//` + 单个制表符，并自动补一行前置空 `//`。
按上面的形态写入后跑 `gofmt -w .`，以 gofmt 的输出为最终形态。

- 导出符号必须有文档注释且以符号名开头，`go doc` 与 IDE 悬浮提示据此渲染；内部符号注释从简
- 结构体字段的含义用行尾注释（如 `Src string // 小图链接 宽720px`），对应 Python 的 `Attributes:`

## 命名约定

| 类型 | 约定 | 示例 |
| ------ | ------ | ------ |
| 公开方法/类型 | PascalCase（对齐 Python 版 snake_case 的 PascalCase 转换） | `GetThreads` |
| 私有符号 | 小写驼峰 | `fetchFID` |
| 常量 | PascalCase 或驼峰（遵循 Go 惯用法） | `LatestVersion` |

公开方法名与 Python 版对齐：`get_threads` → `GetThreads`；首字母缩写按 Go 惯用法全大写：`get_fid` → `GetFID`、`get_fname` → `GetFName`。

## 错误处理规范

- API 方法以 `(T, error)` 双返回值暴露，内部失败通过 `error` 返回，结果结构体携带 `Err` 字段对齐 Python 的 `.err`
- 服务端返回非 0 错误码时抛 `*exception.TiebaServerError{Code, Msg}`
- 解析得到非预期值时抛 `*exception.TiebaValueError{Msg}`
- 网络状态异常时抛 `*exception.HTTPStatusError{Code, Msg}`
- 使用 `fmt.Errorf("...: %w", err)` 包装错误以保留错误链

## 日志规范

客户端方法没有装饰器可用，日志由方法内部手工记录，必须遵守下列约定。`logging_args_test.go` 会解析
`client.go` 的语法树自动校验前两条，违反即测试失败。

- 一个方法内所有 `c.logCallError` 站点必须使用**同一组调用方实参**（方法签名里除 `ctx` 外的参数），
  不能使用中途解析出的 `fid`、`fname` 等中间值。Python 的 `handle_exception` 每次调用只产生一条日志、
  参数恒为调用方实参，此约定用于对齐该行为
- 返回 `exception.BoolResponse` 的写操作必须补上 `c.logCallSuccess`，且参数与失败日志完全一致；
  读接口不记录成功日志
- 可选参数（来自 `XxxArgs` 结构体的字段）用 `logging.PyKw{Name: ..., Value: ...}` 标记；未标记的值
  按位置参数渲染，标记的值进 `kwargs`
- 新增异常类型时实现 `PyArgs() []any`，使其在日志中按 Python 的 `str(err)` 渲染
- 不要用字符串拼接手工拼日志正文，统一走 `c.logCallError` / `c.logCallSuccess`

## 测试编写规范

- 单元测试直接构造解析输入，不依赖真实网络
- 加密与 protobuf 解析用黄金向量对照（`api/*/testdata/*.hex` 由 Python 版生成，逐字节校验）
- 表驱动测试优先

## 依赖更新

`.github/dependabot.yml` 每周检查 `gomod` 与 `github-actions` 两个生态，更新 PR 一律提到 `develop`。
这两类改动不影响使用者，可给 PR 打 `skip-changelog` 标签跳过变更清单检查。

## 新增 API 时的检查清单

- [ ] 在 `api/<name>/` 创建子包，含 `api.go`（`PackProto`/`ParseBody`/`RequestURL`/`Request*`）与 `classdef.go`（`XxxFromProto`/`XxxFromJSON`/`XxxFromXML`）
- [ ] 在 `client.go` 中新增公开方法，接入 `tryInitWebsocket` / `forceWebsocket` 降级逻辑
- [ ] 在 `client.go` 顶部 import 中导入新 API 模块
- [ ] 如有新的 `.proto`，放入 `api/<name>/protobuf/`（通用类型放入 `protobuf/`），运行 `go run ./tools/genproto`
- [ ] 如有新的枚举类型，添加到 `enums/` 并附带 `XxxFrom` 回退构造
- [ ] 方法内所有日志站点使用同一组调用方实参；写操作补上 `c.logCallSuccess`
- [ ] 如有新的错误类型，实现 `PyArgs() []any`
- [ ] 运行 `gofmt -w .`、`go mod tidy`、`go build ./...`、`go vet ./...`、`go test ./...` 确保全绿
