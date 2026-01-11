package handlers

import (
	"net/http"
	"strconv"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
	"github.com/google/uuid"
)

func UpdateMbzIdHandler(store db.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("UpdateMbzIdHandler: Received request to set update MusicBrainz ID")

		err := r.ParseForm()
		if err != nil {
			l.Debug().Msg("UpdateMbzIdHandler: Failed to parse form")
			utils.WriteError(w, "form is invalid", http.StatusBadRequest)
			return
		}

		// Parse query parameters
		artistIDStr := r.FormValue("artist_id")
		albumIDStr := r.FormValue("album_id")
		trackIDStr := r.FormValue("track_id")
		mbzidStr := r.FormValue("mbz_id")

		if mbzidStr == "" || (artistIDStr == "" && albumIDStr == "" && trackIDStr == "") {
			l.Debug().Msg("UpdateMbzIdHandler: Request is missing required parameters")
			utils.WriteError(w, "mbzid and artist_id, album_id, or track_id must be provided", http.StatusBadRequest)
			return
		}
		if utils.MoreThanOneString(artistIDStr, albumIDStr, trackIDStr) {
			l.Debug().Msg("UpdateMbzIdHandler: Request has more than one of artist_id, album_id, and track_id")
			utils.WriteError(w, "only one of artist_id, album_id, or track_id can be provided at a time", http.StatusBadRequest)
			return
		}
		var mbzid uuid.UUID
		if mbzid, err = uuid.Parse(mbzidStr); err != nil {
			l.Debug().Msg("UpdateMbzIdHandler: Provided MusicBrainz ID is invalid")
			utils.WriteError(w, "provided musicbrainz id is invalid", http.StatusBadRequest)
			return
		}

		if artistIDStr != "" {
			var artistID int
			artistID, err = strconv.Atoi(artistIDStr)
			if err != nil {
				l.Debug().AnErr("error", err).Msg("UpdateMbzIdHandler: Invalid artist id")
				utils.WriteError(w, "invalid artist_id", http.StatusBadRequest)
				return
			}
			err = store.UpdateArtist(ctx, db.UpdateArtistOpts{
				ID:            int32(artistID),
				MusicBrainzID: mbzid,
			})
			if err != nil {
				l.Error().Err(err).Msg("UpdateMbzIdHandler: Failed to update musicbrainz id")
				utils.WriteError(w, "failed to update musicbrainz id", http.StatusInternalServerError)
				return
			}
		} else if albumIDStr != "" {
			var albumID int
			albumID, err = strconv.Atoi(albumIDStr)
			if err != nil {
				l.Debug().AnErr("error", err).Msg("UpdateMbzIdHandler: Invalid album id")
				utils.WriteError(w, "invalid artist_id", http.StatusBadRequest)
				return
			}
			err = store.UpdateAlbum(ctx, db.UpdateAlbumOpts{
				ID:            int32(albumID),
				MusicBrainzID: mbzid,
			})
			if err != nil {
				l.Error().Err(err).Msg("UpdateMbzIdHandler: Failed to update musicbrainz id")
				utils.WriteError(w, "failed to update musicbrainz id", http.StatusInternalServerError)
				return
			}
		} else if trackIDStr != "" {
			var trackID int
			trackID, err = strconv.Atoi(trackIDStr)
			if err != nil {
				l.Debug().AnErr("error", err).Msg("UpdateMbzIdHandler: Invalid track id")
				utils.WriteError(w, "invalid artist_id", http.StatusBadRequest)
				return
			}
			err = store.UpdateTrack(ctx, db.UpdateTrackOpts{
				ID:            int32(trackID),
				MusicBrainzID: mbzid,
			})
			if err != nil {
				l.Error().Err(err).Msg("UpdateMbzIdHandler: Failed to update musicbrainz id")
				utils.WriteError(w, "failed to update musicbrainz id", http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
