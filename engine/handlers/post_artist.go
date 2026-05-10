package handlers

import (
	"net/http"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
)

// CreateArtistAliasHandler creates a new alias for a given artist.
func CreateArtistAliasHandler(store db.ArtistStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		artistID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("CreateArtistAliasHandler: Invalid artist id")
			utils.WriteError(w, "invalid artist id", http.StatusBadRequest)
			return
		}

		body, err := utils.DecodeBody[struct {
			Alias string `json:"alias"`
		}](r)
		if err != nil || body.Alias == "" {
			if err != nil {
				l.Debug().AnErr("error", err).Msg("Error decoding request body")
			}
			l.Debug().Msg("CreateArtistAliasHandler: Invalid or missing alias in request body")
			utils.WriteError(w, "alias must be provided", http.StatusBadRequest)
			return
		}

		if err = store.SaveArtistAliases(ctx, artistID, []string{body.Alias}, "Manual"); err != nil {
			l.Error().Err(err).Msg("CreateArtistAliasHandler: Failed to save artist alias")
			utils.WriteError(w, "failed to save alias", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}
