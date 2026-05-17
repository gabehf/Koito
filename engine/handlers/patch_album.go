package handlers

import (
	"net/http"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
	"github.com/google/uuid"
)

// SetPrimaryAlbumAliasHandler sets the primary alias for a given album.
func SetPrimaryAlbumAliasHandler(store db.AlbumStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		albumID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("SetPrimaryAlbumAliasHandler: Invalid album id")
			utils.WriteError(w, "invalid album id", http.StatusBadRequest)
			return
		}

		body, err := utils.DecodeBody[struct {
			Alias string `json:"alias"`
		}](r)
		if err != nil || body.Alias == "" {
			l.Debug().Msg("SetPrimaryAlbumAliasHandler: Invalid or missing alias in request body")
			utils.WriteError(w, "alias must be provided", http.StatusBadRequest)
			return
		}

		if err = store.SetPrimaryAlbumAlias(ctx, albumID, body.Alias); err != nil {
			l.Error().Err(err).Msg("SetPrimaryAlbumAliasHandler: Failed to set album primary alias")
			utils.WriteError(w, "failed to set primary alias", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// SetPrimaryAlbumArtistHandler sets whether a given artist is the primary artist for an album.
func SetPrimaryAlbumArtistHandler(store db.ArtistStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		albumID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("SetPrimaryAlbumArtistHandler: Invalid album id")
			utils.WriteError(w, "invalid album id", http.StatusBadRequest)
			return
		}

		artistID, err := utils.ParseIDParam(r, "artist_id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("SetPrimaryAlbumArtistHandler: Invalid artist id")
			utils.WriteError(w, "invalid artist id", http.StatusBadRequest)
			return
		}

		body, err := utils.DecodeBody[struct {
			IsPrimary *bool `json:"is_primary"`
		}](r)
		if err != nil || body.IsPrimary == nil {
			l.Debug().Msg("SetPrimaryAlbumArtistHandler: Invalid or missing is_primary in request body")
			utils.WriteError(w, "is_primary must be provided", http.StatusBadRequest)
			return
		}

		if err = store.SetPrimaryAlbumArtist(ctx, albumID, artistID, *body.IsPrimary); err != nil {
			l.Error().Err(err).Msg("SetPrimaryAlbumArtistHandler: Failed to set primary album artist")
			utils.WriteError(w, "failed to set primary album artist", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// UpdateAlbumHandler updates a given album.
func UpdateAlbumHandler(store db.AlbumStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		albumID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("UpdateAlbumHandler: Invalid album id")
			utils.WriteError(w, "invalid album id", http.StatusBadRequest)
			return
		}

		body, err := utils.DecodeBody[struct {
			MBID             *string `json:"mbid"`
			IsVariousArtists *bool   `json:"is_various_artists"`
		}](r)
		if err != nil {
			l.Debug().Msg("UpdateAlbumHandler: Invalid request body")
			utils.WriteError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if body.MBID == nil && body.IsVariousArtists == nil {
			l.Debug().Msg("UpdateAlbumHandler: Request body contains no updatable fields")
			utils.WriteError(w, "no updatable fields provided", http.StatusBadRequest)
			return
		}

		updateOpts := db.UpdateAlbumOpts{ID: albumID}

		if body.MBID != nil {
			mbid, err := uuid.Parse(*body.MBID)
			if err != nil {
				l.Debug().Msg("UpdateAlbumHandler: Provided MusicBrainz ID is invalid")
				utils.WriteError(w, "provided musicbrainz id is invalid", http.StatusBadRequest)
				return
			}
			updateOpts.MusicBrainzID = mbid
		}

		if body.IsVariousArtists != nil {
			updateOpts.VariousArtistsUpdate = true
			updateOpts.VariousArtistsValue = *body.IsVariousArtists
		}

		if err = store.UpdateAlbum(ctx, updateOpts); err != nil {
			l.Error().Err(err).Msg("UpdateAlbumHandler: Failed to update album")
			utils.WriteError(w, "failed to update album", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
