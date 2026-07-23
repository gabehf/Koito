package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/models"
	"github.com/gabehf/koito/internal/utils"
	"github.com/google/uuid"
)

const (
	lbzFeedbackDefaultCount = 25
	lbzFeedbackMaxCount     = 100
)

type lbzFeedbackItem struct {
	UserID        string  `json:"user_id"`
	RecordingMSID *string `json:"recording_msid"`
	RecordingMBID *string `json:"recording_mbid"`
	Score         int     `json:"score"`
	Created       int64   `json:"created,omitempty"`
}

// LbzGetUserFeedbackHandler returns paginated feedback.
//
// The {user_name} path segment is accepted for ListenBrainz API compatibility but
// ignored for now (Koito is single-user; feedback is always the admin's).
// TODO: validate/match user_name when multi-user support is added.
func LbzGetUserFeedbackHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		user, ok := lbzFeedbackUser(w, r, store)
		if !ok {
			return
		}

		offset := 0
		if raw := r.URL.Query().Get("offset"); raw != "" {
			v, err := strconv.Atoi(raw)
			if err != nil || v < 0 {
				utils.WriteError(w, "offset must be a non-negative integer", http.StatusBadRequest)
				return
			}
			offset = v
		}

		count := lbzFeedbackDefaultCount
		if raw := r.URL.Query().Get("count"); raw != "" {
			v, err := strconv.Atoi(raw)
			if err != nil || v < 1 {
				utils.WriteError(w, "count must be a positive integer", http.StatusBadRequest)
				return
			}
			count = v
		}
		count = min(count, lbzFeedbackMaxCount)

		var scoreFilter *int
		if raw := r.URL.Query().Get("score"); raw != "" {
			v, err := strconv.Atoi(raw)
			if err != nil || (v != 1 && v != -1) {
				utils.WriteError(w, "score must be 1 or -1", http.StatusBadRequest)
				return
			}
			scoreFilter = &v
		}

		items, total, err := store.ListUserFeedback(ctx, db.ListUserFeedbackOpts{
			UserID: user.ID,
			Score:  scoreFilter,
			Limit:  count,
			Offset: offset,
		})
		if err != nil {
			l.Err(err).Msg("LbzGetUserFeedbackHandler: Failed to list feedback")
			utils.WriteError(w, "failed to retrieve feedback", http.StatusInternalServerError)
			return
		}

		feedback := make([]lbzFeedbackItem, len(items))
		for i, item := range items {
			feedback[i] = toLbzFeedbackItem(user.Username, item)
		}
		utils.WriteJSON(w, http.StatusOK, map[string]any{
			"count":       len(feedback),
			"feedback":    feedback,
			"offset":      offset,
			"total_count": total,
		})
	}
}

// LbzGetFeedbackForRecordingsHandler returns feedback scores for the given recordings.
//
// The {user_name} path segment is accepted for ListenBrainz API compatibility but
// ignored for now (Koito is single-user; feedback is always the admin's).
// TODO: validate/match user_name when multi-user support is added.
func LbzGetFeedbackForRecordingsHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		user, ok := lbzFeedbackUser(w, r, store)
		if !ok {
			return
		}

		mbids, err := parseRecordingMBIDs(r)
		if err != nil {
			utils.WriteError(w, err.Error(), http.StatusBadRequest)
			return
		}

		items, err := store.GetFeedbackForRecordings(ctx, db.GetFeedbackForRecordingsOpts{
			UserID: user.ID,
			MBIDs:  mbids,
		})
		if err != nil {
			l.Err(err).Msg("LbzGetFeedbackForRecordingsHandler: Failed to get feedback")
			utils.WriteError(w, "failed to retrieve feedback", http.StatusInternalServerError)
			return
		}

		feedback := make([]lbzFeedbackItem, len(items))
		for i, item := range items {
			feedback[i] = toLbzFeedbackItem(user.Username, item)
		}
		utils.WriteJSON(w, http.StatusOK, map[string]any{"feedback": feedback})
	}
}

func lbzFeedbackUser(w http.ResponseWriter, r *http.Request, store db.UserStore) (*models.User, bool) {
	u, err := store.GetAdminUser(r.Context())
	if err != nil || u == nil {
		logger.FromContext(r.Context()).Err(err).Msg("lbzFeedbackUser: Failed to get admin user")
		utils.WriteError(w, "failed to look up user", http.StatusInternalServerError)
		return nil, false
	}
	return u, true
}

func toLbzFeedbackItem(username string, item db.RecordingFeedback) lbzFeedbackItem {
	out := lbzFeedbackItem{UserID: username, Score: item.Score, Created: item.CreatedAt}
	if item.RecordingMBID != nil {
		s := item.RecordingMBID.String()
		out.RecordingMBID = &s
	}
	return out
}

func parseRecordingMBIDs(r *http.Request) ([]uuid.UUID, error) {
	var raw []string
	if r.Method == http.MethodPost {
		body, err := utils.DecodeBody[struct {
			RecordingMBIDs []string `json:"recording_mbids"`
		}](r)
		if err != nil {
			return nil, errors.New("invalid JSON sent")
		}
		raw = body.RecordingMBIDs
	} else if param := r.URL.Query().Get("recording_mbids"); param != "" {
		raw = strings.Split(param, ",")
	}

	mbids := make([]uuid.UUID, 0, len(raw))
	for _, s := range raw {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		id, err := uuid.Parse(s)
		if err != nil {
			return nil, errors.New("invalid recording_mbid")
		}
		mbids = append(mbids, id)
	}
	return mbids, nil
}
