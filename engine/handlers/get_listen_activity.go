package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
)

func GetListenActivityHandler(store db.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("GetListenActivityHandler: Received request to retrieve listen activity")

		rangeStr := r.URL.Query().Get("range")
		_range, err := strconv.Atoi(rangeStr)
		if err != nil && rangeStr != "" {
			l.Debug().AnErr("error", err).Msg("GetListenActivityHandler: Invalid range parameter")
			utils.WriteError(w, "invalid range parameter", http.StatusBadRequest)
			return
		}

		monthStr := r.URL.Query().Get("month")
		month, err := strconv.Atoi(monthStr)
		if err != nil && monthStr != "" {
			l.Debug().AnErr("error", err).Msg("GetListenActivityHandler: Invalid month parameter")
			utils.WriteError(w, "invalid month parameter", http.StatusBadRequest)
			return
		}

		yearStr := r.URL.Query().Get("year")
		year, err := strconv.Atoi(yearStr)
		if err != nil && yearStr != "" {
			l.Debug().AnErr("error", err).Msg("GetListenActivityHandler: Invalid year parameter")
			utils.WriteError(w, "invalid year parameter", http.StatusBadRequest)
			return
		}

		artistIdStr := r.URL.Query().Get("artist_id")
		artistId, err := strconv.Atoi(artistIdStr)
		if err != nil && artistIdStr != "" {
			l.Debug().AnErr("error", err).Msg("GetListenActivityHandler: Invalid artist ID parameter")
			utils.WriteError(w, "invalid artist ID parameter", http.StatusBadRequest)
			return
		}

		albumIdStr := r.URL.Query().Get("album_id")
		albumId, err := strconv.Atoi(albumIdStr)
		if err != nil && albumIdStr != "" {
			l.Debug().AnErr("error", err).Msg("GetListenActivityHandler: Invalid album ID parameter")
			utils.WriteError(w, "invalid album ID parameter", http.StatusBadRequest)
			return
		}

		trackIdStr := r.URL.Query().Get("track_id")
		trackId, err := strconv.Atoi(trackIdStr)
		if err != nil && trackIdStr != "" {
			l.Debug().AnErr("error", err).Msg("GetListenActivityHandler: Invalid track ID parameter")
			utils.WriteError(w, "invalid track ID parameter", http.StatusBadRequest)
			return
		}

		var step db.StepInterval
		switch strings.ToLower(r.URL.Query().Get("step")) {
		case "day":
			step = db.StepDay
		case "week":
			step = db.StepWeek
		case "month":
			step = db.StepMonth
		case "year":
			step = db.StepYear
		default:
			l.Debug().Msgf("GetListenActivityHandler: Using default value '%s' for step", db.StepDefault)
			step = db.StepDay
		}

		opts := db.ListenActivityOpts{
			Step:     step,
			Range:    _range,
			Month:    month,
			Year:     year,
			AlbumID:  int32(albumId),
			ArtistID: int32(artistId),
			TrackID:  int32(trackId),
		}

		l.Debug().Msgf("GetListenActivityHandler: Retrieving listen activity with options: %+v", opts)

		activity, err := store.GetListenActivity(ctx, opts)
		if err != nil {
			l.Err(err).Msg("GetListenActivityHandler: Failed to retrieve listen activity")
			utils.WriteError(w, "failed to retrieve listen activity", http.StatusInternalServerError)
			return
		}

		l.Debug().Msg("GetListenActivityHandler: Successfully retrieved listen activity")
		utils.WriteJSON(w, http.StatusOK, activity)
	}
}
