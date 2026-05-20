// Package imagecache handles the downloading and cacheing of images.
package imagecache

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/images"
	"github.com/google/uuid"
	"github.com/h2non/bimg"
	"golang.org/x/sync/singleflight"
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

// DownloadImage downloads an image from the given URL, then saves it to the cache at source quality.
func DownloadImage(imgid uuid.UUID, url string) error {
	err := images.ValidateImageURL(url)
	if err != nil {
		return fmt.Errorf("DownloadImage: %w", err)
	}
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("DownloadImage: http.Get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("DownloadImage: failed to download image, status: %s", resp.Status)
	}

	err = compressAndSaveImage(imgid, ImageSizeSource, resp.Body)
	if err != nil {
		return fmt.Errorf("DownloadImage: %w", err)
	}
	return nil
}

// SaveImage saves the image at source quality to the cache.
func SaveImage(imgid uuid.UUID, image io.Reader) error {
	return compressAndSaveImage(imgid, ImageSizeSource, image)
}

type ImageInfo struct {
	Path string
	Mime string
}

// GetImage takes the UUID image id and image filename e.g. "128x128.webp" and returns the path and mime type of the
// image in the cache. If the image does not exist in the cache in either the desired or source sizes,
// returns a fs.ErrNotExist error.
func GetImage(imgid uuid.UUID, image string) (*ImageInfo, error) {

	var resizeGroup singleflight.Group

	imageSize, err := ParseImageSize(strings.Split(image, ".")[0])
	if err != nil {
		return nil, errors.New("GetImage: failed to parse image size")
	}

	desiredImgPath := BuildImagePath(imgid, imageSize)
	if _, err := os.Stat(desiredImgPath); errors.Is(err, fs.ErrNotExist) {
		sourcePath := BuildImagePath(imgid, ImageSizeSource)

		if _, err = os.Stat(sourcePath); errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("GetImage: image does not exist in cache: %w", err)
		}

		imageBuf, err := os.ReadFile(sourcePath)
		if err != nil {
			return nil, fmt.Errorf("GetImage: failed to read source file: %w", err)
		}

		resizeKey := imgid.String() + "/" + imageSize.String()
		_, err, _ = resizeGroup.Do(resizeKey, func() (any, error) {
			return nil, compressAndSaveImage(imgid, imageSize, bytes.NewReader(imageBuf))
		})
		if err != nil {
			return nil, fmt.Errorf("GetImage: failed to cache source file in desired size: %w", err)
		}

	} else if err != nil {
		return nil, fmt.Errorf("GetImage: failed to access desired image in cache: %w", err)
	}

	return &ImageInfo{
		Path: desiredImgPath,
		Mime: "image/webp",
	}, nil
}

// Returns the ImageSize from a filename such as "1000x1000.webp", or an error if the size cannot be parsed.
func ParseImageSize(s string) (ImageSize, error) {
	s = strings.Split(s, ".")[0]
	switch ImageSize(s) {
	case ImageSizeXS, ImageSizeSmall, ImageSizeMedium, ImageSizeLarge, ImageSizeXL:
		return ImageSize(s), nil
	default:
		return "", fmt.Errorf("ParseImageSize: invalid image size %q", s)
	}
}

func compressAndSaveImage(imgid uuid.UUID, size ImageSize, image io.Reader) error {
	compressed, err := compressImage(size, image)
	if err != nil {
		return fmt.Errorf("CompressAndSaveImage: %w", err)
	}

	err = saveImage(imgid, size, compressed)
	if err != nil {
		return fmt.Errorf("CompressAndSaveImage: %w", err)
	}
	return nil
}

// saveImage saves an image to the imgid[:2]/imgid/{size}.ext path
func saveImage(imgid uuid.UUID, size ImageSize, data io.Reader) error {
	imagePath := BuildImagePath(imgid, size)

	// Ensure the cache directory exists
	err := os.MkdirAll(path.Dir(imagePath), 0744)
	if err != nil {
		return fmt.Errorf("saveImage: failed to create full image cache directory: %w", err)
	}

	// Create a file in the cache directory
	file, err := os.Create(imagePath)
	if err != nil {
		return fmt.Errorf("saveImage: failed to create image file: %w", err)
	}
	defer file.Close()

	// Save the image to the file
	_, err = io.Copy(file, data)
	if err != nil {
		return fmt.Errorf("saveImage: failed to save image: %w", err)
	}

	return nil
}

func compressImage(size ImageSize, data io.Reader) (io.Reader, error) {
	imgBytes, err := io.ReadAll(data)
	if err != nil {
		return nil, fmt.Errorf("compressImage: io.ReadAll: %w", err)
	}
	px := size.Width()

	if size == ImageSizeSource {
		ogsize, err := bimg.Size(imgBytes)
		if err != nil {
			return nil, fmt.Errorf("compressImage: bimg.Size: %w", err)
		}
		tmpx := min(ogsize.Width, ogsize.Height)
		px = min(tmpx, ImageSizeSource.Width())
	}

	quality := 85
	if size == ImageSizeSource {
		quality = 100
	}

	// Resize with bimg
	imgBytes, err = bimg.NewImage(imgBytes).Process(bimg.Options{
		Width:         px,
		Height:        px,
		Crop:          true,
		Quality:       quality,
		StripMetadata: true,
		Type:          bimg.WEBP,
	})
	if err != nil {
		return nil, fmt.Errorf("compressImage: bimg.NewImage: %w", err)
	}
	if len(imgBytes) == 0 {
		return nil, fmt.Errorf("compressImage: failed to compress image: %w", err)
	}
	return bytes.NewReader(imgBytes), nil
}

func DeleteImage(filename uuid.UUID) error {

	err := os.RemoveAll(filepath.Dir(BuildImagePath(filename, ImageSizeSource)))
	if err != nil {
		return fmt.Errorf("DeleteImage: %w", err)
	}

	return nil
}

func BuildImagePath(imageid uuid.UUID, size ImageSize) string {
	return filepath.Join(cfg.ConfigDir(), ImageCacheDir, imageid.String()[:2], imageid.String(), string(size)+".webp")
}

func ForEachCachedImageID(foreach func(imageID uuid.UUID)) error {
	cacheDir := filepath.Join(cfg.ConfigDir(), ImageCacheDir)
	prefixDirs, err := os.ReadDir(cacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("ForEachCachedImage: reading cache dir: %w", err)
	}

	for _, prefixEntry := range prefixDirs {
		if !prefixEntry.IsDir() {
			continue
		}
		prefixPath := filepath.Join(cacheDir, prefixEntry.Name())

		uuidDirs, err := os.ReadDir(prefixPath)
		if err != nil {
			return fmt.Errorf("ForEachCachedImage: reading prefix dir %s: %w", prefixEntry.Name(), err)
		}

		for _, uuidEntry := range uuidDirs {
			if !uuidEntry.IsDir() {
				continue
			}
			uuidStr := uuidEntry.Name()
			imgid, err := uuid.Parse(uuidStr)
			if err != nil {
				// non-uuid filaname -> ignore
				continue
			}
			foreach(imgid)
		}
	}
	return nil
}
