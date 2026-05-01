package handlers

import (
	"net/http"
	"strconv"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
)

func UpdateTrackHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("UpdateTrackHandler: Received request")

		if err := r.ParseForm(); err != nil {
			l.Debug().AnErr("error", err).Msg("UpdateTrackHandler: Failed to parse form")
			utils.WriteError(w, "form is invalid", http.StatusBadRequest)
			return
		}

		idStr := r.Form.Get("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			l.Debug().AnErr("error", err).Msg("UpdateTrackHandler: Invalid id parameter")
			utils.WriteError(w, "id is invalid", http.StatusBadRequest)
			return
		}

		var updateOpts = db.UpdateTrackOpts{
			ID: int32(id),
		}

		if formVal, ok := r.Form["add_artist"]; ok {
			var artists []int32
			for _, val := range formVal {
				if id, err := strconv.Atoi(val); err != nil {
					l.Debug().AnErr("error", err).Msg("UpdateTrackHandler: ID of artist to add is invalid")
					utils.WriteError(w, "ID of artist to add is invalid", http.StatusBadRequest)
					return
				} else {
					artists = append(artists, int32(id))
				}
			}
			updateOpts.AddArtists = artists
		}

		if formVal, ok := r.Form["remove_artist"]; ok {
			var artists []int32
			for _, val := range formVal {
				if id, err := strconv.Atoi(val); err != nil {
					l.Debug().Msg("UpdateTrackHandler: ID of artist to remove is invalid")
					utils.WriteError(w, "ID of artist to remove is invalid", http.StatusBadRequest)
					return
				} else {
					artists = append(artists, int32(id))
				}
			}
			updateOpts.RemoveArtists = artists
		}

		if err = store.UpdateTrack(ctx, updateOpts); err != nil {
			l.Debug().AnErr("error", err).Msg("UpdateTrackHandler: Failed to update track")
			utils.WriteError(w, "failed to update track", http.StatusBadRequest)
			return
		}

		l.Debug().Msg("UpdateTrackHandler: Successfully updated track")

		w.WriteHeader(http.StatusOK)
	}
}

func UpdateAlbumHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("UpdateAlbumHandler: Received request")

		if err := r.ParseForm(); err != nil {
			l.Debug().AnErr("error", err).Msg("UpdateAlbumHandler: Failed to parse form")
			utils.WriteError(w, "form is invalid", http.StatusBadRequest)
			return
		}

		idStr := r.Form.Get("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			l.Debug().AnErr("error", err).Msg("UpdateAlbumHandler: Invalid id parameter")
			utils.WriteError(w, "id is invalid", http.StatusBadRequest)
			return
		}

		var updateOpts = db.UpdateAlbumOpts{
			ID: int32(id),
		}

		if r.Form.Has("is_various_artists") {
			valStr := r.Form.Get("is_various_artists")
			VariousArtistsValue, err := strconv.ParseBool(valStr)
			if err != nil {
				l.Debug().AnErr("error", err).Msg("UpdateAlbumHandler: Various artists setting is invalid")
				utils.WriteError(w, "Various artists setting is invalid", http.StatusBadRequest)
				return
			}
			updateOpts.VariousArtistsUpdate = true
			updateOpts.VariousArtistsValue = VariousArtistsValue
		}

		if err = store.UpdateAlbum(ctx, updateOpts); err != nil {
			l.Debug().AnErr("error", err).Msg("UpdateAlbumHandler: Failed to update album")
			utils.WriteError(w, "failed to update album", http.StatusBadRequest)
			return
		}

		l.Debug().Msg("UpdateAlbumHandler: Successfully updated album")

		w.WriteHeader(http.StatusOK)
	}
}
