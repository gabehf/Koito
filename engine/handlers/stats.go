package handlers

import (
	"net/http"

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

		tf := TimeframeFromRequest(r)

		l.Debug().Msg("StatsHandler: Fetching statistics")

		listens, err := store.CountListens(r.Context(), tf)
		if err != nil {
			l.Err(err).Msg("StatsHandler: Failed to fetch listen count")
			utils.WriteError(w, "failed to get listens: "+err.Error(), http.StatusInternalServerError)
			return
		}

		tracks, err := store.CountTracks(r.Context(), tf)
		if err != nil {
			l.Err(err).Msg("StatsHandler: Failed to fetch track count")
			utils.WriteError(w, "failed to get tracks: "+err.Error(), http.StatusInternalServerError)
			return
		}

		albums, err := store.CountAlbums(r.Context(), tf)
		if err != nil {
			l.Err(err).Msg("StatsHandler: Failed to fetch album count")
			utils.WriteError(w, "failed to get albums: "+err.Error(), http.StatusInternalServerError)
			return
		}

		artists, err := store.CountArtists(r.Context(), tf)
		if err != nil {
			l.Err(err).Msg("StatsHandler: Failed to fetch artist count")
			utils.WriteError(w, "failed to get artists: "+err.Error(), http.StatusInternalServerError)
			return
		}

		timeListenedS, err := store.CountTimeListened(r.Context(), tf)
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
