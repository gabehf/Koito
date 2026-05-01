package handlers

import (
	"net/http"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
)

func GetListensHandler(store db.ListenStore) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("GetListensHandler: Received request to retrieve listens")

		opts := OptsFromRequest(r)
		l.Debug().Msgf("GetListensHandler: Retrieving listens with options: %+v", opts)

		listens, err := store.GetListensPaginated(ctx, opts)
		if err != nil {
			l.Err(err).Msg("GetListensHandler: Failed to retrieve listens")
			utils.WriteError(w, "failed to get listens: "+err.Error(), http.StatusBadRequest)
			return
		}

		l.Debug().Msg("GetListensHandler: Successfully retrieved listens")
		utils.WriteJSON(w, http.StatusOK, listens)
	}
}
