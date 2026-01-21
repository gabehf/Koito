package catalog

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/images"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
	"github.com/google/uuid"
	"github.com/h2non/bimg"
)

type ImageSize string

const (
	ImageSizeSmall  ImageSize = "small"
	ImageSizeMedium ImageSize = "medium"
	ImageSizeLarge  ImageSize = "large"
	// imageSizeXL     ImageSize = "xl"
	ImageSizeFull ImageSize = "full"

	ImageCacheDir = "image_cache"
)

func ImageSourceSize() (size ImageSize) {
	if cfg.FullImageCacheEnabled() {
		size = ImageSizeFull
	} else {
		size = ImageSizeLarge
	}
	return
}

func ParseImageSize(size string) (ImageSize, error) {
	switch strings.ToLower(size) {
	case "small":
		return ImageSizeSmall, nil
	case "medium":
		return ImageSizeMedium, nil
	case "large":
		return ImageSizeLarge, nil
	// case "xl":
	// 	return imageSizeXL, nil
	case "full":
		return ImageSizeFull, nil
	default:
		return "", fmt.Errorf("unknown image size: %s", size)
	}
}
func GetImageSize(size ImageSize) int {
	var px int
	switch size {
	case "small":
		px = 48
	case "medium":
		px = 256
	case "large":
		px = 500
	case "xl":
		px = 1000
	}
	return px
}

func SourceImageDir() string {
	if cfg.FullImageCacheEnabled() {
		return path.Join(cfg.ConfigDir(), ImageCacheDir, "full")
	} else {
		return path.Join(cfg.ConfigDir(), ImageCacheDir, "large")
	}
}

// DownloadAndCacheImage downloads an image from the given URL, then calls CompressAndSaveImage.
func DownloadAndCacheImage(ctx context.Context, id uuid.UUID, url string, size ImageSize) error {
	l := logger.FromContext(ctx)
	err := images.ValidateImageURL(url)
	if err != nil {
		return fmt.Errorf("DownloadAndCacheImage: %w", err)
	}
	l.Debug().Msgf("Downloading image for ID %s", id)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("DownloadAndCacheImage: http.Get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("DownloadAndCacheImage: failed to download image, status: %s", resp.Status)
	}

	err = CompressAndSaveImage(ctx, id.String(), size, resp.Body)
	if err != nil {
		return fmt.Errorf("DownloadAndCacheImage: %w", err)
	}
	return nil
}

// Compresses an image to the specified size, then saves it to the correct cache folder.
func CompressAndSaveImage(ctx context.Context, filename string, size ImageSize, body io.Reader) error {
	l := logger.FromContext(ctx)

	if size == ImageSizeFull {
		err := saveImage(filename, size, body)
		if err != nil {
			return fmt.Errorf("CompressAndSaveImage: %w", err)
		}
		return nil
	}

	l.Debug().Msg("Creating resized image")
	compressed, err := compressImage(size, body)
	if err != nil {
		return fmt.Errorf("CompressAndSaveImage: %w", err)
	}

	err = saveImage(filename, size, compressed)
	if err != nil {
		return fmt.Errorf("CompressAndSaveImage: %w", err)
	}
	return nil
}

// SaveImage saves an image to the image_cache/{size} folder
func saveImage(filename string, size ImageSize, data io.Reader) error {
	configDir := cfg.ConfigDir()
	cacheDir := filepath.Join(configDir, ImageCacheDir)

	// Ensure the cache directory exists
	err := os.MkdirAll(filepath.Join(cacheDir, string(size)), 0744)
	if err != nil {
		return fmt.Errorf("saveImage: failed to create full image cache directory: %w", err)
	}

	// Create a file in the cache directory
	imagePath := filepath.Join(cacheDir, string(size), filename)
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
	px := GetImageSize(size)
	// Resize with bimg
	imgBytes, err = bimg.NewImage(imgBytes).Process(bimg.Options{
		Width:         px,
		Height:        px,
		Crop:          true,
		Quality:       85,
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
	configDir := cfg.ConfigDir()
	cacheDir := filepath.Join(configDir, ImageCacheDir)

	// err := os.Remove(path.Join(cacheDir, "xl", filename.String()))
	// if err != nil && !os.IsNotExist(err) {
	// 	return err
	// }
	err := os.Remove(path.Join(cacheDir, "full", filename.String()))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("DeleteImage: %w", err)
	}
	err = os.Remove(path.Join(cacheDir, "large", filename.String()))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("DeleteImage: %w", err)
	}
	err = os.Remove(path.Join(cacheDir, "medium", filename.String()))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("DeleteImage: %w", err)
	}
	err = os.Remove(path.Join(cacheDir, "small", filename.String()))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("DeleteImage: %w", err)
	}
	return nil
}

// Finds any images in all image_cache folders and deletes them if they are not associated with
// an album or artist.
func PruneOrphanedImages(ctx context.Context, store db.DB) error {
	l := logger.FromContext(ctx)

	configDir := cfg.ConfigDir()
	cacheDir := filepath.Join(configDir, ImageCacheDir)

	count := 0
	// go through every folder to find orphaned images
	// store already processed images to speed up pruining
	memo := make(map[string]bool)
	for _, dir := range []string{"large", "medium", "small", "full"} {
		c, err := pruneDirImgs(ctx, store, path.Join(cacheDir, dir), memo)
		if err != nil {
			return fmt.Errorf("PruneOrphanedImages: %w", err)
		}
		count += c
	}
	l.Info().Msgf("Purged %d images", count)
	return nil
}

// returns the number of pruned images
func pruneDirImgs(ctx context.Context, store db.DB, path string, memo map[string]bool) (int, error) {
	l := logger.FromContext(ctx)
	count := 0
	files, err := os.ReadDir(path)
	if err != nil {
		l.Info().Msgf("Failed to read from directory %s; skipping for prune", path)
		files = []os.DirEntry{}
	}
	for _, file := range files {
		fn := file.Name()
		imageid, err := uuid.Parse(fn)
		if err != nil {
			l.Debug().Msgf("Filename does not appear to be UUID: %s", fn)
			continue
		}
		exists, err := store.ImageHasAssociation(ctx, imageid)
		if err != nil {
			return 0, fmt.Errorf("pruneDirImages: %w", err)
		} else if exists {
			continue
		}
		// image does not have association
		l.Debug().Msgf("Deleting image: %s", imageid)
		err = DeleteImage(imageid)
		if err != nil {
			l.Err(err).Msg("Error purging orphaned images")
		}
		if memo != nil {
			memo[fn] = true
		}
		count++
	}
	return count, nil
}

func FetchMissingArtistImages(ctx context.Context, store db.DB) error {
	l := logger.FromContext(ctx)
	l.Info().Msg("FetchMissingArtistImages: Starting backfill of missing artist images")

	var from int32 = 0

	for {
		l.Debug().Int32("ID", from).Msg("Fetching artist images to backfill from ID")
		artists, err := store.ArtistsWithoutImages(ctx, from)
		if err != nil {
			return fmt.Errorf("FetchMissingArtistImages: failed to fetch artists for image backfill: %w", err)
		}

		if len(artists) == 0 {
			if from == 0 {
				l.Info().Msg("FetchMissingArtistImages: No artists with missing images found")
			} else {
				l.Info().Msg("FetchMissingArtistImages: Finished fetching missing artist images")
			}
			return nil
		}

		for _, artist := range artists {
			from = artist.ID

			l.Debug().
				Str("title", artist.Name).
				Msg("FetchMissingArtistImages: Attempting to fetch missing artist image")

			var aliases []string
			if aliasrow, err := store.GetAllArtistAliases(ctx, artist.ID); err != nil {
				aliases = utils.FlattenAliases(aliasrow)
			} else {
				aliases = []string{artist.Name}
			}

			var imgid uuid.UUID
			imgUrl, imgErr := images.GetArtistImage(ctx, images.ArtistImageOpts{
				Aliases: aliases,
			})
			if imgErr == nil && imgUrl != "" {
				imgid = uuid.New()
				err = store.UpdateArtist(ctx, db.UpdateArtistOpts{
					ID:       artist.ID,
					Image:    imgid,
					ImageSrc: imgUrl,
				})
				if err != nil {
					l.Err(err).
						Str("title", artist.Name).
						Msg("FetchMissingArtistImages: Failed to update artist with image in database")
					continue
				}
				l.Info().
					Str("name", artist.Name).
					Msg("FetchMissingArtistImages: Successfully fetched missing artist image")
			} else {
				l.Err(err).
					Str("name", artist.Name).
					Msg("FetchMissingArtistImages: Failed to fetch artist image")
			}
		}
	}
}
func FetchMissingAlbumImages(ctx context.Context, store db.DB) error {
	l := logger.FromContext(ctx)
	l.Info().Msg("FetchMissingAlbumImages: Starting backfill of missing album images")

	var from int32 = 0

	for {
		l.Debug().Int32("ID", from).Msg("Fetching album images to backfill from ID")
		albums, err := store.AlbumsWithoutImages(ctx, from)
		if err != nil {
			return fmt.Errorf("FetchMissingAlbumImages: failed to fetch albums for image backfill: %w", err)
		}

		if len(albums) == 0 {
			if from == 0 {
				l.Info().Msg("FetchMissingAlbumImages: No albums with missing images found")
			} else {
				l.Info().Msg("FetchMissingAlbumImages: Finished fetching missing album images")
			}
			return nil
		}

		for _, album := range albums {
			from = album.ID

			l.Debug().
				Str("title", album.Title).
				Msg("FetchMissingAlbumImages: Attempting to fetch missing album image")

			var imgid uuid.UUID
			imgUrl, imgErr := images.GetAlbumImage(ctx, images.AlbumImageOpts{
				Artists:      utils.FlattenSimpleArtistNames(album.Artists),
				Album:        album.Title,
				ReleaseMbzID: album.MbzID,
			})
			if imgErr == nil && imgUrl != "" {
				imgid = uuid.New()
				err = store.UpdateAlbum(ctx, db.UpdateAlbumOpts{
					ID:       album.ID,
					Image:    imgid,
					ImageSrc: imgUrl,
				})
				if err != nil {
					l.Err(err).
						Str("title", album.Title).
						Msg("FetchMissingAlbumImages: Failed to update album with image in database")
					continue
				}
				l.Info().
					Str("name", album.Title).
					Msg("FetchMissingAlbumImages: Successfully fetched missing album image")
			} else {
				l.Err(err).
					Str("name", album.Title).
					Msg("FetchMissingAlbumImages: Failed to fetch album image")
			}
		}
	}
}
