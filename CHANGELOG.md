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

### 新增

- `Client.GetBawuList`：吧务团队成员列表，含吧务助手、还能任命与招募的名额，
  以及每个成员是否为超级会员
- `Client.SearchInForumPosts`：在指定吧内搜索帖子。与 `SearchGlobal` 共用同一接口，
  区别是带上 `fname` 限定范围

### 修复

- `Client.GetFans` 与 `Client.GetFollows` 改走网页主域。两者路径在 `tieba.baidu.com`
  下，此前用 app 会话请求会被 hostOverride 改写到移动端 API 域，
  服务端返回空列表或 `110001`
- `Client.GetRecovers` 与 `Client.GetRecoverInfo` 补上客户端要求的 `tbs` 参数。
  缺少它时服务端返回 `300000 param invalid`
- `Client.GetCID` 补上客户端恒发的 `tbs`
- `Client.HandleUnblockAppeals` 按页面脚本对齐请求体：`fn` 传吧名，
  `refuse_reason` 在通过时为「已解封」

## [1.8.0] - 2026-10-30

本版对齐上游 aiotieba `6a32de11..3922a9f1`（2026-08-31 → 2026-10-25，80 个提交）。

### 破坏性变更

- **用户类型的 `IsGod` 字段全部移除**：`classdef.UserInfo`、`get_comments` 的 `UserInfoC` / `UserInfoCp` / `UserInfoCt`、
  `get_posts` 的 `UserInfoP` / `UserInfoPt`、`get_threads` 的 `UserInfoT`、`get_uinfo_getuserinfo_app` 的
  `UserInfoGuinfoApp`、`profile` 的 `UserInfoPF`、`tieba_uid2user_info` 的 `UserInfoTUid`。
  proto 的 `new_god_data` 与 `is_godthread_recommend` 一并删除
- **搜索接口整组改名**：`Client.SearchExact` → `SearchInForum`，`api/search_exact` → `api/search_in_forum`，
  `ExactSearch(s)` → `SearchInForum(s)`，`PageExsch` → `PageFsch`，
  `enums.SearchType` → `SearchInForumType`，`enums.GlobalSearchSortType` → `SearchGlobalType`，
  `searchglobal.GlobalSearchPost(s)` → `SearchGlobal(s)`
- **`get_replys` 回包重排**：`Reply` 删除 `TID` / `PPID` / `IsComment` / `PostUser` / `ThreadUser`，
  `Text` 由字段改为方法；tid 与 ppid 改由 `Thread.TID`、`Post.PID` 承载，`IsComment` 由 `ContentType` 取代
- **`get_ats` 重写**：`At` 删除 `IsComment` / `IsThread`，新增 `FID` / `Post` / `Thread` / `ContentType`；
  新增 `FragEmojiAt` / `FragImageAt` / `FragAtAt` / `ContentsAt` / `PostAt` / `ThreadAt`
- **`get_user_contents`**：`UserPost` 删除 `IsComment`，新增 `Thread` 与 `ContentType`；
  `ContentsUp.Texts` 由 `[]classdef.Fragment` 改为 `[]any`（`FragAtUp` 无法实现封闭的 `Fragment` 接口）
- **`get_user_contents_pc` 重写**：`PcUserPost` 删除 `TID`，新增 `Thread` / `ContentType` /
  `Forum`；新增 `ForumPcup` / `ThreadPcup` / `FragImagePcup` / `FragAtPcup` / `FragVideoPcup`；
  `FragVoiceUp` 改名为 `FragVoicePcup`
- **`get_threads`** 删除整条 feed 路径：`FragImageFeed` / `FragEmojiFeed` / `ContentsTFromFeed` /
  `ThreadFromFeed` / `ThreadsFromFeed` 与 proto 的 `PageData`
- 其余字段移除：`get_posts.CommentP` 的 `Agree` / `Disagree`、`get_posts.PageP` 的 `TotalCount`、
  `get_user_forum_info` 的 `MemberRank` / `Identify`、`get_dislike_forums` 的 `IsFollowed`、
  `get_posts.UserInfoPt` 的 `IsVIP`
- 字段改名与类型变化：`get_blocks.Block.NickNameOld` → `NickNameNew`（新增 `BlockTime`）、
  `get_forum_level.IsLike` 由 `int64` 改为 `bool`、`get_square_forums.PostNum` → `ThreadNum`
- **`Client.GetSelfFollowForums` 去掉 `pn` / `rn`**：上游已改为一次返回全部关注贴吧
- **websocket 会话按版本分桶**：`Client.wsCore` 变为 `map[string]*core.WsCore`，
  `core.NewWsCore` 增加 `clientVersion` 参数，`Client.WSCore()` 返回默认版本，
  新增 `Client.WSCoreFor`；`Client.InitWebsocket` 改为可变参数选项（`WithWebsocketVersion` / `WithWebsocketFull`）
- `core.WsCore.Connect` 只把会话置为 `Connecting`，新增 `MarkOpen` 由调用方在密钥上传成功后调用。
  判断会话可用性用 `wsCore.Status() == enums.WsStatusOpen`
- `core.MsgIDManager` 移除 `GetRecordID`，改用 `GetLocalMsgID`；新增 `PrivChatCursors` 与四个游标更新方法
- `api/send_msg` / `api/set_msg_readed` 的 `ParseBody` 改为返回 `SendMsgResult` / `SetMsgReadedResult`
- `core.BLCPCore.GenerateLCMToken` 改为返回 `(string, error)`；`UK` 的返回类型由 `string` 改为 `int64`；
  移除 `JoinChatRoom` 与 `FetchMcastMsgClientRequest`（能力移入 `api/join_chatroom` 与 `api/fetch_chatroom_msg`）
- `api/init_websocket` 改名为 `api/upload_ws_sec`，`PackProto` 增加 `clientVersion` 与 `full` 参数，
  `ParseBody` 改为返回 `InitWsInfo`
- `Account` 序列化键改名：`z_id` → `zid`、`aes_ecb_sec_key` → `ws_sec_key`、`aes_cbc_sec_key` → `zid_sec_key`，
  新增 `baidu_id`。`FromDict` 继续接受旧键名

### 新增

- `Client.GetRecoverInfo`、`Client.GetChatroomDetail`、`Client.FetchChatroomMsg`、`Client.InitBaiduID`、`Client.WSCoreFor`
- 新包 `api/get_chatroom_detail`、`api/fetch_chatroom_msg`、`api/join_chatroom`、`api/init_baidu_id`
- `enums.ContentType`（对应上游由 `ObjType` 改名而来的枚举）、`enums.BawuType` 补到 11 个成员
- `helper.TbNum2Int` / `helper.VipInfo2Bool` / `helper.BodySnippet`、`crypto.BDUKToUserID`
- websocket 支持按客户端版本建立多个会话，`upload_ws_sec` 支持上游的 min / full 两套设备指纹
- `get_comments` 的 `ContentsC.Imgs`、`get_forum` 的 `Admins`、`get_bawu_perm` 的 `User`、
  `get_comments.ThreadC.FloorNum`、`get_posts.ThreadP.CollectNum`、`get_self_follow_forums` 的四个计数、
  `classdef.FragAt.Portrait`

### 变更

- `consts.LatestVersion` 由 `22.10.1.0` 升至 `22.12.1.0`，`ChatVersion` 由 `12.68.1.0` 升至 `22.12.1.0`；
  新增 `LastReplyersVersion` 与 `UserPostsVersion`
- `IsVIP` 改由 `User.is_mem` 判定，不再依赖 `new_tshow_icon`；JSON 侧改用 `helper.VipInfo2Bool`
  （`v_status` 非零且 `e_time` 未过期）
- `get_bawu_info` 改按 `role_info` 内层的 `role_name` 分桶，无法识别的角色跳过；`All` 变为服务端原始顺序
- `get_self_follow_forums` 的请求体不再带 `tbs` / `page_no` / `res_num`
- `sync` 端点由 `/c/s/sync` 改为 `/c/s/realtimeSync`
- websocket 默认值对齐上游：`WSKeepAlive` 300s → 900s，`WSHeartbeat` 0 → 180s，并真正实现心跳与读截止时间
- BLCP 心跳间隔改由服务端 `lcm.next_interval_ms` 下发，未下发时回退 60 秒
- `Client.Close` 会关闭全部版本的 websocket 会话；断连时等待中的请求返回 `core.ErrWebsocketDisconnected`
- **请求头不再自报库身份**：HTTP 与 websocket 握手的 `User-Agent` 由 `aiotieba/<库版本>` 改为真机形态。
  app / app_proto 会话发 iOS 客户端 UA（含由账号派生的 `uniqueId`，与 `CommonReq.unique_id` 同源），
  web 会话发 webview UA。原先的写法等于每个请求都向服务端声明「这不是官方客户端」，且所有使用者
  共用同一串指纹
- `AddPost` 的请求体改为贴吧 iOS `22.12.1.0` 客户端的实际形态：设备档案由安卓 `12.35.1.0`
  （`SM-G988N` / Android 9）换成 iOS，补齐 `m_*` 调用链四元组、`caid`、`unique_id`、`idfv`、
  `x_field_59`、页面栈等字段，`post_from` 改为 `pb_reply`。设备级标识全部由账号持久化的
  `android_id` / `uuid` 派生，不含任何真实设备信息；同一账号恒定、不同账号互不相同

### 修复

- `get_replys` 的 `pn` 由 `string` 改回 `int32`。服务端按 varint 解析，原先以字符串发送时翻页参数无效
- `WsCore.Connect` 不再提前把会话置为 `Open`：此前并发调用 `Client.InitWebsocket` 时，第二个调用会因
  「会话未关闭」而跳过密钥上传，返回一个半初始化的会话
- `MsgIDManager.UpdateMsgID` 现在会登记未知分组。此前非私信分组的 `msg_id` 恒为 `0`
- BLCP 登录返回的 `uk` / `bduk` / `trigger_id` 现会写回客户端的用户信息，并按十进制渲染。
  此前 `SendChatroomMsg` 读到的是 `2.679882777e+09` 这类科学计数法（解析失败后恒为零）
- `get_lcm_token` 失败时向上返回错误，不再带着空 token 继续三步登录
- `get_uinfo_panel` 的计数不再把不带后缀的数字放大一万倍；`search_global` 的 `post_num` 支持 `1.2W` 缩写
- `get_rank_users` 的 `total_page` 取错字段，导致 `has_more` 几乎恒真
- `get_statistics` 的计数按上游反转为从旧到新
- `agree` 的表单字段改为 `content_type`，并修正主题帖与回复的类型值颠倒（上游：回复=1 楼中楼=2 主题帖=3）
- `GetUserPosts` 此前从不初始化 websocket，永远走 HTTP

### 内部

- 文档更正：本库并非上游的纯移植。`AGENTS.md`、`README.md`、`docs/api.md` 现明确区分与上游同源的接口和
  本库独有的 `AddPost`、`Share`、`wsframe` 实现，并说明后者不参与上游同步。
- 接口计数与目录树按当前代码更新：`Client` 共 123 个接口方法，「帖子与回复」补入 `Share`、
  「消息与聊天」补入两个聊天室接口，`api/` 下 104 个子包（103 个接口 + 1 个共享类型包）。
- 新增 `helper/deviceid`：`AddPost` 与 `Share` 共用同一套由账号派生设备标识的实现
  （`cuid` / `caid` / `unique_id` / `x_field_59` / UA / 时间戳），此前只存在于 `api/share` 包内。
- 依赖升级：`golang.org/x/net` `v0.60.0` → `v0.61.0`，`golang.org/x/image` `v0.46.0` → `v0.47.0`，
  `golang.org/x/sys` `v0.48.0` → `v0.49.0`（间接）
- `AGENTS.md` 约定功能分支的寿命：一处改动一个分支，合入 `develop` 后立刻删除源分支

## [1.7.0] - 2026-10-22

### 变更

- 最低 Go 版本由 `1.26.0` 提升至 `1.27.3`，CI 矩阵随之收敛为 1.27。
  原因是 `crypto/x509` 的证书链验证缺陷（CVE-2026-111943 / GO-2026-6809）：
  `Verify` 未拦截经由已撤销中间证书的链，Go 1.26.10 与 1.27.3 修复。
  本库唯一执行证书校验的路径是 BLCP 的 TLS 连接（`core/blcp.go` 的 `Connect`），
  其余 HTTP 与 WebSocket 通路均不校验证书，不额外扩大暴露面。
  使用 Go 1.26 的调用方在默认的 `GOTOOLCHAIN=auto` 下会自动下载 1.27.3，无需改代码；
  关闭自动工具链切换的构建需要手动升级工具链

## [1.6.0] - 2026-10-16

### 新增

- 新增 `Client.Share`，上报主题帖的分享行为，**会驱动帖子的分享数**。
  该接口只上报行为、不返回分享链接，并且没有 WebSocket 通路，仅走 HTTP。

  参数为 `Share(ctx, tid, fid, shareChannel)`：`tid` 是被分享的主题帖，
  `fid` 是该帖所在吧的 fid（分享前会先调一次 `getShareInfo`，它需要 `forum_id`），
  `shareChannel` 对应抓包中的 `share_channel`，实测服务端不校验该字段，默认传 0。

  实现要点：请求体 `data.log_id` 承载的是**被分享帖子的 tid**，不是随机数 ——
  服务端只校验这一个字段，填错则回 `errorno=0` 却不记账。详见 `docs/share.md` 第九节。

### 内部

- 新增 `core.HttpCore.LastCall()`，记录最近一次 app 调用（路径、服务端 logid、
  耗时、错误码），供 `Share` 申报 `CommonReq` 的 `m_api` / `m_logid` / `m_cost` /
  `m_result` 四元组。纯增量，不改变任何既有请求的行为。

## [1.5.1] - 2026-09-29

### 修复

- JSON 接口解析失败时，错误消息里带上响应体开头一段（折叠空白后截到 200 个字符）。
  上游返回 HTML（风控页 / 网关页）而不是 JSON 时，原来只报
  `invalid character '<' looking for beginning of value`，分不出是哪种情况，
  拿到这条错误的人无从判断该重试还是该停手；现在能直接看出返回的是什么。
  **错误消息因此变长，且会包含响应体的开头**——把错误原文写进自己日志的调用方请注意

## [1.5.0] - 2026-09-25

### 新增

- `GetComments` 新增 `GetCommentsArgs.Sort`，与 Python 版 `get_comments` 的 `sort` 参数对齐。
  服务端当前不采纳该参数——传时间顺序 / 时间倒序 / 热门序得到的返回顺序完全相同，HTTP 与
  WebSocket 两条传输都如此。保留是为了与上游保持一致，服务端实现后即自动生效

### 变更

- 客户端版本号（`consts.LatestVersion`）由 `22.6.5.1` 升至 `22.10.1.0`，与上游 Python 版对齐。
  所有使用该常量的接口都会以新版本号上报；`get_comments` 也从旧版格式（`12.64.1.1`）切到该版本

### 修复

- 修复 WebSocket 传输从未生效的问题。此前握手委托给 `gorilla/websocket`，而它拒绝调用方设置
  贴吧 IM 要求的非标准握手头 `Sec-WebSocket-Extensions: im_version=2.3`，导致握手必然失败：
  `WithTryWebsocket(true)` 与 `InitWebsocket` 恒为失败，所有能走 WebSocket 的接口都静默回退到
  HTTP。**依赖 WebSocket 的 `SendMsg`、`SetMsgReaded`、`GetGroupMsg` 在此前版本中完全不可用**，
  现已恢复
- 握手与 WebSocket 帧层改为自行实现（与 Python 版的做法一致），不再依赖
  `github.com/gorilla/websocket`

### 内部

- 重新生成受版本号影响的 15 个请求黄金向量（`api/*/testdata/req*.hex`）：先用 Python 绑定复现
  旧向量、确认解码与重编码逐字节无损，再按新版本号重出，保证 Go 与 Python 的输出仍逐字节一致
- `AGENTS.md` 与 `.github/CONTRIBUTING.md` 补充仓库操作约定：已发布的 tag 不可移动（Go 模块代理与
  校验和数据库均不可撤销）、`master` 的 HEAD 停在最新 release tag 上、分支保护与合并方式的现状，
  以及发版 PR 的命名与 tag 形态
- `AGENTS.md` 新增「与上游同步」：写明移植基线、同步只跟上游对百度接口行为的适配（上游没有
  CHANGELOG，改为按路径过滤 `git log`）、基线是参考而非权威，以及不要引入上游 tag 的原因
  （本库 module path 无 `/vN` 后缀，只能有 v0 / v1）
- `CI.yml` 显式声明 `permissions: contents: read`；`.gitignore` 中的 `test` 规则锚定为根目录
  `/test`，避免以后新建的 `test/` 目录被静默忽略

## [1.4.1] - 2026-09-25

### 内部

- `AGENTS.md` 新增「注释与中文文案」：记录中文文案的来源基线（迁移前 Python 源码
  `0847e2aa~1`）、读取方式，以及 `api/*/_api.py` 中只有 `search_global` 有 docstring 等事实
- `.github/CONTRIBUTING.md` 新增「注释规范」：说明「参数」段的书写形态，以及 Go 1.19 起
  `gofmt` 会把缩进注释行重排为 preformatted 代码块这一行为
- 新增 `SECURITY.md`，安全漏洞改走私密漏洞报告而非公开 issue
- `.github/ISSUE_TEMPLATE/` 补上默认标签与 `config.yml`（关闭空白 issue，提供文档与安全报告入口），
  并移除与已关闭的 Discussions 功能冲突的 `discussion.md`
- `dependabot.yml` 的检查频率由每天改为每周，减少更新 PR 的噪音
## [1.4.0] - 2026-09-15

### 变更

- 项目协议由 Unlicense（公有领域）改为 MIT。此前已发布的版本（含 v1.3.1）仍为 Unlicense，
  已获得的授权不受影响；自本次改动起，分发或修改需要保留版权声明与许可声明
- `go.mod` 的 `go` 指令由 `1.27.1` 下调为 `1.26.0`：Go 1.26 及以上都能引入本库，1.27.x 用户
  不受影响。1.26.0 是依赖（`golang.org/x/image`、`x/net`、`x/sys`）声明的最低版本

### 内部

- 修正 `go.mod` 中 `github.com/rs/zerolog` 与 `gopkg.in/natefinch/lumberjack.v2` 的 `// indirect`
  错标——两者都被 `logging` 包直接导入，此前因缺少一次 `go mod tidy` 而被记为间接依赖
- CI 增加 `go mod tidy -diff` 守卫（`CI.yml` 与 `release.yml`），go.mod 不整洁会直接失败
- CI 由单版本改为 Go 1.26 / 1.27 双版本矩阵，并新增同名的 `Test` 聚合任务以维持分支保护
  所需的检查名（矩阵会让检查名变成 `Test (1.26)`）
- `PULL_REQUEST_TEMPLATE.md` 与 `CONTRIBUTING.md` 的检查清单补上 `go mod tidy`

## [1.3.1] - 2026-09-15

### 修复

- `AddPost` 补上成功日志。Python 版 `add_post` 的装饰器标注了 `ok_log_level=logging.INFO`，
  Go 版此前只在失败时记录，成功时少输出一行

### 内部

- 新增 `TestBoolResponseMethodsLogSuccess`，断言返回 `exception.BoolResponse` 的公开方法
  都会记录成功日志（唯一例外 `JoinChatroom`，对应 Python 未标注 `ok_log_level` 的 `join_chatroom`）

## [1.3.0] - 2026-09-15

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
- 新增 `changelog.yml` 守住变更清单；Release 说明改为从 `CHANGELOG.md` 生成，
  并附上变更清单入口与与上一版本的对比链接
- README 的 API 列表拆分为 `docs/api.md`，逐条列出接口与用途

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
