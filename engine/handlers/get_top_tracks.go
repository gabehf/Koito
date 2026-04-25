package handlers

import (
	"net/http"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
)

func GetTopTracksHandler(store db.TrackStore) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("GetTopTracksHandler: Received request to retrieve top tracks")

		opts := OptsFromRequest(r)
		l.Debug().Msgf("GetTopTracksHandler: Retrieving top tracks with options: %+v", opts)

		tracks, err := store.GetTopTracksPaginated(ctx, opts)
		if err != nil {
			l.Err(err).Msg("GetTopTracksHandler: Failed to retrieve top tracks")
			utils.WriteError(w, "failed to get tracks", http.StatusBadRequest)
			return
		}

		l.Debug().Msg("GetTopTracksHandler: Successfully retrieved top tracks")
		utils.WriteJSON(w, http.StatusOK, tracks)
	}
}
