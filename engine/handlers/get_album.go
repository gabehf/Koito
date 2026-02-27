package handlers

import (
	"net/http"
	"strconv"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
)

func GetAlbumHandler(store AlbumStore) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("GetAlbumHandler: Received request to retrieve album")

		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			l.Debug().Msg("GetAlbumHandler: Missing album ID in request")
			utils.WriteError(w, "id must be provided", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			l.Debug().AnErr("error", err).Msg("GetAlbumHandler: Invalid album ID")
			utils.WriteError(w, "id is invalid", http.StatusBadRequest)
			return
		}

		l.Debug().Msgf("GetAlbumHandler: Retrieving album with ID %d", id)

		album, err := store.GetAlbum(ctx, db.GetAlbumOpts{ID: int32(id)})
		if err != nil {
			l.Err(err).Msgf("GetAlbumHandler: Failed to retrieve album with ID %d", id)
			utils.WriteError(w, "album with specified id could not be found", http.StatusNotFound)
			return
		}

		l.Debug().Msgf("GetAlbumHandler: Successfully retrieved album with ID %d", id)
		utils.WriteJSON(w, http.StatusOK, album)
	}
}
