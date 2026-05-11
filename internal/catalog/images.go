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

	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/images"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/models"
	"github.com/gabehf/koito/internal/utils"
	"github.com/google/uuid"
	"github.com/h2non/bimg"
)

// DownloadAndCacheImage downloads an image from the given URL, then saves it to the cache at source quality.
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

	err = CompressAndSaveImage(ctx, id, size, resp.Body)
	if err != nil {
		return fmt.Errorf("DownloadAndCacheImage: %w", err)
	}
	return nil
}

// Compresses an image to the specified size, then saves it to the correct cache folder.
// If called with ImageSizeSource, losslessly compresses the image and caches it at source
// resolution or 1000x1000, whichever is smaller.
func CompressAndSaveImage(ctx context.Context, imgid uuid.UUID, size ImageSize, body io.Reader) error {
	l := logger.FromContext(ctx)

	l.Debug().Msgf("Creating resized image at size %d x %d", size.Width(), size.Width())
	compressed, err := compressImage(size, body)
	if err != nil {
		return fmt.Errorf("CompressAndSaveImage: %w", err)
	}

	err = saveImage(imgid, size, compressed)
	if err != nil {
		return fmt.Errorf("CompressAndSaveImage: %w", err)
	}
	return nil
}

// SaveImage saves an image to the imgid[:2]/imgid/{size}.webp path
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

	err := os.Remove(BuildImagePath(filename, ImageSizeSource))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("DeleteImage: %w", err)
	}
	// err = os.Remove(BuildImagePath(filename, ImageSizeXL))
	// if err != nil && !os.IsNotExist(err) {
	// 	return fmt.Errorf("DeleteImage: %w", err)
	// }
	err = os.Remove(BuildImagePath(filename, ImageSizeLarge))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("DeleteImage: %w", err)
	}
	err = os.Remove(BuildImagePath(filename, ImageSizeMedium))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("DeleteImage: %w", err)
	}
	err = os.Remove(BuildImagePath(filename, ImageSizeSmall))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("DeleteImage: %w", err)
	}
	err = os.Remove(BuildImagePath(filename, ImageSizeXS))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("DeleteImage: %w", err)
	}
	return nil
}

// Finds any images in all image_cache folders and deletes them if they are not associated with
// an album or artist.
func PruneOrphanedImages(ctx context.Context, store db.ImageStore) error {
	l := logger.FromContext(ctx)
	cacheDir := filepath.Join(cfg.ConfigDir(), ImageCacheDir)
	count := 0
	memo := make(map[string]bool)

	prefixDirs, err := os.ReadDir(cacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("PruneOrphanedImages: reading cache dir: %w", err)
	}

	for _, prefixEntry := range prefixDirs {
		if !prefixEntry.IsDir() {
			continue
		}
		prefixPath := filepath.Join(cacheDir, prefixEntry.Name())

		uuidDirs, err := os.ReadDir(prefixPath)
		if err != nil {
			return fmt.Errorf("PruneOrphanedImages: reading prefix dir %s: %w", prefixEntry.Name(), err)
		}

		for _, uuidEntry := range uuidDirs {
			if !uuidEntry.IsDir() {
				continue
			}
			uuidStr := uuidEntry.Name()
			imgid, err := uuid.Parse(uuidStr)
			if err != nil {
				continue
			} else if imgid == uuid.Nil {
				// default image - dont prune
				continue
			}

			exists, seen := memo[uuidStr]
			if !seen {
				exists, err = store.ImageHasAssociation(ctx, imgid)
				if err != nil {
					return fmt.Errorf("PruneOrphanedImages: checking store for %s: %w", imgid, err)
				}
				memo[uuidStr] = exists
			}

			if !exists {
				uuidPath := filepath.Join(prefixPath, uuidStr)
				if err := os.RemoveAll(uuidPath); err != nil {
					return fmt.Errorf("PruneOrphanedImages: removing %s: %w", uuidPath, err)
				}
				count++
			}
		}
	}

	l.Info().Msgf("Purged %d images", count)
	return nil
}

func FetchMissingArtistImages(ctx context.Context, store db.ArtistStore) error {
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
			if aliasrow, err := store.GetAllArtistAliases(ctx, artist.ID); err == nil {
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
func FetchMissingAlbumImages(ctx context.Context, store db.AlbumStore) error {
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

func BuildImageList(imageid *uuid.UUID) models.ImageList {
	if imageid == nil || *imageid == uuid.Nil {
		return models.ImageList{}
	} else {
		return models.ImageList{
			XS:     fmt.Sprintf("/images/%s/%s.webp", imageid.String(), ImageSizeXS),
			Small:  fmt.Sprintf("/images/%s/%s.webp", imageid.String(), ImageSizeSmall),
			Medium: fmt.Sprintf("/images/%s/%s.webp", imageid.String(), ImageSizeMedium),
			Large:  fmt.Sprintf("/images/%s/%s.webp", imageid.String(), ImageSizeLarge),
			XL:     fmt.Sprintf("/images/%s/%s.webp", imageid.String(), ImageSizeXL),
		}
	}
}

func BuildImagePath(imageid uuid.UUID, size ImageSize) string {
	return path.Join(cfg.ConfigDir(), ImageCacheDir, imageid.String()[:2], imageid.String(), string(size)+".webp")
}

func MigrateImageCache(ctx context.Context, store db.ImageStore) error {
	l := logger.FromContext(ctx)
	cacheDir := path.Join(cfg.ConfigDir(), ImageCacheDir)

	legacyLargeDir := path.Join(cacheDir, "large")
	if _, err := os.Stat(legacyLargeDir); os.IsNotExist(err) {
		return nil // nothing to migrate
	}

	uploadedImages, err := store.GetUserUploadedImageIDs(ctx)
	if err != nil {
		return fmt.Errorf("MigrateImageCache: failed to fetch uploaded image IDs: %w", err)
	}

	l.Debug().Msgf("MigrateImageCache: %d user uploaded images must be migrated", len(uploadedImages))
	l.Debug().Any("UUIDs", uploadedImages).Msg("Image IDs to be migrated")

	count := 0

	migrated := make(map[uuid.UUID]bool, len(uploadedImages))
	for _, id := range uploadedImages {
		migrated[id] = false
	}

	dirs := []string{"full", "large", "medium", "small"}

	for _, dir := range dirs {
		dirPath := path.Join(cacheDir, dir)
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("MigrateImageCache: failed to read directory %s: %w", dirPath, err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			filename := entry.Name()
			filePath := path.Join(dirPath, entry.Name())

			if filename == "default_img" {
				if err := os.Remove(filePath); err != nil {
					l.Err(err).Msgf("MigrateImageCache: failed to remove default_img file %s", filePath)
				}
				continue
			}

			imageid, err := uuid.Parse(filename)
			if err != nil {
				l.Err(err).Msgf("MigrateImageCache: found non-uuid filename '%s', this must be removed manually", filePath)
			}

			alreadyMigrated, isUploaded := migrated[imageid]
			if isUploaded && !alreadyMigrated {
				destPath := BuildImagePath(imageid, ImageSizeSource)
				if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
					l.Err(err).Msgf("MigrateImageCache: failed to create destination dir for %s", destPath)
				} else if err := os.Rename(filePath, destPath); err != nil {
					l.Err(err).Msgf("MigrateImageCache: failed to migrate image %s to %s", filePath, destPath)
				} else {
					migrated[imageid] = true
					count++
				}
			} else {
				if err := os.Remove(filePath); err != nil {
					l.Err(err).Msgf("MigrateImageCache: failed to remove file %s", filePath)
				}
			}
		}
	}

	for _, dir := range dirs {
		dirPath := path.Join(cacheDir, dir)
		if err := os.Remove(dirPath); err != nil && !os.IsNotExist(err) {
			l.Err(err).Msgf("MigrateImageCache: failed to remove directory %s, it must be removed manually", dirPath)
		}
	}

	l.Info().Msgf("MigrateImageCache: %d user uploaded images were migrated", count)

	return nil
}
