package handlers

import (
	"net/http"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
)

func DeleteTrackHandler(store db.TrackStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("DeleteTrackHandler: Received request to delete track")

		trackID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("DeleteTrackHandler: Invalid track id")
			utils.WriteError(w, "invalid track id", http.StatusBadRequest)
			return
		}

		l.Debug().Msgf("DeleteTrackHandler: Deleting track with ID %d", trackID)

		err = store.DeleteTrack(ctx, int32(trackID))
		if err != nil {
			l.Err(err).Msg("DeleteTrackHandler: Failed to delete track")
			utils.WriteError(w, "failed to delete track", http.StatusInternalServerError)
			return
		}

		l.Debug().Msgf("DeleteTrackHandler: Successfully deleted track with ID %d", trackID)
		w.WriteHeader(http.StatusNoContent)
	}
}

// DeleteTrackAliasHandler deletes an alias for a given track.
func DeleteTrackAliasHandler(store db.TrackStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		trackID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("DeleteTrackAliasHandler: Invalid track id")
			utils.WriteError(w, "invalid track id", http.StatusBadRequest)
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

		if err = store.DeleteTrackAlias(ctx, trackID, body.Alias); err != nil {
			l.Error().Err(err).Msg("DeleteTrackAliasHandler: Failed to delete track alias")
			utils.WriteError(w, "failed to delete alias", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// DeleteTrackArtistHandler removes an artist from a given track.
func DeleteTrackArtistHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		trackID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("RemoveTrackArtistHandler: Invalid track id")
			utils.WriteError(w, "invalid track id", http.StatusBadRequest)
			return
		}

		artistID, err := utils.ParseIDParam(r, "artist_id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("RemoveTrackArtistHandler: Invalid artist id")
			utils.WriteError(w, "invalid artist id", http.StatusBadRequest)
			return
		}

		if err = store.UpdateTrack(ctx, db.UpdateTrackOpts{
			ID:            trackID,
			RemoveArtists: []int32{artistID},
		}); err != nil {
			l.Error().Err(err).Msg("RemoveTrackArtistHandler: Failed to remove artist from track")
			utils.WriteError(w, "failed to remove artist from track", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
