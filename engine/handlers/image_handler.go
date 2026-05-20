package handlers

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/gabehf/koito/imagecache"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"golang.org/x/sync/singleflight"
)

func ImageHandler(store db.ImageStore) http.HandlerFunc {
	var downloadGroup singleflight.Group
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)
		imageIdStr := chi.URLParam(r, "image_id")
		filename := chi.URLParam(r, "filename")

		l.Debug().Msgf("ImageHandler: Received request to retrieve image '%s' with id '%s'", filename, imageIdStr)

		imgid, err := uuid.Parse(imageIdStr)
		if err != nil {
			l.Debug().Msgf("ImageHandler: Invalid image id '%s'", imageIdStr)
			http.NotFound(w, r)
			return
		}

		image, err := imagecache.GetImage(imgid, filename)
		if errors.Is(err, fs.ErrNotExist) {
			l.Debug().Err(err).Msgf("ImageHandler: Could not find requested image %s. Attempting to download from source", imgid.String())
			image, err = imageHandlerRedownload(w, r, l, store, &downloadGroup, imgid, filename)
			if err != nil {
				return
			}
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		l.Debug().Msgf("ImageHandler: Serving image from path '%s'", image.Path)
		w.Header().Set("Content-Type", image.Mime)
		w.Header().Set("Cache-Control", "public, max-age=2592000")
		http.ServeFile(w, r, image.Path)
	}
}

func imageHandlerRedownload(w http.ResponseWriter, r *http.Request, l *zerolog.Logger, store db.ImageStore, downloadGroup *singleflight.Group, imgid uuid.UUID, filename string) (*imagecache.ImageInfo, error) {
	ctx := r.Context()

	src, err := store.GetImageSource(ctx, imgid)
	if errors.Is(err, db.ErrNotFound) {
		l.Debug().Msgf("Image with id '%s' not found", imgid.String())
		http.NotFound(w, r)
		return nil, err
	} else if err != nil {
		l.Debug().Err(err).Msgf("Failed to get image source for image '%s'", imgid.String())
		w.WriteHeader(http.StatusInternalServerError)
		return nil, err
	} else {
		_, err, _ = downloadGroup.Do(imgid.String(), func() (any, error) {
			if err := imagecache.DownloadImage(imgid, src); err != nil {
				l.Err(err).Msg("ImageHandler: Failed to redownload missing image")
				imageSize, err := imagecache.ParseImageSize(filename)
				if err != nil {
					http.NotFound(w, r)
					return nil, err
				}
				serveDefaultImage(w, imageSize)
				return nil, err
			}
			return nil, nil
		})
		if err != nil {
			// errors are already handled in the func
			return nil, err
		}
	}

	image, err := imagecache.GetImage(imgid, filename)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}
	return image, nil
}

func defaultSVG(size imagecache.ImageSize) []byte {
	px := size.Width()

	icon := `<path d="M263.999786,3641.05125 L263.999786,3650.60392 C263.999786,3652.8603 262.302786,3654.83873 260.103786,3654.89616 C257.683786,3654.95873 255.733786,3652.8162 256.029786,3650.287 C256.233786,3648.54036 257.568786,3647.07167 259.257786,3646.7609 C260.262786,3646.57527 261.215786,3646.77629 261.999786,3647.2409 L261.999786,3642.07688 C261.999786,3641.51073 261.551786,3641.05125 260.999786,3641.05125 L252.999786,3641.05125 C252.447786,3641.05125 251.999786,3641.51073 251.999786,3642.07688 L251.999786,3654.70642 C251.999786,3656.9628 250.302786,3658.94123 248.103786,3658.99866 C245.683786,3659.06123 243.733786,3656.9187 244.029786,3654.3895 C244.233786,3652.64286 245.568786,3651.17417 247.257786,3650.86341 C248.262786,3650.67777 249.215786,3650.87879 249.999786,3651.3434 L249.999786,3641.05125 C249.999786,3639.91793 250.895786,3639 251.999786,3639 L261.999786,3639 C263.104786,3639 263.999786,3639.91793 263.999786,3641.05125" transform="scale(0.5) translate(-244, -3639) translate(9, 12)" fill="%s"/>`

	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg"
        width="%d"
        height="%d"
        viewBox="0 0 20 20"
        style="background:#eeeeee">
        %s
    </svg>`, px, px, fmt.Sprintf(icon, "#aaaaaa"))

	return []byte(svg)
}

func serveDefaultImage(w http.ResponseWriter, size imagecache.ImageSize) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	w.Write(defaultSVG(size))
}

// finds the item associated with the image id, downloads it, and saves it in the source path, returning the path to the image
func downloadMissingImage(ctx context.Context, store db.ImageStore, id uuid.UUID) error {
	return nil
}
