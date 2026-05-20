package catalog

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/gabehf/koito/imagecache"
	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/images"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/models"
	"github.com/gabehf/koito/internal/utils"
	"github.com/google/uuid"
)

// Finds any images in all image_cache folders and deletes them if they are not associated with
// an album or artist.
func PruneOrphanedImages(ctx context.Context, store db.ImageStore) error {
	l := logger.FromContext(ctx)
	count := 0
	memo := make(map[uuid.UUID]bool)

	imagecache.ForEachCachedImageID(func(imageID uuid.UUID) {
		exists, seen := memo[imageID]
		if !seen {
			exists, err := store.ImageHasAssociation(ctx, imageID)
			if err != nil {
				l.Err(err).Msgf("Failed to query image association for image with ID %s", imageID.String())
			}
			memo[imageID] = exists
		}

		if !exists {
			err := imagecache.DeleteImage(imageID)
			if err != nil {
				l.Err(err).Msgf("Failed to delete image with ID %s", imageID.String())
			}
			count++
		}
	})

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

// TODO: move this function into models
func BuildImageList(imageid *uuid.UUID) models.ImageList {
	if imageid == nil || *imageid == uuid.Nil {
		return models.ImageList{}
	} else {
		return models.ImageList{
			XS:     fmt.Sprintf("/image/%s/%s.webp", imageid.String(), imagecache.ImageSizeXS),
			Small:  fmt.Sprintf("/image/%s/%s.webp", imageid.String(), imagecache.ImageSizeSmall),
			Medium: fmt.Sprintf("/image/%s/%s.webp", imageid.String(), imagecache.ImageSizeMedium),
			Large:  fmt.Sprintf("/image/%s/%s.webp", imageid.String(), imagecache.ImageSizeLarge),
			XL:     fmt.Sprintf("/image/%s/%s.webp", imageid.String(), imagecache.ImageSizeXL),
		}
	}
}

func MigrateImageCache(ctx context.Context, store db.ImageStore) error {
	l := logger.FromContext(ctx)
	cacheDir := path.Join(cfg.ConfigDir(), imagecache.ImageCacheDir)

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
				destPath := imagecache.BuildImagePath(imageid, imagecache.ImageSizeSource)
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
