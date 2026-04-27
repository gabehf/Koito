package handlers

import (
	"net/http"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
)

func GetTopArtistsHandler(store db.ArtistStore) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("GetTopArtistsHandler: Received request to retrieve top artists")

		opts := OptsFromRequest(r)
		l.Debug().Msgf("GetTopArtistsHandler: Retrieving top artists with options: %+v", opts)

		artists, err := store.GetTopArtistsPaginated(ctx, opts)
		if err != nil {
			l.Err(err).Msg("GetTopArtistsHandler: Failed to retrieve top artists")
			utils.WriteError(w, "failed to get artists", http.StatusBadRequest)
			return
		}

		l.Debug().Msg("GetTopArtistsHandler: Successfully retrieved top artists")
		utils.WriteJSON(w, http.StatusOK, artists)
	}
}
