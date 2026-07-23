package handlers

import (
	"net/http"

	"github.com/gabehf/koito/engine/middleware"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
)

// CreateTrackAliasHandler creates a new alias for a given track.
func CreateTrackAliasHandler(store db.TrackStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		trackID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("CreateTrackAliasHandler: Invalid track id")
			utils.WriteError(w, "invalid track id", http.StatusBadRequest)
			return
		}

		body, err := utils.DecodeBody[struct {
			Alias string `json:"alias"`
		}](r)
		if err != nil || body.Alias == "" {
			l.Debug().Msg("CreateTrackAliasHandler: Invalid or missing alias in request body")
			utils.WriteError(w, "alias must be provided", http.StatusBadRequest)
			return
		}

		if err = store.SaveTrackAliases(ctx, trackID, []string{body.Alias}, "Manual"); err != nil {
			l.Error().Err(err).Msg("CreateTrackAliasHandler: Failed to save track alias")
			utils.WriteError(w, "failed to save alias", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

// AddTrackArtistsHandler adds artists to a given track.
func AddTrackArtistsHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		trackID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("AddTrackArtistsHandler: Invalid track id")
			utils.WriteError(w, "invalid track id", http.StatusBadRequest)
			return
		}

		body, err := utils.DecodeBody[struct {
			Artists []int32 `json:"artist_ids"`
		}](r)
		if err != nil || len(body.Artists) == 0 {
			l.Debug().Msg("AddTrackArtistsHandler: Invalid or missing artist_ids in request body")
			utils.WriteError(w, "artist_ids must be provided", http.StatusBadRequest)
			return
		}

		if err = store.UpdateTrack(ctx, db.UpdateTrackOpts{
			ID:         trackID,
			AddArtists: body.Artists,
		}); err != nil {
			l.Error().Err(err).Msg("AddTrackArtistsHandler: Failed to add artists to track")
			utils.WriteError(w, "failed to add artists to track", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// SubmitTrackFeedbackHandler sets the current user's feedback score for a track.
func SubmitTrackFeedbackHandler(store db.FeedbackStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		u := middleware.GetUserFromContext(ctx)
		if u == nil {
			utils.WriteError(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		trackID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("SubmitTrackFeedbackHandler: Invalid track id")
			utils.WriteError(w, "invalid track id", http.StatusBadRequest)
			return
		}

		body, err := utils.DecodeBody[struct {
			Score int `json:"score"`
		}](r)
		if err != nil || (body.Score != 1 && body.Score != -1 && body.Score != 0) {
			utils.WriteError(w, "score must be 1, 0, or -1", http.StatusBadRequest)
			return
		}

		if err := store.SetTrackFeedback(ctx, u.ID, trackID, body.Score); err != nil {
			l.Err(err).Msg("SubmitTrackFeedbackHandler: Failed to save feedback")
			utils.WriteError(w, "failed to save feedback", http.StatusInternalServerError)
			return
		}

		utils.WriteJSON(w, http.StatusOK, body)
	}
}
