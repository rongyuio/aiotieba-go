// Package getimages 实现 aiotieba 的 get_images API。
//
// 对应 Python 包 aiotieba.api.get_images。解码后的图像为标准库的 image.Image，而非 numpy 数组。
package getimages

import "image"

// Image 图像。
type Image struct {
	Img image.Image // 图像
	Err error       // 捕获的异常
}

// ImageBytes 图像原始字节流。
type ImageBytes struct {
	Data []byte // 图像原始字节流
	Err  error  // 捕获的异常
}
