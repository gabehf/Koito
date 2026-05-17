package handlers

import (
	"net/http"
	"time"

	"github.com/gabehf/koito/engine/middleware"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
)

func SubmitListenWithIDHandler(store db.ListenStore) http.HandlerFunc {
	var defaultClientStr = "Koito Web UI"
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)
		l.Debug().Msg("SubmitListenWithIDHandler: Got request")

		u := middleware.GetUserFromContext(ctx)
		if u == nil {
			l.Debug().Msg("SubmitListenWithIDHandler: Unauthorized request (user context is nil)")
			utils.WriteError(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		body, err := utils.DecodeBody[struct {
			TrackID int32  `json:"track_id"`
			Unix    int64  `json:"unix"`
			Client  string `json:"client"`
		}](r)
		if err != nil || body.TrackID == 0 || body.Unix == 0 {
			l.Debug().Msg("SubmitListenWithIDHandler: Invalid or missing required fields in request body")
			utils.WriteError(w, "track_id and unix (timestamp) must be provided", http.StatusBadRequest)
			return
		}

		if time.Now().Unix() < body.Unix {
			l.Debug().Msg("SubmitListenWithIDHandler: Timestamp is in the future")
			utils.WriteError(w, "invalid timestamp", http.StatusBadRequest)
			return
		}

		client := body.Client
		if client == "" {
			client = defaultClientStr
		}

		if err = store.SaveListen(ctx, db.SaveListenOpts{
			TrackID: body.TrackID,
			Time:    time.Unix(body.Unix, 0),
			UserID:  u.ID,
			Client:  client,
		}); err != nil {
			l.Err(err).Msg("SubmitListenWithIDHandler: Failed to submit listen")
			utils.WriteError(w, "failed to submit listen", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}
