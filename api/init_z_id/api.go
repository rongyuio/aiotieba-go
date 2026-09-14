// Package initzid 实现 aiotieba 的 init_z_id API。
//
// 对应 Python 包 aiotieba.api.init_z_id。
package initzid

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/rongyuio/aiotieba-go/consts"
	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/helper"
	"github.com/rongyuio/aiotieba-go/helper/crypto"
)

// sofire z_id 服务的常量，对应 init_z_id/_api.py。
const (
	// SofireHost 是 z_id 服务的主机名。
	SofireHost = "sofire.baidu.com"
	appKey     = "200033"
	secKey     = "ea737e4f435b53786043369d2e5ace4f"
	cdmVersion = "4.4.1.3"
)

type initZIDParams struct {
	ModuleSection []initZIDModule `json:"module_section"`
}

type initZIDModule struct {
	ZID string `json:"zid"`
}

// Request 获取账号的 z_id，对应 request。
func Request(ctx context.Context, httpCore *core.HttpCore) (string, error) {
	account := httpCore.Account

	xyus := strings.ToUpper(md5Hex(account.AndroidID()+account.UUID())) + "|0"
	xyusMD5 := md5Hex(xyus)
	currentTS := strconv.FormatInt(time.Now().Unix(), 10)

	// 请求体是紧凑 JSON，先 gzip 压缩再做 AES-CBC 加密。
	reqBody := []byte(helper.PackJSON(initZIDParams{
		ModuleSection: []initZIDModule{{ZID: xyus}},
	}))
	compressed, err := gzipCompress(reqBody)
	if err != nil {
		return "", err
	}

	encrypted, err := crypto.CBCEncryptZeroIV(account.AESCBCSecKey(), compressed)
	if err != nil {
		return "", err
	}

	// 压缩后请求体的 MD5 会作为后缀附加在末尾。
	sum := md5.Sum(compressed)
	payload := make([]byte, 0, len(encrypted)+len(sum))
	payload = append(payload, encrypted...)
	payload = append(payload, sum[:]...)

	pathCombineMD5 := md5Hex(appKey + currentTS + secKey)
	skey, err := crypto.Rc442(xyusMD5, account.AESCBCSecKey())
	if err != nil {
		return "", err
	}
	// binascii.b2a_base64 会在末尾追加一个换行符。
	skeyB64 := base64.StdEncoding.EncodeToString(skey) + "\n"

	target := &url.URL{
		Scheme:   "https",
		Host:     SofireHost,
		Path:     "/c/11/z/100/" + appKey + "/" + currentTS + "/" + pathCombineMD5,
		RawQuery: "skey=" + url.QueryEscape(skeyB64),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target.String(), bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("core: building the init_z_id request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("x-device-id", xyusMD5)
	req.Header.Set("User-Agent", "x6/"+appKey+"/"+consts.LegacyVersion+"/"+cdmVersion)
	req.Header.Set("x-plu-ver", "x6/"+cdmVersion)

	body, err := httpCore.NetCore.SendRequest(req)
	if err != nil {
		return "", err
	}
	res, err := helper.ParseJSONMap(body)
	if err != nil {
		return "", err
	}

	resQuerySkey, err := base64.StdEncoding.DecodeString(helper.JSONStr(res, "skey"))
	if err != nil {
		return "", fmt.Errorf("core: decoding the init_z_id skey: %w", err)
	}
	resAESKey, err := crypto.Rc442(xyusMD5, resQuerySkey)
	if err != nil {
		return "", err
	}
	resData, err := base64.StdEncoding.DecodeString(helper.JSONStr(res, "data"))
	if err != nil {
		return "", fmt.Errorf("core: decoding the init_z_id data: %w", err)
	}

	// 解密后去掉末尾 16 字节的 MD5 后缀，再进行去填充。
	decrypted, err := crypto.CBCDecryptRaw(resAESKey, make([]byte, 16), resData)
	if err != nil {
		return "", err
	}
	if len(decrypted) < 16 {
		return "", fmt.Errorf("core: the init_z_id response is too short")
	}
	unpadded, err := crypto.PKCS7Unpad(decrypted[:len(decrypted)-16])
	if err != nil {
		return "", err
	}

	resMap, err := helper.ParseJSONMap(unpadded)
	if err != nil {
		return "", err
	}
	return helper.JSONStr(resMap, "token"), nil
}

func md5Hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

// gzipCompress 对应 gzip.compress(data, compresslevel=6, mtime=0)。
func gzipCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw, err := gzip.NewWriterLevel(&buf, 6)
	if err != nil {
		return nil, fmt.Errorf("core: creating the gzip writer: %w", err)
	}
	if _, err := zw.Write(data); err != nil {
		return nil, fmt.Errorf("core: compressing the request body: %w", err)
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("core: compressing the request body: %w", err)
	}
	return buf.Bytes(), nil
}
