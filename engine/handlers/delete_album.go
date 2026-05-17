package handlers

import (
	"net/http"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
)

func DeleteAlbumHandler(store db.AlbumStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("DeleteAlbumHandler: Received request to delete album")

		albumID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("DeleteAlbumHandler: Invalid album id")
			utils.WriteError(w, "invalid album id", http.StatusBadRequest)
			return
		}

		l.Debug().Msgf("DeleteAlbumHandler: Deleting album with ID %d", albumID)

		err = store.DeleteAlbum(ctx, int32(albumID))
		if err != nil {
			l.Err(err).Msg("DeleteAlbumHandler: Failed to delete album")
			utils.WriteError(w, "failed to delete album", http.StatusInternalServerError)
			return
		}

		l.Debug().Msgf("DeleteAlbumHandler: Successfully deleted album with ID %d", albumID)
		w.WriteHeader(http.StatusNoContent)
	}
}

// DeleteAlbumAliasHandler deletes an alias for a given album.
func DeleteAlbumAliasHandler(store db.AlbumStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		albumID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("DeleteAlbumAliasHandler: Invalid album id")
			utils.WriteError(w, "invalid album id", http.StatusBadRequest)
			return
		}

		body, err := utils.DecodeBody[struct {
			Alias string `json:"alias"`
		}](r)
		if err != nil || body.Alias == "" {
			l.Debug().Msg("DeleteArtistAliasHandler: Invalid or missing alias in request body")
			utils.WriteError(w, "alias must be provided", http.StatusBadRequest)
			return
		}

		if err = store.DeleteAlbumAlias(ctx, albumID, body.Alias); err != nil {
			l.Error().Err(err).Msg("DeleteAlbumAliasHandler: Failed to delete album alias")
			utils.WriteError(w, "failed to delete alias", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
