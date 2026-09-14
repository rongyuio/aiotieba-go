// Package aiotieba is a Go port of the aiotieba library: an asynchronous client
// for the Baidu Tieba APIs.
//
// The public method names follow the Python client (aiotieba.client.Client) so
// that both implementations can be compared side by side, while the internals
// use idiomatic Go: context.Context, explicit error returns and structs.
package aiotieba

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/rongyuio/aiotieba/api/add_bawu"
	"github.com/rongyuio/aiotieba/api/add_bawu_blacklist"
	"github.com/rongyuio/aiotieba/api/add_blacklist_old"
	"github.com/rongyuio/aiotieba/api/add_poll"
	"github.com/rongyuio/aiotieba/api/agree"
	"github.com/rongyuio/aiotieba/api/block"
	"github.com/rongyuio/aiotieba/api/classdef"
	"github.com/rongyuio/aiotieba/api/del_bawu"
	"github.com/rongyuio/aiotieba/api/del_bawu_blacklist"
	"github.com/rongyuio/aiotieba/api/del_blacklist_old"
	"github.com/rongyuio/aiotieba/api/del_post"
	"github.com/rongyuio/aiotieba/api/del_posts"
	"github.com/rongyuio/aiotieba/api/del_thread"
	"github.com/rongyuio/aiotieba/api/del_threads"
	"github.com/rongyuio/aiotieba/api/dislike_forum"
	"github.com/rongyuio/aiotieba/api/follow_forum"
	"github.com/rongyuio/aiotieba/api/follow_user"
	"github.com/rongyuio/aiotieba/api/get_ats"
	"github.com/rongyuio/aiotieba/api/get_bawu_blacklist"
	"github.com/rongyuio/aiotieba/api/get_bawu_info"
	"github.com/rongyuio/aiotieba/api/get_bawu_memberlist"
	"github.com/rongyuio/aiotieba/api/get_bawu_perm"
	"github.com/rongyuio/aiotieba/api/get_bawu_postlogs"
	"github.com/rongyuio/aiotieba/api/get_bawu_userlogs"
	"github.com/rongyuio/aiotieba/api/get_blacklist"
	"github.com/rongyuio/aiotieba/api/get_blacklist_old"
	"github.com/rongyuio/aiotieba/api/get_blocks"
	"github.com/rongyuio/aiotieba/api/get_cid"
	"github.com/rongyuio/aiotieba/api/get_comments"
	"github.com/rongyuio/aiotieba/api/get_dislike_forums"
	"github.com/rongyuio/aiotieba/api/get_fans"
	"github.com/rongyuio/aiotieba/api/get_fid"
	"github.com/rongyuio/aiotieba/api/get_follow_forums"
	"github.com/rongyuio/aiotieba/api/get_follow_forums_pc"
	"github.com/rongyuio/aiotieba/api/get_follows"
	"github.com/rongyuio/aiotieba/api/get_forum"
	"github.com/rongyuio/aiotieba/api/get_forum_detail"
	"github.com/rongyuio/aiotieba/api/get_forum_level"
	"github.com/rongyuio/aiotieba/api/get_group_msg"
	"github.com/rongyuio/aiotieba/api/get_images"
	"github.com/rongyuio/aiotieba/api/get_last_replyers"
	"github.com/rongyuio/aiotieba/api/get_member_users"
	"github.com/rongyuio/aiotieba/api/get_posts"
	"github.com/rongyuio/aiotieba/api/get_rank_forums"
	"github.com/rongyuio/aiotieba/api/get_rank_users"
	"github.com/rongyuio/aiotieba/api/get_recom_status"
	"github.com/rongyuio/aiotieba/api/get_recovers"
	"github.com/rongyuio/aiotieba/api/get_replys"
	"github.com/rongyuio/aiotieba/api/get_roomlist_by_fid"
	"github.com/rongyuio/aiotieba/api/get_self_follow_forums"
	"github.com/rongyuio/aiotieba/api/get_selfinfo_initNickname"
	"github.com/rongyuio/aiotieba/api/get_square_forums"
	"github.com/rongyuio/aiotieba/api/get_statistics"
	"github.com/rongyuio/aiotieba/api/get_tab_map"
	"github.com/rongyuio/aiotieba/api/get_threads"
	"github.com/rongyuio/aiotieba/api/get_uinfo_getUserInfo_web"
	"github.com/rongyuio/aiotieba/api/get_uinfo_getuserinfo_app"
	"github.com/rongyuio/aiotieba/api/get_uinfo_panel"
	"github.com/rongyuio/aiotieba/api/get_uinfo_userCard"
	"github.com/rongyuio/aiotieba/api/get_uinfo_user_json"
	"github.com/rongyuio/aiotieba/api/get_unblock_appeals"
	"github.com/rongyuio/aiotieba/api/get_user_contents"
	getusercontentsposts "github.com/rongyuio/aiotieba/api/get_user_contents/get_posts"
	getusercontentsthreads "github.com/rongyuio/aiotieba/api/get_user_contents/get_threads"
	"github.com/rongyuio/aiotieba/api/get_user_contents_pc"
	"github.com/rongyuio/aiotieba/api/get_user_forum_info"
	"github.com/rongyuio/aiotieba/api/good"
	"github.com/rongyuio/aiotieba/api/handle_unblock_appeals"
	"github.com/rongyuio/aiotieba/api/init_websocket"
	"github.com/rongyuio/aiotieba/api/init_z_id"
	"github.com/rongyuio/aiotieba/api/login"
	"github.com/rongyuio/aiotieba/api/move"
	"github.com/rongyuio/aiotieba/api/profile"
	"github.com/rongyuio/aiotieba/api/profile/get_homepage"
	"github.com/rongyuio/aiotieba/api/profile/get_uinfo_profile"
	"github.com/rongyuio/aiotieba/api/recommend"
	"github.com/rongyuio/aiotieba/api/recover"
	"github.com/rongyuio/aiotieba/api/remove_fan"
	"github.com/rongyuio/aiotieba/api/search_exact"
	"github.com/rongyuio/aiotieba/api/search_global"
	"github.com/rongyuio/aiotieba/api/send_chatroom_msg"
	"github.com/rongyuio/aiotieba/api/send_msg"
	"github.com/rongyuio/aiotieba/api/set_bawu_perm"
	"github.com/rongyuio/aiotieba/api/set_blacklist"
	"github.com/rongyuio/aiotieba/api/set_msg_readed"
	"github.com/rongyuio/aiotieba/api/set_nickname_old"
	"github.com/rongyuio/aiotieba/api/set_profile"
	"github.com/rongyuio/aiotieba/api/set_thread_privacy"
	"github.com/rongyuio/aiotieba/api/sign_forum"
	"github.com/rongyuio/aiotieba/api/sign_forums"
	"github.com/rongyuio/aiotieba/api/sign_growth"
	syncapi "github.com/rongyuio/aiotieba/api/sync"
	"github.com/rongyuio/aiotieba/api/tieba_uid2user_info"
	"github.com/rongyuio/aiotieba/api/top"
	"github.com/rongyuio/aiotieba/api/unblock"
	"github.com/rongyuio/aiotieba/api/undislike_forum"
	"github.com/rongyuio/aiotieba/api/unfollow_forum"
	"github.com/rongyuio/aiotieba/api/unfollow_user"
	"github.com/rongyuio/aiotieba/api/ungood"
	"github.com/rongyuio/aiotieba/config"
	"github.com/rongyuio/aiotieba/consts"
	"github.com/rongyuio/aiotieba/core"
	"github.com/rongyuio/aiotieba/enums"
	"github.com/rongyuio/aiotieba/exception"
	"github.com/rongyuio/aiotieba/helper"
	"github.com/rongyuio/aiotieba/logging"
)

// bLCPQueueLength is the capacity of the BLCP notification queue, mirroring the
// Python default of 100.
const bLCPQueueLength = 100

// Option configures a Client.
type Option func(*clientOptions)

type clientOptions struct {
	account *core.Account
	timeout config.TimeoutConfig
	proxy   *config.ProxyConfig
	tryWS   bool
}

// WithAccount sets the account of the client, overriding BDUSS and STOKEN.
func WithAccount(account *core.Account) Option {
	return func(o *clientOptions) { o.account = account }
}

// WithTryWebsocket enables the websocket transport with an HTTP fallback.
func WithTryWebsocket(try bool) Option {
	return func(o *clientOptions) { o.tryWS = try }
}

// WithProxy sets the proxy configuration.
func WithProxy(proxy *config.ProxyConfig) Option {
	return func(o *clientOptions) { o.proxy = proxy }
}

// WithProxyFromEnv uses the proxy described by the environment variables.
func WithProxyFromEnv() Option {
	return func(o *clientOptions) { o.proxy = config.FromEnv() }
}

// WithTimeout sets the timeout configuration.
func WithTimeout(timeout config.TimeoutConfig) Option {
	return func(o *clientOptions) { o.timeout = timeout }
}

// Client is the entry point of the library. It mirrors aiotieba.client.Client.
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

// New creates a client.
//
// BDUSS must be empty or 192 characters long and STOKEN must be empty or 64
// characters long, like the Python constructor.
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

// Close releases the network resources of the client. It mirrors __aexit__.
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

// Account returns the account of the client.
func (c *Client) Account() *core.Account {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.account
}

// SetAccount swaps the account of every session, mirroring the Python account
// setter.
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

// User returns the cached user information of the account. It mirrors the
// self_info property.
func (c *Client) User() classdef.UserInfo {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.user
}

// HTTPCore exposes the HTTP session of the client.
func (c *Client) HTTPCore() *core.HttpCore { return c.httpCore }

// WSCore exposes the websocket session of the client.
func (c *Client) WSCore() *core.WsCore { return c.wsCore }

// BLCPCore exposes the BLCP session of the client.
func (c *Client) BLCPCore() *core.BLCPCore { return c.blcpCore }

// InitWebsocket connects the websocket session and uploads the secret key,
// mirroring Client.init_websocket. It is a no-op when the session is not closed.
func (c *Client) InitWebsocket(ctx context.Context) (bool, error) {
	if c.wsCore.Status() != enums.WsStatusClosed {
		return true, nil
	}
	if err := c.wsCore.Connect(ctx); err != nil {
		return false, err
	}
	if err := c.uploadSecKey(); err != nil {
		// The Python client resets the status to CLOSED when the upload fails so
		// that the next call retries the handshake.
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

// GetThreadsArgs holds the optional arguments of GetThreads.
type GetThreadsArgs struct {
	Pn     int
	Rn     int
	Sort   enums.ThreadSortType
	IsGood bool
}

// DefaultGetThreadsArgs returns the Python defaults of GetThreads.
func DefaultGetThreadsArgs() GetThreadsArgs {
	return GetThreadsArgs{Pn: 1, Rn: 30, Sort: enums.ThreadSortReply}
}

// GetThreads returns the thread list of a forum, mirroring Client.get_threads.
//
// It uses the websocket transport when the session is open, falling back to the
// app HTTP API otherwise.
func (c *Client) GetThreads(ctx context.Context, fname string, args GetThreadsArgs) (getthreads.Threads, error) {
	c.tryInitWebsocket(ctx)

	var (
		threads getthreads.Threads
		err     error
	)
	pn, rn, sort := int32(args.Pn), int32(args.Rn), int32(args.Sort)
	if c.wsCore.Status() == enums.WsStatusOpen {
		threads, err = getthreads.RequestWS(c.wsCore, fname, pn, rn, sort, args.IsGood, consts.LegacyVersion)
	} else {
		threads, err = getthreads.RequestHTTP(ctx, c.httpCore, fname, pn, rn, sort, args.IsGood, consts.LegacyVersion)
	}
	if err != nil {
		c.logCallError("get_threads", err, "fname", fname, "pn", args.Pn)
		return threads, err
	}
	return threads, nil
}

// ForumRef identifies a forum by name or by fid. It mirrors the
// `fname_or_fid: str | int` argument of the Python client, where the name takes
// precedence for get_forum and the fid takes precedence for get_forum_detail.
type ForumRef struct {
	FName string
	FID   int64
}

// ByFName references a forum by its name.
func ByFName(fname string) ForumRef { return ForumRef{FName: fname} }

// ByFID references a forum by its fid.
func ByFID(fid int64) ForumRef { return ForumRef{FID: fid} }

// GetFID resolves a forum name to its fid, using the forum cache first. It
// mirrors Client.get_fid.
func (c *Client) GetFID(ctx context.Context, fname string) (int64, error) {
	fid, err := c.fetchFID(ctx, fname)
	if err != nil {
		c.logCallError("get_fid", err, "fname", fname)
		return 0, err
	}
	return fid, nil
}

// GetFName resolves a fid to its forum name, using the forum cache first. It
// mirrors Client.get_fname.
func (c *Client) GetFName(ctx context.Context, fid int64) (string, error) {
	fname, err := c.fetchFName(ctx, fid)
	if err != nil {
		c.logCallError("get_fname", err, "fid", fid)
		return "", err
	}
	return fname, nil
}

// GetForum returns the information of a forum, mirroring Client.get_forum.
func (c *Client) GetForum(ctx context.Context, ref ForumRef) (getforum.Forum, error) {
	fname := ref.FName
	if fname == "" {
		var err error
		if fname, err = c.fetchFName(ctx, ref.FID); err != nil {
			c.logCallError("get_forum", err, "fid", ref.FID)
			return getforum.Forum{}, err
		}
	}

	forum, err := getforum.Request(ctx, c.httpCore, fname)
	if err != nil {
		c.logCallError("get_forum", err, "fname", fname)
		return getforum.Forum{}, err
	}
	return forum, nil
}

// GetForumDetail returns the information of a forum, mirroring
// Client.get_forum_detail.
func (c *Client) GetForumDetail(ctx context.Context, ref ForumRef) (getforumdetail.ForumDetail, error) {
	c.tryInitWebsocket(ctx)

	fid := ref.FID
	if fid == 0 {
		var err error
		if fid, err = c.fetchFID(ctx, ref.FName); err != nil {
			c.logCallError("get_forum_detail", err, "fname", ref.FName)
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
		c.logCallError("get_forum_detail", err, "fid", fid)
		return detail, err
	}
	return detail, nil
}

// fetchFIDOrFID mirrors the `fid = fname_or_fid if isinstance(fname_or_fid, int)
// else await self.__get_fid(fname_or_fid)` idiom.
func (c *Client) fetchFIDOrFID(ctx context.Context, ref ForumRef) (int64, error) {
	if ref.FName != "" {
		return c.fetchFID(ctx, ref.FName)
	}
	return ref.FID, nil
}

// fetchFNameOrFName mirrors the `fname = fname_or_fid if
// isinstance(fname_or_fid, str) else await self.__get_fname(fname_or_fid)`
// idiom.
func (c *Client) fetchFNameOrFName(ctx context.Context, ref ForumRef) (string, error) {
	if ref.FName != "" {
		return ref.FName, nil
	}
	return c.fetchFName(ctx, ref.FID)
}

// fetchFID mirrors the private Client.__get_fid.
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

// fetchFName mirrors the private Client.__get_fname.
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

// InitTbs loads the tbs token of the account when it is missing. It mirrors the
// private Client.__init_tbs.
func (c *Client) InitTbs(ctx context.Context) error {
	if c.account.Tbs() != "" {
		return nil
	}
	return c.Login(ctx)
}

// Login refreshes the cached user information and the tbs token. It mirrors the
// private Client.__login.
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

// InitClientID loads the client id of the account when it is missing. It
// mirrors the private Client.__init_client_id.
func (c *Client) InitClientID(ctx context.Context) error {
	if c.account.ClientID() != "" {
		return nil
	}
	return c.sync(ctx)
}

// InitSampleID loads the sample id of the account when it is missing. It
// mirrors the private Client.__init_sample_id.
func (c *Client) InitSampleID(ctx context.Context) error {
	if c.account.SampleID() != "" {
		return nil
	}
	return c.sync(ctx)
}

// InitZID loads the z_id of the account when it is missing. It mirrors the
// private Client.__init_z_id.
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

// sync mirrors the private Client.__sync.
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

// UserRef identifies a user by numeric id, portrait or user name. It mirrors the
// `id_: str | int` argument of the Python client, where a string is either a
// portrait ("tb.1.xxx") or a user name.
type UserRef struct {
	UserID   int64
	Portrait string
	UserName string
}

// ByUserID references a user by their numeric id.
func ByUserID(id int64) UserRef { return UserRef{UserID: id} }

// ByPortrait references a user by their portrait.
func ByPortrait(portrait string) UserRef { return UserRef{Portrait: portrait} }

// ByUserName references a user by their user name.
func ByUserName(userName string) UserRef { return UserRef{UserName: userName} }

// IsZero reports whether the reference carries no identifier.
func (r UserRef) IsZero() bool {
	return r.UserID == 0 && r.Portrait == "" && r.UserName == ""
}

// isSubset reports whether every bit of a is set in b, mirroring the Python
// `(a | b) == b` idiom used to test whether only certain fields were requested.
func isSubset(a, b enums.ReqUInfo) bool { return a|b == b }

// GetUserInfo returns the information of a user, mirroring Client.get_user_info.
//
// The Python client returns a different specialised type per endpoint; this port
// normalises every branch into classdef.UserInfo so callers see one stable type
// while the request itself stays identical. Branches whose endpoint has not been
// migrated yet fail with exception.ErrNotMigrated.
func (c *Client) GetUserInfo(ctx context.Context, ref UserRef, require enums.ReqUInfo) (classdef.UserInfo, error) {
	if ref.IsZero() {
		logging.GetLogger().Warn("GetUserInfo: empty input")
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
		c.logCallError("get_user_info", err, "require", require)
		return classdef.UserInfo{}, err
	}
	return user, nil
}

// getUserInfoByIDOrPortrait mirrors the numeric-id and portrait branches.
func (c *Client) getUserInfoByIDOrPortrait(ctx context.Context, ref UserRef, require enums.ReqUInfo) (classdef.UserInfo, error) {
	if ref.Portrait != "" {
		// The reference is a portrait.
		if isSubset(require, enums.ReqUInfoBasic) && require&enums.ReqUInfoUserID == 0 {
			return c.getUinfoPanel(ctx, ref.Portrait)
		}
		if isSubset(require, enums.ReqUInfoNickName|enums.ReqUInfoTiebaUID) {
			return c.getUinfoUserCard(ctx, ref.Portrait)
		}
		return c.getUinfoProfile(ctx, ref)
	}

	// The reference is a numeric id.
	if isSubset(require, enums.ReqUInfoBasic) {
		return c.getUinfoGetUserInfoApp(ctx, ref.UserID)
	}
	if c.account.BDUSS() != "" && require&(enums.ReqUInfoTiebaUID|enums.ReqUInfoOther) == 0 {
		return c.getUinfoGetUserInfoWeb(ctx, ref.UserID)
	}
	return c.getUinfoProfile(ctx, ref)
}

// getUserInfoByName mirrors the user-name branch.
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

// getUinfoGetUserInfoApp mirrors the private _get_uinfo_getuserinfo.
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
	// The app endpoint reports ids above math.MaxInt32 as negative.
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

// getUinfoGetUserInfoWeb mirrors the private _get_uinfo_getUserInfo.
func (c *Client) getUinfoGetUserInfoWeb(ctx context.Context, userID int64) (classdef.UserInfo, error) {
	user, err := getuserinfoweb.Request(ctx, c.httpCore, userID)
	if err != nil {
		return classdef.UserInfo{}, err
	}
	return classdef.UserInfo{
		// The endpoint does not echo the id, so the requested one is used.
		UserID:      userID,
		Portrait:    user.Portrait,
		UserName:    user.UserName,
		NickNameNew: user.NickNameNew,
	}, nil
}

// getUinfoUserJSON mirrors the private _get_uinfo_user_json.
func (c *Client) getUinfoUserJSON(ctx context.Context, userName string) (classdef.UserInfo, error) {
	user, err := getuserjson.Request(ctx, c.httpCore, userName)
	if err != nil {
		return classdef.UserInfo{}, err
	}
	return classdef.UserInfo{
		UserID:   user.UserID,
		Portrait: user.Portrait,
		// The endpoint does not echo the name.
		UserName: userName,
	}, nil
}

// getUinfoPanel mirrors the private _get_uinfo_panel.
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

// getUinfoUserCard mirrors the private _get_uinfo_userCard.
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

// getUinfoProfile mirrors the private _get_uinfo_profile.
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

// GetHomepage returns the posts of a user home page, mirroring
// Client.get_homepage.
func (c *Client) GetHomepage(ctx context.Context, id UserRef, pn int32) (profile.Homepage, error) {
	userID, err := c.resolveUserID(ctx, id)
	if err != nil {
		c.logCallError("get_homepage", err)
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
		c.logCallError("get_homepage", err, "user_id", userID, "pn", pn)
		return profile.Homepage{}, err
	}
	return homepage, nil
}

// GetTabMap returns the mapping from tab name to tab id, mirroring
// Client.get_tab_map.
func (c *Client) GetTabMap(ctx context.Context, ref ForumRef) (gettabmap.TabMap, error) {
	fname, err := c.fetchFNameOrFName(ctx, ref)
	if err != nil {
		c.logCallError("get_tab_map", err)
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
		c.logCallError("get_tab_map", err, "fname", fname)
		return gettabmap.TabMap{}, err
	}
	return tabMap, nil
}

// SendMsg sends a private message, mirroring Client.send_msg.
//
// The API is websocket only. The returned msg id is recorded in the message id
// manager so that subsequent reads continue from it.
func (c *Client) SendMsg(ctx context.Context, id UserRef, content string) (exception.BoolResponse, error) {
	userID, err := c.resolveUserID(ctx, id)
	if err != nil {
		c.logCallError("send_msg", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.forceWebsocket(ctx); err != nil {
		c.logCallError("send_msg", err)
		return exception.BoolResponse{Err: err}, err
	}

	msgID, err := sendmsg.Request(c.wsCore, userID, content)
	if err != nil {
		c.logCallError("send_msg", err, "user_id", userID)
		return exception.BoolResponse{Err: err}, err
	}

	midManager := c.wsCore.MsgIDManager()
	midManager.UpdateMsgID(midManager.PrivGID, int(msgID))

	return exception.BoolResponse{}, nil
}

// SetBlacklist sets the new user blacklist for a user, mirroring
// Client.set_blacklist.
func (c *Client) SetBlacklist(ctx context.Context, id UserRef, btype enums.BlacklistType) (exception.BoolResponse, error) {
	userID, err := c.resolveUserID(ctx, id)
	if err != nil {
		c.logCallError("set_blacklist", err)
		return exception.BoolResponse{Err: err}, err
	}

	c.tryInitWebsocket(ctx)

	if c.wsCore.Status() == enums.WsStatusOpen {
		err = setblacklist.RequestWS(c.wsCore, userID, btype)
	} else {
		err = setblacklist.RequestHTTP(ctx, c.httpCore, userID, btype)
	}
	if err != nil {
		c.logCallError("set_blacklist", err, "user_id", userID)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// GetAts returns the @ notifications of the account, mirroring Client.get_ats.
func (c *Client) GetAts(ctx context.Context, pn int64) (getats.Ats, error) {
	ats, err := getats.Request(ctx, c.httpCore, pn)
	if err != nil {
		c.logCallError("get_ats", err)
		return getats.Ats{}, err
	}
	return ats, nil
}

// GetBlacklist returns the new-style user blacklist, mirroring
// Client.get_blacklist.
func (c *Client) GetBlacklist(ctx context.Context) (getblacklist.BlacklistUsers, error) {
	users, err := getblacklist.Request(ctx, c.httpCore)
	if err != nil {
		c.logCallError("get_blacklist", err)
		return getblacklist.BlacklistUsers{}, err
	}
	return users, nil
}

// resolveUserIDOrSelf resolves a user reference to a numeric id, using the
// account itself when the reference is empty, mirroring the `id_ is None`
// branch of the fan/follow APIs.
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

// GetFans returns the fans of a user, mirroring Client.get_fans. An empty id
// refers to the account itself.
func (c *Client) GetFans(ctx context.Context, id UserRef, pn int64) (getfans.Fans, error) {
	userID, err := c.resolveUserIDOrSelf(ctx, id)
	if err != nil {
		c.logCallError("get_fans", err)
		return getfans.Fans{}, err
	}
	fans, err := getfans.Request(ctx, c.httpCore, userID, pn)
	if err != nil {
		c.logCallError("get_fans", err, "user_id", userID)
		return getfans.Fans{}, err
	}
	return fans, nil
}

// GetFollows returns the follow list of a user, mirroring Client.get_follows.
// An empty id refers to the account itself.
func (c *Client) GetFollows(ctx context.Context, id UserRef, pn int64) (getfollows.Follows, error) {
	userID, err := c.resolveUserIDOrSelf(ctx, id)
	if err != nil {
		c.logCallError("get_follows", err)
		return getfollows.Follows{}, err
	}
	follows, err := getfollows.Request(ctx, c.httpCore, userID, pn)
	if err != nil {
		c.logCallError("get_follows", err, "user_id", userID)
		return getfollows.Follows{}, err
	}
	return follows, nil
}

// GetFollowForums returns the forums followed by a user, mirroring
// Client.get_follow_forums.
func (c *Client) GetFollowForums(ctx context.Context, id UserRef, pn, rn int64) (getfollowforums.FollowForums, error) {
	userID, err := c.resolveUserID(ctx, id)
	if err != nil {
		c.logCallError("get_follow_forums", err)
		return getfollowforums.FollowForums{}, err
	}
	forums, err := getfollowforums.Request(ctx, c.httpCore, userID, pn, rn)
	if err != nil {
		c.logCallError("get_follow_forums", err, "user_id", userID)
		return getfollowforums.FollowForums{}, err
	}
	return forums, nil
}

// GetRoomlistByFID returns the chatrooms of a forum, mirroring
// Client.get_roomlist_by_fid.
func (c *Client) GetRoomlistByFID(ctx context.Context, fid int64) (getroomlistbyfid.RoomList, error) {
	roomList, err := getroomlistbyfid.Request(ctx, c.httpCore, fid)
	if err != nil {
		c.logCallError("get_roomlist_by_fid", err, "fid", fid)
		return getroomlistbyfid.RoomList{}, err
	}
	return roomList, nil
}

// GetStatistics returns the 24-day statistics of the forum backend, mirroring
// Client.get_statistics.
func (c *Client) GetStatistics(ctx context.Context, ref ForumRef) (getstatistics.Statistics, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("get_statistics", err)
		return getstatistics.Statistics{}, err
	}
	stats, err := getstatistics.Request(ctx, c.httpCore, fid)
	if err != nil {
		c.logCallError("get_statistics", err, "fid", fid)
		return getstatistics.Statistics{}, err
	}
	return stats, nil
}

// GetRecomStatus returns the monthly recommend quota, mirroring
// Client.get_recom_status.
func (c *Client) GetRecomStatus(ctx context.Context, ref ForumRef) (getrecomstatus.RecomStatus, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("get_recom_status", err)
		return getrecomstatus.RecomStatus{}, err
	}
	status, err := getrecomstatus.Request(ctx, c.httpCore, fid)
	if err != nil {
		c.logCallError("get_recom_status", err, "fid", fid)
		return getrecomstatus.RecomStatus{}, err
	}
	return status, nil
}

// GetUserForumInfo returns the information of a user in a forum, mirroring
// Client.get_user_forum_info.
func (c *Client) GetUserForumInfo(ctx context.Context, ref ForumRef, id UserRef) (getuserforuminfo.UserForumInfo, error) {
	if ref.FName == "" && ref.FID == 0 || id.IsZero() {
		logging.GetLogger().Warn("GetUserForumInfo: null input")
		return getuserforuminfo.UserForumInfo{}, nil
	}
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("get_user_forum_info", err)
		return getuserforuminfo.UserForumInfo{}, err
	}
	portrait, err := c.resolvePortrait(ctx, id)
	if err != nil {
		c.logCallError("get_user_forum_info", err)
		return getuserforuminfo.UserForumInfo{}, err
	}
	if portrait == "" {
		return getuserforuminfo.UserForumInfo{}, nil
	}
	info, err := getuserforuminfo.Request(ctx, c.httpCore, fid, portrait)
	if err != nil {
		c.logCallError("get_user_forum_info", err, "fid", fid)
		return getuserforuminfo.UserForumInfo{}, err
	}
	return info, nil
}

// SearchExact searches within a forum, mirroring Client.search_exact.
func (c *Client) SearchExact(ctx context.Context, ref ForumRef, query string, pn, rn int64, searchType enums.SearchType, onlyThread bool) (searchexact.ExactSearches, error) {
	fname, err := c.fetchFNameOrFName(ctx, ref)
	if err != nil {
		c.logCallError("search_exact", err)
		return searchexact.ExactSearches{}, err
	}
	searches, err := searchexact.Request(ctx, c.httpCore, fname, query, pn, rn, searchType, onlyThread)
	if err != nil {
		c.logCallError("search_exact", err, "fname", fname)
		return searchexact.ExactSearches{}, err
	}
	return searches, nil
}

// GetBawuPerm returns the permissions assigned to a moderator, mirroring
// Client.get_bawu_perm.
func (c *Client) GetBawuPerm(ctx context.Context, ref ForumRef, id UserRef) (getbawuperm.BawuPerm, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("get_bawu_perm", err)
		return getbawuperm.BawuPerm{}, err
	}
	portrait, err := c.resolvePortrait(ctx, id)
	if err != nil {
		c.logCallError("get_bawu_perm", err)
		return getbawuperm.BawuPerm{}, err
	}
	perm, err := getbawuperm.Request(ctx, c.httpCore, fid, portrait)
	if err != nil {
		c.logCallError("get_bawu_perm", err, "fid", fid)
		return getbawuperm.BawuPerm{}, err
	}
	return perm, nil
}

// GetFollowForumsPc returns the forums followed by a user, mirroring
// Client.get_follow_forums_pc.
func (c *Client) GetFollowForumsPc(ctx context.Context, id UserRef, pn, rn int64) (getfollowforumspc.PcFollowForums, error) {
	portrait, err := c.resolvePortrait(ctx, id)
	if err != nil {
		c.logCallError("get_follow_forums_pc", err)
		return getfollowforumspc.PcFollowForums{}, err
	}
	forums, err := getfollowforumspc.Request(ctx, c.httpCore, portrait, pn, rn)
	if err != nil {
		c.logCallError("get_follow_forums_pc", err, "portrait", portrait)
		return getfollowforumspc.PcFollowForums{}, err
	}
	return forums, nil
}

// GetSelfFollowForums returns the forums followed by the account, mirroring
// Client.get_self_follow_forums.
func (c *Client) GetSelfFollowForums(ctx context.Context, pn, rn int64) (getselffollowforums.SelfFollowForums, error) {
	forums, err := getselffollowforums.Request(ctx, c.httpCore, pn, rn)
	if err != nil {
		c.logCallError("get_self_follow_forums", err)
		return getselffollowforums.SelfFollowForums{}, err
	}
	return forums, nil
}

// GetUnblockAppeals returns the unblock appeal list, mirroring
// Client.get_unblock_appeals.
func (c *Client) GetUnblockAppeals(ctx context.Context, ref ForumRef, pn, rn int64) (getunblockappeals.Appeals, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("get_unblock_appeals", err)
		return getunblockappeals.Appeals{}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("get_unblock_appeals", err)
		return getunblockappeals.Appeals{}, err
	}
	appeals, err := getunblockappeals.Request(ctx, c.httpCore, fid, pn, rn)
	if err != nil {
		c.logCallError("get_unblock_appeals", err, "fid", fid)
		return getunblockappeals.Appeals{}, err
	}
	return appeals, nil
}

// SearchGlobal searches the whole forum site, mirroring Client.search_global.
func (c *Client) SearchGlobal(ctx context.Context, word string, pn, rn, sort int64) (searchglobal.GlobalSearches, error) {
	searches, err := searchglobal.Request(ctx, c.httpCore, word, pn, rn, sort)
	if err != nil {
		c.logCallError("search_global", err, "word", word)
		return searchglobal.GlobalSearches{}, err
	}
	return searches, nil
}

// GetRecovers returns the recoverable posts, mirroring Client.get_recovers.
func (c *Client) GetRecovers(ctx context.Context, ref ForumRef, pn, rn int64, id UserRef) (getrecovers.Recovers, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("get_recovers", err)
		return getrecovers.Recovers{}, err
	}
	var userID int64
	if !id.IsZero() {
		if userID, err = c.resolveUserID(ctx, id); err != nil {
			c.logCallError("get_recovers", err)
			return getrecovers.Recovers{}, err
		}
	}
	recovers, err := getrecovers.Request(ctx, c.httpCore, fid, userID, pn, rn)
	if err != nil {
		c.logCallError("get_recovers", err, "fid", fid)
		return getrecovers.Recovers{}, err
	}
	return recovers, nil
}

// mergeUserInto copies the non-zero fields of src into dst, mirroring the
// `self._user |= user` merge of the Python client.
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

// initSelfinfoInitNickname mirrors the private Client.__init_selfinfo_initNickname.
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

// GetSelfInfo returns the cached information of the account, filling the missing
// fields on demand. It mirrors Client.get_self_info.
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

// GetBawuBlacklist returns the forum-backend blacklist, mirroring
// Client.get_bawu_blacklist.
func (c *Client) GetBawuBlacklist(ctx context.Context, ref ForumRef, pn int64) (getbawublacklist.BawuBlacklistUsers, error) {
	fname, err := c.fetchFNameOrFName(ctx, ref)
	if err != nil {
		c.logCallError("get_bawu_blacklist", err)
		return getbawublacklist.BawuBlacklistUsers{}, err
	}
	users, err := getbawublacklist.Request(ctx, c.httpCore, fname, pn)
	if err != nil {
		c.logCallError("get_bawu_blacklist", err, "fname", fname)
		return getbawublacklist.BawuBlacklistUsers{}, err
	}
	return users, nil
}

// GetBawuMemberlist returns the member list of a forum, mirroring
// Client.get_bawu_memberlist.
func (c *Client) GetBawuMemberlist(ctx context.Context, ref ForumRef, pn int64, searchValue string) (getbawumemberlist.BawuListMemberUsers, error) {
	fname, err := c.fetchFNameOrFName(ctx, ref)
	if err != nil {
		c.logCallError("get_bawu_memberlist", err)
		return getbawumemberlist.BawuListMemberUsers{}, err
	}
	users, err := getbawumemberlist.Request(ctx, c.httpCore, fname, pn, searchValue)
	if err != nil {
		c.logCallError("get_bawu_memberlist", err, "fname", fname)
		return getbawumemberlist.BawuListMemberUsers{}, err
	}
	return users, nil
}

// GetBawuPostlogs returns the post-management logs, mirroring
// Client.get_bawu_postlogs.
func (c *Client) GetBawuPostlogs(
	ctx context.Context, ref ForumRef, pn int64, searchValue string, searchType enums.BawuSearchType,
	startDT, endDT *time.Time, opType int64,
) (getbawupostlogs.BawuPostLogs, error) {
	fname, err := c.fetchFNameOrFName(ctx, ref)
	if err != nil {
		c.logCallError("get_bawu_postlogs", err)
		return getbawupostlogs.BawuPostLogs{}, err
	}
	logs, err := getbawupostlogs.Request(ctx, c.httpCore, fname, pn, searchValue, searchType, startDT, endDT, opType)
	if err != nil {
		c.logCallError("get_bawu_postlogs", err, "fname", fname)
		return getbawupostlogs.BawuPostLogs{}, err
	}
	return logs, nil
}

// GetBawuUserlogs returns the user-management logs, mirroring
// Client.get_bawu_userlogs.
func (c *Client) GetBawuUserlogs(
	ctx context.Context, ref ForumRef, pn int64, searchValue string, searchType enums.BawuSearchType,
	startDT, endDT *time.Time, opType int64,
) (getbawuuserlogs.BawuUserLogs, error) {
	fname, err := c.fetchFNameOrFName(ctx, ref)
	if err != nil {
		c.logCallError("get_bawu_userlogs", err)
		return getbawuuserlogs.BawuUserLogs{}, err
	}
	logs, err := getbawuuserlogs.Request(ctx, c.httpCore, fname, pn, searchValue, searchType, startDT, endDT, opType)
	if err != nil {
		c.logCallError("get_bawu_userlogs", err, "fname", fname)
		return getbawuuserlogs.BawuUserLogs{}, err
	}
	return logs, nil
}

// GetMemberUsers returns the latest members of a forum, mirroring
// Client.get_member_users.
func (c *Client) GetMemberUsers(ctx context.Context, ref ForumRef, pn int64) (getmemberusers.MemberUsers, error) {
	fname, err := c.fetchFNameOrFName(ctx, ref)
	if err != nil {
		c.logCallError("get_member_users", err)
		return getmemberusers.MemberUsers{}, err
	}
	users, err := getmemberusers.Request(ctx, c.httpCore, fname, pn)
	if err != nil {
		c.logCallError("get_member_users", err, "fname", fname)
		return getmemberusers.MemberUsers{}, err
	}
	return users, nil
}

// GetRankForums returns the sign-in ranking of forums, mirroring
// Client.get_rank_forums.
func (c *Client) GetRankForums(ctx context.Context, ref ForumRef, pn int64, rankType enums.RankForumType) (getrankforums.RankForums, error) {
	fname, err := c.fetchFNameOrFName(ctx, ref)
	if err != nil {
		c.logCallError("get_rank_forums", err)
		return getrankforums.RankForums{}, err
	}
	forums, err := getrankforums.Request(ctx, c.httpCore, fname, pn, rankType)
	if err != nil {
		c.logCallError("get_rank_forums", err, "fname", fname)
		return getrankforums.RankForums{}, err
	}
	return forums, nil
}

// GetRankUsers returns the level-ranking users of a forum, mirroring
// Client.get_rank_users.
func (c *Client) GetRankUsers(ctx context.Context, ref ForumRef, pn int64) (getrankusers.RankUsers, error) {
	fname, err := c.fetchFNameOrFName(ctx, ref)
	if err != nil {
		c.logCallError("get_rank_users", err)
		return getrankusers.RankUsers{}, err
	}
	users, err := getrankusers.Request(ctx, c.httpCore, fname, pn)
	if err != nil {
		c.logCallError("get_rank_users", err, "fname", fname)
		return getrankusers.RankUsers{}, err
	}
	return users, nil
}

// GetBlocks returns the blocked users pending unblock, mirroring
// Client.get_blocks.
func (c *Client) GetBlocks(ctx context.Context, ref ForumRef, name string, pn int64) (getblocks.Blocks, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("get_blocks", err)
		return getblocks.Blocks{}, err
	}
	blocks, err := getblocks.Request(ctx, c.httpCore, fid, name, pn)
	if err != nil {
		c.logCallError("get_blocks", err, "fid", fid)
		return getblocks.Blocks{}, err
	}
	return blocks, nil
}

// GetImage fetches and decodes a static image, mirroring Client.get_image.
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

// GetImageBytes fetches the raw bytes of a static image, mirroring
// Client.get_image_bytes.
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

// Disagree dislikes a thread or reply, mirroring Client.disagree.
func (c *Client) Disagree(ctx context.Context, tid, pid int64, isComment bool) (exception.BoolResponse, error) {
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("disagree", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := agree.Request(ctx, c.httpCore, tid, pid, isComment, true, false); err != nil {
		c.logCallError("disagree", err, "tid", tid, "pid", pid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// Unagree removes a like, mirroring Client.unagree.
func (c *Client) Unagree(ctx context.Context, tid, pid int64, isComment bool) (exception.BoolResponse, error) {
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("unagree", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := agree.Request(ctx, c.httpCore, tid, pid, isComment, false, true); err != nil {
		c.logCallError("unagree", err, "tid", tid, "pid", pid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// Undisagree removes a dislike, mirroring Client.undisagree.
func (c *Client) Undisagree(ctx context.Context, tid, pid int64, isComment bool) (exception.BoolResponse, error) {
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("undisagree", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := agree.Request(ctx, c.httpCore, tid, pid, isComment, true, true); err != nil {
		c.logCallError("undisagree", err, "tid", tid, "pid", pid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// GetPortrait fetches the portrait image of a user, mirroring Client.get_portrait.
func (c *Client) GetPortrait(ctx context.Context, id UserRef, size string) (getimages.Image, error) {
	portrait, err := c.resolvePortrait(ctx, id)
	if err != nil {
		c.logCallError("get_portrait", err)
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
		logging.GetLogger().Warn("get_portrait: invalid size", "size", size)
		return getimages.Image{}, nil
	}

	u, err := url.Parse("http://tb.himg.baidu.com/sys/portrait" + path + "/item/" + portrait)
	if err != nil {
		c.logCallError("get_portrait", err)
		return getimages.Image{Err: err}, err
	}
	img, err := getimages.Request(ctx, c.httpCore, u)
	if err != nil {
		c.logCallError("get_portrait", err, "portrait", portrait)
		return getimages.Image{Err: err}, err
	}
	return img, nil
}

// GetSelfPosts returns the posts of the account, mirroring Client.get_self_posts.
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

// GetSelfThreads returns the threads of the account, mirroring
// Client.get_self_threads.
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

// GetUserPosts returns the posts of a user, mirroring Client.get_user_posts.
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

// GetUserPostsPc returns the posts of a user over the pc channel, mirroring
// Client.get_user_posts_pc.
func (c *Client) GetUserPostsPc(ctx context.Context, id UserRef, pn, rn int64) (getusercontentpc.PcUserPosts, error) {
	portrait, err := c.resolvePortrait(ctx, id)
	if err != nil {
		c.logCallError("get_user_posts_pc", err)
		return getusercontentpc.PcUserPosts{}, err
	}
	posts, err := getusercontentpc.Request(ctx, c.httpCore, portrait, pn, rn)
	if err != nil {
		c.logCallError("get_user_posts_pc", err, "portrait", portrait)
		return getusercontentpc.PcUserPosts{}, err
	}
	return posts, nil
}

// GetUserThreads returns the threads of a user, mirroring
// Client.get_user_threads.
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

// Hash2Image fetches the image of a Baidu image hash, mirroring
// Client.hash2image.
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
		logging.GetLogger().Warn("hash2image: invalid size", "size", size)
		return getimages.Image{}, nil
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		c.logCallError("hash2image", err)
		return getimages.Image{Err: err}, err
	}
	img, err := getimages.Request(ctx, c.httpCore, u)
	if err != nil {
		c.logCallError("hash2image", err, "hash", rawHash)
		return getimages.Image{Err: err}, err
	}
	return img, nil
}

// HideThread hides a thread, mirroring Client.hide_thread.
func (c *Client) HideThread(ctx context.Context, ref ForumRef, tid int64) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("hide_thread", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("hide_thread", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := delthread.Request(ctx, c.httpCore, fid, tid, true); err != nil {
		c.logCallError("hide_thread", err, "fid", fid, "tid", tid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// UnhideThread unhides a thread, mirroring Client.unhide_thread.
func (c *Client) UnhideThread(ctx context.Context, ref ForumRef, tid int64) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("unhide_thread", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("unhide_thread", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := recover.Request(ctx, c.httpCore, fid, tid, 0, true); err != nil {
		c.logCallError("unhide_thread", err, "fid", fid, "tid", tid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// SetThreadPrivate hides a thread or a reply, mirroring Client.set_thread_private.
func (c *Client) SetThreadPrivate(ctx context.Context, ref ForumRef, tid, pid int64) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("set_thread_private", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := setthreadprivacy.Request(ctx, c.httpCore, fid, tid, pid, true); err != nil {
		c.logCallError("set_thread_private", err, "fid", fid, "tid", tid, "pid", pid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// SetThreadPublic unhides a thread or a reply, mirroring Client.set_thread_public.
func (c *Client) SetThreadPublic(ctx context.Context, ref ForumRef, tid, pid int64) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("set_thread_public", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := setthreadprivacy.Request(ctx, c.httpCore, fid, tid, pid, false); err != nil {
		c.logCallError("set_thread_public", err, "fid", fid, "tid", tid, "pid", pid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// RecoverPost recovers a reply, mirroring Client.recover_post.
func (c *Client) RecoverPost(ctx context.Context, ref ForumRef, pid int64) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("recover_post", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("recover_post", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := recover.Request(ctx, c.httpCore, fid, 0, pid, false); err != nil {
		c.logCallError("recover_post", err, "fid", fid, "pid", pid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// RecoverThread recovers a thread, mirroring Client.recover_thread.
func (c *Client) RecoverThread(ctx context.Context, ref ForumRef, tid int64) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("recover_thread", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("recover_thread", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := recover.Request(ctx, c.httpCore, fid, tid, 0, false); err != nil {
		c.logCallError("recover_thread", err, "fid", fid, "tid", tid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// TiebaUid2UserInfo resolves a tieba uid to a user, mirroring
// Client.tieba_uid2user_info.
func (c *Client) TiebaUid2UserInfo(ctx context.Context, tiebaUID int64) (tiebauid2userinfo.UserInfoTUid, error) {
	c.tryInitWebsocket(ctx)

	if c.wsCore.Status() == enums.WsStatusOpen {
		return tiebauid2userinfo.RequestWS(c.wsCore, tiebaUID)
	}
	return tiebauid2userinfo.RequestHTTP(ctx, c.httpCore, tiebaUID)
}

// Untop untops a thread, mirroring Client.untop.
func (c *Client) Untop(ctx context.Context, ref ForumRef, tid int64, isVIP bool) (exception.BoolResponse, error) {
	fname, fid, err := c.resolveForumBoth(ctx, ref)
	if err != nil {
		c.logCallError("untop", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("untop", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := top.Request(ctx, c.httpCore, fname, fid, tid, isVIP, false); err != nil {
		c.logCallError("untop", err, "fid", fid, "tid", tid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// JoinChatroom joins a chatroom, mirroring Client.join_chatroom.
func (c *Client) JoinChatroom(ctx context.Context, roomID int64) (exception.BoolResponse, error) {
	if c.user.UserID == 0 {
		if _, err := c.GetSelfInfo(ctx, enums.ReqUInfoAll); err != nil {
			c.logCallError("join_chatroom", err)
			return exception.BoolResponse{Err: err}, err
		}
	}
	if err := c.initBLCP(ctx); err != nil {
		c.logCallError("join_chatroom", err)
		return exception.BoolResponse{Err: err}, err
	}

	if _, err := c.blcpCore.JoinChatRoom(ctx, roomID); err != nil {
		err = fmt.Errorf("aiotieba: 加入房间失败: %w", err)
		c.logCallError("join_chatroom", err, "room_id", roomID)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// initBLCP brings the BLCP session to the logged-in state, mirroring the
// private Client._init_blcp.
func (c *Client) initBLCP(ctx context.Context) error {
	if c.blcpCore.Status() == -1 {
		if err := c.blcpCore.Connect(ctx); err != nil {
			return err
		}
	}
	if c.blcpCore.Status() == 0 {
		if err := c.blcpCore.Login(ctx); err != nil {
			return err
		}
	}
	if c.blcpCore.Status() != 1 {
		return errors.New("aiotieba: BLCP 登录失败")
	}
	return nil
}

// SendChatroomMsg sends a plain text message to a forum group, mirroring
// Client.send_chatroom_msg.
//
// The cached self information (c.user) must carry user_id and portrait, which
// the Python client populates through get_self_info.
func (c *Client) SendChatroomMsg(ctx context.Context, chatroomID, fid int64, text string, atUserIDs []int64, robotc int64) (exception.BoolResponse, error) {
	if c.user.UserID == 0 || c.user.Portrait == "" {
		err := errors.New("aiotieba: 自账号信息未加载，请先调用 GetSelfInfo 或 Login")
		c.logCallError("send_chatroom_msg", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.initBLCP(ctx); err != nil {
		c.logCallError("send_chatroom_msg", err)
		return exception.BoolResponse{Err: err}, err
	}

	levelInfo, err := getforumlevel.RequestHTTP(ctx, c.httpCore, fid)
	if err != nil {
		c.logCallError("send_chatroom_msg", err, "fid", fid)
		return exception.BoolResponse{Err: err}, err
	}

	// Resolve the @ targets, mirroring the atdata construction.
	atdata := []map[string]any{}
	for i, userID := range atUserIDs {
		user, err := c.getUinfoProfile(ctx, ByUserID(userID))
		if err != nil {
			c.logCallError("send_chatroom_msg", err, "at_user_id", userID)
			return exception.BoolResponse{Err: err}, err
		}
		if user.Portrait == "" || user.NickName() == "" {
			if user, err = c.getUinfoProfile(ctx, ByUserID(userID)); err != nil {
				c.logCallError("send_chatroom_msg", err, "at_user_id", userID)
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
		c.logCallError("send_chatroom_msg", err, "chatroom_id", chatroomID)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// SetMsgReaded marks a private message as read, mirroring
// Client.set_msg_readed.
func (c *Client) SetMsgReaded(ctx context.Context, message getgroupmsg.WsMessage) (exception.BoolResponse, error) {
	if err := c.forceWebsocket(ctx); err != nil {
		c.logCallError("set_msg_readed", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := setmsgreaded.Request(c.wsCore, message); err != nil {
		c.logCallError("set_msg_readed", err, "msg_id", message.MsgID)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// GetGroupMsg returns the messages of the given websocket groups, mirroring
// Client.get_group_msg.
//
// The API is only available over the websocket transport, so a connection is
// established unconditionally (the _force_websocket decorator of the Python
// client) and its failure is reported to the caller.
func (c *Client) GetGroupMsg(ctx context.Context, groupIDs []int64, getType int64) (getgroupmsg.WsMsgGroups, error) {
	if err := c.forceWebsocket(ctx); err != nil {
		c.logCallError("get_group_msg", err)
		return getgroupmsg.WsMsgGroups{}, err
	}

	groups, err := getgroupmsg.Request(c.wsCore, groupIDs, getType)
	if err != nil {
		c.logCallError("get_group_msg", err, "group_ids", groupIDs)
		return getgroupmsg.WsMsgGroups{}, err
	}
	return groups, nil
}

// GetBawuInfo returns the moderator team of a forum, mirroring
// Client.get_bawu_info.
func (c *Client) GetBawuInfo(ctx context.Context, ref ForumRef) (getbawuinfo.BawuInfo, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("get_bawu_info", err)
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
		c.logCallError("get_bawu_info", err, "fid", fid)
		return getbawuinfo.BawuInfo{}, err
	}
	return info, nil
}

// GetBlacklistOld returns the legacy user blacklist, mirroring
// Client.get_blacklist_old.
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
		c.logCallError("get_blacklist_old", err, "pn", pn)
		return getblacklistold.BlacklistOldUsers{}, err
	}
	return users, nil
}

// AddPoll casts a vote, mirroring Client.add_poll.
func (c *Client) AddPoll(ctx context.Context, tid int64, options []int64) (exception.BoolResponse, error) {
	c.tryInitWebsocket(ctx)

	var err error
	if c.wsCore.Status() == enums.WsStatusOpen {
		err = addpoll.RequestWS(c.wsCore, tid, options)
	} else {
		err = addpoll.RequestHTTP(ctx, c.httpCore, tid, options)
	}
	if err != nil {
		c.logCallError("add_poll", err, "tid", tid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// GetLastReplyers returns the threads of a forum together with their last
// replier, mirroring Client.get_last_replyers.
//
// The legacy endpoint is mainly used to detect thread necromancy and does not
// expose the full thread information.
func (c *Client) GetLastReplyers(ctx context.Context, ref ForumRef, pn, rn int32, sort enums.ThreadSortType, isGood bool) (getlastreplyers.ThreadsLP, error) {
	fname, err := c.fetchFNameOrFName(ctx, ref)
	if err != nil {
		c.logCallError("get_last_replyers", err)
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
		c.logCallError("get_last_replyers", err, "fname", fname, "pn", pn)
		return getlastreplyers.ThreadsLP{}, err
	}
	return threads, nil
}

// GetReplys returns the replies received by the logged in account, mirroring
// Client.get_replys.
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
		c.logCallError("get_replys", err, "pn", pn)
		return getreplys.Replys{}, err
	}
	return replys, nil
}

// GetSquareForums returns the forum square list, mirroring
// Client.get_square_forums.
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
		c.logCallError("get_square_forums", err, "cname", cname, "pn", pn)
		return getsquareforums.SquareForums{}, err
	}
	return forums, nil
}

// GetDislikeForums returns the forums hidden from the home page
// recommendations, mirroring Client.get_dislike_forums.
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
		c.logCallError("get_dislike_forums", err, "pn", pn)
		return getdislikeforums.DislikeForums{}, err
	}
	return forums, nil
}

// TiebaUID2UserInfo returns the information of a user by their tieba uid,
// mirroring Client.tieba_uid2user_info.
//
// Note that tieba_uid differs from the legacy user_id.
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
		c.logCallError("tieba_uid2user_info", err, "tieba_uid", tiebaUID)
		return tiebauid2userinfo.UserInfoTUid{}, err
	}
	return user, nil
}

// resolveForumBoth mirrors the `if isinstance(fname_or_fid, str)` idiom used by
// the APIs that need both the forum name and the fid.
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

// resolvePortrait mirrors the `user = await self.get_user_info(id_,
// ReqUInfo.PORTRAIT)` idiom.
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

// resolveUserID mirrors the ReqUInfo.USER_ID variant of the same idiom.
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

// resolveUserName mirrors the ReqUInfo.USER_NAME variant of the same idiom.
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

// AddBawu adds a moderator to a forum, mirroring Client.add_bawu.
func (c *Client) AddBawu(ctx context.Context, ref ForumRef, id UserRef, bawuType enums.BawuType) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("add_bawu", err)
		return exception.BoolResponse{Err: err}, err
	}
	userName, err := c.resolveUserName(ctx, id)
	if err != nil {
		c.logCallError("add_bawu", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("add_bawu", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := addbawu.Request(ctx, c.httpCore, fid, userName, bawuType); err != nil {
		c.logCallError("add_bawu", err, "fid", fid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// DelBawu removes a moderator from a forum, mirroring Client.del_bawu.
func (c *Client) DelBawu(ctx context.Context, ref ForumRef, id UserRef, bawuType enums.BawuType) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("del_bawu", err)
		return exception.BoolResponse{Err: err}, err
	}
	portrait, err := c.resolvePortrait(ctx, id)
	if err != nil {
		c.logCallError("del_bawu", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := delbawu.Request(ctx, c.httpCore, fid, portrait, bawuType); err != nil {
		c.logCallError("del_bawu", err, "fid", fid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// SetBawuPerm assigns permissions to a moderator, mirroring
// Client.set_bawu_perm.
func (c *Client) SetBawuPerm(ctx context.Context, ref ForumRef, id UserRef, perms enums.BawuPermType) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("set_bawu_perm", err)
		return exception.BoolResponse{Err: err}, err
	}
	portrait, err := c.resolvePortrait(ctx, id)
	if err != nil {
		c.logCallError("set_bawu_perm", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := setbawuperm.Request(ctx, c.httpCore, fid, portrait, perms); err != nil {
		c.logCallError("set_bawu_perm", err, "fid", fid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// Block bans a user from a forum, mirroring Client.block.
func (c *Client) Block(ctx context.Context, ref ForumRef, id UserRef, day int64, reason string) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("block", err)
		return exception.BoolResponse{Err: err}, err
	}
	portrait, err := c.resolvePortrait(ctx, id)
	if err != nil {
		c.logCallError("block", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("block", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := block.Request(ctx, c.httpCore, fid, portrait, day, reason); err != nil {
		c.logCallError("block", err, "fid", fid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// Unblock lifts a ban from a user, mirroring Client.unblock.
func (c *Client) Unblock(ctx context.Context, ref ForumRef, id UserRef) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("unblock", err)
		return exception.BoolResponse{Err: err}, err
	}
	userID, err := c.resolveUserID(ctx, id)
	if err != nil {
		c.logCallError("unblock", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("unblock", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := unblock.Request(ctx, c.httpCore, fid, userID); err != nil {
		c.logCallError("unblock", err, "fid", fid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// AddBawuBlacklist adds a user to the forum blacklist, mirroring
// Client.add_bawu_blacklist.
func (c *Client) AddBawuBlacklist(ctx context.Context, ref ForumRef, id UserRef) (exception.BoolResponse, error) {
	fname, err := c.fetchFNameOrFName(ctx, ref)
	if err != nil {
		c.logCallError("add_bawu_blacklist", err)
		return exception.BoolResponse{Err: err}, err
	}
	userID, err := c.resolveUserID(ctx, id)
	if err != nil {
		c.logCallError("add_bawu_blacklist", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("add_bawu_blacklist", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := addbawublacklist.Request(ctx, c.httpCore, fname, userID); err != nil {
		c.logCallError("add_bawu_blacklist", err, "fname", fname)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// DelBawuBlacklist removes a user from the forum blacklist, mirroring
// Client.del_bawu_blacklist.
func (c *Client) DelBawuBlacklist(ctx context.Context, ref ForumRef, id UserRef) (exception.BoolResponse, error) {
	fname, err := c.fetchFNameOrFName(ctx, ref)
	if err != nil {
		c.logCallError("del_bawu_blacklist", err)
		return exception.BoolResponse{Err: err}, err
	}
	userID, err := c.resolveUserID(ctx, id)
	if err != nil {
		c.logCallError("del_bawu_blacklist", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("del_bawu_blacklist", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := delbawublacklist.Request(ctx, c.httpCore, fname, userID); err != nil {
		c.logCallError("del_bawu_blacklist", err, "fname", fname)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// DelThread removes a thread, mirroring Client.del_thread.
func (c *Client) DelThread(ctx context.Context, ref ForumRef, tid int64) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("del_thread", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("del_thread", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := delthread.Request(ctx, c.httpCore, fid, tid, false); err != nil {
		c.logCallError("del_thread", err, "fid", fid, "tid", tid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// DelThreads removes several threads, mirroring Client.del_threads.
func (c *Client) DelThreads(ctx context.Context, ref ForumRef, tids []int64, block bool) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("del_threads", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("del_threads", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := delthreads.Request(ctx, c.httpCore, fid, tids, block); err != nil {
		c.logCallError("del_threads", err, "fid", fid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// DelPost removes a reply, mirroring Client.del_post.
func (c *Client) DelPost(ctx context.Context, ref ForumRef, tid, pid int64) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("del_post", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("del_post", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := delpost.Request(ctx, c.httpCore, fid, tid, pid); err != nil {
		c.logCallError("del_post", err, "fid", fid, "tid", tid, "pid", pid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// DelPosts removes several replies, mirroring Client.del_posts.
func (c *Client) DelPosts(ctx context.Context, ref ForumRef, tid int64, pids []int64, block bool) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("del_posts", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("del_posts", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := delposts.Request(ctx, c.httpCore, fid, tid, pids, block); err != nil {
		c.logCallError("del_posts", err, "fid", fid, "tid", tid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// Recover recovers a thread or a reply, mirroring Client.recover.
func (c *Client) Recover(ctx context.Context, ref ForumRef, tid, pid int64, isHide bool) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("recover", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("recover", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := recover.Request(ctx, c.httpCore, fid, tid, pid, isHide); err != nil {
		c.logCallError("recover", err, "fid", fid, "tid", tid, "pid", pid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// GetCID returns the id of a good category, mirroring Client.get_cid.
func (c *Client) GetCID(ctx context.Context, ref ForumRef, cname string) (exception.IntResponse, error) {
	cid, err := c.fetchCID(ctx, ref, cname)
	if err != nil {
		c.logCallError("get_cid", err, "cname", cname)
		return exception.IntResponse{Err: err}, err
	}
	return exception.IntResponse{Value: int(cid)}, nil
}

// fetchCID mirrors the private Client.__get_cid.
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

// Good marks a thread as excellent, mirroring Client.good.
func (c *Client) Good(ctx context.Context, ref ForumRef, tid int64, cname string) (exception.BoolResponse, error) {
	fname, fid, err := c.resolveForumBoth(ctx, ref)
	if err != nil {
		c.logCallError("good", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("good", err)
		return exception.BoolResponse{Err: err}, err
	}

	cid, err := c.fetchCID(ctx, ref, cname)
	if err != nil {
		c.logCallError("good", err, "cname", cname)
		return exception.BoolResponse{Err: err}, err
	}

	if err := good.Request(ctx, c.httpCore, fname, fid, tid, cid); err != nil {
		c.logCallError("good", err, "fid", fid, "tid", tid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// Ungood removes the excellent mark of a thread, mirroring Client.ungood.
func (c *Client) Ungood(ctx context.Context, ref ForumRef, tid int64) (exception.BoolResponse, error) {
	fname, fid, err := c.resolveForumBoth(ctx, ref)
	if err != nil {
		c.logCallError("ungood", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("ungood", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := ungood.Request(ctx, c.httpCore, fname, fid, tid); err != nil {
		c.logCallError("ungood", err, "fid", fid, "tid", tid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// Top tops a thread, mirroring Client.top.
func (c *Client) Top(ctx context.Context, ref ForumRef, tid int64, isVIP bool) (exception.BoolResponse, error) {
	fname, fid, err := c.resolveForumBoth(ctx, ref)
	if err != nil {
		c.logCallError("top", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("top", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := top.Request(ctx, c.httpCore, fname, fid, tid, isVIP, true); err != nil {
		c.logCallError("top", err, "fid", fid, "tid", tid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// Move moves a thread to another tab, mirroring Client.move.
func (c *Client) Move(ctx context.Context, ref ForumRef, tid, toTabID, fromTabID int64) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("move", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("move", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := move.Request(ctx, c.httpCore, fid, tid, toTabID, fromTabID); err != nil {
		c.logCallError("move", err, "fid", fid, "tid", tid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// Recommend pushes a thread to the personalised home page, mirroring
// Client.recommend.
func (c *Client) Recommend(ctx context.Context, ref ForumRef, tid int64) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("recommend", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := recommend.Request(ctx, c.httpCore, fid, tid); err != nil {
		c.logCallError("recommend", err, "fid", fid, "tid", tid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// SetThreadPrivacy hides or unhides a thread or a reply, mirroring
// Client.set_thread_privacy.
func (c *Client) SetThreadPrivacy(ctx context.Context, ref ForumRef, tid, pid int64, isHide bool) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("set_thread_privacy", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := setthreadprivacy.Request(ctx, c.httpCore, fid, tid, pid, isHide); err != nil {
		c.logCallError("set_thread_privacy", err, "fid", fid, "tid", tid, "pid", pid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// HandleUnblockAppeals accepts or refuses unblock appeals, mirroring
// Client.handle_unblock_appeals.
func (c *Client) HandleUnblockAppeals(ctx context.Context, ref ForumRef, appealIDs []int64, refuse bool) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("handle_unblock_appeals", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("handle_unblock_appeals", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := handleunblockappeals.Request(ctx, c.httpCore, fid, appealIDs, refuse); err != nil {
		c.logCallError("handle_unblock_appeals", err, "fid", fid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// Agree likes a thread or a reply, mirroring Client.agree.
func (c *Client) Agree(ctx context.Context, tid, pid int64, isComment bool) (exception.BoolResponse, error) {
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("agree", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := agree.Request(ctx, c.httpCore, tid, pid, isComment, false, false); err != nil {
		c.logCallError("agree", err, "tid", tid, "pid", pid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// FollowUser follows a user, mirroring Client.follow_user.
func (c *Client) FollowUser(ctx context.Context, id UserRef) (exception.BoolResponse, error) {
	portrait, err := c.resolvePortrait(ctx, id)
	if err != nil {
		c.logCallError("follow_user", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("follow_user", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := followuser.Request(ctx, c.httpCore, portrait); err != nil {
		c.logCallError("follow_user", err, "portrait", portrait)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// UnfollowUser unfollows a user, mirroring Client.unfollow_user.
func (c *Client) UnfollowUser(ctx context.Context, id UserRef) (exception.BoolResponse, error) {
	portrait, err := c.resolvePortrait(ctx, id)
	if err != nil {
		c.logCallError("unfollow_user", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("unfollow_user", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := unfollowuser.Request(ctx, c.httpCore, portrait); err != nil {
		c.logCallError("unfollow_user", err, "portrait", portrait)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// RemoveFan removes a fan, mirroring Client.remove_fan.
func (c *Client) RemoveFan(ctx context.Context, id UserRef) (exception.BoolResponse, error) {
	userID, err := c.resolveUserID(ctx, id)
	if err != nil {
		c.logCallError("remove_fan", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("remove_fan", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := removefan.Request(ctx, c.httpCore, userID); err != nil {
		c.logCallError("remove_fan", err, "user_id", userID)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// AddBlacklistOld adds a user to the legacy user blacklist, mirroring
// Client.add_blacklist_old.
func (c *Client) AddBlacklistOld(ctx context.Context, id UserRef) (exception.BoolResponse, error) {
	userID, err := c.resolveUserID(ctx, id)
	if err != nil {
		c.logCallError("add_blacklist_old", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := addblacklistold.Request(ctx, c.httpCore, userID); err != nil {
		c.logCallError("add_blacklist_old", err, "user_id", userID)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// DelBlacklistOld removes a user from the legacy user blacklist, mirroring
// Client.del_blacklist_old.
func (c *Client) DelBlacklistOld(ctx context.Context, id UserRef) (exception.BoolResponse, error) {
	userID, err := c.resolveUserID(ctx, id)
	if err != nil {
		c.logCallError("del_blacklist_old", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := delblacklistold.Request(ctx, c.httpCore, userID); err != nil {
		c.logCallError("del_blacklist_old", err, "user_id", userID)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// FollowForum follows a forum, mirroring Client.follow_forum.
func (c *Client) FollowForum(ctx context.Context, ref ForumRef) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("follow_forum", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("follow_forum", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := followforum.Request(ctx, c.httpCore, fid); err != nil {
		c.logCallError("follow_forum", err, "fid", fid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// UnfollowForum unfollows a forum, mirroring Client.unfollow_forum.
func (c *Client) UnfollowForum(ctx context.Context, ref ForumRef) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("unfollow_forum", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("unfollow_forum", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := unfollowforum.Request(ctx, c.httpCore, fid); err != nil {
		c.logCallError("unfollow_forum", err, "fid", fid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// DislikeForum hides a forum from the home page recommendations, mirroring
// Client.dislike_forum.
func (c *Client) DislikeForum(ctx context.Context, ref ForumRef) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("dislike_forum", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := dislikeforum.Request(ctx, c.httpCore, fid); err != nil {
		c.logCallError("dislike_forum", err, "fid", fid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// UndislikeForum restores a forum on the home page recommendations, mirroring
// Client.undislike_forum.
func (c *Client) UndislikeForum(ctx context.Context, ref ForumRef) (exception.BoolResponse, error) {
	fid, err := c.fetchFIDOrFID(ctx, ref)
	if err != nil {
		c.logCallError("undislike_forum", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := undislikeforum.Request(ctx, c.httpCore, fid); err != nil {
		c.logCallError("undislike_forum", err, "fid", fid)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// SetProfile updates the profile of the logged in account, mirroring
// Client.set_profile.
func (c *Client) SetProfile(ctx context.Context, nickName, sign string, gender enums.Gender) (exception.BoolResponse, error) {
	if err := setprofile.Request(ctx, c.httpCore, nickName, sign, gender); err != nil {
		c.logCallError("set_profile", err)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// SetNicknameOld updates the legacy nickname, mirroring Client.set_nickname_old.
func (c *Client) SetNicknameOld(ctx context.Context, nickName string) (exception.BoolResponse, error) {
	if err := setnicknameold.Request(ctx, c.httpCore, nickName); err != nil {
		c.logCallError("set_nickname_old", err)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// SignForum signs a single forum, mirroring Client.sign_forum.
func (c *Client) SignForum(ctx context.Context, ref ForumRef) (exception.BoolResponse, error) {
	fname, err := c.fetchFNameOrFName(ctx, ref)
	if err != nil {
		c.logCallError("sign_forum", err)
		return exception.BoolResponse{Err: err}, err
	}
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("sign_forum", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := signforum.Request(ctx, c.httpCore, fname); err != nil {
		c.logCallError("sign_forum", err, "fname", fname)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// SignForums signs every followed forum, mirroring Client.sign_forums.
func (c *Client) SignForums(ctx context.Context) (exception.BoolResponse, error) {
	if err := signforums.Request(ctx, c.httpCore); err != nil {
		c.logCallError("sign_forums", err)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// SignGrowth completes the growth-level sign task, mirroring
// Client.sign_growth.
func (c *Client) SignGrowth(ctx context.Context) (exception.BoolResponse, error) {
	if err := c.InitTbs(ctx); err != nil {
		c.logCallError("sign_growth", err)
		return exception.BoolResponse{Err: err}, err
	}

	if err := signgrowth.RequestWeb(ctx, c.httpCore, "page_sign"); err != nil {
		c.logCallError("sign_growth", err)
		return exception.BoolResponse{Err: err}, err
	}
	return exception.BoolResponse{}, nil
}

// GetPostsArgs holds the optional arguments of GetPosts.
type GetPostsArgs struct {
	Pn                 int
	Rn                 int
	Sort               enums.PostSortType
	OnlyThreadAuthor   bool
	WithComments       bool
	CommentSortByAgree bool
	CommentRn          int
}

// DefaultGetPostsArgs returns the Python defaults of GetPosts.
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

// GetPosts returns the floors of a thread, mirroring Client.get_posts.
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
		c.logCallError("get_posts", err, "tid", tid, "pn", args.Pn)
		return posts, err
	}
	return posts, nil
}

// GetCommentsArgs holds the optional arguments of GetComments.
type GetCommentsArgs struct {
	Pn        int
	IsComment bool
}

// DefaultGetCommentsArgs returns the Python defaults of GetComments.
func DefaultGetCommentsArgs() GetCommentsArgs {
	return GetCommentsArgs{Pn: 1}
}

// GetComments returns the comments of a floor, mirroring Client.get_comments.
func (c *Client) GetComments(ctx context.Context, tid, pid int64, args GetCommentsArgs) (getcomments.Comments, error) {
	c.tryInitWebsocket(ctx)

	var (
		comments getcomments.Comments
		err      error
	)
	pn := int32(args.Pn)
	if c.wsCore.Status() == enums.WsStatusOpen {
		comments, err = getcomments.RequestWS(c.wsCore, tid, pid, pn, args.IsComment)
	} else {
		comments, err = getcomments.RequestHTTP(ctx, c.httpCore, tid, pid, pn, args.IsComment)
	}
	if err != nil {
		c.logCallError("get_comments", err, "tid", tid, "pid", pid, "pn", args.Pn)
		return comments, err
	}
	return comments, nil
}

// tryInitWebsocket mirrors the _try_websocket decorator: when the websocket
// initialisation fails it is logged and the call silently falls back to HTTP,
// exactly like the Python handle_exception wrapper.
func (c *Client) tryInitWebsocket(ctx context.Context) {
	if !c.tryWS {
		return
	}
	if _, err := c.InitWebsocket(ctx); err != nil {
		c.logCallError("init_websocket", err)
	}
}

// forceWebsocket establishes the websocket connection, mirroring the
// _force_websocket decorator. Unlike tryInitWebsocket the failure is returned to
// the caller instead of being swallowed.
func (c *Client) forceWebsocket(ctx context.Context) error {
	_, err := c.InitWebsocket(ctx)
	return err
}

// logCallError records a failed API call. The Python client logs every failure
// through its handle_exception decorator, so the Go port keeps the same
// behaviour at the client boundary.
func (c *Client) logCallError(apiName string, err error, args ...any) {
	attrs := append([]any{"api", apiName, "err", err}, args...)
	logging.GetLogger().Warn("tieba api call failed", attrs...)
}
