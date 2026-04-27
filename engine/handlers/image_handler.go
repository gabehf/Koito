package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/gabehf/koito/internal/catalog"
	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func ImageHandler(store db.ImageStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logger.FromContext(r.Context())
		size := chi.URLParam(r, "size")
		filename := chi.URLParam(r, "filename")

		l.Debug().Msgf("ImageHandler: Received request to retrieve image with size '%s' and filename '%s'", size, filename)

		imageSize, err := catalog.ParseImageSize(size)
		if err != nil {
			l.Debug().Msg("ImageHandler: Invalid image size parameter")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		imgid, err := uuid.Parse(filename)
		if err != nil {
			l.Debug().Msg("ImageHandler: Invalid image filename, serving default image")
			serveDefaultImage(w, r, imageSize)
			return
		}

		desiredImgPath := filepath.Join(cfg.ConfigDir(), catalog.ImageCacheDir, size, filepath.Clean(filename))

		if _, err := os.Stat(desiredImgPath); os.IsNotExist(err) {
			l.Debug().Msg("ImageHandler: Image not found in desired size, attempting to retrieve source image")

			fullSizePath := filepath.Join(cfg.ConfigDir(), catalog.ImageCacheDir, string(catalog.ImageSizeFull), filepath.Clean(filename))
			largeSizePath := filepath.Join(cfg.ConfigDir(), catalog.ImageCacheDir, string(catalog.ImageSizeLarge), filepath.Clean(filename))

			// this if statement flow is terrible but whatever
			var sourcePath string
			if _, err = os.Stat(fullSizePath); os.IsNotExist(err) {
				if _, err = os.Stat(largeSizePath); os.IsNotExist(err) {
					l.Warn().Msgf("ImageHandler: Could not find requested image %s. Attempting to download from source", imgid.String())
					sourcePath, err = downloadMissingImage(r.Context(), store, imgid)
					if err != nil {
						l.Err(err).Msg("ImageHandler: Failed to redownload missing image")
						w.WriteHeader(http.StatusInternalServerError)
					}
				} else if err != nil {
					l.Err(err).Msg("ImageHandler: Failed to access source image file at large size")
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					sourcePath = largeSizePath
				}
			} else if err != nil {
				l.Err(err).Msg("ImageHandler: Failed to access source image file at full size")
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				sourcePath = fullSizePath
			}

			l.Debug().Msgf("ImageHandler: Found source image file at path '%s'", sourcePath)

			imageBuf, err := os.ReadFile(sourcePath)
			if err != nil {
				l.Err(err).Msg("ImageHandler: Failed to read source image file")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			err = catalog.CompressAndSaveImage(r.Context(), imgid.String(), imageSize, bytes.NewReader(imageBuf))
			if err != nil {
				l.Err(err).Msg("ImageHandler: Failed to save compressed image to cache")
			}
		} else if err != nil {
			l.Err(err).Msg("ImageHandler: Failed to access desired image file")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		l.Debug().Msgf("ImageHandler: Serving image from path '%s'", desiredImgPath)
		http.ServeFile(w, r, desiredImgPath)
	}
}

func serveDefaultImage(w http.ResponseWriter, r *http.Request, size catalog.ImageSize) {
	var lock sync.Mutex
	l := logger.FromContext(r.Context())

	l.Debug().Msgf("serveDefaultImage: Serving default image at size '%s'", size)

	defaultImagePath := filepath.Join(cfg.ConfigDir(), catalog.ImageCacheDir, string(size), "default_img")
	if _, err := os.Stat(defaultImagePath); os.IsNotExist(err) {
		l.Debug().Msg("serveDefaultImage: Default image does not exist in cache at desired size")
		defaultImagePath := filepath.Join(catalog.SourceImageDir(), "default_img")
		if _, err = os.Stat(defaultImagePath); os.IsNotExist(err) {
			l.Debug().Msg("serveDefaultImage: Default image does not exist in source directory, attempting to move...")
			err = os.MkdirAll(filepath.Dir(defaultImagePath), 0744)
			if err != nil {
				l.Err(err).Msg("serveDefaultImage: Error when attempting to create image_cache/full directory")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			lock.Lock()
			err = utils.CopyFile(path.Join("assets", "default_img"), defaultImagePath)
			if err != nil {
				l.Err(err).Msg("serveDefaultImage: Error when copying default image from assets")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			lock.Unlock()
		} else if err != nil {
			l.Err(err).Msg("serveDefaultImage: Error when attempting to read default image in cache")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		file, err := os.Open(path.Join(catalog.SourceImageDir(), "default_img"))
		if err != nil {
			l.Err(err).Msg("serveDefaultImage: Error when reading default image from source directory")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = catalog.CompressAndSaveImage(r.Context(), "default_img", size, file)
		if err != nil {
			l.Err(err).Msg("serveDefaultImage: Error when caching default image at desired size")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		l.Err(err).Msg("serveDefaultImage: Error when attempting to read default image in cache")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Debug().Msgf("serveDefaultImage: Successfully serving default image at size '%s'", size)
	http.ServeFile(w, r, path.Join(cfg.ConfigDir(), catalog.ImageCacheDir, string(size), "default_img"))
}

// finds the item associated with the image id, downloads it, and saves it in the source path, returning the path to the image
func downloadMissingImage(ctx context.Context, store db.ImageStore, id uuid.UUID) (string, error) {
	src, err := store.GetImageSource(ctx, id)
	if err != nil {
		return "", fmt.Errorf("downloadMissingImage: %w", err)
	}
	var size catalog.ImageSize
	if cfg.FullImageCacheEnabled() {
		size = catalog.ImageSizeFull
	} else {
		size = catalog.ImageSizeLarge
	}
	err = catalog.DownloadAndCacheImage(ctx, id, src, size)
	if err != nil {
		return "", fmt.Errorf("downloadMissingImage: %w", err)
	}
	return path.Join(catalog.SourceImageDir(), id.String()), nil
}
