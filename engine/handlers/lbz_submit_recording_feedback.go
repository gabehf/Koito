package handlers

import (
	"errors"
	"net/http"

	"github.com/gabehf/koito/engine/middleware"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
	"github.com/google/uuid"
)

// LbzSubmitRecordingFeedbackHandler submits love/hate feedback for a recording.
func LbzSubmitRecordingFeedbackHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		u := middleware.GetUserFromContext(ctx)
		if u == nil {
			utils.WriteError(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		body, err := utils.DecodeBody[struct {
			RecordingMBID string `json:"recording_mbid"`
			Score         int    `json:"score"`
		}](r)
		if err != nil {
			utils.WriteError(w, "invalid JSON sent", http.StatusBadRequest)
			return
		}
		if body.Score != 1 && body.Score != -1 && body.Score != 0 {
			utils.WriteError(w, "score must be 1, 0, or -1", http.StatusBadRequest)
			return
		}

		mbid, err := uuid.Parse(body.RecordingMBID)
		if err != nil {
			utils.WriteError(w, "recording_mbid is required", http.StatusBadRequest)
			return
		}

		track, err := store.GetTrack(ctx, db.GetTrackOpts{MusicBrainzID: mbid})
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				utils.WriteError(w, "track with specified recording_mbid could not be found", http.StatusNotFound)
				return
			}
			l.Err(err).Msg("LbzSubmitRecordingFeedbackHandler: Failed to resolve track")
			utils.WriteError(w, "failed to resolve track", http.StatusInternalServerError)
			return
		}

		if err := store.SetTrackFeedback(ctx, u.ID, track.ID, body.Score); err != nil {
			l.Err(err).Msg("LbzSubmitRecordingFeedbackHandler: Failed to save feedback")
			utils.WriteError(w, "failed to save feedback", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	}
}
