# API 参考

本文件列出 aiotieba-go 的全部公开接口及其用途。

- 方法名与 Python 版 [aiotieba](https://github.com/lumina37/aiotieba) 一一对齐：`snake_case` →
  `PascalCase`，首字母缩写按 Go 惯用法全大写（`get_fid` → `GetFID`、`get_fname` → `GetFName`）
- 所有 `Client` 方法的第一个参数都是 `context.Context`
- 所有方法返回 `(T, error)`，结果结构体同时携带 `Err` 字段，对齐 Python 的 `.err`
- 完整签名与类型定义见 [Go Reference](https://pkg.go.dev/github.com/rongyuio/aiotieba-go)

## 构造与选项

`New` 的可选项类型为 `Option`，由下列函数构造。

| 接口 | 用途 |
| --- | --- |
| `New` | 创建客户端，校验 BDUSS / STOKEN 并初始化各会话 |
| `WithAccount` | 用已有的 `*Account` 覆盖 BDUSS / STOKEN |
| `WithTryWebsocket` | 允许接口在可用时优先走 WebSocket，失败自动降级 HTTP |
| `WithProxy` | 指定 HTTP 代理 |
| `WithProxyFromEnv` | 从环境变量读取代理配置 |
| `WithTimeout` | 覆盖各阶段的超时配置 |

## 参数构造辅助

| 接口 | 用途 |
| --- | --- |
| `ByFName` | 构造 `ForumRef`：用贴吧名引用一个吧 |
| `ByFID` | 构造 `ForumRef`：用 fid 引用一个吧 |
| `ByUserID` | 构造 `UserRef`：用数字 id 引用一个用户 |
| `ByPortrait` | 构造 `UserRef`：用 portrait 引用一个用户 |
| `ByUserName` | 构造 `UserRef`：用用户名引用一个用户 |
| `DefaultGetThreadsArgs` | 返回 `GetThreads` 的 Python 默认可选参数 |
| `DefaultGetPostsArgs` | 返回 `GetPosts` 的 Python 默认可选参数 |
| `DefaultGetCommentsArgs` | 返回 `GetComments` 的 Python 默认可选参数 |

`GetThreadsArgs` / `GetPostsArgs` / `GetCommentsArgs` 分别是对应接口的可选参数结构体，
零值并不等于 Python 的默认值，请以 `Default*Args` 的返回值为准。

## 会话与生命周期

| 方法 | 用途 |
| --- | --- |
| `Account` | 返回客户端的账号 |
| `SetAccount` | 替换所有会话的账号 |
| `User` | 返回账号的缓存用户信息 |
| `InitWebsocket` | 初始化 WebSocket 会话；返回 true 表示无须执行 |
| `InitTbs` | 在账号缺少 tbs token 时加载它 |
| `Login` | 刷新缓存的用户信息与 tbs token |
| `InitClientID` | 在账号缺少 client id 时加载它 |
| `InitSampleID` | 在账号缺少 sample id 时加载它 |
| `InitZID` | 在账号缺少 z_id 时加载它 |
| `HTTPCore` | 返回客户端的 HTTP 会话 |
| `WSCore` | 返回客户端的 WebSocket 会话 |
| `BLCPCore` | 返回客户端的 BLCP 会话 |
| `Close` | 释放客户端的网络资源 |

## 帖子与回复

| 方法 | 用途 |
| --- | --- |
| `GetThreads` | 获取首页帖子 |
| `GetPosts` | 获取主题帖内回复 |
| `GetComments` | 获取楼中楼回复 |
| `GetLastReplyers` | 通过旧版接口获取带最后回复人的首页帖子 |
| `AddPost` | 回复主题帖 |
| `DelPost` | 删除回复 |
| `DelPosts` | 批量删除回复 |
| `DelThread` | 删除主题帖 |
| `DelThreads` | 批量删除主题帖 |
| `HideThread` | 屏蔽主题帖 |
| `UnhideThread` | 解除主题帖屏蔽 |
| `SetThreadPrivate` | 隐藏主题帖 |
| `SetThreadPublic` | 公开主题帖 |
| `SetThreadPrivacy` | 隐藏或公开主题帖或回复；Go 版提供的合并入口 |
| `Recover` | 恢复主题帖或回复；isHide 为 true 则取消屏蔽，为 false 则撤销删帖 |
| `RecoverPost` | 恢复回复 |
| `RecoverThread` | 恢复主题帖 |
| `GetRecovers` | 获取吧务后台待恢复帖子列表 |
| `Good` | 加精主题帖 |
| `Ungood` | 撤精主题帖 |
| `Top` | 置顶主题帖 |
| `Untop` | 撤销置顶主题帖 |
| `Move` | 将主题帖移动至另一分区 |
| `Recommend` | 大吧主首页推荐 |
| `Agree` | 点赞主题帖或回复 |
| `Disagree` | 点踩主题帖或回复 |
| `Unagree` | 取消点赞主题帖或回复 |
| `Undisagree` | 取消点踩主题帖或回复 |
| `AddPoll` | 投票 |

## 用户

| 方法 | 用途 |
| --- | --- |
| `GetUserInfo` | 获取用户信息 |
| `GetSelfInfo` | 获取本账号信息 |
| `GetHomepage` | 获取用户个人页信息 |
| `GetFans` | 获取粉丝列表 |
| `GetFollows` | 获取关注列表 |
| `FollowUser` | 关注用户 |
| `UnfollowUser` | 取关用户 |
| `RemoveFan` | 移除粉丝 |
| `GetUserPosts` | 获取用户发布的回复列表 |
| `GetUserPostsPc` | 获取用户发布的回复列表（网页端接口） |
| `GetUserThreads` | 获取用户发布的主题帖列表 |
| `GetSelfPosts` | 获取当前用户发布的回复列表 |
| `GetSelfThreads` | 获取当前用户发布的主题帖列表 |
| `TiebaUID2UserInfo` | 通过 tieba_uid 获取用户信息 |
| `GetPortrait` | 获取用户头像 |
| `GetImage` | 从链接获取静态图像 |
| `GetImageBytes` | 从链接获取静态图像的原始字节流 |
| `Hash2Image` | 通过百度图库 hash 获取静态图像 |
| `SetProfile` | 设置主页信息 |
| `SetNicknameOld` | 设置旧版昵称 |

## 吧（论坛）

| 方法 | 用途 |
| --- | --- |
| `GetForum` | 获取贴吧信息；优先按贴吧名解析 |
| `GetForumDetail` | 获取贴吧信息；优先按 fid 解析 |
| `GetFID` | 通过贴吧名获取 forum_id |
| `GetFName` | 通过 forum_id 获取贴吧名 |
| `GetTabMap` | 获取分区名到分区 id 的映射 |
| `GetCID` | 通过精华分区名获取精华分区 id |
| `GetFollowForums` | 获取用户关注贴吧列表 |
| `GetFollowForumsPc` | 获取用户关注贴吧列表（网页端接口） |
| `GetSelfFollowForums` | 获取本账号关注贴吧列表 |
| `FollowForum` | 关注贴吧 |
| `UnfollowForum` | 取关贴吧 |
| `GetSquareForums` | 获取吧广场列表 |
| `GetRankForums` | 获取吧签到排行表 |
| `GetRankUsers` | 获取等级排行榜用户列表 |
| `DislikeForum` | 屏蔽贴吧，使其不再出现在首页推荐列表中 |
| `UndislikeForum` | 解除贴吧的首页推荐屏蔽 |
| `GetDislikeForums` | 获取首页推荐屏蔽的贴吧列表 |
| `GetUserForumInfo` | 获取用户在某吧内的信息 |
| `GetMemberUsers` | 获取最新关注用户列表 |

## 吧务管理

| 方法 | 用途 |
| --- | --- |
| `GetBawuInfo` | 获取吧务团队信息 |
| `GetBawuMemberlist` | 获取吧务后台吧会员列表 |
| `GetBawuPerm` | 获取指定吧务已分配的权限 |
| `SetBawuPerm` | 为指定吧务分配权限 |
| `AddBawu` | 添加吧务 |
| `DelBawu` | 删除吧务 |
| `GetBawuPostlogs` | 获取吧务后台帖子管理日志表 |
| `GetBawuUserlogs` | 获取吧务后台用户管理日志表 |
| `GetBawuBlacklist` | 获取吧务后台黑名单列表 |
| `AddBawuBlacklist` | 添加贴吧黑名单 |
| `DelBawuBlacklist` | 移出贴吧黑名单 |
| `Block` | 封禁用户 |
| `Unblock` | 解封用户 |
| `GetBlocks` | 获取吧务后台待解封用户列表 |
| `GetUnblockAppeals` | 获取吧务后台申诉请求列表 |
| `HandleUnblockAppeals` | 拒绝或通过解封申诉 |
| `SetBlacklist` | 设置新版用户黑名单 |
| `GetBlacklist` | 获取完整的新版用户黑名单列表 |
| `GetBlacklistOld` | 获取旧版用户黑名单列表 |
| `AddBlacklistOld` | 添加旧版用户黑名单 |
| `DelBlacklistOld` | 移除旧版用户黑名单 |
| `GetStatistics` | 获取吧务后台中最近 24 天的统计数据 |
| `GetRecomStatus` | 获取大吧主推荐功能的月度配额状态 |

## 消息与聊天

| 方法 | 用途 |
| --- | --- |
| `SendMsg` | 发送私信 |
| `GetGroupMsg` | 获取分组信息（仅 WebSocket） |
| `SetMsgReaded` | 将一条私信设为已读 |
| `SendChatroomMsg` | 向吧群发送信息，仅限简单文本；@ 他人需指定 atUserIDs |
| `JoinChatroom` | 加入聊天室 |
| `GetAts` | 获取 @ 信息 |
| `GetReplys` | 获取回复信息 |
| `GetRoomlistByFID` | 获取某吧所有群聊 |

## 签到

| 方法 | 用途 |
| --- | --- |
| `SignForum` | 单个贴吧签到 |
| `SignForums` | 一键签到 |
| `SignGrowth` | 用户成长等级任务：签到 |

## 搜索

| 方法 | 用途 |
| --- | --- |
| `SearchExact` | 贴吧内搜索，可限定分区与排序方式 |
| `SearchGlobal` | 全站主题帖关键词搜索，不限定贴吧 |

## 日志

日志接口位于 `logging` 包，控制台输出与 Python 版 aiotieba 对齐。用法与输出示例见
[README 的日志一节](../README.md#日志)。

| 接口 | 用途 |
| --- | --- |
| `logging.GetLogger` | 获取日志记录器（零值可用，懒加载默认实例） |
| `logging.SetLogger` | 替换为自定义的 zerolog 记录器 |
| `logging.SetLevel` | 设置日志级别，默认 Debug |
| `logging.EnableFileLog` | 追加写入 `log/<程序名>.log`，仅首次调用生效 |
| `logging.PyRepr` / `PyArgs` / `PyErr` | 把参数与异常渲染成 Python 风格文本 |
| `logging.PyKw` | 标记关键字参数，使其渲染进 `kwargs` |
