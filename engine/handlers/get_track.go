package handlers

import (
	"net/http"
	"strconv"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
)

func GetTrackHandler(store db.TrackStore) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("GetTrackHandler: Received request to retrieve track")

		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			l.Debug().Msg("GetTrackHandler: Missing track ID in request")
			utils.WriteError(w, "id must be provided", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			l.Debug().AnErr("error", err).Msg("GetTrackHandler: Invalid track ID")
			utils.WriteError(w, "id is invalid", http.StatusBadRequest)
			return
		}

		l.Debug().Msgf("GetTrackHandler: Retrieving track with ID %d", id)

		track, err := store.GetTrack(ctx, db.GetTrackOpts{ID: int32(id)})
		if err != nil {
			l.Err(err).Msgf("GetTrackHandler: Failed to retrieve track with ID %d", id)
			utils.WriteError(w, "track with specified id could not be found", http.StatusNotFound)
			return
		}

		l.Debug().Msgf("GetTrackHandler: Successfully retrieved track with ID %d", id)
		utils.WriteJSON(w, http.StatusOK, track)
	}
}
