package handlers

import (
	"net/http"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/memkv"
	"github.com/gabehf/koito/internal/models"
	"github.com/gabehf/koito/internal/utils"
)

type NowPlayingResponse struct {
	CurrentlyPlaying bool         `json:"currently_playing"`
	StartedAt        int64        `json:"started_at"`
	Track            models.Track `json:"track"`
}

func NowPlayingHandler(store db.TrackStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("NowPlayingHandler: Got request")

		// Hardcoded user id as 1. Not great but it works until (if) multi-user is supported.
		if nowPlayingDataI, ok := memkv.Store.Get("1"); !ok {
			utils.WriteJSON(w, http.StatusOK, NowPlayingResponse{CurrentlyPlaying: false})
		} else if nowPlayingData, ok := nowPlayingDataI.(models.NowPlayingData); !ok {
			l.Debug().Msg("NowPlayingHandler: Failed type assertion for nowPlayingDataI")
			utils.WriteError(w, "internal server error", http.StatusInternalServerError)
		} else {
			track, err := store.GetTrack(ctx, db.GetTrackOpts{ID: nowPlayingData.TrackID})
			if err != nil {
				l.Error().Err(err).Msg("NowPlayingHandler: Failed to get track from database")
				utils.WriteError(w, "failed to fetch currently playing track from database", http.StatusInternalServerError)
			} else {
				utils.WriteJSON(w, http.StatusOK, NowPlayingResponse{CurrentlyPlaying: true, StartedAt: nowPlayingData.StartTime.Unix(), Track: *track})
			}
		}
	}
}
