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

	weekStr := r.URL.Query().Get("week")
	week, _ := strconv.Atoi(weekStr)
	monthStr := r.URL.Query().Get("month")
	month, _ := strconv.Atoi(monthStr)
	yearStr := r.URL.Query().Get("year")
	year, _ := strconv.Atoi(yearStr)
	fromStr := r.URL.Query().Get("from")
	from, _ := strconv.Atoi(fromStr)
	toStr := r.URL.Query().Get("to")
	to, _ := strconv.Atoi(toStr)

	artistIdStr := r.URL.Query().Get("artist_id")
	artistId, _ := strconv.Atoi(artistIdStr)
	albumIdStr := r.URL.Query().Get("album_id")
	albumId, _ := strconv.Atoi(albumIdStr)
	trackIdStr := r.URL.Query().Get("track_id")
	trackId, _ := strconv.Atoi(trackIdStr)

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
	default:
		l.Debug().Msgf("OptsFromRequest: Using default value '%s' for period", db.PeriodDay)
		period = db.PeriodDay
	}

	l.Debug().Msgf("OptsFromRequest: Parsed options: limit=%d, page=%d, week=%d, month=%d, year=%d, from=%d, to=%d, artist_id=%d, album_id=%d, track_id=%d, period=%s",
		limit, page, week, month, year, from, to, artistId, albumId, trackId, period)

	return db.GetItemsOpts{
		Limit:    limit,
		Period:   period,
		Page:     page,
		Week:     week,
		Month:    month,
		Year:     year,
		From:     from,
		To:       to,
		ArtistID: artistId,
		AlbumID:  albumId,
		TrackID:  trackId,
	}
}
