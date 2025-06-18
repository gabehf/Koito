package handlers

import (
	"net/http"

	"github.com/gabehf/koito/engine/middleware"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/export"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
)

func ExportHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", `attachment; filename="koito_export.json"`)
		ctx := r.Context()
		l := logger.FromContext(ctx)
		l.Debug().Msg("ExportHandler: Recieved request for export file")
		u := middleware.GetUserFromContext(ctx)
		if u == nil {
			l.Debug().Msg("ExportHandler: Unauthorized access")
			utils.WriteError(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		err := export.ExportData(ctx, u, store, w)
		if err != nil {
			l.Err(err).Msg("ExportHandler: Failed to create export file")
			utils.WriteError(w, "failed to create export file", http.StatusInternalServerError)
			return
		}
	}
}
