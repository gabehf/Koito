package catalog

import (
	"fmt"
)

type ImageSize string

const (
	ImageSizeXS     ImageSize = "64x64"
	ImageSizeSmall  ImageSize = "128x128"
	ImageSizeMedium ImageSize = "300x300"
	ImageSizeLarge  ImageSize = "640x640"
	ImageSizeXL     ImageSize = "1000x1000"
	ImageSizeSource ImageSize = "source"

	ImageCacheDir = "image_cache"
)

func (s ImageSize) Width() int {
	switch s {
	case ImageSizeXS:
		return 64
	case ImageSizeSmall:
		return 128
	case ImageSizeMedium:
		return 300
	case ImageSizeLarge:
		return 640
	case ImageSizeXL:
		return 1000
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
	case ImageSizeXS, ImageSizeSmall, ImageSizeMedium, ImageSizeLarge, ImageSizeXL:
		return ImageSize(s), nil
	default:
		return "", fmt.Errorf("ParseImageSize: invalid image size %q", s)
	}
}
