package core

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/helper"
	tbcrypto "github.com/rongyuio/aiotieba-go/helper/crypto"
)

// pimBaseURL 是贴吧 IM 参数服务的地址。
const pimBaseURL = "https://pim.baidu.com"

// pimUA 是所有 pim.baidu.com 请求使用的 User-Agent。
const pimUA = "okhttp/3.11.0"

// GenerateLCMToken 请求 LCM token。
//
// 请求 LCM token，对应 BLCPCore.generate_lcm_token。
//
// 失败必须向上抛：拿不到 token 时后续三步登录只会收到含糊的服务端错误，
// 真正的原因（网络 / 反爬 / 回包缺字段）会彻底丢失。
func (b *BLCPCore) GenerateLCMToken(ctx context.Context, cuidGalaxy2 string) (string, error) {
	ts := time.Now().UnixMilli()
	data := map[string]any{
		"app_id":      fmt.Sprint(ChatAppID),
		"app_version": consts.ChatVersion,
		"cuid":        cuidGalaxy2,
		"device_type": "android",
		"manufacture": "",
		"model_type":  "",
		"request_id":  fmt.Sprint(ts),
		"sdk_version": blcpSDKVer,
		"sign":        md5Hex(fmt.Sprint(ChatAppID) + cuidGalaxy2 + "android" + fmt.Sprint(ts)),
		"ts":          ts,
		"user_key":    "",
	}
	body, err := b.postPIM(ctx, "/rest/5.0/generate_lcm_token", "application/json", string(marshalJSON(data)), nil)
	if err != nil {
		return "", fmt.Errorf("core: requesting the LCM token: %w", err)
	}
	var reply map[string]any
	if err := json.Unmarshal(body, &reply); err != nil {
		return "", fmt.Errorf("core: decoding the LCM token response: %w", err)
	}
	token, _ := reply["token"].(string)
	if token == "" {
		return "", fmt.Errorf("core: the LCM token response carries no token%s", helper.BodySnippet(body))
	}
	return token, nil
}

// GroupChat 模拟正常请求，暂不清楚作用。
//
// 对应 BLCPCore.groupchat，永不失败。
func (b *BLCPCore) GroupChat(ctx context.Context) {
	ts := time.Now().Unix()
	data := map[string]any{
		"method":    "get_joined_groups",
		"appid":     ChatAppID,
		"timestamp": ts,
		"sign":      md5Hex(fmt.Sprint(ts) + b.Account.BDUSS() + fmt.Sprint(ChatAppID)),
	}
	_, _ = b.postPIM(ctx, "/rest/2.0/im/groupchat", "application/x-www-form-urlencoded",
		encodeValues(data), b.cookieHeaders())
}

// GroupChatV1 模拟正常请求，暂不清楚作用。
//
// 对应 BLCPCore.groupchatv1，永不失败。
func (b *BLCPCore) GroupChatV1(ctx context.Context) {
	cuid, _ := b.Account.CuidGalaxy2()
	ts := time.Now().Unix()
	data := map[string]any{
		"method":      "get_joined_groups",
		"group_type":  3,
		"appid":       ChatAppID,
		"source":      0,
		"cuid":        cuid,
		"app_version": consts.ChatVersion,
		"sdk_version": fmt.Sprint(ChatSDKVersion),
		"timestamp":   ts,
		"device_type": 2,
		"sign":        md5Hex(fmt.Sprint(ts) + b.Account.BDUSS() + fmt.Sprint(ChatAppID)),
	}
	_, _ = b.postPIM(ctx, "/rest/2.0/im/groupchatv1", "application/x-www-form-urlencoded",
		encodeValues(data), b.cookieHeaders())
}

// EnterChatroomClientRequest 模拟正常请求，暂不清楚作用。
//
// 对应 BLCPCore.enter_chatroom_client_request。
func (b *BLCPCore) EnterChatroomClientRequest(ctx context.Context, cuidGalaxy2 string, roomID int64, accountType int) (map[string]any, error) {
	data := map[string]any{
		"appid":        ChatAppID,
		"room_id":      roomID,
		"app_version":  consts.ChatVersion,
		"cuid":         cuidGalaxy2,
		"device_id":    cuidGalaxy2,
		"sdk_version":  ChatSDKVersion,
		"timestamp":    time.Now().Unix(),
		"account_type": accountType,
	}
	body, err := b.postPIM(ctx, "/rest/3.0/im/chatroom/enter_chatroom_client", "application/json",
		string(marshalJSON(signDict(data, nil))), b.cookieHeaders())
	if err != nil {
		return nil, fmt.Errorf("core: entering the chatroom: %w", err)
	}
	var reply map[string]any
	if err := json.Unmarshal(body, &reply); err != nil {
		return nil, fmt.Errorf("core: decoding the chatroom response: %w", err)
	}
	return reply, nil
}

// SendLcmRaw 发送一次业务面 Lcm JSON 请求，返回帧头错误码、错误描述与响应报文体。
//
// 与 SendLcm 不同，它不向请求体注入 client_logid 与 rpc：上游 join_chatroom 的请求体
// 只有 method、mcast_id、appid、uk 四个字段，多出来的字段会改变服务端行为。
func (b *BLCPCore) SendLcmRaw(serviceID, methodID int64, requestData map[string]any) (int32, string, map[string]any, error) {
	req := NewBLCPData(serviceID, methodID)
	req.RPCBody = BuildRPCBody(serviceID, methodID, req.CorrelationID, 0, 1)
	req.LCMBody = marshalJSON(requestData)

	rpc, body, err := b.exchange(req)
	if err != nil {
		return 0, "", nil, err
	}
	response := rpc.GetResponse()
	if response == nil {
		return 0, "", nil, errors.New("core: the BLCP response carries no RpcResponseMeta")
	}

	var payload map[string]any
	if body != nil {
		payload = body.JSON
	}
	return response.GetErrorCode(), response.GetErrorText(), payload, nil
}

// ExitChatRoom 待实现。Python 原版同样未实现该方法。
func (b *BLCPCore) ExitChatRoom(chatroomID int64, roomType int) error {
	return nil
}

// StartHeartbeat 启动周期性 Heartbeat goroutine，对应 BLCPCore.start_heartbeat。
func (b *BLCPCore) StartHeartbeat() {
	b.mu.Lock()
	if b.cancelHB != nil {
		b.cancelHB()
	}
	ctx, cancel := context.WithCancel(context.Background())
	b.cancelHB = cancel
	b.mu.Unlock()

	go func() {
		for {
			// 间隔由服务端下发，每轮重新取，收到新值后立刻生效。
			timer := time.NewTimer(b.heartbeatInterval())
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				_ = b.Heartbeat()
			}
		}
	}()
}

func (b *BLCPCore) postPIM(ctx context.Context, path, contentType, payload string, extra map[string]string) ([]byte, error) {
	u, err := url.Parse(pimBaseURL + path)
	if err != nil {
		return nil, fmt.Errorf("core: parsing %s: %w", path, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), strings.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("core: building the pim request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", pimUA)
	req.Header.Set("Host", "pim.baidu.com")
	for k, v := range extra {
		req.Header.Set(k, v)
	}

	resp, err := b.NetCore.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("core: pim returned status %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func (b *BLCPCore) cookieHeaders() map[string]string {
	return map[string]string{"Cookie": "BDUSS=" + b.Account.BDUSS()}
}

// signDict 用给定盐值对参数字典签名。
//
// 注意：Python 的辅助函数把 dict 传给 sign()，而 sign() 会遍历排序后的键，键长超过
// 两个字符时会抛错。本移植版实现了预期行为（按键排序条目），而不是复现该崩溃。
func signDict(data map[string]any, salt []byte) map[string]any {
	params := make([]tbcrypto.Param, 0, len(data))
	for k, v := range data {
		params = append(params, tbcrypto.Param{Key: k, Value: v})
	}
	out := make(map[string]any, len(data)+1)
	for k, v := range data {
		out[k] = v
	}
	out["sign"] = tbcrypto.ComputeSign(params, salt)
	return out
}

func encodeValues(data map[string]any) string {
	values := make(url.Values, len(data))
	for k, v := range data {
		values.Set(k, fmt.Sprint(v))
	}
	return values.Encode()
}

func md5Hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}
