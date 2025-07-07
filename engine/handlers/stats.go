package handlers

import (
	"net/http"
	"strings"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
)

type StatsResponse struct {
	ListenCount     int64 `json:"listen_count"`
	TrackCount      int64 `json:"track_count"`
	AlbumCount      int64 `json:"album_count"`
	ArtistCount     int64 `json:"artist_count"`
	MinutesListened int64 `json:"minutes_listened"`
}

func StatsHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logger.FromContext(r.Context())

		l.Debug().Msg("StatsHandler: Received request to retrieve statistics")

		var period db.Period
		switch strings.ToLower(r.URL.Query().Get("period")) {
		case "day":
			period = db.PeriodDay
		case "week":
			period = db.PeriodWeek
		case "month":
			period = db.PeriodMonth
		case "year":
			period = db.PeriodYear
		case "all_time":
			period = db.PeriodAllTime
		default:
			l.Debug().Msgf("StatsHandler: Using default value '%s' for period", db.PeriodDay)
			period = db.PeriodDay
		}

		l.Debug().Msgf("StatsHandler: Fetching statistics for period '%s'", period)

		listens, err := store.CountListens(r.Context(), period)
		if err != nil {
			l.Err(err).Msg("StatsHandler: Failed to fetch listen count")
			utils.WriteError(w, "failed to get listens: "+err.Error(), http.StatusInternalServerError)
			return
		}

		tracks, err := store.CountTracks(r.Context(), period)
		if err != nil {
			l.Err(err).Msg("StatsHandler: Failed to fetch track count")
			utils.WriteError(w, "failed to get tracks: "+err.Error(), http.StatusInternalServerError)
			return
		}

		albums, err := store.CountAlbums(r.Context(), period)
		if err != nil {
			l.Err(err).Msg("StatsHandler: Failed to fetch album count")
			utils.WriteError(w, "failed to get albums: "+err.Error(), http.StatusInternalServerError)
			return
		}

		artists, err := store.CountArtists(r.Context(), period)
		if err != nil {
			l.Err(err).Msg("StatsHandler: Failed to fetch artist count")
			utils.WriteError(w, "failed to get artists: "+err.Error(), http.StatusInternalServerError)
			return
		}

		timeListenedS, err := store.CountTimeListened(r.Context(), period)
		if err != nil {
			l.Err(err).Msg("StatsHandler: Failed to fetch time listened")
			utils.WriteError(w, "failed to get time listened: "+err.Error(), http.StatusInternalServerError)
			return
		}

		l.Debug().Msg("StatsHandler: Successfully fetched statistics")
		utils.WriteJSON(w, http.StatusOK, StatsResponse{
			ListenCount:     listens,
			TrackCount:      tracks,
			AlbumCount:      albums,
			ArtistCount:     artists,
			MinutesListened: timeListenedS / 60,
		})
	}
}
