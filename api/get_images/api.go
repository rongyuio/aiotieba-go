package getimages

import (
	"bytes"
	"context"
	"image"
	"net/url"
	"strings"

	// Register the image decoders used by the client. The blank imports mirror
	// the formats accepted by the Python _headers_checker.
	_ "golang.org/x/image/bmp"
	_ "image/jpeg"
	_ "image/png"

	"github.com/rongyuio/aiotieba-go/core"
	"github.com/rongyuio/aiotieba-go/exception"
)

// contentTypeOK mirrors the Python _headers_checker: the content type must end
// with jpeg, png or bmp.
func contentTypeOK(contentType string) bool {
	return strings.HasSuffix(contentType, "jpeg") ||
		strings.HasSuffix(contentType, "png") ||
		strings.HasSuffix(contentType, "bmp")
}

// requestBytes fetches the raw image bytes, mirroring _request_bytes.
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

// RequestBytes mirrors request_bytes.
func RequestBytes(ctx context.Context, httpCore *core.HttpCore, u *url.URL) (ImageBytes, error) {
	body, err := requestBytes(ctx, httpCore, u)
	if err != nil {
		return ImageBytes{Err: err}, err
	}
	return ImageBytes{Data: body}, nil
}

// Request mirrors request: it fetches and decodes the image.
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
