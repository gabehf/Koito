package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gabehf/koito/engine/middleware"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
	"github.com/google/uuid"
)

type lbzGetFeedbackResponse struct {
	Count      int                  `json:"count"`
	Offset     int                  `json:"offset"`
	TotalCount int64                `json:"total_count"`
	Feedback   []LbzFeedbackPayload `json:"feedback"`
}

type lbzRecordingFeedbackRequest struct {
	RecordingMbid string `json:"recording_mbid"`
	RecordingMsid string `json:"recording_msid"`
	Score         int    `json:"score"`
}

func LbzGetFeedbackHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		u, ok := lbzSelfUser(w, r)
		if !ok {
			return
		}

		count, err := lbzCount(r)
		if err != nil {
			utils.WriteError(w, err.Error(), http.StatusBadRequest)
			return
		}
		offset, err := lbzOptionalInt(r, "offset")
		if err != nil || offset < 0 {
			utils.WriteError(w, "offset must be a non-negative integer", http.StatusBadRequest)
			return
		}

		items, total, err := store.GetTrackFeedback(ctx, u.ID, count, int(offset))
		if err != nil {
			l.Err(err).Msg("LbzGetFeedbackHandler: Failed to get feedback")
			utils.WriteError(w, "failed to get feedback", http.StatusInternalServerError)
			return
		}

		resp := lbzGetFeedbackResponse{
			Count:      len(items),
			Offset:     int(offset),
			TotalCount: total,
			Feedback:   make([]LbzFeedbackPayload, 0, len(items)),
		}
		for _, item := range items {
			md := lbzTrackMetadataFromDetail(&item.Track)
			resp.Feedback = append(resp.Feedback, LbzFeedbackPayload{
				Created:       item.CreatedAt.Unix(),
				RecordingMbid: uuidPtrToString(item.Track.TrackMbid),
				Score:         item.Score,
				TrackMetadata: &md,
			})
		}
		utils.WriteJSON(w, http.StatusOK, resp)
	}
}

// LbzRecordingFeedbackHandler stores a love (1), hate (-1) or cleared (0) rating for a track.
// Only recording_mbid identifies a track: Koito does not issue recording_msid values, so
// feedback can only target tracks that carry a MusicBrainz recording ID.
func LbzRecordingFeedbackHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		u := middleware.GetUserFromContext(ctx)
		if u == nil {
			utils.WriteError(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var req lbzRecordingFeedbackRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteError(w, "failed to decode request", http.StatusBadRequest)
			return
		}
		if req.Score < -1 || req.Score > 1 {
			utils.WriteError(w, "score must be -1, 0 or 1", http.StatusBadRequest)
			return
		}
		if req.RecordingMbid == "" {
			utils.WriteError(w, "recording_mbid is required; recording_msid is not supported", http.StatusBadRequest)
			return
		}
		mbid, err := uuid.Parse(req.RecordingMbid)
		if err != nil {
			utils.WriteError(w, "recording_mbid must be a valid MusicBrainz ID", http.StatusBadRequest)
			return
		}

		track, err := store.GetTrack(ctx, db.GetTrackOpts{MusicBrainzID: mbid})
		if errors.Is(err, db.ErrNotFound) {
			utils.WriteError(w, "recording not found", http.StatusNotFound)
			return
		}
		if err != nil {
			l.Err(err).Msg("LbzRecordingFeedbackHandler: Failed to look up recording")
			utils.WriteError(w, "failed to save feedback", http.StatusInternalServerError)
			return
		}

		err = store.SaveTrackFeedback(ctx, db.SaveTrackFeedbackOpts{
			UserID:  u.ID,
			TrackID: track.ID,
			Score:   req.Score,
		})
		if err != nil {
			l.Err(err).Msg("LbzRecordingFeedbackHandler: Failed to save feedback")
			utils.WriteError(w, "failed to save feedback", http.StatusInternalServerError)
			return
		}

		utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
