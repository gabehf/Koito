package handlers

import (
	"net/http"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
)

func DeleteArtistHandler(store db.ArtistStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("DeleteArtistHandler: Received request to delete artist")

		artistID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("DeleteArtistHandler: Invalid artist id")
			utils.WriteError(w, "invalid artist id", http.StatusBadRequest)
			return
		}

		l.Debug().Msgf("DeleteArtistHandler: Deleting artist with ID %d", artistID)

		err = store.DeleteArtist(ctx, int32(artistID))
		if err != nil {
			l.Err(err).Msg("DeleteArtistHandler: Failed to delete artist")
			utils.WriteError(w, "failed to delete artist", http.StatusInternalServerError)
			return
		}

		l.Debug().Msgf("DeleteArtistHandler: Successfully deleted artist with ID %d", artistID)
		w.WriteHeader(http.StatusNoContent)
	}
}

// DeleteArtistAliasHandler deletes an alias for a given artist.
func DeleteArtistAliasHandler(store db.ArtistStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		artistID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("DeleteArtistAliasHandler: Invalid artist id")
			utils.WriteError(w, "invalid artist id", http.StatusBadRequest)
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

		if err = store.DeleteArtistAlias(ctx, artistID, body.Alias); err != nil {
			l.Error().Err(err).Msg("DeleteArtistAliasHandler: Failed to delete artist alias")
			utils.WriteError(w, "failed to delete alias", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
