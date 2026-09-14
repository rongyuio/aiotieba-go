package crypto

import (
	"cmp"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"slices"
)

// Param 是一个参与签名的表单参数。
//
// 值要么是字符串，要么是整数，对应 Python 类型 `tuple[str, str | int]`。
type Param struct {
	Key   string
	Value any
}

// ComputeSign 计算贴吧客户端签名。
//
// 参数:
//
//	data 参数元组列表
//	salt 计算签名使用的盐值
//
// 对应 aiotieba.helper.crypto.sign.compute_sign：按 key 排序后以 `key=value` 拼接（无分隔符），
// 追加盐值后取 MD5 十六进制摘要。
func ComputeSign(data []Param, salt []byte) string {
	sorted := slices.Clone(data)
	slices.SortFunc(sorted, func(a, b Param) int { return cmp.Compare(a.Key, b.Key) })

	h := md5.New()
	for _, p := range sorted {
		fmt.Fprintf(h, "%s=%v", p.Key, p.Value)
	}
	h.Write(salt)
	return hex.EncodeToString(h.Sum(nil))
}

// Sign 为参数元组列表添加贴吧客户端签名。
//
// 参数:
//
//	data 参数元组列表
//	salt 计算签名使用的盐值
//
// 把 ComputeSign 生成的签名追加到 data，对应 aiotieba.helper.crypto.sign.sign。
//
// PC 网页端签名算法（对应 search_global._api.py 的 _pc_sign）：取除 sign/sig 外的全部参数，
// 按 key 升序排序后逐个以 "key=value" 无分隔拼接，末尾拼接密钥，整体 UTF-8 编码后取 MD5 十六进制
// （32 位）。使用 crypto.PCSalt 作为盐值即为 PC 网页端签名。
func Sign(data []Param, salt []byte) []Param {
	return append(data, Param{Key: "sign", Value: ComputeSign(data, salt)})
}
