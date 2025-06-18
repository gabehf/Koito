package handlers

import (
	"net/http"
	"strconv"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/models"
	"github.com/gabehf/koito/internal/utils"
)

// GetAliasesHandler retrieves all aliases for a given artist or album ID.
func GetAliasesHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msgf("GetAliasesHandler: Got request with params: '%s'", r.URL.Query().Encode())

		// Parse query parameters
		artistIDStr := r.URL.Query().Get("artist_id")
		albumIDStr := r.URL.Query().Get("album_id")
		trackIDStr := r.URL.Query().Get("track_id")

		if artistIDStr == "" && albumIDStr == "" && trackIDStr == "" {
			l.Debug().Msgf("GetAliasesHandler: Request is missing required parameters")
			utils.WriteError(w, "artist_id, album_id, or track_id must be provided", http.StatusBadRequest)
			return
		}
		if utils.MoreThanOneString(artistIDStr, albumIDStr, trackIDStr) {
			l.Debug().Msgf("GetAliasesHandler: Request is has more than one of artist_id, album_id, and track_id")
			utils.WriteError(w, "only one of artist_id, album_id, or track_id can be provided at a time", http.StatusBadRequest)
			return
		}

		var aliases []models.Alias

		if artistIDStr != "" {
			artistID, err := strconv.Atoi(artistIDStr)
			if err != nil {
				l.Debug().AnErr("error", err).Msg("GetAliasesHandler: Invalid artist id")
				utils.WriteError(w, "invalid artist_id", http.StatusBadRequest)
				return
			}
			aliases, err = store.GetAllArtistAliases(ctx, int32(artistID))
			if err != nil {
				l.Err(err).Msg("GetAliasesHandler: Failed to get artist aliases")
				utils.WriteError(w, "failed to retrieve aliases", http.StatusInternalServerError)
				return
			}
		} else if albumIDStr != "" {
			albumID, err := strconv.Atoi(albumIDStr)
			if err != nil {
				l.Debug().AnErr("error", err).Msg("GetAliasesHandler: Invalid album id")
				utils.WriteError(w, "invalid album_id", http.StatusBadRequest)
				return
			}
			aliases, err = store.GetAllAlbumAliases(ctx, int32(albumID))
			if err != nil {
				l.Err(err).Msg("GetAliasesHandler: Failed to get album aliases")
				utils.WriteError(w, "failed to retrieve aliases", http.StatusInternalServerError)
				return
			}
		} else if trackIDStr != "" {
			trackID, err := strconv.Atoi(trackIDStr)
			if err != nil {
				l.Debug().AnErr("error", err).Msg("GetAliasesHandler: Invalid track id")
				utils.WriteError(w, "invalid track_id", http.StatusBadRequest)
				return
			}
			aliases, err = store.GetAllTrackAliases(ctx, int32(trackID))
			if err != nil {
				l.Err(err).Msg("GetAliasesHandler: Failed to get track aliases")
				utils.WriteError(w, "failed to retrieve aliases", http.StatusInternalServerError)
				return
			}
		}
		utils.WriteJSON(w, http.StatusOK, aliases)
	}
}

// DeleteAliasHandler deletes an alias for a given artist or album ID.
func DeleteAliasHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("DeleteAliasHandler: Got request")

		err := r.ParseForm()
		if err != nil {
			l.Debug().Msg("DeleteAliasHandler: Failed to parse form")
			utils.WriteError(w, "form is invalid", http.StatusBadRequest)
			return
		}

		// Parse query parameters
		artistIDStr := r.FormValue("artist_id")
		albumIDStr := r.FormValue("album_id")
		trackIDStr := r.FormValue("track_id")
		alias := r.FormValue("alias")

		if alias == "" || (artistIDStr == "" && albumIDStr == "" && trackIDStr == "") {
			l.Debug().Msg("DeleteAliasHandler: Request is missing required parameters")
			utils.WriteError(w, "alias and artist_id, album_id, or track_id must be provided", http.StatusBadRequest)
			return
		}
		if utils.MoreThanOneString(artistIDStr, albumIDStr, trackIDStr) {
			l.Debug().Msg("DeleteAliasHandler: Request has more than one of artist_id, album_id, and track_id")
			utils.WriteError(w, "only one of artist_id, album_id, or track_id can be provided at a time", http.StatusBadRequest)
			return
		}

		if artistIDStr != "" {
			var artistID int
			artistID, err = strconv.Atoi(artistIDStr)
			if err != nil {
				l.Debug().AnErr("error", err).Msg("DeleteAliasHandler: Invalid artist id")
				utils.WriteError(w, "invalid artist_id", http.StatusBadRequest)
				return
			}
			err = store.DeleteArtistAlias(ctx, int32(artistID), alias)
			if err != nil {
				l.Error().Err(err).Msg("DeleteAliasHandler: Failed to delete artist alias")
				utils.WriteError(w, "failed to delete alias", http.StatusInternalServerError)
				return
			}
		} else if albumIDStr != "" {
			var albumID int
			albumID, err = strconv.Atoi(albumIDStr)
			if err != nil {
				l.Debug().AnErr("error", err).Msg("DeleteAliasHandler: Invalid album id")
				utils.WriteError(w, "invalid album_id", http.StatusBadRequest)
				return
			}
			err = store.DeleteAlbumAlias(ctx, int32(albumID), alias)
			if err != nil {
				l.Error().Err(err).Msg("DeleteAliasHandler: Failed to delete album alias")
				utils.WriteError(w, "failed to delete alias", http.StatusInternalServerError)
				return
			}
		} else if trackIDStr != "" {
			var trackID int
			trackID, err = strconv.Atoi(trackIDStr)
			if err != nil {
				l.Debug().AnErr("error", err).Msg("DeleteAliasHandler: Invalid track id")
				utils.WriteError(w, "invalid track_id", http.StatusBadRequest)
				return
			}
			err = store.DeleteTrackAlias(ctx, int32(trackID), alias)
			if err != nil {
				l.Error().Err(err).Msg("DeleteAliasHandler: Failed to delete track alias")
				utils.WriteError(w, "failed to delete alias", http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// CreateAliasHandler creates new aliases for a given artist, album, or track.
func CreateAliasHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("CreateAliasHandler: Got request")

		err := r.ParseForm()
		if err != nil {
			l.Debug().AnErr("error", err).Msg("CreateAliasHandler: Failed to parse form")
			utils.WriteError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		alias := r.FormValue("alias")
		if alias == "" {
			l.Debug().Msg("CreateAliasHandler: Alias parameter missing")
			utils.WriteError(w, "alias must be provided", http.StatusBadRequest)
			return
		}

		artistIDStr := r.FormValue("artist_id")
		albumIDStr := r.FormValue("album_id")
		trackIDStr := r.FormValue("track_id")

		if artistIDStr == "" && albumIDStr == "" && trackIDStr == "" {
			l.Debug().Msg("CreateAliasHandler: Missing ID parameter")
			utils.WriteError(w, "artist_id, album_id, or track_id must be provided", http.StatusBadRequest)
			return
		}
		if utils.MoreThanOneString(artistIDStr, albumIDStr, trackIDStr) {
			l.Debug().Msg("CreateAliasHandler: Multiple ID parameters provided")
			utils.WriteError(w, "only one of artist_id, album_id, or track_id can be provided", http.StatusBadRequest)
			return
		}

		var id int
		if artistIDStr != "" {
			id, err = strconv.Atoi(artistIDStr)
			if err != nil {
				l.Debug().AnErr("error", err).Msg("CreateAliasHandler: Invalid artist id")
				utils.WriteError(w, "invalid artist_id", http.StatusBadRequest)
				return
			}
			err = store.SaveArtistAliases(ctx, int32(id), []string{alias}, "Manual")
			if err != nil {
				l.Error().Err(err).Msg("CreateAliasHandler: Failed to save artist alias")
				utils.WriteError(w, "failed to save alias", http.StatusInternalServerError)
				return
			}
		} else if albumIDStr != "" {
			id, err = strconv.Atoi(albumIDStr)
			if err != nil {
				l.Debug().AnErr("error", err).Msg("CreateAliasHandler: Invalid album id")
				utils.WriteError(w, "invalid album_id", http.StatusBadRequest)
				return
			}
			err = store.SaveAlbumAliases(ctx, int32(id), []string{alias}, "Manual")
			if err != nil {
				l.Error().Err(err).Msg("CreateAliasHandler: Failed to save album alias")
				utils.WriteError(w, "failed to save alias", http.StatusInternalServerError)
				return
			}
		} else if trackIDStr != "" {
			id, err = strconv.Atoi(trackIDStr)
			if err != nil {
				l.Debug().AnErr("error", err).Msg("CreateAliasHandler: Invalid track id")
				utils.WriteError(w, "invalid track_id", http.StatusBadRequest)
				return
			}
			err = store.SaveTrackAliases(ctx, int32(id), []string{alias}, "Manual")
			if err != nil {
				l.Error().Err(err).Msg("CreateAliasHandler: Failed to save track alias")
				utils.WriteError(w, "failed to save alias", http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusCreated)
	}
}

// sets the primary alias for albums, artists, and tracks
func SetPrimaryAliasHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("SetPrimaryAliasHandler: Got request")

		err := r.ParseForm()
		if err != nil {
			l.Debug().Msg("SetPrimaryAliasHandler: Failed to parse form")
			utils.WriteError(w, "form is invalid", http.StatusBadRequest)
			return
		}

		// Parse query parameters
		artistIDStr := r.FormValue("artist_id")
		albumIDStr := r.FormValue("album_id")
		trackIDStr := r.FormValue("track_id")
		alias := r.FormValue("alias")

		l.Debug().Msgf("Alias: %s", alias)

		if alias == "" {
			l.Debug().Msg("SetPrimaryAliasHandler: Missing alias parameter")
			utils.WriteError(w, "alias must be provided", http.StatusBadRequest)
			return
		}
		if artistIDStr == "" && albumIDStr == "" && trackIDStr == "" {
			l.Debug().Msg("SetPrimaryAliasHandler: Missing ID parameter")
			utils.WriteError(w, "artist_id, album_id, or track_id must be provided", http.StatusBadRequest)
			return
		}
		if utils.MoreThanOneString(artistIDStr, albumIDStr, trackIDStr) {
			l.Debug().Msg("SetPrimaryAliasHandler: Multiple ID parameters provided")
			utils.WriteError(w, "only one of artist_id, album_id, or track_id can be provided", http.StatusBadRequest)
			return
		}

		var id int
		if artistIDStr != "" {
			id, err = strconv.Atoi(artistIDStr)
			if err != nil {
				l.Debug().AnErr("error", err).Msg("SetPrimaryAliasHandler: Invalid artist id")
				utils.WriteError(w, "invalid artist_id", http.StatusBadRequest)
				return
			}
			err = store.SetPrimaryArtistAlias(ctx, int32(id), alias)
			if err != nil {
				l.Error().Err(err).Msg("SetPrimaryAliasHandler: Failed to set artist primary alias")
				utils.WriteError(w, "failed to set primary alias", http.StatusInternalServerError)
				return
			}
		} else if albumIDStr != "" {
			id, err = strconv.Atoi(albumIDStr)
			if err != nil {
				l.Debug().AnErr("error", err).Msg("SetPrimaryAliasHandler: Invalid album id")
				utils.WriteError(w, "invalid album_id", http.StatusBadRequest)
				return
			}
			err = store.SetPrimaryAlbumAlias(ctx, int32(id), alias)
			if err != nil {
				l.Error().Err(err).Msg("SetPrimaryAliasHandler: Failed to set album primary alias")
				utils.WriteError(w, "failed to set primary alias", http.StatusInternalServerError)
				return
			}
		} else if trackIDStr != "" {
			id, err = strconv.Atoi(trackIDStr)
			if err != nil {
				l.Debug().AnErr("error", err).Msg("SetPrimaryAliasHandler: Invalid track id")
				utils.WriteError(w, "invalid track_id", http.StatusBadRequest)
				return
			}
			err = store.SetPrimaryTrackAlias(ctx, int32(id), alias)
			if err != nil {
				l.Error().Err(err).Msg("SetPrimaryAliasHandler: Failed to set track primary alias")
				utils.WriteError(w, "failed to set primary alias", http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
