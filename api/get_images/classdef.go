// Package getimages implements the get_images API of aiotieba.
//
// It mirrors the Python package aiotieba.api.get_images. The decoded image is a
// standard library image.Image instead of a numpy array.
package getimages

import "image"

// Image mirrors Image: a decoded image.
type Image struct {
	Img image.Image
	Err error
}

// ImageBytes mirrors ImageBytes: the raw image byte stream.
type ImageBytes struct {
	Data []byte
	Err  error
}
