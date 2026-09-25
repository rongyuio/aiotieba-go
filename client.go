// Package aiotieba 是 aiotieba 库的 Go 移植版本，即百度贴吧 API 的异步客户端。
//
// 公开方法名与 Python 客户端（aiotieba.client.Client）保持一致，便于两种实现对照；
// 内部则使用 Go 惯用法：context.Context、显式 error 返回与结构体。
package aiotieba

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/rongyuio/aiotieba-go/api/add_bawu"
	"github.com/rongyuio/aiotieba-go/api/add_bawu_blacklist"
	"github.com/rongyuio/aiotieba-go/api/add_blacklist_old"
	"github.com/rongyuio/aiotieba-go/api/add_poll"
	"github.com/rongyuio/aiotieba-go/api/add_post"
	"github.com/rongyuio/aiotieba-go/api/agree"
	"github.com/rongyuio/aiotieba-go/api/block"
	"github.com/rongyuio/aiotieba-go/api/classdef"
	"github.com/rongyuio/aiotieba-go/api/del_bawu"
	"github.com/rongyuio/aiotieba-go/api/del_bawu_blacklist"
	"github.com/rongyuio/aiotieba-go/api/del_blacklist_old"
	"github.com/rongyuio/aiotieba-go/api/del_post"
	"github.com/rongyuio/aiotieba-go/api/del_posts"
	"github.com/rongyuio/aiotieba-go/api/del_thread"
	"github.com/rongyuio/aiotieba-go/api/del_threads"
	"github.com/rongyuio/aiotieba-go/api/dislike_forum"
	"github.com/rongyuio/aiotieba-go/api/follow_forum"
	"github.com/rongyuio/aiotieba-go/api/follow_user"
	"github.com/rongyuio/aiotieba-go/api/get_ats"
	"github.com/rongyuio/aiotieba-go/api/get_bawu_blacklist"
	"github.com/rongyuio/aiotieba-go/api/get_bawu_info"
	"github.com/rongyuio/aiotieba-go/api/get_bawu_memberlist"
	"github.com/rongyuio/aiotieba-go/api/get_bawu_perm"
	"github.com/rongyuio/aiotieba-go/api/get_bawu_postlogs"
	"github.com/rongyuio/aiotieba-go/api/get_bawu_userlogs"
	"github.com/rongyuio/aiotieba-go/api/get_blacklist"
	"github.com/rongyuio/aiotieba-go/api/get_blacklist_old"
	"github.com/rongyuio/aiotieba-go/api/get_blocks"
	"github.com/rongyuio/aiotieba-go/api/get_cid"
	"github.com/rongyuio/aiotieba-go/api/get_comments"
	"github.com/rongyuio/aiotieba-go/api/get_dislike_forums"
	"github.com/rongyuio/aiotieba-go/api/get_fans"
	"github.com/rongyuio/aiotieba-go/api/get_fid"
	"github.com/rongyuio/aiotieba-go/api/get_follow_forums"
	"github.com/rongyuio/aiotieba-go/api/get_follow_forums_pc"
	"github.com/rongyuio/aiotieba-go/api/get_follows"
	"github.com/rongyuio/aiotieba-go/api/get_forum"
	"github.com/rongyuio/aiotieba-go/api/get_forum_detail"
	"github.com/rongyuio/aiotieba-go/api/get_forum_level"
	"github.com/rongyuio/aiotieba-go/api/get_group_msg"
	"github.com/rongyuio/aiotieba-go/api/get_images"
	"github.com/rongyuio/aiotieba-go/api/get_last_replyers"
	"github.com/rongyuio/aiotieba-go/api/get_member_users"
	"github.com/rongyuio/aiotieba-go/api/get_posts"
	"github.com/rongyuio/aiotieba-go/api/get_rank_forums"
	"github.com/rongyuio/aiotieba-go/api/get_rank_users"
	"github.com/rongyuio/aiotieba-go/api/get_recom_status"
	"github.com/rongyuio/aiotieba-go/api/get_recovers"
	"github.com/rongyuio/aiotieba-go/api/get_replys"
	"github.com/rongyuio/aiotieba-go/api/get_roomlist_by_fid"
	"github.com/rongyuio/aiotieba-go/api/get_self_follow_forums"
	"github.com/rongyuio/aiotieba-go/api/get_selfinfo_initNickname"
	"github.com/rongyuio/aiotieba-go/api/get_square_forums"
	"github.com/rongyuio/aiotieba-go/api/get_statistics"
	"github.com/rongyuio/aiotieba-go/api/get_tab_map"
	"github.com/rongyuio/aiotieba-go/api/get_threads"
	"github.com/rongyuio/aiotieba-go/api/get_uinfo_getUserInfo_web"
	"github.com/rongyuio/aiotieba-go/api/get_uinfo_getuserinfo_app"
	"github.com/rongyuio/aiotieba-go/api/get_uinfo_panel"
	"github.com/rongyuio/aiotieba-go/api/get_uinfo_userCard"
	"github.com/rongyuio/aiotieba-go/api/get_uinfo_user_json"
	"github.com/rongyuio/aiotieba-go/api/get_unblock_appeals"
	"github.com/rongyuio/aiotieba-go/api/get_user_contents"
	getusercontentsposts "github.com/rongyuio/aiotieba-go/api/get_user_contents/get_posts"
	getusercontentsthreads "github.com/rongyuio/aiotieba-go/api/get_user_contents/get_threads"
	"github.com/rongyuio/aiotieba-go/api/get_user_contents_pc"
	"github.com/rongyuio/aiotieba-go/api/get_user_forum_info"
	"github.com/rongyuio/aiotieba-go/api/good"
	"github.com/rongyuio/aiotieba-go/api/handle_unblock_appeals"
	"github.com/rongyuio/aiotieba-go/api/init_websocket"
	"github.com/rongyuio/aiotieba-go/api/init_z_id"
	"github.com/rongyuio/aiotieba-go/api/login"
	"github.com/rongyuio/aiotieba-go/api/move"
	"github.com/rongyuio/aiotieba-go/api/profile"
	"github.com/rongyuio/aiotieba-go/api/profile/get_homepage"
	"github.com/rongyuio/aiotieba-go/api/profile/get_uinfo_profile"
	"github.com/rongyuio/aiotieba-go/api/recommend"
	"github.com/rongyuio/aiotieba-go/api/recover"
	"github.com/rongyuio/aiotieba-go/api/remove_fan"
	"github.com/rongyuio/aiotieba-go/api/search_exact"
	"github.com/rongyuio/aiotieba-go/api/search_global"
	"github.com/rongyuio/aiotieba-go/api/send_chatroom_msg"
	"github.com/rongyuio/aiotieba-go/api/send_msg"
	"github.com/rongyuio/aiotieba-go/api/set_bawu_perm"
	"github.com/rongyuio/aiotieba-go/api/set_blacklist"
	"github.com/rongyuio/aiotieba-go/api/set_msg_readed"
	"github.com/rongyuio/aiotieba-go/api/set_nickname_old"
	"github.com/rongyuio/aiotieba-go/api/set_profile"
	"github.com/rongyuio/aiotieba-go/api/set_thread_privacy"
	"github.com/rongyuio/aiotieba-go/api/sign_forum"
	"github.com/rongyuio/aiotieba-go/api/sign_forums"
	"github.com/rongyuio/aiotieba-go/api/sign_growth"
	syncapi "github.com/rongyuio/aiotieba-go/api/sync"
	"github.com/rongyuio/aiotieba-go/api/tieba_uid2user_info"
	"github.com/rongyuio/aiotieba-go/api/top"
	"github.com/rongyuio/aiotieba-go/api/unblock"
	"github.com/rongyuio/aiotieba-go/api/undislike_forum"
	"github.com/rongyuio/aiotieba-go/api/unfollow_forum"
	"github.com/rongyuio/aiotieba-go/api/unfollow_user"
	"github.com/rongyuio/aiotieba-go/api/ungood"
	"github.com/rongyuio/aiotieba-go/config"
	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/enums"
	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/logging"
)

// bLCPQueueLength 是 BLCP 通知队列的容量，与 Python 默认值 100 一致。
const bLCPQueueLength = 100

// Option 用于配置 Client。
type Option func(*clientOptions)

type clientOptions struct {
	account *core.Account
	timeout config.TimeoutConfig
	proxy   *config.ProxyConfig
	tryWS   bool
}

// WithAccount 设置客户端账号，会覆盖 BDUSS 与 STOKEN。
func WithAccount(account *core.Account) Option {
	return func(o *clientOptions) { o.account = account }
}

// WithTryWebsocket 启用 websocket 传输，失败时回退到 HTTP。
func WithTryWebsocket(try bool) Option {
	return func(o *clientOptions) { o.tryWS = try }
}

// WithProxy 设置代理配置。
func WithProxy(proxy *config.ProxyConfig) Option {
	return func(o *clientOptions) { o.proxy = proxy }
}

// WithProxyFromEnv 使用环境变量描述的代理。
func WithProxyFromEnv() Option {
	return func(o *clientOptions) { o.proxy = config.FromEnv() }
}

// WithTimeout 设置超时配置。
func WithTimeout(timeout config.TimeoutConfig) Option {
	return func(o *clientOptions) { o.timeout = timeout }
}

// Client 贴吧客户端，是库的入口。它对应 aiotieba.client.Client。
type Client struct {
	mu sync.Mutex

	account *core.Account
	timeout config.TimeoutConfig
	proxy   *config.ProxyConfig
	tryWS   bool

	httpCore *core.HttpCore
	wsCore   *core.WsCore
	blcpCore *core.BLCPCore
	user     classdef.UserInfo
}

// New 创建客户端。account 会覆盖 BDUSS 与 STOKEN；try_ws 表示尝试使用 websocket 接口。
//
// 参数:
//
//	bduss BDUSS
//	stoken STOKEN
//	opts 可选配置，见 Option：account 会覆盖前两个参数，try_ws 尝试使用 websocket 接口，proxy 代理配置，timeout 超时配置
//
// BDUSS 必须为空或 192 个字符，STOKEN 必须为空或 64 个字符，与 Python 构造函数一致。
func New(bduss, stoken string, opts ...Option) (*Client, error) {
	options := clientOptions{timeout: config.DefaultTimeoutConfig()}
	for _, opt := range opts {
		opt(&options)
	}

	account := options.account
	if account == nil {
		var err error
		account, err = core.NewAccount(bduss, stoken)
		if err != nil {
			return nil, err
		}
	}

	netCore := core.NewNetCore(options.proxy, options.timeout)
	return &Client{
		account:  account,
		timeout:  options.timeout,
		proxy:    options.proxy,
		tryWS:    options.tryWS,
		httpCore: core.NewHttpCore(account, netCore),
		wsCore:   core.NewWsCore(account, netCore),
		blcpCore: core.NewBLCPCore(account, netCore, bLCPQueueLength),
	}, nil
}

// Close 释放客户端的网络资源，对应 __aexit__。
func (c *Client) Close() error {
	var errs []error
	if err := c.wsCore.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := c.blcpCore.Close(); err != nil {
		errs = append(errs, err)
	}
	c.httpCore.NetCore.Close()
	return errors.Join(errs...)
}

// Account 返回客户端的账号。
func (c *Client) Account() *core.Account {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.account
}

// SetAccount 替换所有会话的账号，对应 Python 的 account setter。
func (c *Client) SetAccount(newAccount *core.Account) error {
	if newAccount == nil {
		return errors.New("aiotieba: the new account is nil")
	}
	c.mu.Lock()
	c.account = newAccount
	c.mu.Unlock()

	c.httpCore.SetAccount(newAccount)
	c.wsCore.SetAccount(newAccount)
	c.blcpCore.SetAccount(newAccount)
	return nil
}

// User 返回账号的缓存用户信息，对应 self_info 属性。
func (c *Client) User() classdef.UserInfo {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.user
}

// HTTPCore 返回客户端的 HTTP 会话。
func (c *Client) HTTPCore() *core.HttpCore { return c.httpCore }

// WSCore 返回客户端的 websocket 会话。
func (c *Client) WSCore() *core.WsCore { return c.wsCore }

// BLCPCore 返回客户端的 BLCP 会话。
func (c *Client) BLCPCore() *core.BLCPCore { return c.blcpCore }

// InitWebsocket 初始化 websocket。返回 true 表示无须执行，false 表示失败。
//
// 连接 websocket 会话并上传密钥，对应 Client.init_websocket。会话未关闭时不执行任何操作。
func (c *Client) InitWebsocket(ctx context.Context) (bool, error) {
	if c.wsCore.Status() != enums.WsStatusClosed {
		return true, nil
	}
	if err := c.wsCore.Connect(ctx); err != nil {
		return false, err
	}
	if err := c.uploadSecKey(); err != nil {
		// 上传失败时 Python 客户端会把状态重置为 CLOSED，以便下次调用重试握手。
		_ = c.wsCore.Close()
		return false, err
	}
	return true, nil
}

func (c *Client) uploadSecKey() error {
	groups, err := initwebsocket.Request(c.wsCore)
	if err != nil {
		return fmt.Errorf("aiotieba: uploading the websocket sec key: %w", err)
	}

	manager := c.wsCore.MsgIDManager()
	if manager == nil {
		return errors.New("aiotieba: the websocket session is not initialised")
	}
	for _, group := range groups {
		if enums.GroupType(int(group.GroupType)) == enums.GroupTypePrivateMsg {
			manager.PrivGID = int(group.GroupID)
		}
		manager.GID2MID[int(group.GroupID)] = &core.MsgIDPair{
			LastID: int(group.LastMsgID),
			CurrID: int(group.LastMsgID),
		}
	}
	return nil
}

// GetThreadsArgs 是 GetThreads 的可选参数。
type GetThreadsArgs struct {
	Pn     int                  // 页码
	Rn     int                  // 请求的条目数 Max to 100
	Sort   enums.ThreadSortType // HOT热门排序 REPLY按回复时间 CREATE按发布时间 FOLLOW关注的人
	IsGood bool                 // True则获取精品区帖子 False则获取普通区帖子
}

// DefaultGetThreadsArgs 返回 GetThreads 的 Python 默认值。
func DefaultGetThreadsArgs() GetThreadsArgs {
	return GetThreadsArgs{Pn: 1, Rn: 30, Sort: enums.ThreadSortReply}
}

// GetThreads 获取首页帖子。
//
// 参数:
//
//	ref 目标贴吧名或fid 优先贴吧名
//	args 可选参数，详见 GetThreadsArgs
//
// 对应 Client.get_threads。会话打开时使用 websocket 传输，否则回退到 app HTTP API。
func (c *Client) GetThreads(ctx context.Context, ref ForumRef, args GetThreadsArgs) (getthreads.Threads, error) {
	fname, err := c.fetchFNameOrFName(ctx, ref)
	if err != nil {
		c.logCallError("get_threads", err, ref, logging.PyKw{Name: "pn", Value: args.Pn})
		return getthreads.Threads{}, err
	}

	c.tryInitWebsocket(ctx)

	var (
		threads getthreads.Threads
		reqErr  error
	)
	pn, rn, sort := int32(args.Pn), int32(args.Rn), int32(args.Sort)
	if c.wsCore.Status() == enums.WsStatusOpen {
		threads, reqErr = getthreads.RequestWS(c.wsCore, fname, pn, rn, sort, args.IsGood, consts.LegacyVersion)
	} else {
		threads, reqErr = getthreads.RequestHTTP(ctx, c.httpCore, fname, pn, rn, sort, args.IsGood, consts.LegacyVersion)
	}
	if reqErr != nil {
		c.logCallError("get_threads", reqErr, ref, logging.PyKw{Name: "pn", Value: args.Pn})
		return threads, reqErr
	}
	return threads, nil
}

// ForumRef 通过贴吧名或 fid 标识一个贴吧，对应 Python 客户端的
// `fname_or_fid: str | int` 参数：GetForum 优先使用贴吧名，GetForumDetail 优先使用 fid。
type ForumRef struct {
	FName string
	FID   int64
}

// PyRepr 让日志输出贴吧名（优先）或 fid，而不是 Go 结构体的 %+v 形式。
func (r ForumRef) PyRepr() string {
	if r.FName != "" {
		return logging.PyRepr(r.FName)
	}
	return logging.PyRepr(r.FID)
}

// ByFName 通过贴吧名引用一个贴吧。
func ByFName(fname string) ForumRef { return ForumRef{FName: fname} }

// ByFID 通过 fid 引用一个贴吧。
func ByFID(fid int64) ForumRef { return ForumRef{FID: fid} }

// GetFID 通过贴吧名获取 forum_id。
//
// 参数:
//
//	fname 贴吧名
//
// 优先使用吧信息缓存，对应 Client.get_fid。
func (c *Client) GetFID(ctx context.Context, fname string) (int64, error) {
	fid, err := c.fetchFID(ctx, fname)
	if err != nil {
		c.logCallError("get_fid", err, fname)
		return 0, err
	}
	return fid, nil
}

// GetFName 通过 forum_id 获取贴吧名。
//
// 参数:
//
//	fid forum_id
//
// 优先使用吧信息缓存，对应 Client.get_fname。
func (c *Client) GetFName(ctx context.Context, fid int64) (string, error) {
	fname, err := c.fetchFName(ctx, fid)
	if err != nil {
		c.logCallError("get_fname", err, fid)
		return "", err
	}
	return fname, nil
}

// GetForum 获取贴吧信息。
// 此接口较 GetForumDetail 更强大。
//
// 参数:
//
//	ref 目标贴吧名或fid 优先贴吧名
//
// 对应 Client.get_forum。
func (c *Client) GetForum(ctx context.Context, ref ForumRef) (getforum.Forum, error) {
	fname := ref.FName
	if fname == "" {
		var err error
		if fname, err = c.fetchFName(ctx, ref.FID); err != nil {
			c.logCallError("get_forum", err, ref)
			return getforum.Forum{}, err
		}
	}

	forum, err := getforum.Request(ctx, c.httpCore, fname)
	if err != nil {
		c.logCallError("get_forum", err, ref)
		return getforum.Forum{}, err
	}
	return forum, nil
}

// GetForumDetail 获取贴吧信息。
//
// 参数:
//
//	ref 目标贴吧名或fid 优先fid
//
// 对应 Client.get_forum_detail。
func (c *Client) GetForumDetail(ctx context.Context, ref ForumRef) (getforumdetail.ForumDetail, error) {
	c.tryInitWebsocket(ctx)

	fid := ref.FID
	if fid == 0 {
		var err error
		if fid, err = c.fetchFID(ctx, ref.FName); err != nil {
			c.logCallError("get_forum_detail", err, ref)
			return getforumdetail.ForumDetail{}, err
		}
	}

	var (
		detail getforumdetail.ForumDetail
		err    error
	)
	if c.wsCore.Status() == enums.WsStatusOpen {
		detail, err = getforumdetail.RequestWS(c.wsCore, fid)
	} else {
		detail, err = getforumdetail.RequestHTTP(ctx, c.httpCore, fid)
	}
	if err != nil {
		c.logCallError("get_forum_detail", err, ref)
		return detail, err
	}
	return detail, nil
}

// fetchFIDOrFID 对应 Python 的 `fid = fname_or_fid if isinstance(fname_or_fid, int)
// else await self.__get_fid(fname_or_fid)` 写法。
func (c *Client) fetchFIDOrFID(ctx context.Context, ref ForumRef) (int64, error) {
	if ref.FName != "" {
		return c.fetchFID(ctx, ref.FName)
	}
	return ref.FID, nil
}

// fetchFNameOrFName 对应 Python 的 `fname = fname_or_fid if
// isinstance(fname_or_fid, str) else await self.__get_fname(fname_or_fid)` 写法。
func (c *Client) fetchFNameOrFName(ctx context.Context, ref ForumRef) (string, error) {
	if ref.FName != "" {
		return ref.FName, nil
	}
	return c.fetchFName(ctx, ref.FID)
}

// fetchFID 对应私有方法 Client.__get_fid。
func (c *Client) fetchFID(ctx context.Context, fname string) (int64, error) {
	if fid, ok := helper.DefaultForumInfoCache.GetFid(fname); ok && fid != 0 {
		return fid, nil
	}
	fid, err := getfid.Request(ctx, c.httpCore, fname)
	if err != nil {
		return 0, err
	}
	helper.DefaultForumInfoCache.AddForum(fname, fid)
	return fid, nil
}

// fetchFName 对应私有方法 Client.__get_fname。
func (c *Client) fetchFName(ctx context.Context, fid int64) (string, error) {
	if fname, ok := helper.DefaultForumInfoCache.GetFname(fid); ok && fname != "" {
		return fname, nil
	}
	detail, err := c.GetForumDetail(ctx, ByFID(fid))
	if err != nil {
		return "", err
	}
	if detail.FName != "" {
		helper.DefaultForumInfoCache.AddForum(detail.FName, fid)
	}
	return detail.FName, nil
}

// InitTbs 在账号缺少 tbs token 时加载它，对应私有方法 Client.__init_tbs。
func (c *Client) InitTbs(ctx context.Context) error {
	if c.account.Tbs() != "" {
		return nil
	}
	return c.Login(ctx)
}

// Login 刷新缓存的用户信息与 tbs token，对应私有方法 Client.__login。
func (c *Client) Login(ctx context.Context) error {
	user, tbs, err := login.Request(ctx, c.httpCore)
	if err != nil {
		c.logCallError("login", err)
		return err
	}

	c.mu.Lock()
	c.user.UserID = user.UserID
	if user.Portrait != "" {
		c.user.Portrait = user.Portrait
	}
	if user.UserName != "" {
		c.user.UserName = user.UserName
	}
	c.mu.Unlock()
	c.account.SetTbs(tbs)
	return nil
}

// InitClientID 在账号缺少 client id 时加载它，对应私有方法 Client.__init_client_id。
func (c *Client) InitClientID(ctx context.Context) error {
	if c.account.ClientID() != "" {
		return nil
	}
	return c.sync(ctx)
}

// InitSampleID 在账号缺少 sample id 时加载它，对应私有方法 Client.__init_sample_id。
func (c *Client) InitSampleID(ctx context.Context) error {
	if c.account.SampleID() != "" {
		return nil
	}
	return c.sync(ctx)
}

// InitZID 在账号缺少 z_id 时加载它，对应私有方法 Client.__init_z_id。
func (c *Client) InitZID(ctx context.Context) error {
	if c.account.ZID() != "" {
		return nil
	}
	zid, err := initzid.Request(ctx, c.httpCore)
	if err != nil {
		c.logCallError("init_z_id", err)
		return err
	}
	c.account.SetZID(zid)
	return nil
}

// sync 对应私有方法 Client.__sync。
func (c *Client) sync(ctx context.Context) error {
	clientID, sampleID, err := syncapi.Request(ctx, c.httpCore)
	if err != nil {
		c.logCallError("sync", err)
		return err
	}
	c.account.SetClientID(clientID)
	c.account.SetSampleID(sampleID)
	return nil
}

// UserRef 通过数字 id、portrait 或用户名标识一个用户，对应 Python 客户端的
// `id_: str | int` 参数：字符串要么是 portrait（"tb.1.xxx"），要么是用户名。
type UserRef struct {
	UserID   int64
	Portrait string
	UserName string
}

// PyRepr 让日志输出用户名/portrait/user_id，而不是 Go 结构体的 %+v 形式。
func (r UserRef) PyRepr() string {
	switch {
	case r.UserName != "":
		return logging.PyRepr(r.UserName)
	case r.Portrait != "":
		return logging.PyRepr(r.Portrait)
	default:
		return logging.PyRepr(r.UserID)
	}
}

// ByUserID 通过数字 id 引用一个用户。
func ByUserID(id int64) UserRef { return UserRef{UserID: id} }

// ByPortrait 通过 portrait 引用一个用户。
func ByPortrait(portrait string) UserRef { return UserRef{Portrait: portrait} }

// ByUserName 通过用户名引用一个用户。
func ByUserName(userName string) UserRef { return UserRef{UserName: userName} }

// IsZero 报告该引用是否不含任何标识。
func (r UserRef) IsZero() bool {
	return r.UserID == 0 && r.Portrait == "" && r.UserName == ""
}

// isSubset 报告 a 的每一位是否都在 b 中置位，对应 Python 中用于判断只请求了
// 某些字段的 `(a | b) == b` 写法。
func isSubset(a, b enums.ReqUInfo) bool { return a|b == b }

// GetUserInfo 获取用户信息。
//
// 参数:
//
//	ref 用户id user_id / portrait / user_name
//	require 指示需要获取的字段
//
// 对应 Client.get_user_info。
//
// Python 客户端每个端点返回不同的专用类型；本移植版把所有分支统一为 classdef.UserInfo，
// 使调用方看到稳定类型而请求本身保持不变。端点尚未迁移的分支会返回 exception.ErrNotMigrated。
func (c *Client) GetUserInfo(ctx context.Context, ref UserRef, require enums.ReqUInfo) (classdef.UserInfo, error) {
	if ref.IsZero() {
		logging.GetLogger().Warn().Msg("GetUserInfo: empty input")
		return classdef.UserInfo{}, nil
	}

	var (
		user classdef.UserInfo
		err  error
	)
	switch {
	case ref.UserName == "":
		user, err = c.getUserInfoByIDOrPortrait(ctx, ref, require)
	default:
		user, err = c.getUserInfoByName(ctx, ref.UserName, require)
	}
	if err != nil {
		c.logCallError("get_user_info", err, require)
		return classdef.UserInfo{}, err
	}
	return user, nil
}

// getUserInfoByIDOrPortrait 对应数字 id 与 portrait 两个分支。
func (c *Client) getUserInfoByIDOrPortrait(ctx context.Context, ref UserRef, require enums.ReqUInfo) (classdef.UserInfo, error) {
	if ref.Portrait != "" {
		// 该引用是 portrait。
		if isSubset(require, enums.ReqUInfoBasic) && require&enums.ReqUInfoUserID == 0 {
			return c.getUinfoPanel(ctx, ref.Portrait)
		}
		if isSubset(require, enums.ReqUInfoNickName|enums.ReqUInfoTiebaUID) {
			return c.getUinfoUserCard(ctx, ref.Portrait)
		}
		return c.getUinfoProfile(ctx, ref)
	}

	// 该引用是数字 id。
	if isSubset(require, enums.ReqUInfoBasic) {
		return c.getUinfoGetUserInfoApp(ctx, ref.UserID)
	}
	if c.account.BDUSS() != "" && require&(enums.ReqUInfoTiebaUID|enums.ReqUInfoOther) == 0 {
		return c.getUinfoGetUserInfoWeb(ctx, ref.UserID)
	}
	return c.getUinfoProfile(ctx, ref)
}

// getUserInfoByName 对应用户名分支。
func (c *Client) getUserInfoByName(ctx context.Context, userName string, require enums.ReqUInfo) (classdef.UserInfo, error) {
	if isSubset(require, enums.ReqUInfoBasic) {
		return c.getUinfoUserJSON(ctx, userName)
	}
	if require&enums.ReqUInfoNickName != 0 && require&(enums.ReqUInfoUserID|enums.ReqUInfoTiebaUID) == 0 {
		return c.getUinfoPanel(ctx, userName)
	}
	user, err := c.getUinfoUserJSON(ctx, userName)
	if err != nil {
		return classdef.UserInfo{}, err
	}
	return c.getUinfoProfile(ctx, ByPortrait(user.Portrait))
}

// getUinfoGetUserInfoApp 接口 https://tiebac.baidu.com/c/u/user/getuserinfo
// 返回 user_id / portrait / user_name / 性别 / 是否大神 / 是否超会。
//
// 参数:
//
//	userID 用户id user_id
//
// getUinfoGetUserInfoApp 对应私有函数 _get_uinfo_getuserinfo。
func (c *Client) getUinfoGetUserInfoApp(ctx context.Context, userID int64) (classdef.UserInfo, error) {
	if c.wsCore.Status() == enums.WsStatusOpen {
		user, err := getuserinfoapp.RequestWS(c.wsCore, userID)
		if err != nil {
			return classdef.UserInfo{}, err
		}
		return appUserToUserInfo(user), nil
	}

	user, err := getuserinfoapp.RequestHTTP(ctx, c.httpCore, userID)
	if err != nil {
		return classdef.UserInfo{}, err
	}
	// app 端点会把大于 math.MaxInt32 的 id 上报为负数。
	if user.UserID < 0 {
		user.UserID = 0xFFFFFFFF + user.UserID
	}
	return appUserToUserInfo(user), nil
}

func appUserToUserInfo(u getuserinfoapp.UserInfoGuinfoApp) classdef.UserInfo {
	return classdef.UserInfo{
		UserID:      u.UserID,
		Portrait:    u.Portrait,
		UserName:    u.UserName,
		NickNameOld: u.NickNameOld,
		Gender:      u.Gender,
		IsVIP:       u.IsVIP,
		IsGod:       u.IsGod,
	}
}

// getUinfoGetUserInfoWeb 接口 http://tieba.baidu.com/im/pcmsg/query/getUserInfo
// 返回 user_id / portrait / user_name / nick_name_new。该接口需要 BDUSS。
//
// 参数:
//
//	userID 用户id user_id
//
// getUinfoGetUserInfoWeb 对应私有函数 _get_uinfo_getUserInfo。
func (c *Client) getUinfoGetUserInfoWeb(ctx context.Context, userID int64) (classdef.UserInfo, error) {
	user, err := getuserinfoweb.Request(ctx, c.httpCore, userID)
	if err != nil {
		return classdef.UserInfo{}, err
	}
	return classdef.UserInfo{
		// 该端点不回显 id，因此使用请求时传入的 id。
		UserID:      userID,
		Portrait:    user.Portrait,
		UserName:    user.UserName,
		NickNameNew: user.NickNameNew,
	}, nil
}

// getUinfoUserJSON 接口 http://tieba.baidu.com/i/sys/user_json
// 返回 user_id / portrait / user_name。
//
// 参数:
//
//	userName 用户id user_name
//
// getUinfoUserJSON 对应私有函数 _get_uinfo_user_json。
func (c *Client) getUinfoUserJSON(ctx context.Context, userName string) (classdef.UserInfo, error) {
	user, err := getuserjson.Request(ctx, c.httpCore, userName)
	if err != nil {
		return classdef.UserInfo{}, err
	}
	return classdef.UserInfo{
		UserID:   user.UserID,
		Portrait: user.Portrait,
		// 该端点不回显用户名。
		UserName: userName,
	}, nil
}

// getUinfoPanel 接口 https://tieba.baidu.com/home/get/panel
// 返回 portrait / user_name / age / 是否超会 等信息。
//
// 从 2022.08.30 开始服务端不再返回 user_id 字段，请谨慎使用；
// 该接口可判断用户是否被屏蔽；该接口 rps 阈值较低。
//
// 参数:
//
//	nameOrPortrait 用户id user_name / portrait
//
// getUinfoPanel 对应私有函数 _get_uinfo_panel。
func (c *Client) getUinfoPanel(ctx context.Context, nameOrPortrait string) (classdef.UserInfo, error) {
	user, err := getuinfopanel.Request(ctx, c.httpCore, nameOrPortrait)
	if err != nil {
		return classdef.UserInfo{}, err
	}
	return classdef.UserInfo{
		Portrait:    user.Portrait,
		UserName:    user.UserName,
		NickNameNew: user.NickNameNew,
		NickNameOld: user.NickNameOld,
		Gender:      user.Gender,
		Age:         user.Age,
		PostNum:     user.PostNum,
		FanNum:      user.FanNum,
		IsVIP:       user.IsVIP,
	}, nil
}

// getUinfoUserCard 接口 https://tieba.baidu.com/c/u/pc/userCard
// 返回 portrait / tieba_uid / nick_name_new / age / sign / ip 等信息。
//
// 参数:
//
//	portrait 用户portrait
//
// getUinfoUserCard 对应私有函数 _get_uinfo_userCard。
func (c *Client) getUinfoUserCard(ctx context.Context, portrait string) (classdef.UserInfo, error) {
	user, err := getuserinfousercard.Request(ctx, c.httpCore, portrait)
	if err != nil {
		return classdef.UserInfo{}, err
	}
	return classdef.UserInfo{
		Portrait:    user.Portrait,
		NickNameNew: user.NickNameNew,
		TiebaUID:    user.TiebaUID,
		Gender:      user.Gender,
		Age:         user.Age,
		AgreeNum:    user.AgreeNum,
		FanNum:      user.FanNum,
		FollowNum:   user.FollowNum,
		Sign:        user.Sign,
		IP:          user.IP,
	}, nil
}

// getUinfoProfile 接口 https://tiebac.baidu.com/c/u/user/profile
// 返回包含最全面用户信息的 UserInfo。
//
// 参数:
//
//	ref 用户id user_id / portrait
//
// getUinfoProfile 对应私有函数 _get_uinfo_profile。
func (c *Client) getUinfoProfile(ctx context.Context, ref UserRef) (classdef.UserInfo, error) {
	pref := profile.ByUserID(ref.UserID)
	if ref.Portrait != "" {
		pref = profile.ByPortrait(ref.Portrait)
	}

	var (
		user profile.UserInfoPF
		err  error
	)
	if c.wsCore.Status() == enums.WsStatusOpen {
		user, err = getuinfoprofile.RequestWS(c.wsCore, pref)
	} else {
		user, err = getuinfoprofile.RequestHTTP(ctx, c.httpCore, pref)
	}
	if err != nil {
		return classdef.UserInfo{}, err
	}

	return classdef.UserInfo{
		UserID:      user.UserID,
		Portrait:    user.Portrait,
		UserName:    user.UserName,
		NickNameNew: user.NickNameNew,
		TiebaUID:    user.TiebaUID,
		GLevel:      user.GLevel,
		Gender:      user.Gender,
		Age:         user.Age,
		PostNum:     user.PostNum,
		AgreeNum:    user.AgreeNum,
		FanNum:      user.FanNum,
		FollowNum:   user.FollowNum,
		ForumNum:    user.ForumNum,
		Sign:        user.Sign,
		IP:          user.IP,
		Icons:       user.Icons,
		IsVIP:       user.IsVIP,
		IsGod:       user.IsGod,
		IsBlocked:   user.IsBlocked,
		PrivLike:    user.PrivLike,
		PrivReply:   user.PrivReply,
	}, nil
}

// GetHomepage 获取用户个人页信息。
//
// 参数:
//
//	id 用户id user_id / user_name / portrait 优先user_id
//	pn 页码
//
// 对应 Client.get_homepage。
func (c *Client) GetHomepage(ctx context.Context, id UserRef, pn int32) (profile.Homepage, error) {
	userID, err := c.resolveUserID(ctx, id)
	if err != nil {
		c.logCallError("get_homepage", err, id, pn)
		return profile.Homepage{}, err
	}

	c.tryInitWebsocket(ctx)

	var homepage profile.Homepage
	if c.wsCore.Status() == enums.WsStatusOpen {
		homepage, err = gethomepage.RequestWS(c.wsCore, userID, pn)
	} else {
		homepage, err = gethomepage.RequestHTTP(ctx, c.httpCore, userID, pn)
	}
	if err != nil {
		c.logCallError("get_homepage", err, id, pn)
		return profile.Homepage{}, err
	}
	return homepage, nil
}

// GetTabMap 获取分区名到分区 id 的映射。
//
// 参数:
//
//	ref 目标贴吧名或fid 优先贴吧名
//
// 对应 Client.get_tab_map。
func (c *Client) GetTabMap(ctx context.Context, ref ForumRef) (gettabmap.TabMap, error) {
	fname, err := c.fetchFNameOrFName(ctx, ref)
	if err != nil {
		c.logCallError("get_tab_map", err, ref)
		return gettabmap.TabMap{}, err
	}

	c.tryInitWebsocket(ctx)

	var tabMap gettabmap.TabMap
	if c.wsCore.Status() == enums.WsStatusOpen {
		tabMap, err = gettabmap.RequestWS(c.wsCore, fname)
	} else {
		tabMap, err = gettabmap.RequestHTTP(ctx, c.httpCore, fname)
	}
	if err != nil {
		c.logCallError("get_tab_map", err, ref)
		return gettabmap.TabMap{}, err
	}
	return tabMap, nil
}

// SendMsg 发送私信。
//
// 参数:
//
//	id 用户id user_id / user_name / portrait 优先user_id
//	content 发送内容
//
// 对应 Client.send_msg。该接口仅支持 websocket；返回的 msg id 会记录到消息 id 管理器中，
// 后续读取从该位置继续。
func (c *Client) SendMsg(ctx context.Context, id UserRef, content string) (exception.BoolResponse, error) {
	userID, err := c.resolveUserID(ctx, id)
	if err != nil {
		c.logCallError("send_msg", err, id, content)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.forceWebsocket(ctx); err != nil {
		c.logCallError("send_msg", err, id, content)
		return exception.BoolResponse{Err: err}, err
	}

	msgID, err := sendmsg.Request(c.wsCore, userID, content)
	if err != nil {
		c.logCallError("send_msg", err, id, content)
		return exception.BoolResponse{Err: err}, err
	}

	midManager := c.wsCore.MsgIDManager()
	midManager.UpdateMsgID(midManager.PrivGID, int(msgID))

	c.logCallSuccess("send_msg", id, content)
	return exception.BoolResponse{}, nil
}

// SetBlacklist 设置新版用户黑名单。
//
// 参数:
//
//	id 待设置黑名单的用户id user_id / user_name / portrait 优先user_id
//	btype 黑名单类型 默认全屏蔽
//
// 对应 Client.set_blacklist。
func (c *Client) SetBlacklist(ctx context.Context, id UserRef, btype enums.BlacklistType) (exception.BoolResponse, error) {
	userID, err := c.resolveUserID(ctx, id)
	if err != nil {
		c.logCallError("set_blacklist", err, id, btype)
		return exception.BoolResponse{Err: err}, err
	}

	c.tryInitWebsocket(ctx)

	if c.wsCore.Status() == enums.WsStatusOpen {
		err = setblacklist.RequestWS(c.wsCore, userID, btype)
	} else {
		err = setblacklist.RequestHTTP(ctx, c.httpCore, userID, btype)
	}
	if err != nil {
		c.logCallError("set_blacklist", err, id, btype)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("set_blacklist", id, btype)
	return exception.BoolResponse{}, nil
}

// GetAts 获取@信息。
//
// 参数:
//
//	pn 页码
//
// 对应 Client.get_ats。
func (c *Client) GetAts(ctx context.Context, pn int64) (getats.Ats, error) {
	ats, err := getats.Request(ctx, c.httpCore, pn)
	if err != nil {
		c.logCallError("get_ats", err)
		return getats.Ats{}, err
	}
	return ats, nil
}

// GetBlacklist 获取完整的新版用户黑名单列表。
//
// 对应 Client.get_blacklist。
func (c *Client) GetBlacklist(ctx context.Context) (getblacklist.BlacklistUsers, error) {
	users, err := getblacklist.Request(ctx, c.httpCore)
	if err != nil {
		c.logCallError("get_blacklist", err)
		return getblacklist.BlacklistUsers{}, err
	}
	return users, nil
}

// resolveUserIDOrSelf 把用户引用解析为数字 id，引用为空时使用本账号，
// 对应粉丝/关注类接口的 `id_ is None` 分支。
func (c *Client) resolveUserIDOrSelf(ctx context.Context, ref UserRef) (int64, error) {
	if ref.IsZero() {
		user, err := c.GetSelfInfo(ctx, enums.ReqUInfoUserID)
		if err != nil {
			return 0, err
		}
		return user.UserID, nil
	}
	return c.resolveUserID(ctx, ref)
}

// GetFans 获取粉丝列表。
//
// 参数:
//
//	id 用户id user_id / user_name / portrait 优先user_id 为空表示本账号
//	pn 页码
//
// 对应 Client.get_fans。
func (c *Client) GetFans(ctx context.Context, id UserRef, pn int64) (getfans.Fans, error) {
	userID, err := c.resolveUserIDOrSelf(ctx, id)
	if err != nil {
		c.logCallError("get_fans", err, id, pn)
		return getfans.Fans{}, err
	}
	fans, err := getfans.Request(ctx, c.httpCore, userID, pn)
	if err != nil {
		c.logCallError("get_fans", err, id, pn)
		return getfans.Fans{}, err
	}
	return fans, nil
}

// GetFollows 获取关注列表。
//
// 参数:
//
//	id 用户id user_id / user_name / portrait 优先user_id 为空表示本账号
//	pn 页码
//
// 对应 Client.get_follows。
func (c *Client) GetFollows(ctx context.Context, id UserRef, pn int64) (getfollows.Follows, error) {
	userID, err := c.resolveUserIDOrSelf(ctx, id)
	if err != nil {
		c.logCallError("get_follows", err, id, pn)
		return getfollows.Follows{}, err
	}
	follows, err := getfollows.Request(ctx, c.httpCore, userID, pn)
	if err != nil {
		c.logCallError("get_follows", err, id, pn)
		return getfollows.Follows{}, err
	}
	return follows, nil
}

// GetFollowForums 获取用户关注贴吧列表。
//
// 参数:
//
//	id 用户id user_id / user_name / portrait 优先user_id
//	pn 页码
//	rn 请求的条目数 Max to Inf
//
// 对应 Client.get_follow_forums。
func (c *Client) GetFollowForums(ctx context.Context, id UserRef, pn, rn int64) (getfollowforums.FollowForums, error) {
	userID, err := c.resolveUserID(ctx, id)
	if err != nil {
		c.logCallError("get_follow_forums", err, id, pn, rn)
		return getfollowforums.FollowForums{}, err
	}
	forums, err := getfollowforums.Request(ctx, c.httpCore, userID, pn, rn)
	if err != nil {
		c.logCallError("get_follow_forums", err, id, pn, rn)
		return getfollowforums.FollowForums{}, err
	}
	return forums, nil
}

// GetRoomlistByFID 获取某吧所有群聊。
//
// 参数:
//
//	fid 吧id
//
// 对应 Client.get_roomlist_by_fid。
func (c *Client) GetRoomlistByFID(ctx context.Context, fid int64) (getroomlistbyfid.RoomList, error) {
	roomList, err := getroomlistbyfid.Request(ctx, c.httpCore, fid)
	if err != nil {
		c.logCallError("get_roomlist_by_fid", err, fid)
		return getroomlistbyfid.RoomList{}, err
	}
	return roomList, nil
}

// GetStatistics 获取吧务后台中最近 24 天的统计数据。
//
// 参数:
//
//	ref 目标贴吧名或fid 优先fid
//
// 对应 Client.get_statistics。
func (c *Client) GetStatistics(ctx context.Context, ref ForumRef) (getstatistics.Statistics, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("get_statistics", err, ref)
		return getstatistics.Statistics{}, err
	}
	stats, err := getstatistics.Request(ctx, c.httpCore, fid)
	if err != nil {
		c.logCallError("get_statistics", err, ref)
		return getstatistics.Statistics{}, err
	}
	return stats, nil
}

// GetRecomStatus 获取大吧主推荐功能的月度配额状态。
//
// 参数:
//
//	ref 目标贴吧名或fid 优先fid
//
// 对应 Client.get_recom_status。
func (c *Client) GetRecomStatus(ctx context.Context, ref ForumRef) (getrecomstatus.RecomStatus, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("get_recom_status", err, ref)
		return getrecomstatus.RecomStatus{}, err
	}
	status, err := getrecomstatus.Request(ctx, c.httpCore, fid)
	if err != nil {
		c.logCallError("get_recom_status", err, ref)
		return getrecomstatus.RecomStatus{}, err
	}
	return status, nil
}

// GetUserForumInfo 获取用户在某吧内的信息。
//
// 参数:
//
//	ref 目标贴吧名或fid 优先fid
//	id 用户id user_id / user_name / portrait 优先portrait
//
// 对应 Client.get_user_forum_info。
func (c *Client) GetUserForumInfo(ctx context.Context, ref ForumRef, id UserRef) (getuserforuminfo.UserForumInfo, error) {
	if ref.FName == "" && ref.FID == 0 || id.IsZero() {
		logging.GetLogger().Warn().Msg("GetUserForumInfo: null input")
		return getuserforuminfo.UserForumInfo{}, nil
	}
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("get_user_forum_info", err, ref, id)
		return getuserforuminfo.UserForumInfo{}, err
	}
	portrait, err := c.resolvePortrait(ctx, id)
	if err != nil {
		c.logCallError("get_user_forum_info", err, ref, id)
		return getuserforuminfo.UserForumInfo{}, err
	}
	if portrait == "" {
		return getuserforuminfo.UserForumInfo{}, nil
	}
	info, err := getuserforuminfo.Request(ctx, c.httpCore, fid, portrait)
	if err != nil {
		c.logCallError("get_user_forum_info", err, ref, id)
		return getuserforuminfo.UserForumInfo{}, err
	}
	return info, nil
}

// SearchExact 贴吧搜索。
//
// 参数:
//
//	ref 查询的贴吧名或fid 优先贴吧名
//	query 查询文本
//	pn 页码
//	rn 请求的条目数
//	searchType 查询模式 默认查询全部
//	onlyThread 是否仅查询主题帖
//
// 对应 Client.search_exact。
func (c *Client) SearchExact(ctx context.Context, ref ForumRef, query string, pn, rn int64, searchType enums.SearchType, onlyThread bool) (searchexact.ExactSearches, error) {
	fname, err := c.fetchFNameOrFName(ctx, ref)
	if err != nil {
		c.logCallError("search_exact", err, ref, query, pn, rn, searchType, onlyThread)
		return searchexact.ExactSearches{}, err
	}
	searches, err := searchexact.Request(ctx, c.httpCore, fname, query, pn, rn, searchType, onlyThread)
	if err != nil {
		c.logCallError("search_exact", err, ref, query, pn, rn, searchType, onlyThread)
		return searchexact.ExactSearches{}, err
	}
	return searches, nil
}

// GetBawuPerm 获取指定吧务已分配的权限。
//
// 参数:
//
//	ref 目标贴吧名或fid 优先fid
//	id 用户id user_id / user_name / portrait 优先portrait
//
// 对应 Client.get_bawu_perm。
func (c *Client) GetBawuPerm(ctx context.Context, ref ForumRef, id UserRef) (getbawuperm.BawuPerm, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("get_bawu_perm", err, ref, id)
		return getbawuperm.BawuPerm{}, err
	}
	portrait, err := c.resolvePortrait(ctx, id)
	if err != nil {
		c.logCallError("get_bawu_perm", err, ref, id)
		return getbawuperm.BawuPerm{}, err
	}
	perm, err := getbawuperm.Request(ctx, c.httpCore, fid, portrait)
	if err != nil {
		c.logCallError("get_bawu_perm", err, ref, id)
		return getbawuperm.BawuPerm{}, err
	}
	return perm, nil
}

// GetFollowForumsPc 获取用户关注贴吧列表。
//
// 参数:
//
//	id 用户id user_id / user_name / portrait 优先portrait
//	pn 页码
//	rn 请求的条目数 Max to Inf
//
// 对应 Client.get_follow_forums_pc。
func (c *Client) GetFollowForumsPc(ctx context.Context, id UserRef, pn, rn int64) (getfollowforumspc.PcFollowForums, error) {
	portrait, err := c.resolvePortrait(ctx, id)
	if err != nil {
		c.logCallError("get_follow_forums_pc", err, id, pn, rn)
		return getfollowforumspc.PcFollowForums{}, err
	}
	forums, err := getfollowforumspc.Request(ctx, c.httpCore, portrait, pn, rn)
	if err != nil {
		c.logCallError("get_follow_forums_pc", err, id, pn, rn)
		return getfollowforumspc.PcFollowForums{}, err
	}
	return forums, nil
}

// GetSelfFollowForums 获取本账号关注贴吧列表。
//
// 参数:
//
//	pn 页码
//	rn 请求的条目数 Max to 200
//
// 对应 Client.get_self_follow_forums。本接口需要 STOKEN。
func (c *Client) GetSelfFollowForums(ctx context.Context, pn, rn int64) (getselffollowforums.SelfFollowForums, error) {
	forums, err := getselffollowforums.Request(ctx, c.httpCore, pn, rn)
	if err != nil {
		c.logCallError("get_self_follow_forums", err)
		return getselffollowforums.SelfFollowForums{}, err
	}
	return forums, nil
}

// GetUnblockAppeals 获取吧务后台申诉请求列表。
//
// 参数:
//
//	ref 目标贴吧的贴吧名或fid 优先fid
//	pn 页码
//	rn 请求的条目数 Max to 50
//
// 对应 Client.get_unblock_appeals。
func (c *Client) GetUnblockAppeals(ctx context.Context, ref ForumRef, pn, rn int64) (getunblockappeals.Appeals, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("get_unblock_appeals", err, ref, pn, rn)
		return getunblockappeals.Appeals{}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("get_unblock_appeals", err, ref, pn, rn)
		return getunblockappeals.Appeals{}, err
	}
	appeals, err := getunblockappeals.Request(ctx, c.httpCore, fid, pn, rn)
	if err != nil {
		c.logCallError("get_unblock_appeals", err, ref, pn, rn)
		return getunblockappeals.Appeals{}, err
	}
	return appeals, nil
}

// SearchGlobal 全吧搜索，不限定贴吧的全站主题帖关键词搜索。
//
// 参数:
//
//	word 查询文本
//	pn 页码
//	rn 请求的条目数
//	sort 排序方式
//
// 该接口为 PC 网页端搜索接口(逆向所得 非官方开放 API)，走 `subapp_type=pc` 网页端签名通道，复用当前账号的 Cookie(BDUSS) 鉴权。
// 不同于 `search_exact` 所用的 App 表单签名协议，其稳定性与频控策略未经长期验证，请自行控制调用频率。
// 该接口存在与请求参数无关的服务端间歇性错误(如 `TiebaServerError` 300003)，失败会体现在返回值 `.err`，建议调用方按需重试。
// 仅支持搜索主题帖，实测该接口的评论/楼中楼搜索(tt=3)不会生效，服务端会原样返回主题帖结果。
// 若需要某个主题帖下的评论，请在拿到 `tid` 后使用 `GetPosts` 单独查询。
//
// 对应 Client.search_global。
func (c *Client) SearchGlobal(ctx context.Context, word string, pn, rn, sort int64) (searchglobal.GlobalSearches, error) {
	searches, err := searchglobal.Request(ctx, c.httpCore, word, pn, rn, sort)
	if err != nil {
		c.logCallError("search_global", err, word)
		return searchglobal.GlobalSearches{}, err
	}
	return searches, nil
}

// GetRecovers 获取吧务后台待恢复帖子列表。
//
// 参数:
//
//	ref 目标贴吧的贴吧名或fid 优先fid
//	pn 页码
//	rn 请求的条目数 Max to 50
//	id 用于查询的被删帖用户的id user_id / user_name / portrait 优先user_id
//
// 对应 Client.get_recovers。
func (c *Client) GetRecovers(ctx context.Context, ref ForumRef, pn, rn int64, id UserRef) (getrecovers.Recovers, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("get_recovers", err, ref, pn, rn, id)
		return getrecovers.Recovers{}, err
	}
	var userID int64
	if !id.IsZero() {
		if userID, err = c.resolveUserID(ctx, id); err != nil {
			c.logCallError("get_recovers", err, ref, pn, rn, id)
			return getrecovers.Recovers{}, err
		}
	}
	recovers, err := getrecovers.Request(ctx, c.httpCore, fid, userID, pn, rn)
	if err != nil {
		c.logCallError("get_recovers", err, ref, pn, rn, id)
		return getrecovers.Recovers{}, err
	}
	return recovers, nil
}

// mergeUserInto 把 src 的非零字段复制到 dst，对应 Python 客户端的
// `self._user |= user` 合并。
func mergeUserInto(dst *classdef.UserInfo, src classdef.UserInfo) {
	if src.UserID != 0 {
		dst.UserID = src.UserID
	}
	if src.Portrait != "" {
		dst.Portrait = src.Portrait
	}
	if src.UserName != "" {
		dst.UserName = src.UserName
	}
	if src.NickNameOld != "" {
		dst.NickNameOld = src.NickNameOld
	}
	if src.NickNameNew != "" {
		dst.NickNameNew = src.NickNameNew
	}
	if src.TiebaUID != 0 {
		dst.TiebaUID = src.TiebaUID
	}
	if src.GLevel != 0 {
		dst.GLevel = src.GLevel
	}
	if src.Gender != enums.GenderUnknown {
		dst.Gender = src.Gender
	}
	if src.Age != 0 {
		dst.Age = src.Age
	}
	if src.PostNum != 0 {
		dst.PostNum = src.PostNum
	}
	if src.AgreeNum != 0 {
		dst.AgreeNum = src.AgreeNum
	}
	if src.FanNum != 0 {
		dst.FanNum = src.FanNum
	}
	if src.FollowNum != 0 {
		dst.FollowNum = src.FollowNum
	}
	if src.ForumNum != 0 {
		dst.ForumNum = src.ForumNum
	}
	if src.Sign != "" {
		dst.Sign = src.Sign
	}
	if src.IP != "" {
		dst.IP = src.IP
	}
	if len(src.Icons) > 0 {
		dst.Icons = src.Icons
	}
	if src.IsVIP {
		dst.IsVIP = src.IsVIP
	}
	if src.IsGod {
		dst.IsGod = src.IsGod
	}
	if src.IsBlocked {
		dst.IsBlocked = src.IsBlocked
	}
	if src.UK != 0 {
		dst.UK = src.UK
	}
	if src.BDUK != "" {
		dst.BDUK = src.BDUK
	}
	if src.TriggerID != 0 {
		dst.TriggerID = src.TriggerID
	}
	if src.PrivLike != enums.PrivLikeUnknown {
		dst.PrivLike = src.PrivLike
	}
	if src.PrivReply != enums.PrivReplyUnknown {
		dst.PrivReply = src.PrivReply
	}
}

// initSelfinfoInitNickname 填充 user_name / nick_name_old / tieba_uid。
//
// initSelfinfoInitNickname 对应私有方法 Client.__init_selfinfo_initNickname。
func (c *Client) initSelfinfoInitNickname(ctx context.Context) error {
	user, err := getselfinfoinitnickname.Request(ctx, c.httpCore)
	if err != nil {
		c.logCallError("init_selfinfo_initNickname", err)
		return err
	}
	c.mu.Lock()
	mergeUserInto(&c.user, classdef.UserInfo{
		UserName:    user.UserName,
		NickNameOld: user.NickNameOld,
		TiebaUID:    user.TiebaUID,
	})
	c.mu.Unlock()
	return nil
}

// GetSelfInfo 获取本账号信息。
//
// 参数:
//
//	require 指示需要获取的字段
//
// 返回账号的缓存信息，缺失字段按需补齐，对应 Client.get_self_info。
func (c *Client) GetSelfInfo(ctx context.Context, require enums.ReqUInfo) (classdef.UserInfo, error) {
	if c.user.UserID == 0 {
		if require&enums.ReqUInfoBasic != 0 {
			if err := c.Login(ctx); err != nil {
				return c.user, err
			}
		}
	}
	if c.user.TiebaUID == 0 {
		if require == enums.ReqUInfoAll {
			user, err := c.getUinfoProfile(ctx, ByUserID(c.user.UserID))
			if err != nil {
				return c.user, err
			}
			c.mu.Lock()
			mergeUserInto(&c.user, user)
			c.mu.Unlock()
		} else if require&(enums.ReqUInfoTiebaUID|enums.ReqUInfoNickName) != 0 {
			if err := c.initSelfinfoInitNickname(ctx); err != nil {
				return c.user, err
			}
		}
	}
	return c.user, nil
}

// GetBawuBlacklist 获取吧务后台黑名单列表。
//
// 参数:
//
//	ref 目标贴吧的贴吧名或fid 优先贴吧名
//	pn 页码
//
// 对应 Client.get_bawu_blacklist。本接口需要 STOKEN。
func (c *Client) GetBawuBlacklist(ctx context.Context, ref ForumRef, pn int64) (getbawublacklist.BawuBlacklistUsers, error) {
	fname, err := c.fetchFNameOrFName(ctx, ref)
	if err != nil {
		c.logCallError("get_bawu_blacklist", err, ref, pn)
		return getbawublacklist.BawuBlacklistUsers{}, err
	}
	users, err := getbawublacklist.Request(ctx, c.httpCore, fname, pn)
	if err != nil {
		c.logCallError("get_bawu_blacklist", err, ref, pn)
		return getbawublacklist.BawuBlacklistUsers{}, err
	}
	return users, nil
}

// GetBawuMemberlist 获取吧务后台吧会员列表。
//
// 参数:
//
//	ref 目标贴吧名或fid 优先贴吧名
//	pn 页码
//	searchValue 搜索用户名
//
// 对应 Client.get_bawu_memberlist。本接口需要 STOKEN。
func (c *Client) GetBawuMemberlist(ctx context.Context, ref ForumRef, pn int64, searchValue string) (getbawumemberlist.BawuListMemberUsers, error) {
	fname, err := c.fetchFNameOrFName(ctx, ref)
	if err != nil {
		c.logCallError("get_bawu_memberlist", err, ref, pn, searchValue)
		return getbawumemberlist.BawuListMemberUsers{}, err
	}
	users, err := getbawumemberlist.Request(ctx, c.httpCore, fname, pn, searchValue)
	if err != nil {
		c.logCallError("get_bawu_memberlist", err, ref, pn, searchValue)
		return getbawumemberlist.BawuListMemberUsers{}, err
	}
	return users, nil
}

// GetBawuPostlogs 获取吧务后台帖子管理日志表。
//
// 参数:
//
//	ref 目标贴吧名或fid 优先贴吧名
//	pn 页码
//	searchValue 搜索关键字
//	searchType 搜索类型
//	startDT 搜索的起始时间(含)
//	endDT 搜索的结束时间(含)
//	opType 搜索操作类型
//
// 对应 Client.get_bawu_postlogs。本接口需要 STOKEN。
func (c *Client) GetBawuPostlogs(
	ctx context.Context, ref ForumRef, pn int64, searchValue string, searchType enums.BawuSearchType,
	startDT, endDT *time.Time, opType int64,
) (getbawupostlogs.BawuPostLogs, error) {
	fname, err := c.fetchFNameOrFName(ctx, ref)
	if err != nil {
		c.logCallError("get_bawu_postlogs", err, ref, pn, searchValue, searchType, startDT, endDT, opType)
		return getbawupostlogs.BawuPostLogs{}, err
	}
	logs, err := getbawupostlogs.Request(ctx, c.httpCore, fname, pn, searchValue, searchType, startDT, endDT, opType)
	if err != nil {
		c.logCallError("get_bawu_postlogs", err, ref, pn, searchValue, searchType, startDT, endDT, opType)
		return getbawupostlogs.BawuPostLogs{}, err
	}
	return logs, nil
}

// GetBawuUserlogs 获取吧务用户管理日志表。
//
// 参数:
//
//	ref 目标贴吧名或fid 优先贴吧名
//	pn 页码
//	searchValue 搜索关键字
//	searchType 搜索类型
//	startDT 搜索的起始时间(含)
//	endDT 搜索的结束时间(含)
//	opType 搜索操作类型
//
// 对应 Client.get_bawu_userlogs。本接口需要 STOKEN。
func (c *Client) GetBawuUserlogs(
	ctx context.Context, ref ForumRef, pn int64, searchValue string, searchType enums.BawuSearchType,
	startDT, endDT *time.Time, opType int64,
) (getbawuuserlogs.BawuUserLogs, error) {
	fname, err := c.fetchFNameOrFName(ctx, ref)
	if err != nil {
		c.logCallError("get_bawu_userlogs", err, ref, pn, searchValue, searchType, startDT, endDT, opType)
		return getbawuuserlogs.BawuUserLogs{}, err
	}
	logs, err := getbawuuserlogs.Request(ctx, c.httpCore, fname, pn, searchValue, searchType, startDT, endDT, opType)
	if err != nil {
		c.logCallError("get_bawu_userlogs", err, ref, pn, searchValue, searchType, startDT, endDT, opType)
		return getbawuuserlogs.BawuUserLogs{}, err
	}
	return logs, nil
}

// GetMemberUsers 获取最新关注用户列表。
//
// 参数:
//
//	ref 目标贴吧名或fid 优先贴吧名
//	pn 页码
//
// 对应 Client.get_member_users。本接口需要 STOKEN。
func (c *Client) GetMemberUsers(ctx context.Context, ref ForumRef, pn int64) (getmemberusers.MemberUsers, error) {
	fname, err := c.fetchFNameOrFName(ctx, ref)
	if err != nil {
		c.logCallError("get_member_users", err, ref, pn)
		return getmemberusers.MemberUsers{}, err
	}
	users, err := getmemberusers.Request(ctx, c.httpCore, fname, pn)
	if err != nil {
		c.logCallError("get_member_users", err, ref, pn)
		return getmemberusers.MemberUsers{}, err
	}
	return users, nil
}

// GetRankForums 获取吧签到排行表。
//
// 参数:
//
//	ref 目标贴吧名或fid 优先贴吧名
//	pn 页码
//	rankType 榜单类型 默认为周榜
//
// 对应 Client.get_rank_forums。
func (c *Client) GetRankForums(ctx context.Context, ref ForumRef, pn int64, rankType enums.RankForumType) (getrankforums.RankForums, error) {
	fname, err := c.fetchFNameOrFName(ctx, ref)
	if err != nil {
		c.logCallError("get_rank_forums", err, ref, pn, rankType)
		return getrankforums.RankForums{}, err
	}
	forums, err := getrankforums.Request(ctx, c.httpCore, fname, pn, rankType)
	if err != nil {
		c.logCallError("get_rank_forums", err, ref, pn, rankType)
		return getrankforums.RankForums{}, err
	}
	return forums, nil
}

// GetRankUsers 获取等级排行榜用户列表。
//
// 参数:
//
//	ref 目标贴吧名或fid 优先贴吧名
//	pn 页码
//
// 对应 Client.get_rank_users。
func (c *Client) GetRankUsers(ctx context.Context, ref ForumRef, pn int64) (getrankusers.RankUsers, error) {
	fname, err := c.fetchFNameOrFName(ctx, ref)
	if err != nil {
		c.logCallError("get_rank_users", err, ref, pn)
		return getrankusers.RankUsers{}, err
	}
	users, err := getrankusers.Request(ctx, c.httpCore, fname, pn)
	if err != nil {
		c.logCallError("get_rank_users", err, ref, pn)
		return getrankusers.RankUsers{}, err
	}
	return users, nil
}

// GetBlocks 获获取吧务后台待解封用户列表。
//
// 参数:
//
//	ref 目标贴吧的贴吧名或fid 优先fid
//	name 通过被封禁用户的用户名/昵称查询 默认为空即查询全部
//	pn 页码
//
// 对应 Client.get_blocks。
func (c *Client) GetBlocks(ctx context.Context, ref ForumRef, name string, pn int64) (getblocks.Blocks, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("get_blocks", err, ref, name, pn)
		return getblocks.Blocks{}, err
	}
	blocks, err := getblocks.Request(ctx, c.httpCore, fid, name, pn)
	if err != nil {
		c.logCallError("get_blocks", err, ref, name, pn)
		return getblocks.Blocks{}, err
	}
	return blocks, nil
}

// GetImage 从链接获取静态图像。
//
// 参数:
//
//	imgURL 图像链接
//
// 对应 Client.get_image。
func (c *Client) GetImage(ctx context.Context, imgURL string) (getimages.Image, error) {
	u, err := url.Parse(imgURL)
	if err != nil {
		c.logCallError("get_image", err)
		return getimages.Image{Err: err}, err
	}
	img, err := getimages.Request(ctx, c.httpCore, u)
	if err != nil {
		c.logCallError("get_image", err)
		return getimages.Image{Err: err}, err
	}
	return img, nil
}

// GetImageBytes 从链接获取静态图像的原始字节流。
//
// 参数:
//
//	imgURL 图像链接
//
// 对应 Client.get_image_bytes。
func (c *Client) GetImageBytes(ctx context.Context, imgURL string) (getimages.ImageBytes, error) {
	u, err := url.Parse(imgURL)
	if err != nil {
		c.logCallError("get_image_bytes", err)
		return getimages.ImageBytes{Err: err}, err
	}
	b, err := getimages.RequestBytes(ctx, c.httpCore, u)
	if err != nil {
		c.logCallError("get_image_bytes", err)
		return getimages.ImageBytes{Err: err}, err
	}
	return b, nil
}

// Disagree 点踩主题帖或回复。
//
// 参数:
//
//	tid 待点踩的主题帖或回复所在的主题帖的tid
//	pid 待点踩的回复pid
//	isComment pid是否指向楼中楼
//
// 对应 Client.disagree。
func (c *Client) Disagree(ctx context.Context, tid, pid int64, isComment bool) (exception.BoolResponse, error) {
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("disagree", err, tid, pid, isComment)
		return exception.BoolResponse{Err: err}, err
	}
	if err := agree.Request(ctx, c.httpCore, tid, pid, isComment, true, false); err != nil {
		c.logCallError("disagree", err, tid, pid, isComment)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("disagree", tid, pid, isComment)
	return exception.BoolResponse{}, nil
}

// Unagree 取消点赞主题帖或回复。
//
// 参数:
//
//	tid 待取消点赞的主题帖或回复所在的主题帖的tid
//	pid 待取消点赞的回复pid
//	isComment pid是否指向楼中楼
//
// 对应 Client.unagree。
func (c *Client) Unagree(ctx context.Context, tid, pid int64, isComment bool) (exception.BoolResponse, error) {
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("unagree", err, tid, pid, isComment)
		return exception.BoolResponse{Err: err}, err
	}
	if err := agree.Request(ctx, c.httpCore, tid, pid, isComment, false, true); err != nil {
		c.logCallError("unagree", err, tid, pid, isComment)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("unagree", tid, pid, isComment)
	return exception.BoolResponse{}, nil
}

// Undisagree 取消点踩主题帖或回复。
//
// 参数:
//
//	tid 待取消点踩的主题帖或回复所在的主题帖的tid
//	pid 待取消点踩的回复pid
//	isComment pid是否指向楼中楼
//
// 对应 Client.undisagree。
func (c *Client) Undisagree(ctx context.Context, tid, pid int64, isComment bool) (exception.BoolResponse, error) {
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("undisagree", err, tid, pid, isComment)
		return exception.BoolResponse{Err: err}, err
	}
	if err := agree.Request(ctx, c.httpCore, tid, pid, isComment, true, true); err != nil {
		c.logCallError("undisagree", err, tid, pid, isComment)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("undisagree", tid, pid, isComment)
	return exception.BoolResponse{}, nil
}

// GetPortrait 获取用户头像。
//
// 参数:
//
//	id 用户id user_id / user_name / portrait 优先portrait
//	size 获取头像的大小 s为55x55 m为110x110 l为原图
//
// 对应 Client.get_portrait。
func (c *Client) GetPortrait(ctx context.Context, id UserRef, size string) (getimages.Image, error) {
	portrait, err := c.resolvePortrait(ctx, id)
	if err != nil {
		c.logCallError("get_portrait", err, id, size)
		return getimages.Image{Err: err}, err
	}

	var path string
	switch size {
	case "s":
		path = "n"
	case "m":
		path = ""
	case "l":
		path = "h"
	default:
		logging.GetLogger().Warn().Str("size", size).Msg("get_portrait: invalid size")
		return getimages.Image{}, nil
	}

	u, err := url.Parse("http://tb.himg.baidu.com/sys/portrait" + path + "/item/" + portrait)
	if err != nil {
		c.logCallError("get_portrait", err, id, size)
		return getimages.Image{Err: err}, err
	}
	img, err := getimages.Request(ctx, c.httpCore, u)
	if err != nil {
		c.logCallError("get_portrait", err, id, size)
		return getimages.Image{Err: err}, err
	}
	return img, nil
}

// GetSelfPosts 获取当前用户发布的回复列表。
//
// 参数:
//
//	pn 页码
//	rn 请求的条目数 Max to 74
//
// 对应 Client.get_self_posts。
func (c *Client) GetSelfPosts(ctx context.Context, pn, rn int32) (getusercontents.UserPostss, error) {
	c.tryInitWebsocket(ctx)

	user, err := c.GetSelfInfo(ctx, enums.ReqUInfoUserID)
	if err != nil {
		c.logCallError("get_self_posts", err)
		return getusercontents.UserPostss{}, err
	}

	if c.wsCore.Status() == enums.WsStatusOpen {
		return getusercontentsposts.RequestWS(c.wsCore, user.UserID, pn, rn, consts.LatestVersion)
	}
	return getusercontentsposts.RequestHTTP(ctx, c.httpCore, user.UserID, pn, rn, consts.LatestVersion)
}

// GetSelfThreads 获取当前用户发布的主题帖列表。
//
// 参数:
//
//	pn 页码
//	publicOnly 是否仅获取公开主题帖 该选项在获取他人主题帖时无效
//
// 对应 Client.get_self_threads。
func (c *Client) GetSelfThreads(ctx context.Context, pn int32, publicOnly bool) (getusercontents.UserThreads, error) {
	c.tryInitWebsocket(ctx)

	user, err := c.GetSelfInfo(ctx, enums.ReqUInfoUserID)
	if err != nil {
		c.logCallError("get_self_threads", err)
		return getusercontents.UserThreads{}, err
	}

	if c.wsCore.Status() == enums.WsStatusOpen {
		return getusercontentsthreads.RequestWS(c.wsCore, user.UserID, pn, publicOnly)
	}
	return getusercontentsthreads.RequestHTTP(ctx, c.httpCore, user.UserID, pn, publicOnly)
}

// GetUserPosts 获取用户发布的回复列表。
//
// 参数:
//
//	id 用户id user_id / user_name / portrait 优先user_id
//	pn 页码
//	rn 请求的条目数 Max to 74
//
// 对应 Client.get_user_posts。
func (c *Client) GetUserPosts(ctx context.Context, id UserRef, pn, rn int32) (getusercontents.UserPostss, error) {
	userID, err := c.resolveUserID(ctx, id)
	if err != nil {
		c.logCallError("get_user_posts", err)
		return getusercontents.UserPostss{}, err
	}

	const userPostsVersion = "8"
	if c.wsCore.Status() == enums.WsStatusOpen {
		return getusercontentsposts.RequestWS(c.wsCore, userID, pn, rn, userPostsVersion)
	}
	return getusercontentsposts.RequestHTTP(ctx, c.httpCore, userID, pn, rn, userPostsVersion)
}

// GetUserPostsPc 获取用户发布的回复列表。
//
// 参数:
//
//	id 用户id user_id / user_name / portrait 优先portrait
//	pn 页码
//	rn 请求的条目数 Max to 74
//
// 对应 Client.get_user_posts_pc。
func (c *Client) GetUserPostsPc(ctx context.Context, id UserRef, pn, rn int64) (getusercontentpc.PcUserPosts, error) {
	portrait, err := c.resolvePortrait(ctx, id)
	if err != nil {
		c.logCallError("get_user_posts_pc", err, id, pn, rn)
		return getusercontentpc.PcUserPosts{}, err
	}
	posts, err := getusercontentpc.Request(ctx, c.httpCore, portrait, pn, rn)
	if err != nil {
		c.logCallError("get_user_posts_pc", err, id, pn, rn)
		return getusercontentpc.PcUserPosts{}, err
	}
	return posts, nil
}

// GetUserThreads 获取用户发布的主题帖列表。
//
// 参数:
//
//	id 用户id user_id / user_name / portrait 优先user_id
//	pn 页码
//
// 对应 Client.get_user_threads。
func (c *Client) GetUserThreads(ctx context.Context, id UserRef, pn int32) (getusercontents.UserThreads, error) {
	c.tryInitWebsocket(ctx)

	userID, err := c.resolveUserID(ctx, id)
	if err != nil {
		c.logCallError("get_user_threads", err)
		return getusercontents.UserThreads{}, err
	}

	if c.wsCore.Status() == enums.WsStatusOpen {
		return getusercontentsthreads.RequestWS(c.wsCore, userID, pn, false)
	}
	return getusercontentsthreads.RequestHTTP(ctx, c.httpCore, userID, pn, false)
}

// Hash2Image 通过百度图库 hash 获取静态图像。
//
// 参数:
//
//	rawHash 百度图库hash
//	size 获取图像的大小 s为宽720 m为宽960 l为原图
//
// 对应 Client.hash2image。
func (c *Client) Hash2Image(ctx context.Context, rawHash, size string) (getimages.Image, error) {
	var rawURL string
	switch size {
	case "s":
		rawURL = "http://imgsrc.baidu.com/forum/w=720;q=60;g=0/sign=__/" + rawHash + ".jpg"
	case "m":
		rawURL = "http://imgsrc.baidu.com/forum/w=960;q=60;g=0/sign=__/" + rawHash + ".jpg"
	case "l":
		rawURL = "http://imgsrc.baidu.com/forum/pic/item/" + rawHash + ".jpg"
	default:
		logging.GetLogger().Warn().Str("size", size).Msg("hash2image: invalid size")
		return getimages.Image{}, nil
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		c.logCallError("hash2image", err, rawHash, size)
		return getimages.Image{Err: err}, err
	}
	img, err := getimages.Request(ctx, c.httpCore, u)
	if err != nil {
		c.logCallError("hash2image", err, rawHash, size)
		return getimages.Image{Err: err}, err
	}
	return img, nil
}

// HideThread 屏蔽主题帖。
//
// 参数:
//
//	ref 帖子所在贴吧的贴吧名或fid 优先fid
//	tid 待屏蔽的主题帖tid
//
// 对应 Client.hide_thread。
func (c *Client) HideThread(ctx context.Context, ref ForumRef, tid int64) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("hide_thread", err, ref, tid)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("hide_thread", err, ref, tid)
		return exception.BoolResponse{Err: err}, err
	}
	if err := delthread.Request(ctx, c.httpCore, fid, tid, true); err != nil {
		c.logCallError("hide_thread", err, ref, tid)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("hide_thread", ref, tid)
	return exception.BoolResponse{}, nil
}

// UnhideThread 解除主题帖屏蔽。
//
// 参数:
//
//	ref 帖子所在贴吧的贴吧名或fid 优先fid
//	tid 待解除屏蔽的主题帖tid
//
// 对应 Client.unhide_thread。
func (c *Client) UnhideThread(ctx context.Context, ref ForumRef, tid int64) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("unhide_thread", err, ref, tid)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("unhide_thread", err, ref, tid)
		return exception.BoolResponse{Err: err}, err
	}
	if err := recover.Request(ctx, c.httpCore, fid, tid, 0, true); err != nil {
		c.logCallError("unhide_thread", err, ref, tid)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("unhide_thread", ref, tid)
	return exception.BoolResponse{}, nil
}

// SetThreadPrivate 隐藏主题帖。
//
// 参数:
//
//	ref 主题帖所在贴吧的贴吧名或fid 优先fid
//	tid 主题帖tid
//	pid 主题帖pid
//
// 对应 Client.set_thread_private。
func (c *Client) SetThreadPrivate(ctx context.Context, ref ForumRef, tid, pid int64) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("set_thread_private", err, ref, tid, pid)
		return exception.BoolResponse{Err: err}, err
	}
	if err := setthreadprivacy.Request(ctx, c.httpCore, fid, tid, pid, true); err != nil {
		c.logCallError("set_thread_private", err, ref, tid, pid)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("set_thread_private", ref, tid, pid)
	return exception.BoolResponse{}, nil
}

// SetThreadPublic 公开主题帖。
//
// 参数:
//
//	ref 主题帖所在贴吧的贴吧名或fid 优先fid
//	tid 主题帖tid
//	pid 主题帖pid
//
// 对应 Client.set_thread_public。
func (c *Client) SetThreadPublic(ctx context.Context, ref ForumRef, tid, pid int64) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("set_thread_public", err, ref, tid, pid)
		return exception.BoolResponse{Err: err}, err
	}
	if err := setthreadprivacy.Request(ctx, c.httpCore, fid, tid, pid, false); err != nil {
		c.logCallError("set_thread_public", err, ref, tid, pid)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("set_thread_public", ref, tid, pid)
	return exception.BoolResponse{}, nil
}

// RecoverPost 恢复回复。
//
// 参数:
//
//	ref 帖子所在贴吧的贴吧名或fid 优先fid
//	pid 待恢复的回复pid
//
// 对应 Client.recover_post。
func (c *Client) RecoverPost(ctx context.Context, ref ForumRef, pid int64) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("recover_post", err, ref, pid)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("recover_post", err, ref, pid)
		return exception.BoolResponse{Err: err}, err
	}
	if err := recover.Request(ctx, c.httpCore, fid, 0, pid, false); err != nil {
		c.logCallError("recover_post", err, ref, pid)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("recover_post", ref, pid)
	return exception.BoolResponse{}, nil
}

// RecoverThread 恢复主题帖。
//
// 参数:
//
//	ref 帖子所在贴吧的贴吧名或fid 优先fid
//	tid 待恢复的主题帖tid
//
// 对应 Client.recover_thread。
func (c *Client) RecoverThread(ctx context.Context, ref ForumRef, tid int64) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("recover_thread", err, ref, tid)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("recover_thread", err, ref, tid)
		return exception.BoolResponse{Err: err}, err
	}
	if err := recover.Request(ctx, c.httpCore, fid, tid, 0, false); err != nil {
		c.logCallError("recover_thread", err, ref, tid)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("recover_thread", ref, tid)
	return exception.BoolResponse{}, nil
}

// Untop 撤销置顶主题帖。
//
// 参数:
//
//	ref 帖子所在贴吧的贴吧名或fid
//	tid 待撤销置顶的主题帖tid
//	isVIP 是否会员置顶
//
// 对应 Client.untop。
func (c *Client) Untop(ctx context.Context, ref ForumRef, tid int64, isVIP bool) (exception.BoolResponse, error) {
	fname, fid, err := c.resolveForumBoth(ctx, ref)
	if err != nil {
		c.logCallError("untop", err, ref, tid, isVIP)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("untop", err, ref, tid, isVIP)
		return exception.BoolResponse{Err: err}, err
	}
	if err := top.Request(ctx, c.httpCore, fname, fid, tid, isVIP, false); err != nil {
		c.logCallError("untop", err, ref, tid, isVIP)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("untop", ref, tid, isVIP)
	return exception.BoolResponse{}, nil
}

// JoinChatroom 加入聊天室。
//
// 参数:
//
//	roomID 房间id
//
// 对应 Client.join_chatroom。
func (c *Client) JoinChatroom(ctx context.Context, roomID int64) (exception.BoolResponse, error) {
	if c.user.UserID == 0 {
		if _, err := c.GetSelfInfo(ctx, enums.ReqUInfoAll); err != nil {
			c.logCallError("join_chatroom", err, roomID)
			return exception.BoolResponse{Err: err}, err
		}
	}
	if err := c.initBLCP(ctx); err != nil {
		c.logCallError("join_chatroom", err, roomID)
		return exception.BoolResponse{Err: err}, err
	}

	if _, err := c.blcpCore.JoinChatRoom(ctx, roomID); err != nil {
		err = fmt.Errorf("aiotieba: 加入房间失败: %w", err)
		c.logCallError("join_chatroom", err, roomID)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// initBLCP 让 BLCP 会话进入已登录状态，对应私有方法 Client._init_blcp。
func (c *Client) initBLCP(ctx context.Context) error {
	if c.blcpCore.Status() == -1 {
		if err := c.blcpCore.Connect(ctx); err != nil {
			c.logCallError("init_blcp", err)
			return err
		}
	}
	if c.blcpCore.Status() == 0 {
		if err := c.blcpCore.Login(ctx); err != nil {
			c.logCallError("init_blcp", err)
			return err
		}
	}
	if c.blcpCore.Status() != 1 {
		err := errors.New("aiotieba: BLCP 登录失败")
		c.logCallError("init_blcp", err)
		return err
	}
	c.logCallSuccess("init_blcp")
	return nil
}

// SendChatroomMsg 向吧群发送信息，仅限简单文本。如需要@他人需要指定 atUserIDs，如需与 bot 交互需要指定 atUserIDs 和 robotc。
//
// 参数:
//
//	chatroomID 聊天室id
//	fid 吧id
//	text 待发送内容
//	atUserIDs 需要@的人的user_id列表
//	robotc 机器人指令id。机器人靠此分辨指令，而非text内容。
//
// 对应 Client.send_chatroom_msg。缓存的自身信息（c.user）必须包含 user_id 与 portrait，
// Python 客户端通过 get_self_info 填充这两项。
func (c *Client) SendChatroomMsg(ctx context.Context, chatroomID, fid int64, text string, atUserIDs []int64, robotc int64) (exception.BoolResponse, error) {
	if c.user.UserID == 0 || c.user.Portrait == "" {
		err := errors.New("aiotieba: 自账号信息未加载，请先调用 GetSelfInfo 或 Login")
		c.logCallError("send_chatroom_msg", err, chatroomID, fid, text, atUserIDs, robotc)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.initBLCP(ctx); err != nil {
		c.logCallError("send_chatroom_msg", err, chatroomID, fid, text, atUserIDs, robotc)
		return exception.BoolResponse{Err: err}, err
	}

	levelInfo, err := getforumlevel.RequestHTTP(ctx, c.httpCore, fid)
	if err != nil {
		c.logCallError("send_chatroom_msg", err, chatroomID, fid, text, atUserIDs, robotc)
		return exception.BoolResponse{Err: err}, err
	}

	// 解析 @ 目标，对应 atdata 的构造过程。
	atdata := []map[string]any{}
	for i, userID := range atUserIDs {
		user, err := c.getUinfoProfile(ctx, ByUserID(userID))
		if err != nil {
			c.logCallError("send_chatroom_msg", err, chatroomID, fid, text, atUserIDs, robotc)
			return exception.BoolResponse{Err: err}, err
		}
		if user.Portrait == "" || user.NickName() == "" {
			if user, err = c.getUinfoProfile(ctx, ByUserID(userID)); err != nil {
				c.logCallError("send_chatroom_msg", err, chatroomID, fid, text, atUserIDs, robotc)
				return exception.BoolResponse{Err: err}, err
			}
		}
		atdata = append(atdata, map[string]any{
			"at_type":     "user",
			"at_baidu_uk": sendchatroommsg.BDUKFromUserID(strconv.FormatInt(userID, 10)),
			"at_name":     user.NickName(),
			"at_portrait": user.Portrait,
			"position":    strconv.Itoa(i),
		})
	}

	err = sendchatroommsg.Request(
		c.blcpCore,
		chatroomID,
		c.user.UK,
		c.user.UserID,
		c.user.TriggerID,
		c.user.NickName(),
		c.user.Portrait,
		text,
		fid,
		levelInfo.UserLevel,
		c.user.IsVIP,
		int64(c.user.GLevel),
		atdata,
		robotc,
	)
	if err != nil {
		c.logCallError("send_chatroom_msg", err, chatroomID, fid, text, atUserIDs, robotc)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("send_chatroom_msg", chatroomID, fid, text, atUserIDs, robotc)
	return exception.BoolResponse{}, nil
}

// SetMsgReaded 将一条私信设为已读。
//
// 参数:
//
//	message websocket私信消息
//
// 对应 Client.set_msg_readed。
func (c *Client) SetMsgReaded(ctx context.Context, message getgroupmsg.WsMessage) (exception.BoolResponse, error) {
	if err := c.forceWebsocket(ctx); err != nil {
		c.logCallError("set_msg_readed", err, message)
		return exception.BoolResponse{Err: err}, err
	}

	if err := setmsgreaded.Request(c.wsCore, message); err != nil {
		c.logCallError("set_msg_readed", err, message)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("set_msg_readed", message)
	return exception.BoolResponse{}, nil
}

// GetGroupMsg 获取分组信息。
//
// 参数:
//
//	groupIDs 待获取分组的group_id
//	getType 获取类型
//
// 对应 Client.get_group_msg。该接口仅支持 websocket 传输，因此会无条件建立连接
// （对应 Python 客户端的 _force_websocket 装饰器），失败会直接返回给调用方。
func (c *Client) GetGroupMsg(ctx context.Context, groupIDs []int64, getType int64) (getgroupmsg.WsMsgGroups, error) {
	if err := c.forceWebsocket(ctx); err != nil {
		c.logCallError("get_group_msg", err, groupIDs, getType)
		return getgroupmsg.WsMsgGroups{}, err
	}

	groups, err := getgroupmsg.Request(c.wsCore, groupIDs, getType)
	if err != nil {
		c.logCallError("get_group_msg", err, groupIDs, getType)
		return getgroupmsg.WsMsgGroups{}, err
	}
	return groups, nil
}

// GetBawuInfo 获取吧务团队信息。
//
// 参数:
//
//	ref 目标贴吧名或fid 优先fid
//
// 对应 Client.get_bawu_info。
func (c *Client) GetBawuInfo(ctx context.Context, ref ForumRef) (getbawuinfo.BawuInfo, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("get_bawu_info", err, ref)
		return getbawuinfo.BawuInfo{}, err
	}

	c.tryInitWebsocket(ctx)

	var info getbawuinfo.BawuInfo
	if c.wsCore.Status() == enums.WsStatusOpen {
		info, err = getbawuinfo.RequestWS(c.wsCore, fid)
	} else {
		info, err = getbawuinfo.RequestHTTP(ctx, c.httpCore, fid)
	}
	if err != nil {
		c.logCallError("get_bawu_info", err, ref)
		return getbawuinfo.BawuInfo{}, err
	}
	return info, nil
}

// GetBlacklistOld 获取旧版用户黑名单列表。
//
// 参数:
//
//	pn 页码
//	rn 请求的条目数 Max to Inf
//
// 对应 Client.get_blacklist_old。
func (c *Client) GetBlacklistOld(ctx context.Context, pn, rn int32) (getblacklistold.BlacklistOldUsers, error) {
	c.tryInitWebsocket(ctx)

	var (
		users getblacklistold.BlacklistOldUsers
		err   error
	)
	if c.wsCore.Status() == enums.WsStatusOpen {
		users, err = getblacklistold.RequestWS(c.wsCore, pn, rn)
	} else {
		users, err = getblacklistold.RequestHTTP(ctx, c.httpCore, pn, rn)
	}
	if err != nil {
		c.logCallError("get_blacklist_old", err, pn, rn)
		return getblacklistold.BlacklistOldUsers{}, err
	}
	return users, nil
}

// AddPoll 投票。
//
// 参数:
//
//	tid 投票帖的id
//	options 投票选项集合 (1,)对应第一个选项
//
// 对应 Client.add_poll。
func (c *Client) AddPoll(ctx context.Context, tid int64, options []int64) (exception.BoolResponse, error) {
	c.tryInitWebsocket(ctx)

	var err error
	if c.wsCore.Status() == enums.WsStatusOpen {
		err = addpoll.RequestWS(c.wsCore, tid, options)
	} else {
		err = addpoll.RequestHTTP(ctx, c.httpCore, tid, options)
	}
	if err != nil {
		c.logCallError("add_poll", err, tid, options)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("add_poll", tid, options)
	return exception.BoolResponse{}, nil
}

// AddPost 回复主题帖，对应 Client.add_post。
//
// 发帖在贴吧平台属于高风险操作：调用过于频繁可能导致永久封禁，请谨慎使用。
func (c *Client) AddPost(ctx context.Context, ref ForumRef, tid int64, content string) (exception.BoolResponse, error) {
	fname, fid, err := c.resolveForumBoth(ctx, ref)
	if err != nil {
		c.logCallError("add_post", err, ref, tid, content)
		return exception.BoolResponse{Err: err}, err
	}

	if err := c.InitZID(ctx); err != nil {
		c.logCallError("add_post", err, ref, tid, content)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("add_post", err, ref, tid, content)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitClientID(ctx); err != nil {
		c.logCallError("add_post", err, ref, tid, content)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitSampleID(ctx); err != nil {
		c.logCallError("add_post", err, ref, tid, content)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.initSelfinfoInitNickname(ctx); err != nil {
		c.logCallError("add_post", err, ref, tid, content)
		return exception.BoolResponse{Err: err}, err
	}

	showName := c.user.ShowName()

	c.tryInitWebsocket(ctx)
	if c.wsCore.Status() == enums.WsStatusOpen {
		err = addpost.RequestWS(c.wsCore, fname, fid, tid, showName, content)
	} else {
		err = addpost.RequestHTTP(ctx, c.httpCore, fname, fid, tid, showName, content)
	}
	if err != nil {
		c.logCallError("add_post", err, ref, tid, content)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("add_post", ref, tid, content)
	return exception.BoolResponse{}, nil
}

// GetLastReplyers 通过旧版接口获取带最后回复人的首页帖子。
//
// 参数:
//
//	ref 贴吧名或fid 优先贴吧名
//	pn 页码
//	rn 请求的条目数 Max to 100
//	sort HOT热门排序 REPLY按回复时间 CREATE按发布时间 FOLLOW关注的人
//	isGood True则获取精品区帖子 False则获取普通区帖子
//
// 对应 Client.get_last_replyers。该接口主要用于反挖坟，不暴露完整的帖子信息，目前未封装完整的返回信息。
func (c *Client) GetLastReplyers(ctx context.Context, ref ForumRef, pn, rn int32, sort enums.ThreadSortType, isGood bool) (getlastreplyers.ThreadsLP, error) {
	fname, err := c.fetchFNameOrFName(ctx, ref)
	if err != nil {
		c.logCallError("get_last_replyers", err, ref, pn, rn, sort, isGood)
		return getlastreplyers.ThreadsLP{}, err
	}

	c.tryInitWebsocket(ctx)

	var threads getlastreplyers.ThreadsLP
	if c.wsCore.Status() == enums.WsStatusOpen {
		threads, err = getlastreplyers.RequestWS(c.wsCore, fname, pn, rn, sort, isGood)
	} else {
		threads, err = getlastreplyers.RequestHTTP(ctx, c.httpCore, fname, pn, rn, sort, isGood)
	}
	if err != nil {
		c.logCallError("get_last_replyers", err, ref, pn, rn, sort, isGood)
		return getlastreplyers.ThreadsLP{}, err
	}
	return threads, nil
}

// GetReplys 获取回复信息。
//
// 参数:
//
//	pn 页码
//
// 对应 Client.get_replys。
func (c *Client) GetReplys(ctx context.Context, pn int32) (getreplys.Replys, error) {
	c.tryInitWebsocket(ctx)

	var (
		replys getreplys.Replys
		err    error
	)
	if c.wsCore.Status() == enums.WsStatusOpen {
		replys, err = getreplys.RequestWS(c.wsCore, pn)
	} else {
		replys, err = getreplys.RequestHTTP(ctx, c.httpCore, pn)
	}
	if err != nil {
		c.logCallError("get_replys", err, pn)
		return getreplys.Replys{}, err
	}
	return replys, nil
}

// GetSquareForums 获取吧广场列表。
//
// 参数:
//
//	cname 类别名
//	pn 页码
//	rn 请求的条目数 Max to Inf
//
// 对应 Client.get_square_forums。
func (c *Client) GetSquareForums(ctx context.Context, cname string, pn, rn int32) (getsquareforums.SquareForums, error) {
	c.tryInitWebsocket(ctx)

	var (
		forums getsquareforums.SquareForums
		err    error
	)
	if c.wsCore.Status() == enums.WsStatusOpen {
		forums, err = getsquareforums.RequestWS(c.wsCore, cname, pn, rn)
	} else {
		forums, err = getsquareforums.RequestHTTP(ctx, c.httpCore, cname, pn, rn)
	}
	if err != nil {
		c.logCallError("get_square_forums", err, cname, pn, rn)
		return getsquareforums.SquareForums{}, err
	}
	return forums, nil
}

// GetDislikeForums 获取首页推荐屏蔽的贴吧列表。
//
// 参数:
//
//	pn 页码
//	rn 请求的条目数 Max to 20
//
// 对应 Client.get_dislike_forums。
func (c *Client) GetDislikeForums(ctx context.Context, pn, rn int32) (getdislikeforums.DislikeForums, error) {
	c.tryInitWebsocket(ctx)

	var (
		forums getdislikeforums.DislikeForums
		err    error
	)
	if c.wsCore.Status() == enums.WsStatusOpen {
		forums, err = getdislikeforums.RequestWS(c.wsCore, pn, rn)
	} else {
		forums, err = getdislikeforums.RequestHTTP(ctx, c.httpCore, pn, rn)
	}
	if err != nil {
		c.logCallError("get_dislike_forums", err, pn, rn)
		return getdislikeforums.DislikeForums{}, err
	}
	return forums, nil
}

// TiebaUID2UserInfo 通过 tieba_uid 获取用户信息。
//
// 参数:
//
//	tiebaUID 用户id tieba_uid
//
// 对应 Client.tieba_uid2user_info。请注意 tieba_uid 与旧版 user_id 的区别。
func (c *Client) TiebaUID2UserInfo(ctx context.Context, tiebaUID int64) (tiebauid2userinfo.UserInfoTUid, error) {
	c.tryInitWebsocket(ctx)

	var (
		user tiebauid2userinfo.UserInfoTUid
		err  error
	)
	if c.wsCore.Status() == enums.WsStatusOpen {
		user, err = tiebauid2userinfo.RequestWS(c.wsCore, tiebaUID)
	} else {
		user, err = tiebauid2userinfo.RequestHTTP(ctx, c.httpCore, tiebaUID)
	}
	if err != nil {
		c.logCallError("tieba_uid2user_info", err, tiebaUID)
		return tiebauid2userinfo.UserInfoTUid{}, err
	}
	return user, nil
}

// resolveForumBoth 对应需要同时拿到贴吧名与 fid 的接口所使用的
// `if isinstance(fname_or_fid, str)` 写法。
func (c *Client) resolveForumBoth(ctx context.Context, ref ForumRef) (string, int64, error) {
	if ref.FName != "" {
		fid, err := c.fetchFID(ctx, ref.FName)
		if err != nil {
			return "", 0, err
		}
		return ref.FName, fid, nil
	}
	fname, err := c.fetchFName(ctx, ref.FID)
	if err != nil {
		return "", 0, err
	}
	return fname, ref.FID, nil
}

// resolvePortrait 对应 `user = await self.get_user_info(id_, ReqUInfo.PORTRAIT)` 写法。
func (c *Client) resolvePortrait(ctx context.Context, ref UserRef) (string, error) {
	if ref.Portrait != "" {
		return ref.Portrait, nil
	}
	user, err := c.GetUserInfo(ctx, ref, enums.ReqUInfoPortrait)
	if err != nil {
		return "", err
	}
	return user.Portrait, nil
}

// resolveUserID 对应同一写法的 ReqUInfo.USER_ID 变体。
func (c *Client) resolveUserID(ctx context.Context, ref UserRef) (int64, error) {
	if ref.UserID != 0 {
		return ref.UserID, nil
	}
	user, err := c.GetUserInfo(ctx, ref, enums.ReqUInfoUserID)
	if err != nil {
		return 0, err
	}
	return user.UserID, nil
}

// resolveUserName 对应同一写法的 ReqUInfo.USER_NAME 变体。
func (c *Client) resolveUserName(ctx context.Context, ref UserRef) (string, error) {
	if ref.UserName != "" {
		return ref.UserName, nil
	}
	user, err := c.GetUserInfo(ctx, ref, enums.ReqUInfoUserName)
	if err != nil {
		return "", err
	}
	return user.UserName, nil
}

// AddBawu 添加吧务。
//
// 参数:
//
//	ref 目标贴吧名或fid 优先fid
//	id 用户id user_id / user_name / portrait 优先user_name
//	bawuType 吧务类型
//
// 对应 Client.add_bawu。
func (c *Client) AddBawu(ctx context.Context, ref ForumRef, id UserRef, bawuType enums.BawuType) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("add_bawu", err, ref, id, bawuType)
		return exception.BoolResponse{Err: err}, err
	}
	userName, err := c.resolveUserName(ctx, id)
	if err != nil {
		c.logCallError("add_bawu", err, ref, id, bawuType)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("add_bawu", err, ref, id, bawuType)
		return exception.BoolResponse{Err: err}, err
	}

	if err := addbawu.Request(ctx, c.httpCore, fid, userName, bawuType); err != nil {
		c.logCallError("add_bawu", err, ref, id, bawuType)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("add_bawu", ref, id, bawuType)
	return exception.BoolResponse{}, nil
}

// DelBawu 删除吧务。
//
// 参数:
//
//	ref 目标贴吧名或fid 优先fid
//	id 用户id user_id / user_name / portrait 优先portrait
//	bawuType 吧务类型
//
// 对应 Client.del_bawu。
func (c *Client) DelBawu(ctx context.Context, ref ForumRef, id UserRef, bawuType enums.BawuType) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("del_bawu", err, ref, id, bawuType)
		return exception.BoolResponse{Err: err}, err
	}
	portrait, err := c.resolvePortrait(ctx, id)
	if err != nil {
		c.logCallError("del_bawu", err, ref, id, bawuType)
		return exception.BoolResponse{Err: err}, err
	}

	if err := delbawu.Request(ctx, c.httpCore, fid, portrait, bawuType); err != nil {
		c.logCallError("del_bawu", err, ref, id, bawuType)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("del_bawu", ref, id, bawuType)
	return exception.BoolResponse{}, nil
}

// SetBawuPerm 为指定吧务分配权限。
//
// 参数:
//
//	ref 目标贴吧名或fid 优先fid
//	id 用户id user_id / user_name / portrait 优先portrait
//	perms 待分配的权限
//
// 对应 Client.set_bawu_perm。
func (c *Client) SetBawuPerm(ctx context.Context, ref ForumRef, id UserRef, perms enums.BawuPermType) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("set_bawu_perm", err, ref, id, perms)
		return exception.BoolResponse{Err: err}, err
	}
	portrait, err := c.resolvePortrait(ctx, id)
	if err != nil {
		c.logCallError("set_bawu_perm", err, ref, id, perms)
		return exception.BoolResponse{Err: err}, err
	}

	if err := setbawuperm.Request(ctx, c.httpCore, fid, portrait, perms); err != nil {
		c.logCallError("set_bawu_perm", err, ref, id, perms)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("set_bawu_perm", ref, id, perms)
	return exception.BoolResponse{}, nil
}

// Block 封禁用户。
//
// 参数:
//
//	ref 所在贴吧的贴吧名或fid 优先fid
//	id 用户id user_id / user_name / portrait 优先portrait
//	day 封禁天数
//	reason 封禁理由
//
// 对应 Client.block。
func (c *Client) Block(ctx context.Context, ref ForumRef, id UserRef, day int64, reason string) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("block", err, ref, id, day, reason)
		return exception.BoolResponse{Err: err}, err
	}
	portrait, err := c.resolvePortrait(ctx, id)
	if err != nil {
		c.logCallError("block", err, ref, id, day, reason)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("block", err, ref, id, day, reason)
		return exception.BoolResponse{Err: err}, err
	}

	if err := block.Request(ctx, c.httpCore, fid, portrait, day, reason); err != nil {
		c.logCallError("block", err, ref, id, day, reason)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("block", ref, id, day, reason)
	return exception.BoolResponse{}, nil
}

// Unblock 解封用户。
//
// 参数:
//
//	ref 所在贴吧的贴吧名或fid 优先fid
//	id 用户id user_id / user_name / portrait 优先user_id
//
// 对应 Client.unblock。
func (c *Client) Unblock(ctx context.Context, ref ForumRef, id UserRef) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("unblock", err, ref, id)
		return exception.BoolResponse{Err: err}, err
	}
	userID, err := c.resolveUserID(ctx, id)
	if err != nil {
		c.logCallError("unblock", err, ref, id)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("unblock", err, ref, id)
		return exception.BoolResponse{Err: err}, err
	}

	if err := unblock.Request(ctx, c.httpCore, fid, userID); err != nil {
		c.logCallError("unblock", err, ref, id)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("unblock", ref, id)
	return exception.BoolResponse{}, nil
}

// AddBawuBlacklist 添加贴吧黑名单。
//
// 参数:
//
//	ref 目标贴吧的贴吧名或fid 优先贴吧名
//	id 用户id user_id / user_name / portrait 优先user_id
//
// 对应 Client.add_bawu_blacklist。
func (c *Client) AddBawuBlacklist(ctx context.Context, ref ForumRef, id UserRef) (exception.BoolResponse, error) {
	fname, err := c.fetchFNameOrFName(ctx, ref)
	if err != nil {
		c.logCallError("add_bawu_blacklist", err, ref, id)
		return exception.BoolResponse{Err: err}, err
	}
	userID, err := c.resolveUserID(ctx, id)
	if err != nil {
		c.logCallError("add_bawu_blacklist", err, ref, id)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("add_bawu_blacklist", err, ref, id)
		return exception.BoolResponse{Err: err}, err
	}

	if err := addbawublacklist.Request(ctx, c.httpCore, fname, userID); err != nil {
		c.logCallError("add_bawu_blacklist", err, ref, id)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("add_bawu_blacklist", ref, id)
	return exception.BoolResponse{}, nil
}

// DelBawuBlacklist 移出贴吧黑名单。
//
// 参数:
//
//	ref 目标贴吧的贴吧名或fid 优先贴吧名
//	id 用户id user_id / user_name / portrait 优先user_id
//
// 对应 Client.del_bawu_blacklist。
func (c *Client) DelBawuBlacklist(ctx context.Context, ref ForumRef, id UserRef) (exception.BoolResponse, error) {
	fname, err := c.fetchFNameOrFName(ctx, ref)
	if err != nil {
		c.logCallError("del_bawu_blacklist", err, ref, id)
		return exception.BoolResponse{Err: err}, err
	}
	userID, err := c.resolveUserID(ctx, id)
	if err != nil {
		c.logCallError("del_bawu_blacklist", err, ref, id)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("del_bawu_blacklist", err, ref, id)
		return exception.BoolResponse{Err: err}, err
	}

	if err := delbawublacklist.Request(ctx, c.httpCore, fname, userID); err != nil {
		c.logCallError("del_bawu_blacklist", err, ref, id)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("del_bawu_blacklist", ref, id)
	return exception.BoolResponse{}, nil
}

// DelThread 删除主题帖。
//
// 参数:
//
//	ref 帖子所在贴吧的贴吧名或fid 优先fid
//	tid 待删除的主题帖tid
//
// 对应 Client.del_thread。
func (c *Client) DelThread(ctx context.Context, ref ForumRef, tid int64) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("del_thread", err, ref, tid)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("del_thread", err, ref, tid)
		return exception.BoolResponse{Err: err}, err
	}

	if err := delthread.Request(ctx, c.httpCore, fid, tid, false); err != nil {
		c.logCallError("del_thread", err, ref, tid)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("del_thread", ref, tid)
	return exception.BoolResponse{}, nil
}

// DelThreads 批量删除主题帖。
//
// 参数:
//
//	ref 帖子所在贴吧的贴吧名或fid 优先fid
//	tids 待删除的主题帖tid列表. Length Max to 30
//	block 是否同时封一天
//
// 对应 Client.del_threads。部分成功返回 true。
func (c *Client) DelThreads(ctx context.Context, ref ForumRef, tids []int64, block bool) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("del_threads", err, ref, tids, block)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("del_threads", err, ref, tids, block)
		return exception.BoolResponse{Err: err}, err
	}

	if err := delthreads.Request(ctx, c.httpCore, fid, tids, block); err != nil {
		c.logCallError("del_threads", err, ref, tids, block)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("del_threads", ref, tids, block)
	return exception.BoolResponse{}, nil
}

// DelPost 删除回复。
//
// 参数:
//
//	ref 帖子所在贴吧的贴吧名或fid 优先fid
//	tid 所在主题帖tid
//	pid 待删除的回复pid
//
// 对应 Client.del_post。
func (c *Client) DelPost(ctx context.Context, ref ForumRef, tid, pid int64) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("del_post", err, ref, tid, pid)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("del_post", err, ref, tid, pid)
		return exception.BoolResponse{Err: err}, err
	}

	if err := delpost.Request(ctx, c.httpCore, fid, tid, pid); err != nil {
		c.logCallError("del_post", err, ref, tid, pid)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("del_post", ref, tid, pid)
	return exception.BoolResponse{}, nil
}

// DelPosts 批量删除回复。
//
// 参数:
//
//	ref 帖子所在贴吧的贴吧名或fid 优先fid
//	tid 所在主题帖tid
//	pids 待删除的回复pid列表. Length Max to 30
//	block 是否同时封一天
//
// 对应 Client.del_posts。部分成功返回 true。
func (c *Client) DelPosts(ctx context.Context, ref ForumRef, tid int64, pids []int64, block bool) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("del_posts", err, ref, tid, pids, block)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("del_posts", err, ref, tid, pids, block)
		return exception.BoolResponse{Err: err}, err
	}

	if err := delposts.Request(ctx, c.httpCore, fid, tid, pids, block); err != nil {
		c.logCallError("del_posts", err, ref, tid, pids, block)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("del_posts", ref, tid, pids, block)
	return exception.BoolResponse{}, nil
}

// Recover 帖子恢复相关操作。
//
// 参数:
//
//	ref 帖子所在贴吧的贴吧名或fid 优先fid
//	tid 待恢复的主题帖tid
//	pid 待恢复的回复pid
//	isHide True则取消屏蔽主题帖 False则恢复删帖
//
// 对应 Client.recover。
func (c *Client) Recover(ctx context.Context, ref ForumRef, tid, pid int64, isHide bool) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("recover", err, ref, tid, pid, isHide)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("recover", err, ref, tid, pid, isHide)
		return exception.BoolResponse{Err: err}, err
	}

	if err := recover.Request(ctx, c.httpCore, fid, tid, pid, isHide); err != nil {
		c.logCallError("recover", err, ref, tid, pid, isHide)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("recover", ref, tid, pid, isHide)
	return exception.BoolResponse{}, nil
}

// GetCID 通过精华分区名获取精华分区 id。
//
// 参数:
//
//	ref 帖子所在贴吧的贴吧名或fid
//	cname 精华分区名
//
// 对应 Client.get_cid。
func (c *Client) GetCID(ctx context.Context, ref ForumRef, cname string) (exception.IntResponse, error) {
	cid, err := c.fetchCID(ctx, ref, cname)
	if err != nil {
		c.logCallError("get_cid", err, cname)
		return exception.IntResponse{Err: err}, err
	}
	return exception.IntResponse{Value: int(cid)}, nil
}

// fetchCID 对应私有方法 Client.__get_cid。
func (c *Client) fetchCID(ctx context.Context, ref ForumRef, cname string) (int64, error) {
	if cname == "" {
		return 0, nil
	}

	fname := ref.FName
	if fname == "" {
		var err error
		if fname, err = c.fetchFName(ctx, ref.FID); err != nil {
			return 0, err
		}
	}

	cates, err := getcid.Request(ctx, c.httpCore, fname)
	if err != nil {
		return 0, err
	}
	for _, item := range cates {
		if cname == helper.JSONStr(item, "class_name") {
			return helper.JSONInt(item, "class_id"), nil
		}
	}
	return 0, nil
}

// Good 加精主题帖。
//
// 参数:
//
//	ref 帖子所在贴吧的贴吧名或fid
//	tid 待加精的主题帖tid
//	cname 待添加的精华分区名称 默认为''即不分区
//
// 对应 Client.good。
func (c *Client) Good(ctx context.Context, ref ForumRef, tid int64, cname string) (exception.BoolResponse, error) {
	fname, fid, err := c.resolveForumBoth(ctx, ref)
	if err != nil {
		c.logCallError("good", err, ref, tid, cname)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("good", err, ref, tid, cname)
		return exception.BoolResponse{Err: err}, err
	}

	cid, err := c.fetchCID(ctx, ref, cname)
	if err != nil {
		c.logCallError("good", err, ref, tid, cname)
		return exception.BoolResponse{Err: err}, err
	}

	if err := good.Request(ctx, c.httpCore, fname, fid, tid, cid); err != nil {
		c.logCallError("good", err, ref, tid, cname)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("good", ref, tid, cname)
	return exception.BoolResponse{}, nil
}

// Ungood 撤精主题帖。
//
// 参数:
//
//	ref 帖子所在贴吧的贴吧名或fid
//	tid 待撤精的主题帖tid
//
// 对应 Client.ungood。
func (c *Client) Ungood(ctx context.Context, ref ForumRef, tid int64) (exception.BoolResponse, error) {
	fname, fid, err := c.resolveForumBoth(ctx, ref)
	if err != nil {
		c.logCallError("ungood", err, ref, tid)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("ungood", err, ref, tid)
		return exception.BoolResponse{Err: err}, err
	}

	if err := ungood.Request(ctx, c.httpCore, fname, fid, tid); err != nil {
		c.logCallError("ungood", err, ref, tid)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("ungood", ref, tid)
	return exception.BoolResponse{}, nil
}

// Top 置顶主题帖。
//
// 参数:
//
//	ref 帖子所在贴吧的贴吧名或fid
//	tid 待置顶的主题帖tid
//	isVIP 是否会员置顶
//
// 对应 Client.top。
func (c *Client) Top(ctx context.Context, ref ForumRef, tid int64, isVIP bool) (exception.BoolResponse, error) {
	fname, fid, err := c.resolveForumBoth(ctx, ref)
	if err != nil {
		c.logCallError("top", err, ref, tid, isVIP)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("top", err, ref, tid, isVIP)
		return exception.BoolResponse{Err: err}, err
	}

	if err := top.Request(ctx, c.httpCore, fname, fid, tid, isVIP, true); err != nil {
		c.logCallError("top", err, ref, tid, isVIP)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("top", ref, tid, isVIP)
	return exception.BoolResponse{}, nil
}

// Move 将主题帖移动至另一分区。
//
// 参数:
//
//	ref 帖子所在贴吧的贴吧名或fid 优先fid
//	tid 待移动的主题帖tid
//	toTabID 目标分区id
//	fromTabID 来源分区id 默认为0即无分区
//
// 对应 Client.move。
func (c *Client) Move(ctx context.Context, ref ForumRef, tid, toTabID, fromTabID int64) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("move", err, ref, tid, toTabID, fromTabID)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("move", err, ref, tid, toTabID, fromTabID)
		return exception.BoolResponse{Err: err}, err
	}

	if err := move.Request(ctx, c.httpCore, fid, tid, toTabID, fromTabID); err != nil {
		c.logCallError("move", err, ref, tid, toTabID, fromTabID)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("move", ref, tid, toTabID, fromTabID)
	return exception.BoolResponse{}, nil
}

// Recommend 大吧主首页推荐。
//
// 参数:
//
//	ref 帖子所在贴吧的贴吧名或fid 优先fid
//	tid 待推荐的主题帖tid
//
// 对应 Client.recommend。
func (c *Client) Recommend(ctx context.Context, ref ForumRef, tid int64) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("recommend", err, ref, tid)
		return exception.BoolResponse{Err: err}, err
	}

	if err := recommend.Request(ctx, c.httpCore, fid, tid); err != nil {
		c.logCallError("recommend", err, ref, tid)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("recommend", ref, tid)
	return exception.BoolResponse{}, nil
}

// SetThreadPrivacy 隐藏或公开主题帖或回复。isHide 为 true 则隐藏，为 false 则公开。
//
// 本方法是 Go 版提供的合并入口，Python 客户端只有 set_thread_private 与
// set_thread_public 两个独立 API，因此它的日志名 "set_thread_privacy" 在 Python 侧
// 并不存在。除日志名不同外，其余行为（含成功日志）与那两个方法一致。
func (c *Client) SetThreadPrivacy(ctx context.Context, ref ForumRef, tid, pid int64, isHide bool) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("set_thread_privacy", err, ref, tid, pid, isHide)
		return exception.BoolResponse{Err: err}, err
	}

	if err := setthreadprivacy.Request(ctx, c.httpCore, fid, tid, pid, isHide); err != nil {
		c.logCallError("set_thread_privacy", err, ref, tid, pid, isHide)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("set_thread_privacy", ref, tid, pid, isHide)
	return exception.BoolResponse{}, nil
}

// HandleUnblockAppeals 拒绝或通过解封申诉。
//
// 参数:
//
//	ref 申诉所在贴吧的贴吧名或fid 优先fid
//	appealIDs 申诉请求的appeal_id列表. Length Max to 30
//	refuse True则拒绝申诉 False则接受申诉
//
// 对应 Client.handle_unblock_appeals。
func (c *Client) HandleUnblockAppeals(ctx context.Context, ref ForumRef, appealIDs []int64, refuse bool) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("handle_unblock_appeals", err, ref, appealIDs, refuse)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("handle_unblock_appeals", err, ref, appealIDs, refuse)
		return exception.BoolResponse{Err: err}, err
	}

	if err := handleunblockappeals.Request(ctx, c.httpCore, fid, appealIDs, refuse); err != nil {
		c.logCallError("handle_unblock_appeals", err, ref, appealIDs, refuse)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("handle_unblock_appeals", ref, appealIDs, refuse)
	return exception.BoolResponse{}, nil
}

// Agree 点赞主题帖或回复。
//
// 参数:
//
//	tid 待点赞的主题帖或回复所在的主题帖的tid
//	pid 待点赞的回复pid
//	isComment pid是否指向楼中楼
//
// 本接口仍处于测试阶段；高频率调用会导致<发帖秒删>! 请谨慎使用!
//
// 对应 Client.agree。
func (c *Client) Agree(ctx context.Context, tid, pid int64, isComment bool) (exception.BoolResponse, error) {
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("agree", err, tid, pid, isComment)
		return exception.BoolResponse{Err: err}, err
	}

	if err := agree.Request(ctx, c.httpCore, tid, pid, isComment, false, false); err != nil {
		c.logCallError("agree", err, tid, pid, isComment)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("agree", tid, pid, isComment)
	return exception.BoolResponse{}, nil
}

// FollowUser 关注用户。
//
// 参数:
//
//	id 用户id user_id / user_name / portrait 优先portrait
//
// 对应 Client.follow_user。
func (c *Client) FollowUser(ctx context.Context, id UserRef) (exception.BoolResponse, error) {
	portrait, err := c.resolvePortrait(ctx, id)
	if err != nil {
		c.logCallError("follow_user", err, id)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("follow_user", err, id)
		return exception.BoolResponse{Err: err}, err
	}

	if err := followuser.Request(ctx, c.httpCore, portrait); err != nil {
		c.logCallError("follow_user", err, id)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("follow_user", id)
	return exception.BoolResponse{}, nil
}

// UnfollowUser 取关用户。
//
// 参数:
//
//	id 用户id user_id / user_name / portrait 优先portrait
//
// 对应 Client.unfollow_user。
func (c *Client) UnfollowUser(ctx context.Context, id UserRef) (exception.BoolResponse, error) {
	portrait, err := c.resolvePortrait(ctx, id)
	if err != nil {
		c.logCallError("unfollow_user", err, id)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("unfollow_user", err, id)
		return exception.BoolResponse{Err: err}, err
	}

	if err := unfollowuser.Request(ctx, c.httpCore, portrait); err != nil {
		c.logCallError("unfollow_user", err, id)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("unfollow_user", id)
	return exception.BoolResponse{}, nil
}

// RemoveFan 移除粉丝。
//
// 参数:
//
//	id 待移除粉丝的id user_id / user_name / portrait 优先user_id
//
// 对应 Client.remove_fan。
func (c *Client) RemoveFan(ctx context.Context, id UserRef) (exception.BoolResponse, error) {
	userID, err := c.resolveUserID(ctx, id)
	if err != nil {
		c.logCallError("remove_fan", err, id)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("remove_fan", err, id)
		return exception.BoolResponse{Err: err}, err
	}

	if err := removefan.Request(ctx, c.httpCore, userID); err != nil {
		c.logCallError("remove_fan", err, id)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("remove_fan", id)
	return exception.BoolResponse{}, nil
}

// AddBlacklistOld 添加旧版用户黑名单。
//
// 参数:
//
//	id 待添加黑名单的用户id user_id / user_name / portrait 优先user_id
//
// 对应 Client.add_blacklist_old。
func (c *Client) AddBlacklistOld(ctx context.Context, id UserRef) (exception.BoolResponse, error) {
	userID, err := c.resolveUserID(ctx, id)
	if err != nil {
		c.logCallError("add_blacklist_old", err, id)
		return exception.BoolResponse{Err: err}, err
	}

	if err := addblacklistold.Request(ctx, c.httpCore, userID); err != nil {
		c.logCallError("add_blacklist_old", err, id)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("add_blacklist_old", id)
	return exception.BoolResponse{}, nil
}

// DelBlacklistOld 移除旧版用户黑名单。
//
// 参数:
//
//	id 待移除黑名单的用户id user_id / user_name / portrait 优先user_id
//
// 对应 Client.del_blacklist_old。
func (c *Client) DelBlacklistOld(ctx context.Context, id UserRef) (exception.BoolResponse, error) {
	userID, err := c.resolveUserID(ctx, id)
	if err != nil {
		c.logCallError("del_blacklist_old", err, id)
		return exception.BoolResponse{Err: err}, err
	}

	if err := delblacklistold.Request(ctx, c.httpCore, userID); err != nil {
		c.logCallError("del_blacklist_old", err, id)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("del_blacklist_old", id)
	return exception.BoolResponse{}, nil
}

// FollowForum 关注贴吧。
//
// 参数:
//
//	ref 要关注贴吧的贴吧名或fid 优先fid
//
// 对应 Client.follow_forum。
func (c *Client) FollowForum(ctx context.Context, ref ForumRef) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("follow_forum", err, ref)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("follow_forum", err, ref)
		return exception.BoolResponse{Err: err}, err
	}

	if err := followforum.Request(ctx, c.httpCore, fid); err != nil {
		c.logCallError("follow_forum", err, ref)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("follow_forum", ref)
	return exception.BoolResponse{}, nil
}

// UnfollowForum 取关贴吧。
//
// 参数:
//
//	ref 要取关贴吧的贴吧名或fid 优先fid
//
// 对应 Client.unfollow_forum。
func (c *Client) UnfollowForum(ctx context.Context, ref ForumRef) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("unfollow_forum", err, ref)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("unfollow_forum", err, ref)
		return exception.BoolResponse{Err: err}, err
	}

	if err := unfollowforum.Request(ctx, c.httpCore, fid); err != nil {
		c.logCallError("unfollow_forum", err, ref)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("unfollow_forum", ref)
	return exception.BoolResponse{}, nil
}

// DislikeForum 屏蔽贴吧，使其不再出现在首页推荐列表中。
//
// 参数:
//
//	ref 待屏蔽贴吧的贴吧名或fid 优先fid
//
// 对应 Client.dislike_forum。
func (c *Client) DislikeForum(ctx context.Context, ref ForumRef) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("dislike_forum", err, ref)
		return exception.BoolResponse{Err: err}, err
	}

	if err := dislikeforum.Request(ctx, c.httpCore, fid); err != nil {
		c.logCallError("dislike_forum", err, ref)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("dislike_forum", ref)
	return exception.BoolResponse{}, nil
}

// UndislikeForum 解除贴吧的首页推荐屏蔽。
//
// 参数:
//
//	ref 待屏蔽贴吧的贴吧名或fid 优先fid
//
// 对应 Client.undislike_forum。
func (c *Client) UndislikeForum(ctx context.Context, ref ForumRef) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("undislike_forum", err, ref)
		return exception.BoolResponse{Err: err}, err
	}

	if err := undislikeforum.Request(ctx, c.httpCore, fid); err != nil {
		c.logCallError("undislike_forum", err, ref)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("undislike_forum", ref)
	return exception.BoolResponse{}, nil
}

// SetProfile 设置主页信息。
//
// 参数:
//
//	nickName 昵称
//	sign 个性签名
//	gender 性别
//
// 对应 Client.set_profile。
func (c *Client) SetProfile(ctx context.Context, nickName, sign string, gender enums.Gender) (exception.BoolResponse, error) {
	if err := setprofile.Request(ctx, c.httpCore, nickName, sign, gender); err != nil {
		c.logCallError("set_profile", err, nickName, sign, gender)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("set_profile", nickName, sign, gender)
	return exception.BoolResponse{}, nil
}

// SetNicknameOld 设置旧版昵称。
//
// 参数:
//
//	nickName 昵称
//
// 对应 Client.set_nickname_old。
func (c *Client) SetNicknameOld(ctx context.Context, nickName string) (exception.BoolResponse, error) {
	if err := setnicknameold.Request(ctx, c.httpCore, nickName); err != nil {
		c.logCallError("set_nickname_old", err, nickName)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("set_nickname_old", nickName)
	return exception.BoolResponse{}, nil
}

// SignForum 单个贴吧签到。
//
// 参数:
//
//	ref 要签到贴吧的贴吧名或fid 优先贴吧名
//
// 对应 Client.sign_forum。
func (c *Client) SignForum(ctx context.Context, ref ForumRef) (exception.BoolResponse, error) {
	fname, err := c.fetchFNameOrFName(ctx, ref)
	if err != nil {
		c.logCallError("sign_forum", err, ref)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("sign_forum", err, ref)
		return exception.BoolResponse{Err: err}, err
	}

	if err := signforum.Request(ctx, c.httpCore, fname); err != nil {
		c.logCallError("sign_forum", err, ref)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("sign_forum", ref)
	return exception.BoolResponse{}, nil
}

// SignForums 一键签到，对应 Client.sign_forums。
func (c *Client) SignForums(ctx context.Context) (exception.BoolResponse, error) {
	if err := signforums.Request(ctx, c.httpCore); err != nil {
		c.logCallError("sign_forums", err)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("sign_forums")
	return exception.BoolResponse{}, nil
}

// SignGrowth 用户成长等级任务: 签到，对应 Client.sign_growth。
func (c *Client) SignGrowth(ctx context.Context) (exception.BoolResponse, error) {
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("sign_growth", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := signgrowth.RequestWeb(ctx, c.httpCore, "page_sign"); err != nil {
		c.logCallError("sign_growth", err)
		return exception.BoolResponse{Err: err}, err
	}
	c.logCallSuccess("sign_growth")
	return exception.BoolResponse{}, nil
}

// GetPostsArgs 是 GetPosts 的可选参数。
type GetPostsArgs struct {
	Pn                 int                // 页码
	Rn                 int                // 请求的条目数
	Sort               enums.PostSortType // ASC时间顺序 DESC时间倒序 HOT热门序
	OnlyThreadAuthor   bool               // True则只看楼主 False则请求全部
	WithComments       bool               // True则同时请求高赞楼中楼 False则返回的 Post.comments 字段为空
	CommentSortByAgree bool               // True则楼中楼按点赞数顺序 False则楼中楼按时间顺序
	CommentRn          int                // 请求的楼中楼数量 Max to 50
}

// DefaultGetPostsArgs 返回 GetPosts 的 Python 默认值。
func DefaultGetPostsArgs() GetPostsArgs {
	return GetPostsArgs{
		Pn:                 1,
		Rn:                 30,
		Sort:               enums.PostSortAsc,
		OnlyThreadAuthor:   false,
		WithComments:       false,
		CommentSortByAgree: true,
		CommentRn:          4,
	}
}

// GetPosts 获取主题帖内回复。
//
// 参数:
//
//	tid 所在主题帖 tid
//	args 可选参数，详见 GetPostsArgs
//
// 对应 Client.get_posts。
func (c *Client) GetPosts(ctx context.Context, tid int64, args GetPostsArgs) (getposts.Posts, error) {
	c.tryInitWebsocket(ctx)

	var (
		posts getposts.Posts
		err   error
	)
	pn, rn, sort, commentRn := int32(args.Pn), int32(args.Rn), int32(args.Sort), int32(args.CommentRn)
	if c.wsCore.Status() == enums.WsStatusOpen {
		posts, err = getposts.RequestWS(c.wsCore, tid, pn, rn, sort,
			args.OnlyThreadAuthor, args.WithComments, args.CommentSortByAgree, commentRn)
	} else {
		posts, err = getposts.RequestHTTP(ctx, c.httpCore, tid, pn, rn, sort,
			args.OnlyThreadAuthor, args.WithComments, args.CommentSortByAgree, commentRn)
	}
	if err != nil {
		c.logCallError("get_posts", err, tid, logging.PyKw{Name: "pn", Value: args.Pn})
		return posts, err
	}
	return posts, nil
}

// GetCommentsArgs 是 GetComments 的可选参数。
type GetCommentsArgs struct {
	Pn        int                // 页码
	IsComment bool               // pid 是否指向楼中楼 若指向楼中楼则获取其附近的楼中楼列表
	Sort      enums.PostSortType // ASC时间顺序 DESC时间倒序 HOT热门序
}

// DefaultGetCommentsArgs 返回 GetComments 的 Python 默认值。
func DefaultGetCommentsArgs() GetCommentsArgs {
	return GetCommentsArgs{Pn: 1, Sort: enums.PostSortAsc}
}

// GetComments 获取楼中楼回复。
//
// 参数:
//
//	tid 所在主题帖tid
//	pid 所在楼层的pid或楼中楼的pid
//	args 可选参数，详见 GetCommentsArgs
//
// 对应 Client.get_comments。
func (c *Client) GetComments(ctx context.Context, tid, pid int64, args GetCommentsArgs) (getcomments.Comments, error) {
	c.tryInitWebsocket(ctx)

	var (
		comments getcomments.Comments
		err      error
	)
	pn, sort := int32(args.Pn), int32(args.Sort)
	if c.wsCore.Status() == enums.WsStatusOpen {
		comments, err = getcomments.RequestWS(c.wsCore, tid, pid, pn, sort, args.IsComment)
	} else {
		comments, err = getcomments.RequestHTTP(ctx, c.httpCore, tid, pid, pn, sort, args.IsComment)
	}
	if err != nil {
		c.logCallError("get_comments", err, tid, pid, logging.PyKw{Name: "pn", Value: args.Pn})
		return comments, err
	}
	return comments, nil
}

// tryInitWebsocket 对应 _try_websocket 装饰器：websocket 初始化失败时只记录日志，
// 调用会静默回退到 HTTP，与 Python 的 handle_exception 包装完全一致。
func (c *Client) tryInitWebsocket(ctx context.Context) {
	if !c.tryWS {
		return
	}
	if _, err := c.InitWebsocket(ctx); err != nil {
		c.logCallError("init_websocket", err)
	}
}

// forceWebsocket 建立 websocket 连接，对应 _force_websocket 装饰器。
// 与 tryInitWebsocket 不同，失败会返回给调用方而不是被吞掉。
func (c *Client) forceWebsocket(ctx context.Context) error {
	_, err := c.InitWebsocket(ctx)
	return err
}

// logCallError 记录一次失败的 API 调用，对应 Python 的 handle_exception 异常分支。
//
// Python 客户端通过 handle_exception 装饰器记录所有失败，因此 Go 版在客户端边界
// 保持同样的行为。日志正文与 Python 版一致：方括号内是 API 名，随后是异常本身
// （而非一句描述），最后是 Python 风格的参数后缀。
//
//	[sign_forums] (340011, ''). args=() kwargs={}
//
// args 中的值按位置参数渲染；需要落到 kwargs 的可选参数用 logging.PyKw 标记。
func (c *Client) logCallError(apiName string, err error, args ...any) {
	logging.GetLogger().Warn().Msg(
		fmt.Sprintf("[%s] %s. %s", apiName, logging.PyErr(err), logging.PyArgs(args...)),
	)
}

// logCallSuccess 记录一次成功的 API 调用，对应 Python 的 handle_exception 成功分支。
//
// 只有 Python 侧标注了 ok_log_level=logging.INFO 的写操作才会调用它，
// 读取类接口不记录成功日志。
//
//	[sign_forum] Succeeded. args=('盗墓笔记',) kwargs={}
func (c *Client) logCallSuccess(apiName string, args ...any) {
	logging.GetLogger().Info().Msg(
		fmt.Sprintf("[%s] Succeeded. %s", apiName, logging.PyArgs(args...)),
	)
}
