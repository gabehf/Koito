package handlers

import (
	"net/http"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/summary"
	"github.com/gabehf/koito/internal/utils"
)

func SummaryHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)
		l.Debug().Msg("SummaryHandler: Received request to retrieve summary")
		timeframe := TimeframeFromRequest(r)

		summary, err := summary.GenerateSummary(ctx, store, 1, timeframe, "")
		if err != nil {
			l.Err(err).Int("userid", 1).Any("timeframe", timeframe).Msgf("SummaryHandler: Failed to generate summary")
			utils.WriteError(w, "failed to generate summary", http.StatusInternalServerError)
			return
		}

		utils.WriteJSON(w, http.StatusOK, summary)
	}
}
