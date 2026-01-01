// package handlers implements route handlers
package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
)

const defaultLimitSize = 100
const maximumLimit = 500

func OptsFromRequest(r *http.Request) db.GetItemsOpts {
	l := logger.FromContext(r.Context())

	l.Debug().Msg("OptsFromRequest: Parsing query parameters")

	limitStr := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		l.Debug().Msgf("OptsFromRequest: Query parameter 'limit' not specified, using default %d", defaultLimitSize)
		limit = defaultLimitSize
	}
	if limit > maximumLimit {
		l.Debug().Msgf("OptsFromRequest: Limit exceeds maximum %d, using default %d", maximumLimit, defaultLimitSize)
		limit = defaultLimitSize
	}

	pageStr := r.URL.Query().Get("page")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		l.Debug().Msg("OptsFromRequest: Page parameter is less than 1, defaulting to 1")
		page = 1
	}

	artistIdStr := r.URL.Query().Get("artist_id")
	artistId, _ := strconv.Atoi(artistIdStr)
	albumIdStr := r.URL.Query().Get("album_id")
	albumId, _ := strconv.Atoi(albumIdStr)
	trackIdStr := r.URL.Query().Get("track_id")
	trackId, _ := strconv.Atoi(trackIdStr)

	tf := TimeframeFromRequest(r)

	var period db.Period
	switch strings.ToLower(r.URL.Query().Get("period")) {
	case "day":
		period = db.PeriodDay
	case "week":
		period = db.PeriodWeek
	case "month":
		period = db.PeriodMonth
	case "year":
		period = db.PeriodYear
	case "all_time":
		period = db.PeriodAllTime
	}

	l.Debug().Msgf("OptsFromRequest: Parsed options: limit=%d, page=%d, week=%d, month=%d, year=%d, from=%d, to=%d, artist_id=%d, album_id=%d, track_id=%d, period=%s",
		limit, page, tf.Week, tf.Month, tf.Year, tf.FromUnix, tf.ToUnix, artistId, albumId, trackId, period)

	return db.GetItemsOpts{
		Limit:     limit,
		Page:      page,
		Timeframe: tf,
		ArtistID:  artistId,
		AlbumID:   albumId,
		TrackID:   trackId,
	}
}

func TimeframeFromRequest(r *http.Request) db.Timeframe {
	q := r.URL.Query()

	parseInt := func(key string) int {
		v := q.Get(key)
		if v == "" {
			return 0
		}
		i, _ := strconv.Atoi(v)
		return i
	}

	parseInt64 := func(key string) int64 {
		v := q.Get(key)
		if v == "" {
			return 0
		}
		i, _ := strconv.ParseInt(v, 10, 64)
		return i
	}

	return db.Timeframe{
		Period:   db.Period(q.Get("period")),
		Year:     parseInt("year"),
		Month:    parseInt("month"),
		Week:     parseInt("week"),
		FromUnix: parseInt64("from"),
		ToUnix:   parseInt64("to"),
	}
}
