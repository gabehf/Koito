package catalog

import (
	"fmt"
)

type ImageSize string

const (
	ImageSizeXS     ImageSize = "32x32"
	ImageSizeSmall  ImageSize = "128x128"
	ImageSizeMedium ImageSize = "256x256"
	ImageSizeLarge  ImageSize = "640x640"
	// ImageSizeXL     ImageSize = "1000x1000"
	ImageSizeSource ImageSize = "source"

	ImageCacheDir = "image_cache"
)

func (s ImageSize) Width() int {
	switch s {
	case ImageSizeXS:
		return 32
	case ImageSizeSmall:
		return 128
	case ImageSizeMedium:
		return 256
	case ImageSizeLarge:
		return 640
	// case ImageSizeXL:
	// 	return 1000
	case ImageSizeSource:
		return 1200
	default:
		return 0
	}
}
func (s ImageSize) String() string {
	return string(s)
}

func ParseImageSize(s string) (ImageSize, error) {
	switch ImageSize(s) {
	case ImageSizeXS, ImageSizeSmall, ImageSizeMedium, ImageSizeLarge:
		return ImageSize(s), nil
	default:
		return "", fmt.Errorf("ParseImageSize: invalid image size %q", s)
	}
}
