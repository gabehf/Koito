package handlers

import (
	"net/http"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
	"github.com/google/uuid"
)

// SetPrimaryTrackAliasHandler sets the primary alias for a given track.
func SetPrimaryTrackAliasHandler(store db.TrackStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		trackID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("SetPrimaryTrackAliasHandler: Invalid track id")
			utils.WriteError(w, "invalid track id", http.StatusBadRequest)
			return
		}

		body, err := utils.DecodeBody[struct {
			Alias string `json:"alias"`
		}](r)
		if err != nil || body.Alias == "" {
			l.Debug().Msg("SetPrimaryTrackAliasHandler: Invalid or missing alias in request body")
			utils.WriteError(w, "alias must be provided", http.StatusBadRequest)
			return
		}

		if err = store.SetPrimaryTrackAlias(ctx, trackID, body.Alias); err != nil {
			l.Error().Err(err).Msg("SetPrimaryTrackAliasHandler: Failed to set track primary alias")
			utils.WriteError(w, "failed to set primary alias", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// SetPrimaryTrackArtistHandler sets whether a given artist is the primary artist for a track.
func SetPrimaryTrackArtistHandler(store db.ArtistStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		trackID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("SetPrimaryTrackArtistHandler: Invalid track id")
			utils.WriteError(w, "invalid track id", http.StatusBadRequest)
			return
		}

		artistID, err := utils.ParseIDParam(r, "artist_id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("SetPrimaryTrackArtistHandler: Invalid artist id")
			utils.WriteError(w, "invalid artist id", http.StatusBadRequest)
			return
		}

		body, err := utils.DecodeBody[struct {
			IsPrimary *bool `json:"is_primary"`
		}](r)
		if err != nil || body.IsPrimary == nil {
			l.Debug().Msg("SetPrimaryTrackArtistHandler: Invalid or missing is_primary in request body")
			utils.WriteError(w, "is_primary must be provided", http.StatusBadRequest)
			return
		}

		if err = store.SetPrimaryTrackArtist(ctx, trackID, artistID, *body.IsPrimary); err != nil {
			l.Error().Err(err).Msg("SetPrimaryTrackArtistHandler: Failed to set primary track artist")
			utils.WriteError(w, "failed to set primary track artist", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// UpdateTrackHandler updates a given track.
func UpdateTrackHandler(store db.TrackStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		trackID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("UpdateTrackHandler: Invalid track id")
			utils.WriteError(w, "invalid track id", http.StatusBadRequest)
			return
		}

		body, err := utils.DecodeBody[struct {
			MBID *string `json:"mbid"`
		}](r)
		if err != nil {
			l.Debug().Msg("UpdateTrackHandler: Invalid request body")
			utils.WriteError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if body.MBID == nil {
			l.Debug().Msg("UpdateTrackHandler: Request body contains no updatable fields")
			utils.WriteError(w, "no updatable fields provided", http.StatusBadRequest)
			return
		}

		updateOpts := db.UpdateTrackOpts{ID: trackID}

		if body.MBID != nil {
			mbid, err := uuid.Parse(*body.MBID)
			if err != nil {
				l.Debug().Msg("UpdateTrackHandler: Provided MusicBrainz ID is invalid")
				utils.WriteError(w, "provided musicbrainz id is invalid", http.StatusBadRequest)
				return
			}
			updateOpts.MusicBrainzID = mbid
		}

		if err = store.UpdateTrack(ctx, updateOpts); err != nil {
			l.Error().Err(err).Msg("UpdateTrackHandler: Failed to update track")
			utils.WriteError(w, "failed to update track", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
