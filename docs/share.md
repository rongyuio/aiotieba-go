# 贴吧分享接口调研

> **状态**：实现完成并**实测生效**（冷门帖 ShareNum 0 → 1）。原先的"回退、暂缓"结论已作废。
>
> 曾长期查不出原因，根因是 **`log_id(2)` 承载的是被分享帖子的 tid**，不是随机数
> —— 详见第一节与第九节。旧实现生成随机 log_id，等于每次上报一个不存在的帖子，
> 服务端回 `errorno=0` 却从不记账。
>
> 最后更新：2026-10-16

---

## 一、结论速览

| 问题 | 结论 |
|---|---|
| 接口地址 | `POST https://tiebac.baidu.com/c/c/thread/share?cmd=309480&format=protobuf` |
| 请求形态 | multipart 包装的 protobuf |
| 响应形态 | protobuf，固定 16 字节 |
| 是否返回分享链接 | **否**，只上报行为 |
| `share_channel` 是否被校验 | **否**，越界值 999 照样返回 `errorno=0` |
| **服务端实际校验什么** | **只有 `log_id(2)`** —— 见第九节 |
| 是否驱动分享数 | **会**，前提是 `log_id` 等于被分享帖子的 tid |
| 分享链接从哪来 | 客户端**本地拼接**，不需要服务端 |

---

## 二、接口结构

### 2.1 请求 `ShareReqIdl`

```proto
message ShareReqIdl {
    message DataReq {
        CommonReq common = 1;
        int64 log_id = 2;          // **被分享帖子的 tid**（不是随机数，见第九节）
        int32 share_channel = 3;   // 语义未确认；抓包中显式发 0 或 3，服务端不校验
        int32 x_field_4 = 4;       // 抓包中显式发 0
    }
    DataReq data = 1;
}
```

> `share_channel` 与 `x_field_4` 在抓包里都**显式带着零值上线**（wire 上确实是
> `f3=0`、`f4=0` 两个 varint），说明客户端的 proto 对这两个字段用了显式存在性。
> 我们的 `.proto` 是普通 proto3，零值会被省略 —— 但实测这不影响计数（见第九节）。

### 2.2 响应 `ShareResIdl`

```proto
message ShareResIdl {
    Error error = 1;
    message DataRes {
        int64 log_id = 1;   // 回显请求里的 log_id
    }
    DataRes data = 2;
}
```

实际响应的 16 字节（抓包真值）：

```
0a06080012001a00120608c1cdc79527
└ f1 Error(6) : 08 00 errorno=0 | 12 00 errmsg="" | 1a 00 f3=""
└ f2 Data(6)  : 08 c1cdc79527  log_id=10514196161
```

注意 `Error` 消息还有一个项目 `protobuf/Error.proto` 未定义的 `f3` 字段，会落进 `unknownFields`。

### 2.3 请求地址

`RawQuery` 为 `cmd=309480&format=protobuf`。项目内其他 protobuf 接口都只带 `cmd=`，
但抓包的 `meta.json` 里 `"query": "cmd=309480&format=protobuf"` 是直接证据。

---

## 三、抓包样本

`.chlz` 是 Charles 导出的 session，实为 zip，内含 `0-req.dat` / `0-res.dat` / `0-meta.json`。
样本在桌面：`1.chlz` ~ `6.chlz`，以及解包后的 `_chlz/` 目录。

| 样本 | 接口 | 时间 | tid | 说明 |
|---|---|---|---|---|
| 1、2、3 | `c/c/thread/share` | 09-22 | 610597734 / 683531960 / 1907603453 | 三个 share 请求，用于反解字段 |
| 4、5、6 | `c/f/forum/getShareInfo` | 09-23 | 10713887321 / 10974394402 / 8968444745 | 解开字段语义的关键 |

四条链路本身是 iOS 客户端（`_client_type=1`、`_client_version=22.11.1.0`、iPhone 15 Pro、iOS 18.7.2）。

### 3.1 share 请求的字段结构

```
ShareReqIdl
└─ data (f1)
   ├─ common (f1) = CommonReq { ...设备信息... }
   ├─ log_id (f2)      = 10514196161 / 10510412418 / 11044838858
   ├─ share_channel(3) = 0 / 0 / 3
   └─ x_field_4 (f4)   = 0
```

---

## 四、埋点字段语义

`CommonReq` 扩展了 15 个字段来承载这些日志项。经 4/5/6 号样本对照，**全部落地**。

**字段名已按上游 proto 校正。** 本节最初列的是从抓包值形态推断的名字
（`log_source`、`x_field_18` 之类）；后来从
[`clb-128258/tbclient.protobuf`](https://github.com/clb-128258/tbclient.protobuf)
（从贴吧 Android 客户端 22.9.1.0 提取）拿到了权威命名，下表用真名。
**名字倒拼的那些（`diac`/`ledom`/`dnarb`/`noisrev_so`…）是 iOS 侧的对应字段。**

| # | 权威名 | 最初猜的名 | 真义 | 证据 |
|---|---|---|---|---|
| 13 | `subapp_type` | `app_name` | 应用名 `tieba` | 硬编码，无熵 |
| 16 | `m_api` | `log_source` | **上一次调用的 API 路径（含查询串）** | 逐条比对，与上一条 baidu 请求的 URL 路径逐字相同 |
| 17 | `m_logid` | `log_source_id` | **上一次调用服务端返回的 logid** | 与上一条响应里的 `logid` 字段精确吻合（三连验证） |
| 18 | `m_cost` | `x_field_18` | **上一次调用的耗时(ms)** | `m_cost=147.596955`，与上一条 form 请求的 `m_cost` 同量级同格式 |
| 19 | `m_result` | `x_field_19` | **上一次调用的错误码** | getShareInfo 恒返 `error_code=110003` ↔ 该字段恒为 `110003` |
| 23 | `sign` | `x_field_23` | 32 位大写 hex，每次分享不同 | **实测换随机值不影响计数**，见第九节 |
| 27 | `brand_type` | `device_model` | 机型全称 | 抓包中恒等于 f9 `model` |
| 45 | `idfa` | 同 | iOS 广告标识 | 抓包为全 0 UUID（表示不可用） |
| 46 | `diac` | `caid` | 广告标识（名字倒拼） | URL 编码 JSON，含 3 个渠道 + `factors_data` |
| 58 | `shoubai_cuid` | `unique_id` | 客户端 uniqueId | **等于 form 里的 `shoubai_cuid`**，值逐字相同 |
| 59 | `mac` | `x_field_59` | 设备指纹 | **等于 form 里的 `mac`**，值逐字相同 |
| 71 / 72 | `pure_mode` / `xcx_mode` | `x_field_71/72` | 模式开关 | 抓包显式发 0 |
| 88 | `package_version` | `hybrid_version` | 内嵌 webview 版本 | `hybrid-main-pb_1.0.397.1` |
| 89 | `abtest_config_intervention` | `exp_flags` | 实验开关列表 | 两次抓包逐字相同 |

### 4.1 「上次调用三元组」（实为四元组）

`f16 m_api` / `f17 m_logid` / `f18 m_cost` / `f19 m_result` 是配套的一组，
记录**本接口调用之前那一次接口调用的路径、服务端 logid、耗时与错误码**。
这解释了三个看似无规律的现象：

- 为什么 `f16` 在两个样本里是 `c/f/forum/getShareInfo` 而在另一个里是 `c/s/newlog`
  —— 因为客户端上一次调的接口不同
- 为什么 `f19` 恒为 `110003` —— 因为客户端的上一跳 getShareInfo 一直返回这个错误码
- 为什么 `f18` 每次请求都不同 —— 因为那是真实耗时

**但实测这四个字段服务端一个都不校验** —— 逐个换成随机值后计数照样 +1（见第九节）。
实现里仍然如实申报它们（`core.HttpCore.LastCall()`），因为这是真实客户端的行为，
且成本极低；但它不是分享生效的必要条件。

> 旧实现把 `f17` 填成了被分享帖子的 tid —— 那是**错的**，该字段是上一次调用的 logid，
> 与 tid 无关。（真正承载 tid 的是 data 层的 `log_id(2)`，见第九节。）

---

## 五、`getShareInfo` 接口

### 5.1 请求形态

**不是 protobuf**，而是 `application/x-www-form-urlencoded`：

```
POST /c/f/forum/getShareInfo        （无 cmd 参数）
```

业务参数只有 `forum_id` 与 `thread_id`，其余是通用的那一套，外加三个日志字段
`m_api` / `m_cost` / `m_logid` 和一个 `sign`：

- `sign` 由 `crypto.Sign(data, []byte(crypto.AppSalt))` 生成（见 `helper/crypto/sign.go:28`
  与 `core/http.go:99` 的 `AppForm`）
- `m_api` 的值形如 `c/c/thread/share?cmd%3D309480%26format%3Dprotobuf`，即上一次调用的路径

### 5.2 响应

明文 JSON（Charles 已解压，`Content-Encoding: gzip`）：

```json
{"ctime":"0","data":{},"error_code":110003,"error_msg":"内部错误",
 "logid":"1029955438","server_time":13,"time":1790104629}
```

### 5.3 实测结果

**三次全部返回 `error_code=110003`「内部错误」**，真实客户端亦然。
该接口在分享链路里是取分享文案用的，失败不影响分享本身。

---

## 六、实测记录

### 6.1 `share_channel` 不被校验

对同一个帖子依次发送 `shareChannel ∈ {999, 0, 1, 2, 3, 4, 5}`，**七次全部返回 `errorno=0`**，
包括明显越界的 999。结论：服务端不校验该字段，它不是枚举，只是个日志维度。

### 6.2 分享数驱动问题（已解决）

**旧结论**：本实现的请求不驱动分享数，真实客户端可以，原因不明 —— 这是接口暂缓的原因。

**实际原因**：`log_id(2)` 填的是随机数，而它应该是**被分享帖子的 tid**。
服务端拿到一个不存在的 tid，回 `errorno=0` 但谁都不记账。见第九节。

下面这张表是当时的观测，保留作为记录。它本身没得出错误结论，
只是"API 请求不被计入"这个现象被归因错了 —— 不是设备档案不自洽，是目标帖子根本不存在：

| tid | 观测起点 | 后续 | 谁操作过 |
|---|---|---|---|
| 10713887321 | 1 | **4** | 客户端分享 + API 连发 7 次 |
| 10974394402 | 0 | **3** | 客户端分享 |
| 8968444745 | 0 | **1** | 客户端分享 |
| **10608273612** | 0 | **0** | **仅 API 分享 1 次** |
| 11018433275 | 3 | 3 | 无人 |

注意：qq仙境是活跃吧，`10713887321` 的 +3 不能全部归因于我们，也可能有真实用户分享。

**关于"更新延迟"**：早先观察到"20 多分钟后才看到变化"，据此推断有延迟。后来的实测推翻了
这一点 —— 冷门帖分享后立即读取即已 +1，**没有分钟级延迟**。当时那个观察很可能是
读数方法或帖子选择的问题。

### 6.3 API 请求与客户端请求的差异（**均不影响计数**）

传输层差异（当时的实现 vs iOS 抓包）：

| 项 | 旧实现 | 真实客户端 |
|---|---|---|
| HTTP `User-Agent` | `aiotieba/1.4.0` | 完整 iOS Safari UA，含 `tieba/22.12.1.0 uniqueId/…` |
| Cookie | **一个都不带** | 18 项，含 `BDUSS`/`BAIDUCUID`/`CUID`/`caid`/`STDCJUVF`… |
| `f1 _client_type` | 2（Android） | 1（iOS） |
| `f2 _client_version` | 12.35.1.0 | 22.12.1.0 |
| `f9/f26/f25` | SM-G988N / samsung / 9 | iPhone 15 Pro / iPhone / 18.7.2 |
| `f40 q_type` | 0 | 100 |
| `f60 idfv` | `"0"` | UUID |
| `f62 user_agent` | `aiotieba/1.4.0` | 完整 iOS Safari UA |
| `f70 device_score` | `0.4` | `0.800000` |
| `f46 diac`（CAID） | 628 字节，无 `factors_data` | 1435 字节，含 14 项设备画像 |

**这一整张表都被证伪了。** 第九节的单变量替换实测显示：上表每一项换成随机值后，
分享数**照样 +1**。服务端不校验设备档案、Cookie、UA 或 HTTP 头。

保留此表是为了记录"曾经以为的关键差异" —— 它们的共同特征是**看起来很像原因**，
而实际原因是 data 层里一个语义被理解错的字段。

---

## 七、未解决的问题（已全部回答）

1. ~~分享数到底由什么驱动~~ → **`log_id(2)` = 被分享帖子的 tid**，见第九节。
2. ~~API 请求为什么不被计入~~ → log_id 指向不存在的帖子。
3. **`share_channel` 的真实语义** —— 仍然不明。已知它不被校验（越界值 999 照样
   返回 `errorno=0`），抓包中出现 0 与 3，且零值会显式上线。**但不影响计数** ——
   实测把它整个删掉照样 +1。作为未解项保留。
4. ~~分享数的更新机制~~ → **无延迟**；实测同一帖子连续重放会**每次都 +1**，
   即不按 (账号, 帖子) 去重。

---

## 八、复现方法

### 8.1 解包 `.chlz`

`.chlz` 是 zip：

```python
import zipfile
z = zipfile.ZipFile("1.chlz")
for n in z.namelist():
    open(n.replace("/", "_"), "wb").write(z.read(n))
```

### 8.2 解析请求体

- **protobuf 请求**：按 wire format 递归解析（`tag = field<<3 | wiretype`）
- **form 请求**：直接按 `&` 与 `=` 切分
- **响应**：`Content-Encoding: gzip` 的已被 Charles 解压，直接读文本即可

### 8.3 读分享数不需要登录

`GetPosts` 在**未登录**（假 BDUSS）状态下可以正常读取，返回的 `Posts.Thread.ShareNum`
就是分享数。这省掉了登录步骤。

---

## 九、根因：`log_id` 就是 tid

### 9.1 结论

`ShareReqIdl.DataReq.log_id(2)` 承载的是**被分享帖子的 tid**。

抓包里它长得像随机数（`8737665232`），于是被当成"服务端原样回显的日志 id"，
实现里生成了一个随机值 —— 等于每次都在上报一个**不存在的帖子**。
服务端回 `errorno=0`（收下了这条上报），但没有任何帖子可记账，所以分享数从不变化。

**修复是一行**：`LogId: newLogID()` → `LogId: tid`。

### 9.2 怎么定位的

关键前提：**真机请求可以原样重放，且每次都会让目标帖 +1。** 这给了两个东西 ——
一条已知有效的基线，和一个可反复读取的观测量。

方法：拿真机请求当基线，**逐字段换成随机值再重放**，看计数何时断掉。

| 换掉的东西 | 结果 |
|---|---|
| `f10 BDUSS` 清空 | 照样 +1 |
| `f16 m_api`、`f17 m_logid` 换随机 | 照样 +1 |
| `f23 sign` 换随机 | 照样 +1 |
| `f58 shoubai_cuid`、`f59 mac`、`f60 idfv` 换随机 | 照样 +1 |
| `f91`（`view_path_info`）整个删掉 | 照样 +1 |
| Cookie 整个删掉 | 照样 +1 |
| 任意单个 HTTP 头删掉 | 照样 +1 |
| multipart 的 `filename` / part `Content-Type` 改掉 | 照样 +1 |
| **`log_id(2)` 换成另一个帖子的 tid** | **那个帖子 +1，原帖不动** |
| **`log_id(2)` 换成不存在的值** | **谁都不涨** |

**服务端实际只校验 `log_id` 这一个字段。**

一个必要的对照：什么都不做、连续读五次，计数稳定不动 —— 排除"计数本来就在自然增长"。

一个踩过的坑：先按"换头"和"换 body"交叉测试，得到过一组误导性结果。
原因是 `Content-Type` 里的 boundary 必须与 body 的 multipart 分隔符同源，
拆开替换会让服务端解析失败。**换 body 时必须同步换 Content-Type。**

### 9.3 这对既有分析意味着什么

第六节那张"API 请求与客户端请求的差异"表里，**每一项都被证伪了**：
设备档案、Cookie、UA、`m_*` 四元组、`sign` —— 换成随机值后计数照样 +1。

它们看起来都像原因（尤其是"整套安卓指纹配 iOS 的 log_source 路径"那条，
推理很顺），但服务端不看。**真正的原因是 data 层里一个语义被理解错的字段。**

教训：字段语义靠"值形态像什么"来推断是不可靠的 ——
`8737665232` 既是 10 位数又不像任何已知 id，被误判成随机日志 id。
**先找一条可重放的请求做单变量替换，比对着抓包猜语义快一个数量级。**

---

## 十、相关文件

```
抓包样本          C:\Users\Administrator\Desktop\{1..6}.chlz（已不存在）
实现              api/share/（api.go、getshareinfo.go、protobuf/）
                  client.go 的 Share 方法、protobuf/CommonReq.proto 的本地扩展字段
                  api_test.go
                  docs/api.md 与 CHANGELOG.md 的条目
```

> 早先那份"已回退的实现（`git stash@{0}`）"的记录已作废 ——
> 实现现位于分支 `feat/m-call-chain`，且已实测生效。
