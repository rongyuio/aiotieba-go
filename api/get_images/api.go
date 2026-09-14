package getimages

import (
	"bytes"
	"context"
	"image"
	"net/url"
	"strings"

	// 注册客户端用到的图像解码器。这些空导入对应 Python _headers_checker 所接受的格式。
	_ "golang.org/x/image/bmp"
	_ "image/jpeg"
	_ "image/png"

	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
)

// contentTypeOK 对应 Python _headers_checker：content type 必须以 jpeg、png 或 bmp 结尾。
func contentTypeOK(contentType string) bool {
	return strings.HasSuffix(contentType, "jpeg") ||
		strings.HasSuffix(contentType, "png") ||
		strings.HasSuffix(contentType, "bmp")
}

// requestBytes 获取原始图像字节，对应 _request_bytes。
func requestBytes(ctx context.Context, httpCore *core.HttpCore, u *url.URL) ([]byte, error) {
	resp, err := httpCore.WebGet(nil, map[string]string{"Referer": "tieba.baidu.com"}).SetContext(ctx).Get(u.String())
	if err != nil {
		return nil, err
	}

	if ct := resp.Header().Get("Content-Type"); !contentTypeOK(ct) {
		return nil, &exception.ContentTypeError{Msg: "Expect jpeg, png or bmp, got " + ct}
	}
	return resp.Body(), nil
}

// RequestBytes 对应 request_bytes。
func RequestBytes(ctx context.Context, httpCore *core.HttpCore, u *url.URL) (ImageBytes, error) {
	body, err := requestBytes(ctx, httpCore, u)
	if err != nil {
		return ImageBytes{Err: err}, err
	}
	return ImageBytes{Data: body}, nil
}

// Request 对应 request：获取并解码图像。
func Request(ctx context.Context, httpCore *core.HttpCore, u *url.URL) (Image, error) {
	body, err := requestBytes(ctx, httpCore, u)
	if err != nil {
		return Image{Err: err}, err
	}
	img, _, err := image.Decode(bytes.NewReader(body))
	if err != nil {
		derr := &exception.TiebaValueError{Msg: "Error in image.Decode"}
		return Image{Err: derr}, derr
	}
	return Image{Img: img}, nil
}
