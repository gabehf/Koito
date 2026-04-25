package handlers

import (
	"net/http"
	"strconv"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
)

func GetArtistHandler(store db.ArtistStore) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("GetArtistHandler: Received request to retrieve artist")

		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			l.Debug().Msg("GetArtistHandler: Missing artist ID in request")
			utils.WriteError(w, "id must be provided", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			l.Debug().AnErr("error", err).Msg("GetArtistHandler: Invalid artist ID")
			utils.WriteError(w, "id is invalid", http.StatusBadRequest)
			return
		}

		l.Debug().Msgf("GetArtistHandler: Retrieving artist with ID %d", id)

		artist, err := store.GetArtist(ctx, db.GetArtistOpts{ID: int32(id)})
		if err != nil {
			l.Err(err).Msgf("GetArtistHandler: Failed to retrieve artist with ID %d", id)
			utils.WriteError(w, "artist with specified id could not be found", http.StatusNotFound)
			return
		}

		l.Debug().Msgf("GetArtistHandler: Successfully retrieved artist with ID %d", id)
		utils.WriteJSON(w, http.StatusOK, artist)
	}
}
