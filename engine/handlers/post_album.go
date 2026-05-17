package handlers

import (
	"net/http"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
)

// CreateAlbumAliasHandler creates a new alias for a given album.
func CreateAlbumAliasHandler(store db.AlbumStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		albumID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("CreateAlbumAliasHandler: Invalid album id")
			utils.WriteError(w, "invalid album id", http.StatusBadRequest)
			return
		}

		body, err := utils.DecodeBody[struct {
			Alias string `json:"alias"`
		}](r)
		if err != nil || body.Alias == "" {
			l.Debug().Msg("CreateAlbumAliasHandler: Invalid or missing alias in request body")
			utils.WriteError(w, "alias must be provided", http.StatusBadRequest)
			return
		}

		if err = store.SaveAlbumAliases(ctx, albumID, []string{body.Alias}, "Manual"); err != nil {
			l.Error().Err(err).Msg("CreateAlbumAliasHandler: Failed to save album alias")
			utils.WriteError(w, "failed to save alias", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}
