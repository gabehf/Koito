package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gabehf/koito/imagecache"
	"github.com/gabehf/koito/internal/catalog"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/images"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type ReplaceImageResponse struct {
	Success bool   `json:"success"`
	Image   string `json:"image"`
	Message string `json:"message,omitempty"`
}

// ReplaceArtistImageHandler replaces the image for a given artist.
func ReplaceArtistImageHandler(store db.ArtistStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		artistID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("ReplaceArtistImageHandler: Invalid artist id")
			utils.WriteError(w, "invalid artist id", http.StatusBadRequest)
			return
		}

		artist, err := store.GetArtist(ctx, db.GetArtistOpts{ID: artistID})
		if err != nil {
			l.Err(err).Msg("ReplaceArtistImageHandler: Artist with specified ID could not be found")
			utils.WriteError(w, "artist with specified id could not be found", http.StatusBadRequest)
			return
		}

		id, imgsrc, err := resolveImage(r, l)
		if err != nil {
			utils.WriteError(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = store.UpdateArtist(ctx, db.UpdateArtistOpts{
			ID:       artistID,
			Image:    id,
			ImageSrc: imgsrc,
		}); err != nil {
			l.Err(err).Msg("ReplaceArtistImageHandler: Artist image could not be updated")
			utils.WriteError(w, "artist image could not be updated", http.StatusInternalServerError)
			return
		}

		if artist.Image.Small != "" {
			if err = imagecache.DeleteImage(*parseOldImage(artist.Image.Small)); err != nil {
				l.Err(err).Msg("ReplaceArtistImageHandler: Failed to delete old image file")
				utils.WriteError(w, "could not delete old image file", http.StatusInternalServerError)
				return
			}
		}

		utils.WriteJSON(w, http.StatusOK, ReplaceImageResponse{Success: true, Image: id.String()})
	}
}

// ReplaceAlbumImageHandler replaces the image for a given album.
func ReplaceAlbumImageHandler(store db.AlbumStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		albumID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("ReplaceAlbumImageHandler: Invalid album id")
			utils.WriteError(w, "invalid album id", http.StatusBadRequest)
			return
		}

		album, err := store.GetAlbum(ctx, db.GetAlbumOpts{ID: albumID})
		if err != nil {
			l.Err(err).Msg("ReplaceAlbumImageHandler: Album with specified ID could not be found")
			utils.WriteError(w, "album with specified id could not be found", http.StatusBadRequest)
			return
		}

		id, imgsrc, err := resolveImage(r, l)
		if err != nil {
			utils.WriteError(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = store.UpdateAlbum(ctx, db.UpdateAlbumOpts{
			ID:       albumID,
			Image:    id,
			ImageSrc: imgsrc,
		}); err != nil {
			l.Err(err).Msg("ReplaceAlbumImageHandler: Album image could not be updated")
			utils.WriteError(w, "album image could not be updated", http.StatusInternalServerError)
			return
		}

		if album.Image.Small != "" {
			if err = imagecache.DeleteImage(*parseOldImage(album.Image.Small)); err != nil {
				l.Err(err).Msg("ReplaceAlbumImageHandler: Failed to delete old image file")
				utils.WriteError(w, "could not delete old image file", http.StatusInternalServerError)
				return
			}
		}

		utils.WriteJSON(w, http.StatusOK, ReplaceImageResponse{Success: true, Image: id.String()})
	}
}

// resolveImage extracts and caches an image from either a URL or file upload in the request.
// Returns the new image UUID and source string.
func resolveImage(r *http.Request, l *zerolog.Logger) (uuid.UUID, string, error) {
	id := uuid.New()

	fileUrl := r.FormValue("image_url")
	if fileUrl != "" {
		l.Debug().Msg("resolveImage: Image identified as remote file")
		if err := images.ValidateImageURL(fileUrl); err != nil {
			l.Debug().AnErr("error", err).Msg("resolveImage: Invalid image URL")
			return uuid.UUID{}, "", fmt.Errorf("url is invalid or not an image file")
		}
		l.Debug().Msg("resolveImage: Downloading image from source")
		if err := imagecache.DownloadImage(id, fileUrl); err != nil {
			l.Err(err).Msg("resolveImage: Failed to cache image")
			return uuid.UUID{}, "", fmt.Errorf("failed to cache image")
		}
		return id, fileUrl, nil
	}

	l.Debug().Msg("resolveImage: Image identified as uploaded file")
	file, _, err := r.FormFile("image")
	if err != nil {
		l.Err(err).Msg("resolveImage: Invalid file upload")
		return uuid.UUID{}, "", fmt.Errorf("invalid file")
	}
	defer file.Close()

	buf := make([]byte, 512)
	if _, err := file.Read(buf); err != nil {
		l.Err(err).Msg("resolveImage: Could not read file")
		return uuid.UUID{}, "", fmt.Errorf("could not read file")
	}
	if !strings.HasPrefix(http.DetectContentType(buf), "image/") {
		l.Debug().Msg("resolveImage: Uploaded file is not an image")
		return uuid.UUID{}, "", fmt.Errorf("only image uploads are allowed")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		l.Err(err).Msg("resolveImage: Could not seek file")
		return uuid.UUID{}, "", fmt.Errorf("could not seek file")
	}

	l.Debug().Msg("resolveImage: Saving image to cache")
	if err := imagecache.SaveImage(id, file); err != nil {
		l.Err(err).Msg("resolveImage: Could not save file")
		return uuid.UUID{}, "", fmt.Errorf("could not save file")
	}
	return id, catalog.ImageSourceUserUpload, nil
}

// parses the image id from /image/{uuid}/size.webp style links
func parseOldImage(link string) *uuid.UUID {
	ss := strings.Split(link, "/")
	if len(ss) < 4 {
		return nil
	}
	if parsed, err := uuid.Parse(ss[2]); err != nil {
		return nil
	} else {
		return &parsed
	}
}
