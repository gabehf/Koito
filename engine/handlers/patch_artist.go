package handlers

import (
	"net/http"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
	"github.com/google/uuid"
)

func SetPrimaryArtistAliasHandler(store db.ArtistStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		artistID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("SetPrimaryArtistAliasHandler: Invalid artist id")
			utils.WriteError(w, "invalid artist id", http.StatusBadRequest)
			return
		}

		body, err := utils.DecodeBody[struct {
			Alias string `json:"alias"`
		}](r)
		if err != nil || body.Alias == "" {
			l.Debug().Msg("SetPrimaryArtistAliasHandler: Invalid or missing alias in request body")
			utils.WriteError(w, "alias must be provided", http.StatusBadRequest)
			return
		}

		if err = store.SetPrimaryArtistAlias(ctx, artistID, body.Alias); err != nil {
			l.Error().Err(err).Msg("SetPrimaryArtistAliasHandler: Failed to set artist primary alias")
			utils.WriteError(w, "failed to set primary alias", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// UpdateArtistHandler updates a given artist.
func UpdateArtistHandler(store db.ArtistStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		artistID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("UpdateArtistHandler: Invalid artist id")
			utils.WriteError(w, "invalid artist id", http.StatusBadRequest)
			return
		}

		body, err := utils.DecodeBody[struct {
			MBID *string `json:"mbid"`
		}](r)
		if err != nil {
			l.Debug().Msg("UpdateArtistHandler: Invalid request body")
			utils.WriteError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if body.MBID == nil {
			l.Debug().Msg("UpdateArtistHandler: Request body contains no updatable fields")
			utils.WriteError(w, "no updatable fields provided", http.StatusBadRequest)
			return
		}

		updateOpts := db.UpdateArtistOpts{ID: artistID}

		if body.MBID != nil {
			mbid, err := uuid.Parse(*body.MBID)
			if err != nil {
				l.Debug().Msg("UpdateArtistHandler: Provided MusicBrainz ID is invalid")
				utils.WriteError(w, "provided musicbrainz id is invalid", http.StatusBadRequest)
				return
			}
			updateOpts.MusicBrainzID = mbid
		}

		if err = store.UpdateArtist(ctx, updateOpts); err != nil {
			l.Error().Err(err).Msg("UpdateArtistHandler: Failed to update artist")
			utils.WriteError(w, "failed to update artist", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
