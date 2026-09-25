# 安全策略

## 支持范围

只对最新发布版本提供安全修复。项目处于 1.x 阶段，破坏性变更会随次要版本发布，因此请始终使用
[最新版本](https://github.com/rongyuio/aiotieba-go/releases)。

## 报告漏洞

**请不要用公开 issue 报告安全问题。**

请通过 GitHub 的[私密漏洞报告](https://github.com/rongyuio/aiotieba-go/security/advisories/new)提交，
该入口仅维护者可见。

报告中请尽量包含：

- 受影响的版本与运行环境
- 复现步骤或最小复现样例
- 影响评估：能读到什么、能改到什么

## 本库处理账号凭证

aiotieba-go 需要调用方提供 BDUSS、STOKEN 等百度账号凭证。下列情形同样按安全漏洞处理：

- 凭证被日志、错误信息或 panic 输出打印或回显
- 凭证被写入磁盘、临时文件，或被发往非预期的主机
- 请求签名与加密实现存在可被利用的缺陷（`helper/crypto/`、`core/websocket.go`、`core/blcp.go`）

## 处理流程

1. 收到报告后会尽快确认，目标是在 7 天内给出初步判断
2. 修复在 `develop` 上完成后，随下一个版本发布
3. 发布前不公开细节；发布后在对应的 GitHub Security Advisory 与 Release 说明中披露
4. 如需署名，请在报告中说明

## 免责声明

本项目是非官方的第三方实现，与百度及贴吧官方无关。
