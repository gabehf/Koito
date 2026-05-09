package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gabehf/koito/internal/catalog"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/sync/singleflight"
)

func ImageHandler(store db.ImageStore) http.HandlerFunc {

	// singleflight groups ensure that only one resize or download operation is triggered at once
	var (
		downloadGroup singleflight.Group
		resizeGroup   singleflight.Group
	)

	return func(w http.ResponseWriter, r *http.Request) {
		l := logger.FromContext(r.Context())
		imageIdStr := chi.URLParam(r, "image_id")
		size := chi.URLParam(r, "size")
		l.Debug().Msgf("ImageHandler: Received request to retrieve image with size '%s' and id '%s'", size, imageIdStr)

		imageSize, err := catalog.ParseImageSize(size)
		if err != nil {
			l.Debug().Msg("ImageHandler: Invalid image size parameter")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		imgid, err := uuid.Parse(imageIdStr)
		if err != nil {
			l.Debug().Msg("ImageHandler: Invalid image id, serving default image")
			serveDefaultImage(w, imageSize)
			return
		}

		desiredImgPath := catalog.BuildImagePath(imgid, imageSize)
		if _, err := os.Stat(desiredImgPath); os.IsNotExist(err) {
			l.Debug().Msg("ImageHandler: Image not found in desired size, attempting to retrieve source image")
			sourcePath := catalog.BuildImagePath(imgid, catalog.ImageSizeSource)

			if _, err = os.Stat(sourcePath); os.IsNotExist(err) {
				l.Warn().Msgf("ImageHandler: Could not find requested image %s. Attempting to download from source", imgid.String())

				_, err, _ = downloadGroup.Do(imgid.String(), func() (any, error) {
					return nil, downloadMissingImage(r.Context(), store, imgid)
				})
				if err != nil {
					l.Err(err).Msg("ImageHandler: Failed to redownload missing image")
					serveDefaultImage(w, imageSize)
					return
				}
			}

			l.Debug().Msgf("ImageHandler: Found source image file at path '%s'", sourcePath)
			imageBuf, err := os.ReadFile(sourcePath)
			if err != nil {
				l.Err(err).Msg("ImageHandler: Failed to read source image file")
				serveDefaultImage(w, imageSize)
				return
			}

			resizeKey := imgid.String() + "/" + imageSize.String()
			_, err, _ = resizeGroup.Do(resizeKey, func() (any, error) {
				return nil, catalog.CompressAndSaveImage(r.Context(), imgid, imageSize, bytes.NewReader(imageBuf))
			})
			if err != nil {
				l.Err(err).Msg("ImageHandler: Failed to save compressed image to cache")
				w.WriteHeader(http.StatusInternalServerError)
				return
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

func defaultSVG(size catalog.ImageSize) []byte {
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

func serveDefaultImage(w http.ResponseWriter, size catalog.ImageSize) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	w.Write(defaultSVG(size))
}

// finds the item associated with the image id, downloads it, and saves it in the source path, returning the path to the image
func downloadMissingImage(ctx context.Context, store db.ImageStore, id uuid.UUID) error {
	src, err := store.GetImageSource(ctx, id)
	if err != nil {
		return fmt.Errorf("downloadMissingImage: %w", err)
	}
	err = catalog.DownloadAndCacheImage(ctx, id, src, catalog.ImageSizeSource)
	if err != nil {
		return fmt.Errorf("downloadMissingImage: %w", err)
	}
	return nil
}
