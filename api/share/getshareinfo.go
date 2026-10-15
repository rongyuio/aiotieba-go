// getShareInfo 是分享链路的前置调用，对应抓包中的 c/f/forum/getShareInfo。
//
// 它不属于 aiotieba 上游的公开接口，本文件实现它是为了复现真实客户端的调用序列：
//
//	客户端分享前必先调 getShareInfo（取分享文案），随后调 thread/share，
//	并在 thread/share 的 CommonReq 里申报「上一次调用」——
//	m_api = c/f/forum/getShareInfo、m_logid = getShareInfo 响应里的 logid、
//	m_cost = 该次调用耗时、m_result = 该次调用的错误码。
//
// 该接口实测恒返回 error_code=110003「内部错误」（真实客户端亦然），
// 失败不影响分享本身；它的作用是把上一次调用的记录坐实，
// 使 thread/share 的 m_* 四元组成为可校验的真实链条而非编造值。
package share

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// getShareInfoPath 是前置调用的路径。
//
// 注意它**不带 cmd 查询参数**，与项目内其它 protobuf 接口不同 ——
// 抓包中该请求就是一个普通的 form POST，业务参数在表单体里。
const getShareInfoPath = "/c/f/forum/getShareInfo"

// shareInfoResponse 是 getShareInfo 的响应体。
//
// 抓包真值：
//
//	{"ctime":"0","data":{},"error_code":110003,"error_msg":"内部错误",
//	 "logid":"1029955438","server_time":13,"time":1790104629}
type shareInfoResponse struct {
	LogID     string `json:"logid"`
	ErrorCode int32  `json:"error_code"`
}

// getShareInfoURL 返回前置调用的地址。
func getShareInfoURL() *url.URL {
	return &url.URL{Scheme: "https", Host: consts.AppBaseHost, Path: getShareInfoPath}
}

// requestShareInfo 发起 getShareInfo 调用，用于坐实「上一次调用」的记录。
//
// 参数:
//
//	ctx 上下文
//	httpCore 会话核心
//	tid 被分享的主题帖 id
//	fid 该帖所在吧的 fid
//
// 该接口的业务返回值恒为错误码 110003，因此这里**不把它当失败处理**：
// 只要请求本身完成（拿到了响应），m_* 链条就成立了，分享可以继续。
// 只有传输层失败才返回 error —— 那时链上没有任何可申报的记录。
func requestShareInfo(ctx context.Context, httpCore *core.HttpCore, tid, fid int64) error {
	data := []crypto.Param{
		{Key: "forum_id", Value: fid},
		{Key: "thread_id", Value: tid},
	}
	resp, err := httpCore.AppForm(data).
		SetContext(ctx).
		Post(getShareInfoURL().String())
	if err != nil {
		return fmt.Errorf("getShareInfo: %w", err)
	}

	// 解析只为将来可能的日志用途；业务错误码不改变结果。
	var doc shareInfoResponse
	_ = json.Unmarshal(resp.Body(), &doc)
	return nil
}
