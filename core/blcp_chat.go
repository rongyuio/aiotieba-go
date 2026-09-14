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

	"github.com/rongyuio/aiotieba/consts"
	tbcrypto "github.com/rongyuio/aiotieba/helper/crypto"
)

// pimBaseURL is the host of the Tieba IM parameter service.
const pimBaseURL = "https://pim.baidu.com"

// pimUA is the user agent used by every pim.baidu.com request.
const pimUA = "okhttp/3.11.0"

// GenerateLCMToken requests an LCM token, mirroring
// BLCPCore.generate_lcm_token.
//
// Like the Python original every error is swallowed and an empty token is
// returned instead.
func (b *BLCPCore) GenerateLCMToken(ctx context.Context, cuidGalaxy2 string) string {
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
		return ""
	}
	var reply map[string]any
	if err := json.Unmarshal(body, &reply); err != nil {
		return ""
	}
	token, _ := reply["token"].(string)
	return token
}

// GroupChat simulates a normal client request of unknown purpose. It mirrors
// BLCPCore.groupchat and never fails.
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

// GroupChatV1 simulates a normal client request of unknown purpose. It mirrors
// BLCPCore.groupchatv1 and never fails.
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

// EnterChatroomClientRequest simulates a normal client request of unknown
// purpose. It mirrors BLCPCore.enter_chatroom_client_request.
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

// FetchMcastMsgClientRequest fetches chat history. It mirrors
// BLCPCore.fetch_mcast_msg_client_request.
func (b *BLCPCore) FetchMcastMsgClientRequest(ctx context.Context, cuidGalaxy2 string, roomID int64, accountType int) (map[string]any, error) {
	extInfo := url.QueryEscape(string(marshalJSON(map[string]any{
		"last_callback_msg_id": 0,
		"cast_id":              0,
		"local_ts":             0,
		"latest_msg_id":        0,
	})))
	data := map[string]any{
		"appid":        ChatAppID,
		"mcast_id":     roomID,
		"msgid_begin":  0,
		"msgid_end":    9223372036854775807,
		"count":        -60,
		"category":     4,
		"app_version":  consts.ChatVersion,
		"sdk_version":  ChatSDKVersion,
		"device_id":    cuidGalaxy2,
		"device_type":  2,
		"from_action":  1,
		"ext_info":     extInfo,
		"timestamp":    time.Now().Unix(),
		"account_type": accountType,
	}
	body, err := b.postPIM(ctx, "/rest/3.0/im/fetch_mcast_msg_client", "application/json",
		string(marshalJSON(signDict(data, nil))), b.cookieHeaders())
	if err != nil {
		return nil, fmt.Errorf("core: fetching the chat history: %w", err)
	}
	var reply map[string]any
	if err := json.Unmarshal(body, &reply); err != nil {
		return nil, fmt.Errorf("core: decoding the chat history response: %w", err)
	}
	return reply, nil
}

// JoinChatRoom joins a chatroom and starts the heartbeat, mirroring
// BLCPCore.joinChatRoom.
func (b *BLCPCore) JoinChatRoom(ctx context.Context, chatroomID int64) (bool, error) {
	if b.Status() != 1 {
		return false, errors.New("core: BLCP is not logged in")
	}

	req := NewBLCPData(3, 201)
	req.RPCBody = BuildRPCBody(3, 201, req.CorrelationID, 0, 1)
	req.LCMBody = marshalJSON(map[string]any{
		"method":       201,
		"mcast_id":     chatroomID,
		"appid":        ChatAppID,
		"uk":           b.UK(),
		"origin_id":    b.TriggerID(),
		"msg_key":      "k" + fmt.Sprint(time.Now().UnixMilli()*100),
		"sdk_version":  ChatSDKVersion,
		"is_reliable":  false,
		"client_logid": time.Now().UnixMilli() * 1000,
		"rpc":          string(marshalJSON(map[string]any{"rpc_retry_time": 0})),
	})

	rpc, body, err := b.exchange(req)
	if err != nil {
		return false, fmt.Errorf("core: joining the chatroom: %w", err)
	}
	if v, ok := body.JSONValue("err_code"); rpc.GetResponse().GetErrorText() != "success" || !ok || v != "0" {
		return false, errors.New("core: joining the chatroom failed")
	}

	cuidGalaxy2, err := b.Account.CuidGalaxy2()
	if err != nil {
		return false, err
	}
	if _, err := b.FetchMcastMsgClientRequest(ctx, cuidGalaxy2, chatroomID, 1); err != nil {
		return false, err
	}

	b.startHeartbeat()
	return true, nil
}

// ExitChatRoom is not implemented by the Python original either.
func (b *BLCPCore) ExitChatRoom(chatroomID int64, roomType int) error {
	return nil
}

// startHeartbeat starts the periodic Heartbeat goroutine, mirroring
// BLCPCore.__heartbeater.
func (b *BLCPCore) startHeartbeat() {
	b.mu.Lock()
	if b.cancelHB != nil {
		b.cancelHB()
	}
	ctx, cancel := context.WithCancel(context.Background())
	b.cancelHB = cancel
	b.mu.Unlock()

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
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

// signDict signs a parameter dictionary with the given salt.
//
// Note: the Python helper passes a dict to sign(), which iterates the sorted
// keys and would raise on keys longer than two characters. This port implements
// the intended behaviour (sorting the entries by key) instead of reproducing
// the crash.
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
