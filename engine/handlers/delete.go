package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
)

func DeleteTrackHandler(store db.TrackStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("DeleteTrackHandler: Received request to delete track")

		trackIDStr := r.URL.Query().Get("id")
		if trackIDStr == "" {
			l.Debug().Msg("DeleteTrackHandler: Missing track ID in request")
			utils.WriteError(w, "track_id must be provided", http.StatusBadRequest)
			return
		}

		trackID, err := strconv.Atoi(trackIDStr)
		if err != nil {
			l.Debug().AnErr("error", err).Msg("DeleteTrackHandler: Invalid track ID")
			utils.WriteError(w, "invalid id", http.StatusBadRequest)
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

func DeleteListenHandler(store db.ListenStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("DeleteListenHandler: Received request to delete listen record")

		trackIDStr := r.URL.Query().Get("track_id")
		if trackIDStr == "" {
			l.Debug().Msg("DeleteListenHandler: Missing track ID in request")
			utils.WriteError(w, "track_id must be provided", http.StatusBadRequest)
			return
		}

		trackID, err := strconv.Atoi(trackIDStr)
		if err != nil {
			l.Debug().AnErr("error", err).Msg("DeleteListenHandler: Invalid track ID")
			utils.WriteError(w, "invalid id", http.StatusBadRequest)
			return
		}

		unixStr := r.URL.Query().Get("unix")
		if unixStr == "" {
			l.Debug().Msg("DeleteListenHandler: Missing timestamp in request")
			utils.WriteError(w, "unix timestamp must be provided", http.StatusBadRequest)
			return
		}

		unix, err := strconv.ParseInt(unixStr, 10, 64)
		if err != nil {
			l.Debug().AnErr("error", err).Msg("DeleteListenHandler: Invalid timestamp")
			utils.WriteError(w, "invalid unix timestamp", http.StatusBadRequest)
			return
		}

		l.Debug().Msgf("DeleteListenHandler: Deleting listen record for track ID %d at timestamp %d", trackID, unix)

		err = store.DeleteListen(ctx, int32(trackID), time.Unix(unix, 0))
		if err != nil {
			l.Err(err).Msg("DeleteListenHandler: Failed to delete listen record")
			utils.WriteError(w, "failed to delete listen", http.StatusInternalServerError)
			return
		}

		l.Debug().Msgf("DeleteListenHandler: Successfully deleted listen record for track ID %d at timestamp %d", trackID, unix)
		w.WriteHeader(http.StatusNoContent)
	}
}

func DeleteArtistHandler(store db.ArtistStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("DeleteArtistHandler: Received request to delete artist")

		artistIDStr := r.URL.Query().Get("id")
		if artistIDStr == "" {
			l.Debug().Msg("DeleteArtistHandler: Missing artist ID in request")
			utils.WriteError(w, "id must be provided", http.StatusBadRequest)
			return
		}

		artistID, err := strconv.Atoi(artistIDStr)
		if err != nil {
			l.Debug().AnErr("error", err).Msg("DeleteArtistHandler: Invalid artist ID")
			utils.WriteError(w, "invalid id", http.StatusBadRequest)
			return
		}

		l.Debug().Msgf("DeleteArtistHandler: Deleting artist with ID %d", artistID)

		err = store.DeleteArtist(ctx, int32(artistID))
		if err != nil {
			l.Err(err).Msg("DeleteArtistHandler: Failed to delete artist")
			utils.WriteError(w, "failed to delete artist", http.StatusInternalServerError)
			return
		}

		l.Debug().Msgf("DeleteArtistHandler: Successfully deleted artist with ID %d", artistID)
		w.WriteHeader(http.StatusNoContent)
	}
}

func DeleteAlbumHandler(store db.AlbumStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("DeleteAlbumHandler: Received request to delete album")

		albumIDStr := r.URL.Query().Get("id")
		if albumIDStr == "" {
			l.Debug().Msg("DeleteAlbumHandler: Missing album ID in request")
			utils.WriteError(w, "id must be provided", http.StatusBadRequest)
			return
		}

		albumID, err := strconv.Atoi(albumIDStr)
		if err != nil {
			l.Debug().AnErr("error", err).Msg("DeleteAlbumHandler: Invalid album ID")
			utils.WriteError(w, "invalid id", http.StatusBadRequest)
			return
		}

		l.Debug().Msgf("DeleteAlbumHandler: Deleting album with ID %d", albumID)

		err = store.DeleteAlbum(ctx, int32(albumID))
		if err != nil {
			l.Err(err).Msg("DeleteAlbumHandler: Failed to delete album")
			utils.WriteError(w, "failed to delete album", http.StatusInternalServerError)
			return
		}

		l.Debug().Msgf("DeleteAlbumHandler: Successfully deleted album with ID %d", albumID)
		w.WriteHeader(http.StatusNoContent)
	}
}

func PurgeAllDataHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("PurgeAllDataHandler: Received request to purge all data")

		if err := store.PurgeAllData(ctx); err != nil {
			l.Err(err).Msg("PurgeAllDataHandler: Failed to purge all data")
			utils.WriteError(w, "failed to purge data", http.StatusInternalServerError)
			return
		}

		l.Debug().Msg("PurgeAllDataHandler: Successfully purged all data")
		w.WriteHeader(http.StatusNoContent)
	}
}
