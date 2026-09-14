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

## 代码风格

- Go 代码风格遵循标准 `gofmt`（提交前运行 `gofmt -w .` 并确保 `gofmt -l .` 无输出）
- 导出符号必须有文档注释（以名称开头），内部符号尽量简洁
- 导入分组：标准库、第三方、本项目，按路径字典序排列

## 命名约定

| 类型 | 约定 | 示例 |
| ------ | ------ | ------ |
| 公开方法/类型 | PascalCase（对齐 Python 版 snake_case 的 PascalCase 转换） | `GetThreads` |
| 私有符号 | 小写驼峰 | `packProto` |
| 常量 | PascalCase 或驼峰（遵循 Go 惯用法） | `LatestVersion` |

公开方法名与 Python 版对齐：`get_threads` → `GetThreads`；首字母缩写按 Go 惯用法全大写：`get_fid` → `GetFID`、`get_fname` → `GetFName`。

## 错误处理规范

- API 方法以 `(T, error)` 双返回值暴露，内部失败通过 `error` 返回，结果结构体携带 `Err` 字段对齐 Python 的 `.err`
- 服务端返回非 0 错误码时抛 `*exception.TiebaServerError{Code, Msg}`
- 解析得到非预期值时抛 `*exception.TiebaValueError{Msg}`
- 网络状态异常时抛 `*exception.HTTPStatusError{Code, Msg}`
- 使用 `fmt.Errorf("...: %w", err)` 包装错误以保留错误链

## 测试编写规范

- 单元测试直接构造解析输入，不依赖真实网络
- 加密与 protobuf 解析用黄金向量对照（`api/*/testdata/*.hex` 由 Python 版生成，逐字节校验）
- 表驱动测试优先

## 新增 API 时的检查清单

- [ ] 在 `api/<name>/` 创建子包，含 `api.go`（`PackProto`/`ParseBody`/`Request*`）与 `classdef.go`（`XxxFromProto`/`XxxFromJSON`）
- [ ] 在 `client.go` 中新增公开方法，接入 `tryInitWebsocket` / `tryForceWebsocket` 降级逻辑
- [ ] 在 `client.go` 顶部 import 中导入新 API 模块
- [ ] 如有新的 `.proto`，放入 `api/<name>/protobuf/`（通用类型放入 `protobuf/`），运行 `go run ./tools/genproto`
- [ ] 如有新的枚举类型，添加到 `enums/` 并附带 `XxxFrom` 回退构造
- [ ] 运行 `gofmt -w .`、`go build ./...`、`go vet ./...`、`go test ./...` 确保全绿
