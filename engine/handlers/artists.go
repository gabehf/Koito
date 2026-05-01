package handlers

import (
	"net/http"
	"strconv"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/models"
	"github.com/gabehf/koito/internal/utils"
)

func SetPrimaryArtistHandler(store db.ArtistStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// sets the primary alias for albums, artists, and tracks
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("SetPrimaryArtistHandler: Got request")

		r.ParseForm()

		// Parse query parameters
		artistIDStr := r.FormValue("artist_id")
		albumIDStr := r.FormValue("album_id")
		trackIDStr := r.FormValue("track_id")
		isPrimaryStr := r.FormValue("is_primary")

		l.Debug().Str("query", r.Form.Encode()).Msg("Recieved form")

		if artistIDStr == "" {
			l.Debug().Msg("SetPrimaryArtistHandler: artist_id must be provided")
			utils.WriteError(w, "artist_id must be provided", http.StatusBadRequest)
			return
		}

		if isPrimaryStr == "" {
			l.Debug().Msg("SetPrimaryArtistHandler: is_primary must be provided")
			utils.WriteError(w, "is_primary must be provided", http.StatusBadRequest)
			return
		}

		primary, ok := utils.ParseBool(isPrimaryStr)
		if !ok {
			l.Debug().Msg("SetPrimaryArtistHandler: is_primary must be either true or false")
			utils.WriteError(w, "is_primary must be either true or false", http.StatusBadRequest)
			return
		}

		artistId, err := strconv.Atoi(artistIDStr)
		if err != nil {
			l.Debug().Msg("SetPrimaryArtistHandler: artist_id is invalid")
			utils.WriteError(w, "artist_id is invalid", http.StatusBadRequest)
			return
		}

		if albumIDStr == "" && trackIDStr == "" {
			l.Debug().Msg("SetPrimaryArtistHandler: Missing album or track id parameter")
			utils.WriteError(w, "album_id or track_id must be provided", http.StatusBadRequest)
			return
		}
		if utils.MoreThanOneString(albumIDStr, trackIDStr) {
			l.Debug().Msg("SetPrimaryArtistHandler: Multiple ID parameters provided")
			utils.WriteError(w, "only one of album_id or track_id can be provided", http.StatusBadRequest)
			return
		}

		if albumIDStr != "" {
			id, err := strconv.Atoi(albumIDStr)
			if err != nil {
				l.Debug().AnErr("error", err).Msg("SetPrimaryArtistHandler: Invalid album id")
				utils.WriteError(w, "invalid album_id", http.StatusBadRequest)
				return
			}
			err = store.SetPrimaryAlbumArtist(ctx, int32(id), int32(artistId), primary)
			if err != nil {
				l.Error().Err(err).Msg("SetPrimaryArtistHandler: Failed to set album primary alias")
				utils.WriteError(w, "failed to set primary alias", http.StatusInternalServerError)
				return
			}
		} else if trackIDStr != "" {
			id, err := strconv.Atoi(trackIDStr)
			if err != nil {
				l.Debug().AnErr("error", err).Msg("SetPrimaryArtistHandler: Invalid track id")
				utils.WriteError(w, "invalid track_id", http.StatusBadRequest)
				return
			}
			err = store.SetPrimaryTrackArtist(ctx, int32(id), int32(artistId), primary)
			if err != nil {
				l.Error().Err(err).Msg("SetPrimaryArtistHandler: Failed to set track primary alias")
				utils.WriteError(w, "failed to set primary alias", http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
func GetArtistsForItemHandler(store db.ArtistStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("GetArtistsForItemHandler: Received request to retrieve artists for item")

		albumIDStr := r.URL.Query().Get("album_id")
		trackIDStr := r.URL.Query().Get("track_id")

		if albumIDStr == "" && trackIDStr == "" {
			l.Debug().Msg("GetArtistsForItemHandler: Missing album or track ID parameter")
			utils.WriteError(w, "album_id or track_id must be provided", http.StatusBadRequest)
			return
		}

		if utils.MoreThanOneString(albumIDStr, trackIDStr) {
			l.Debug().Msg("GetArtistsForItemHandler: Multiple ID parameters provided")
			utils.WriteError(w, "only one of album_id or track_id can be provided", http.StatusBadRequest)
			return
		}

		var artists []*models.Artist
		var err error

		if albumIDStr != "" {
			albumID, convErr := strconv.Atoi(albumIDStr)
			if convErr != nil {
				l.Debug().AnErr("error", convErr).Msg("GetArtistsForItemHandler: Invalid album ID")
				utils.WriteError(w, "invalid album_id", http.StatusBadRequest)
				return
			}

			l.Debug().Msgf("GetArtistsForItemHandler: Fetching artists for album ID %d", albumID)
			artists, err = store.GetArtistsForAlbum(ctx, int32(albumID))
		} else if trackIDStr != "" {
			trackID, convErr := strconv.Atoi(trackIDStr)
			if convErr != nil {
				l.Debug().AnErr("error", convErr).Msg("GetArtistsForItemHandler: Invalid track ID")
				utils.WriteError(w, "invalid track_id", http.StatusBadRequest)
				return
			}

			l.Debug().Msgf("GetArtistsForItemHandler: Fetching artists for track ID %d", trackID)
			artists, err = store.GetArtistsForTrack(ctx, int32(trackID))
		}

		if err != nil {
			l.Err(err).Msg("GetArtistsForItemHandler: Failed to retrieve artists")
			utils.WriteError(w, "failed to retrieve artists", http.StatusInternalServerError)
			return
		}

		l.Debug().Msg("GetArtistsForItemHandler: Successfully retrieved artists")
		utils.WriteJSON(w, http.StatusOK, artists)
	}
}
