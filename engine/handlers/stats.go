package handlers

import (
	"net/http"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
)

type StatsResponse struct {
	ListenCount     int64   `json:"listen_count"`
	TrackCount      int64   `json:"track_count"`
	AlbumCount      int64   `json:"album_count"`
	ArtistCount     int64   `json:"artist_count"`
	MinutesListened int64   `json:"minutes_listened"`
	DaysActive      int     `json:"days_active"`
	LongestStreak   int     `json:"longest_streak"`
	AvgDailyPlays   float32 `json:"avg_daily_plays"`
	TracksPerArtist float32 `json:"tracks_per_artist"`
	AlbumsPerArtist float32 `json:"albums_per_artist"`
}

type statsStore interface {
	db.ListenStore
	db.TrackStore
	db.AlbumStore
	db.ArtistStore
}

func StatsHandler(store statsStore) http.HandlerFunc {
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

		activeDays, err := store.GetActiveDays(r.Context(), tf.Timezone)
		if err != nil {
			l.Err(err).Msg("StatsHandler: Failed to fetch active days")
			utils.WriteError(w, "failed to get active days: "+err.Error(), http.StatusInternalServerError)
			return
		}

		longestStreak, err := store.GetLongestListenStreak(r.Context(), db.ListenActivityOpts{Timezone: tf.Timezone})
		if err != nil {
			l.Err(err).Msg("StatsHandler: Failed to fetch longest streak")
			utils.WriteError(w, "failed to get longest streak: "+err.Error(), http.StatusInternalServerError)
			return
		}

		l.Debug().Msg("StatsHandler: Successfully fetched statistics")

		resp := &StatsResponse{
			ListenCount:     listens,
			TrackCount:      tracks,
			AlbumCount:      albums,
			ArtistCount:     artists,
			MinutesListened: timeListenedS / 60,
			DaysActive:      activeDays,
			AvgDailyPlays:   safeRatio(listens, int64(activeDays)),
			TracksPerArtist: safeRatio(tracks, artists),
			AlbumsPerArtist: safeRatio(albums, artists),
			LongestStreak:   longestStreak,
		}

		utils.WriteJSON(w, http.StatusOK, resp)

	}
}

func safeRatio(a, b int64) float32 {
	if b == 0 {
		return 0
	}
	return float32(a) / float32(b)
}
