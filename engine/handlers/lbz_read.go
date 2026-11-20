package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gabehf/koito/engine/middleware"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/memkv"
	"github.com/gabehf/koito/internal/models"
	"github.com/gabehf/koito/internal/utils"
	"github.com/go-chi/chi/v5"
)

const (
	lbzDefaultCount = 25
	lbzMaxCount     = 1000
)

// lbzSelfUser returns the authenticated user when the {user_name} path parameter names them.
// Koito has a single user per instance, so every other username is reported as not found.
func lbzSelfUser(w http.ResponseWriter, r *http.Request) (*models.User, bool) {
	u := middleware.GetUserFromContext(r.Context())
	if u == nil || chi.URLParam(r, "user_name") != u.Username {
		utils.WriteError(w, "user not found", http.StatusNotFound)
		return nil, false
	}
	return u, true
}

// lbzCount parses the ListenBrainz `count` parameter, applying its default and upper bound.
func lbzCount(r *http.Request) (int, error) {
	raw := r.URL.Query().Get("count")
	if raw == "" {
		return lbzDefaultCount, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return 0, errors.New("count must be a positive integer")
	}
	return min(n, lbzMaxCount), nil
}

// lbzOptionalInt parses an optional integer query parameter. Absent means 0.
func lbzOptionalInt(r *http.Request, name string) (int64, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return 0, nil
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, errors.New(name + " must be an integer")
	}
	return n, nil
}

func LbzGetListensHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		u, ok := lbzSelfUser(w, r)
		if !ok {
			return
		}

		count, err := lbzCount(r)
		if err != nil {
			utils.WriteError(w, err.Error(), http.StatusBadRequest)
			return
		}
		minTs, err := lbzOptionalInt(r, "min_ts")
		if err != nil {
			utils.WriteError(w, err.Error(), http.StatusBadRequest)
			return
		}
		maxTs, err := lbzOptionalInt(r, "max_ts")
		if err != nil {
			utils.WriteError(w, err.Error(), http.StatusBadRequest)
			return
		}

		listens, err := store.GetListensInRange(ctx, db.ListenRangeOpts{From: minTs, To: maxTs, Limit: count})
		if err != nil {
			l.Err(err).Msg("LbzGetListensHandler: Failed to get listens")
			utils.WriteError(w, "failed to get listens", http.StatusInternalServerError)
			return
		}

		payload := LbzListensPayload{
			Count:   len(listens),
			Listens: make([]LbzListen, 0, len(listens)),
		}
		for _, listen := range listens {
			payload.Listens = append(payload.Listens, LbzListen{
				ListenedAt:    unixPtr(listen.ListenedAt),
				UserName:      u.Username,
				TrackMetadata: lbzTrackMetadataFromDetail(&listen.Track),
			})
		}

		if len(listens) > 0 {
			latest, err := store.GetLatestListenUnix(ctx)
			if err != nil {
				l.Err(err).Msg("LbzGetListensHandler: Failed to get latest listen")
				utils.WriteError(w, "failed to get listens", http.StatusInternalServerError)
				return
			}
			oldest, err := store.GetFirstListenUnix(ctx)
			if err != nil {
				l.Err(err).Msg("LbzGetListensHandler: Failed to get oldest listen")
				utils.WriteError(w, "failed to get listens", http.StatusInternalServerError)
				return
			}
			payload.LatestListenTs = &latest
			payload.OldestListenTs = &oldest
		}

		utils.WriteJSON(w, http.StatusOK, lbzListensResponse{Payload: payload})
	}
}

func LbzPlayingNowHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		u, ok := lbzSelfUser(w, r)
		if !ok {
			return
		}

		listens := []LbzListen{}
		if trackID, ok := nowPlayingTrackID(u.ID); ok {
			track, err := store.GetTrackDetail(ctx, trackID)
			if err != nil {
				l.Err(err).Msg("LbzPlayingNowHandler: Failed to get currently playing track")
				utils.WriteError(w, "failed to get currently playing track", http.StatusInternalServerError)
				return
			}
			playingNow := true
			listens = append(listens, LbzListen{
				PlayingNow:    &playingNow,
				UserName:      u.Username,
				TrackMetadata: lbzTrackMetadataFromDetail(track),
			})
		}

		utils.WriteJSON(w, http.StatusOK, lbzPlayingNowResponse{Payload: lbzPlayingNowPayload{
			Count:      len(listens),
			Listens:    listens,
			PlayingNow: len(listens) > 0,
			UserID:     strconv.Itoa(int(u.ID)),
		}})
	}
}

func LbzListenCountHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		if _, ok := lbzSelfUser(w, r); !ok {
			return
		}

		count, err := store.CountListens(ctx, db.Timeframe{Period: db.PeriodAllTime})
		if err != nil {
			l.Err(err).Msg("LbzListenCountHandler: Failed to count listens")
			utils.WriteError(w, "failed to count listens", http.StatusInternalServerError)
			return
		}

		var resp lbzListenCountResponse
		resp.Payload.Count = count
		utils.WriteJSON(w, http.StatusOK, resp)
	}
}

// LbzFollowingHandler reports no followed users: Koito has no social graph.
func LbzFollowingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := lbzSelfUser(w, r)
		if !ok {
			return
		}
		utils.WriteJSON(w, http.StatusOK, lbzFollowingResponse{Following: []string{}, User: u.Username})
	}
}

// nowPlayingTrackID returns the track most recently submitted as playing_now for the user.
func nowPlayingTrackID(userID int32) (int32, bool) {
	v, ok := memkv.Store.Get(strconv.Itoa(int(userID)))
	if !ok {
		return 0, false
	}
	id, ok := v.(int32)
	return id, ok
}
