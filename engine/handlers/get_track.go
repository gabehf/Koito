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

		id, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("GetTrackHandler: Invalid album id")
			utils.WriteError(w, "invalid album id", http.StatusBadRequest)
			return
		}

		l.Debug().Msgf("GetTrackHandler: Retrieving track with ID %d", id)

		track, err := store.GetTrack(ctx, db.GetTrackOpts{ID: id})
		if err != nil {
			l.Err(err).Msgf("GetTrackHandler: Failed to retrieve track with ID %d", id)
			utils.WriteError(w, "track with specified id could not be found", http.StatusNotFound)
			return
		}

		l.Debug().Msgf("GetTrackHandler: Successfully retrieved track with ID %d", id)
		utils.WriteJSON(w, http.StatusOK, track)
	}
}

// GetTrackAliasesHandler retrieves all aliases for a given track.
func GetTrackAliasesHandler(store db.TrackStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		trackID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("GetTrackAliasesHandler: Invalid track id")
			utils.WriteError(w, "invalid track id", http.StatusBadRequest)
			return
		}

		aliases, err := store.GetAllTrackAliases(ctx, trackID)
		if err != nil {
			l.Err(err).Msg("GetTrackAliasesHandler: Failed to get track aliases")
			utils.WriteError(w, "failed to retrieve aliases", http.StatusInternalServerError)
			return
		}

		utils.WriteJSON(w, http.StatusOK, aliases)
	}
}

// GetArtistsForTrackHandler retrieves all artists for a given track.
func GetArtistsForTrackHandler(store db.ArtistStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		trackID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("GetArtistsForTrackHandler: Invalid track id")
			utils.WriteError(w, "invalid track id", http.StatusBadRequest)
			return
		}

		artists, err := store.GetArtistsForTrack(ctx, trackID)
		if err != nil {
			l.Err(err).Msg("GetArtistsForTrackHandler: Failed to retrieve artists")
			utils.WriteError(w, "failed to retrieve artists", http.StatusInternalServerError)
			return
		}

		utils.WriteJSON(w, http.StatusOK, artists)
	}
}

// GetTrackInterestHandler retrieves interest data for a given track.
func GetTrackInterestHandler(store db.ListenStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		trackID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("GetTrackInterestHandler: Invalid track id")
			utils.WriteError(w, "invalid track id", http.StatusBadRequest)
			return
		}

		buckets, err := strconv.Atoi(r.URL.Query().Get("buckets"))
		if err != nil {
			l.Debug().Msg("GetTrackInterestHandler: Buckets is not an integer")
			utils.WriteError(w, "parameter 'buckets' must be an integer", http.StatusBadRequest)
			return
		}

		interest, err := store.GetInterest(ctx, db.GetInterestOpts{
			TrackID: trackID,
			Buckets: buckets,
		})
		if err != nil {
			l.Err(err).Msg("GetTrackInterestHandler: Failed to query interest")
			utils.WriteError(w, "failed to retrieve interest: "+err.Error(), http.StatusInternalServerError)
			return
		}

		utils.WriteJSON(w, http.StatusOK, interest)
	}
}
