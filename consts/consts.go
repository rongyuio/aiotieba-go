// Package consts 存放各贴吧 API 共用的版本号与基础域名。
//
// 对应 Python 模块 aiotieba.const。
package consts

const (
	// Version 库版本号，对应 aiotieba.__version__。
	Version = "1.7.0"

	// LatestVersion 通常用于大部分API。
	LatestVersion = "22.10.1.0"
	// LegacyVersion 通常用于部分依赖旧版格式的API (`get_threads`...)。
	LegacyVersion = "12.64.1.1"
	// ChatVersion 用于聊天 / BLCP 协议。
	ChatVersion = "12.68.1.0"
	// IOSVersion 是 iOS 客户端的版本号，用于 UA 与 _client_version。
	// 与 LatestVersion 分开：改这个只影响对齐 iOS 的调用点，
	// 不动其余接口已经稳定的版本号。
	IOSVersion = "22.12.1.0"

	// AppBaseHost app HTTP API 的域名。
	AppBaseHost = "tiebac.baidu.com"
	// WebBaseHost web HTTP API 的域名。
	WebBaseHost = "tieba.baidu.com"
)
