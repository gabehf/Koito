package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/models"
	"github.com/gabehf/koito/internal/utils"
	"github.com/google/uuid"
)

type LbzStatsResponse struct {
	Payload LbzStatsPayload `json:"payload"`
}

type LbzActivityPayload struct {
	FromTs            int64                  `json:"from_ts"`
	ToTs              int64                  `json:"to_ts"`
	LastUpdated       int64                  `json:"last_updated"`
	Range             string                 `json:"range"`
	ListeningActivity []LbzListeningActivity `json:"listening_activity"`
}

type LbzActivityResponse struct {
	Payload LbzActivityPayload `json:"payload"`
}

// lbzRangeStart returns the start of a ListenBrainz stats range, relative to now.
// The zero time means no lower bound (all_time).
func lbzRangeStart(rangeName string, now time.Time) (time.Time, error) {
	switch rangeName {
	case "all_time", "":
		return time.Time{}, nil
	case "this_week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		return time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, now.Location()), nil
	case "this_month":
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()), nil
	case "this_year":
		return time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()), nil
	case "week":
		return now.AddDate(0, 0, -7), nil
	case "month":
		return now.AddDate(0, -1, 0), nil
	case "quarter":
		return now.AddDate(0, -3, 0), nil
	case "half_yearly":
		return now.AddDate(0, -6, 0), nil
	case "year":
		return now.AddDate(-1, 0, 0), nil
	default:
		return time.Time{}, errors.New("invalid range")
	}
}

// lbzStatsRequest parses the query shared by the stats endpoints. The ListenBrainz offset
// is mapped onto Koito's pages, so it must be a multiple of count.
func lbzStatsRequest(r *http.Request, now time.Time) (rangeName string, from time.Time, offset, count, page int, err error) {
	rangeName = r.URL.Query().Get("range")
	if rangeName == "" {
		rangeName = "all_time"
	}
	from, err = lbzRangeStart(rangeName, now)
	if err != nil {
		return
	}

	count, err = lbzCount(r)
	if err != nil {
		return
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		offset, err = strconv.Atoi(raw)
		if err != nil || offset < 0 {
			err = errors.New("offset must be a non-negative integer")
			return
		}
	}
	if offset%count != 0 {
		err = errors.New("offset must be a multiple of count")
		return
	}
	page = offset/count + 1
	return
}

func lbzTimeframe(from time.Time) db.Timeframe {
	if from.IsZero() {
		return db.Timeframe{Period: db.PeriodAllTime}
	}
	return db.Timeframe{FromUnix: from.Unix()}
}

func lbzArtistNames(artists []*models.Artist) (string, []string) {
	names := make([]string, 0, len(artists))
	mbids := []string{}
	for _, a := range artists {
		names = append(names, a.Name)
		if s := uuidPtrToString(a.MbzID); s != nil {
			mbids = append(mbids, *s)
		}
	}
	return strings.Join(names, ", "), mbids
}

func LbzStatsArtistsHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		if _, ok := lbzSelfUser(w, r); !ok {
			return
		}
		now := time.Now()
		rangeName, from, offset, count, page, err := lbzStatsRequest(r, now)
		if err != nil {
			utils.WriteError(w, err.Error(), http.StatusBadRequest)
			return
		}

		res, err := store.GetTopArtistsPaginated(ctx, db.GetItemsOpts{Limit: count, Page: page, Timeframe: lbzTimeframe(from)})
		if err != nil {
			l.Err(err).Msg("LbzStatsArtistsHandler: Failed to get top artists")
			utils.WriteError(w, "failed to get artists", http.StatusInternalServerError)
			return
		}

		entries := make([]LbzStatsEntry, 0, len(res.Items))
		for _, item := range res.Items {
			a := item.Item
			entries = append(entries, LbzStatsEntry{
				ArtistName:  a.Name,
				ArtistMbids: lbzMbidStrings(a.MbzID),
				ListenCount: a.ListenCount,
			})
		}
		total := res.TotalCount
		payload := lbzStatsPayload(rangeName, from, offset, count, now)
		payload.Artists = entries
		payload.TotalArtistCount = &total
		utils.WriteJSON(w, http.StatusOK, LbzStatsResponse{Payload: payload})
	}
}

func LbzStatsReleasesHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		if _, ok := lbzSelfUser(w, r); !ok {
			return
		}
		now := time.Now()
		rangeName, from, offset, count, page, err := lbzStatsRequest(r, now)
		if err != nil {
			utils.WriteError(w, err.Error(), http.StatusBadRequest)
			return
		}

		res, err := store.GetTopAlbumsPaginated(ctx, db.GetItemsOpts{Limit: count, Page: page, Timeframe: lbzTimeframe(from)})
		if err != nil {
			l.Err(err).Msg("LbzStatsReleasesHandler: Failed to get top releases")
			utils.WriteError(w, "failed to get releases", http.StatusInternalServerError)
			return
		}

		entries := make([]LbzStatsEntry, 0, len(res.Items))
		for _, item := range res.Items {
			a := item.Item
			artists, err := store.GetArtistsForAlbum(ctx, a.ID)
			if err != nil {
				l.Err(err).Msg("LbzStatsReleasesHandler: Failed to get artists for release")
				utils.WriteError(w, "failed to get releases", http.StatusInternalServerError)
				return
			}
			artistName, artistMbids := lbzArtistNames(artists)
			title := a.Title
			entries = append(entries, LbzStatsEntry{
				ArtistName:  artistName,
				ArtistMbids: artistMbids,
				ReleaseName: &title,
				ReleaseMbid: uuidPtrToString(a.MbzID),
				ListenCount: a.ListenCount,
			})
		}
		total := res.TotalCount
		payload := lbzStatsPayload(rangeName, from, offset, count, now)
		payload.Releases = entries
		payload.TotalReleaseCount = &total
		utils.WriteJSON(w, http.StatusOK, LbzStatsResponse{Payload: payload})
	}
}

func LbzStatsRecordingsHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		if _, ok := lbzSelfUser(w, r); !ok {
			return
		}
		now := time.Now()
		rangeName, from, offset, count, page, err := lbzStatsRequest(r, now)
		if err != nil {
			utils.WriteError(w, err.Error(), http.StatusBadRequest)
			return
		}

		res, err := store.GetTopTracksPaginated(ctx, db.GetItemsOpts{Limit: count, Page: page, Timeframe: lbzTimeframe(from)})
		if err != nil {
			l.Err(err).Msg("LbzStatsRecordingsHandler: Failed to get top recordings")
			utils.WriteError(w, "failed to get recordings", http.StatusInternalServerError)
			return
		}

		entries := make([]LbzStatsEntry, 0, len(res.Items))
		for _, item := range res.Items {
			t := item.Item
			d, err := store.GetTrackDetail(ctx, t.ID)
			if err != nil {
				l.Err(err).Msg("LbzStatsRecordingsHandler: Failed to get track detail")
				utils.WriteError(w, "failed to get recordings", http.StatusInternalServerError)
				return
			}
			trackName := d.TrackTitle
			entries = append(entries, LbzStatsEntry{
				ArtistName:    strings.Join(d.ArtistNames, ", "),
				ArtistMbids:   uuidsToStrings(d.ArtistMbids),
				TrackName:     &trackName,
				RecordingMbid: uuidPtrToString(d.TrackMbid),
				ReleaseName:   nonEmptyPtr(d.ReleaseTitle),
				ReleaseMbid:   uuidPtrToString(d.ReleaseMbid),
				ListenCount:   t.ListenCount,
			})
		}
		total := res.TotalCount
		payload := lbzStatsPayload(rangeName, from, offset, count, now)
		payload.Recordings = entries
		payload.TotalRecordingCount = &total
		utils.WriteJSON(w, http.StatusOK, LbzStatsResponse{Payload: payload})
	}
}

func lbzStatsPayload(rangeName string, from time.Time, offset, count int, now time.Time) LbzStatsPayload {
	var fromTs int64
	if !from.IsZero() {
		fromTs = from.Unix()
	}
	return LbzStatsPayload{
		Count:       count,
		Offset:      offset,
		Range:       rangeName,
		FromTs:      fromTs,
		ToTs:        now.Unix(),
		LastUpdated: now.Unix(),
	}
}

func nonEmptyPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// lbzActivitySpec maps a ListenBrainz activity range onto Koito's listen activity options.
func lbzActivitySpec(rangeName string, now time.Time, firstListen time.Time) (db.ListenActivityOpts, error) {
	opts := db.ListenActivityOpts{Timezone: time.UTC}
	switch rangeName {
	case "all_time", "":
		opts.Step = db.StepYear
		opts.Range = now.Year() - firstListen.Year() + 1
	case "this_year":
		opts.Step = db.StepMonth
		opts.Year = now.Year()
	case "year":
		opts.Step = db.StepMonth
		opts.Range = 12
	case "half_yearly":
		opts.Step = db.StepMonth
		opts.Range = 6
	case "quarter":
		opts.Step = db.StepMonth
		opts.Range = 3
	case "this_month":
		opts.Step = db.StepDay
		opts.Year = now.Year()
		opts.Month = int(now.Month())
	case "month":
		opts.Step = db.StepDay
		opts.Range = 30
	case "this_week":
		opts.Step = db.StepDay
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		opts.Range = weekday
	case "week":
		opts.Step = db.StepDay
		opts.Range = 7
	default:
		return opts, errors.New("invalid range")
	}
	return opts, nil
}

func lbzStepLabel(step db.StepInterval, t time.Time) string {
	switch step {
	case db.StepYear:
		return t.Format("2006")
	case db.StepMonth:
		return t.Format("2006-01")
	default:
		return t.Format("2006-01-02")
	}
}

func LbzListeningActivityHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		if _, ok := lbzSelfUser(w, r); !ok {
			return
		}
		now := time.Now().UTC()
		rangeName := r.URL.Query().Get("range")
		if rangeName == "" {
			rangeName = "all_time"
		}

		payload := LbzActivityPayload{
			Range:             rangeName,
			ToTs:              now.Unix(),
			LastUpdated:       now.Unix(),
			ListeningActivity: []LbzListeningActivity{},
		}

		firstUnix, err := store.GetFirstListenUnix(ctx)
		if err != nil {
			// no listens yet: nothing to report
			utils.WriteJSON(w, http.StatusOK, LbzActivityResponse{Payload: payload})
			return
		}

		opts, err := lbzActivitySpec(rangeName, now, time.Unix(firstUnix, 0).UTC())
		if err != nil {
			utils.WriteError(w, err.Error(), http.StatusBadRequest)
			return
		}

		items, err := store.GetListenActivity(ctx, opts)
		if err != nil {
			l.Err(err).Msg("LbzListeningActivityHandler: Failed to get listen activity")
			utils.WriteError(w, "failed to get listen activity", http.StatusInternalServerError)
			return
		}
		items = processActivity(items, opts)

		for _, item := range items {
			end := addStep(item.Start, opts.Step)
			payload.ListeningActivity = append(payload.ListeningActivity, LbzListeningActivity{
				FromTs:      item.Start.Unix(),
				ToTs:        end.Unix(),
				ListenCount: item.Listens,
				TimeRange:   lbzStepLabel(opts.Step, item.Start),
			})
		}
		if len(items) > 0 {
			payload.FromTs = items[0].Start.Unix()
		}

		utils.WriteJSON(w, http.StatusOK, LbzActivityResponse{Payload: payload})
	}
}

// lbzMbidStrings returns the non-nil MBIDs as strings, never nil.
func lbzMbidStrings(ids ...*uuid.UUID) []string {
	out := []string{}
	for _, id := range ids {
		if s := uuidPtrToString(id); s != nil {
			out = append(out, *s)
		}
	}
	return out
}
