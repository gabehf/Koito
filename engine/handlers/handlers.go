// package handlers implements route handlers
package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

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
		From:     int64(from),
		To:       int64(to),
		ArtistID: artistId,
		AlbumID:  albumId,
		TrackID:  trackId,
	}
}

// Takes a request and returns a db.Timeframe representing the week, month, year, period, or unix
// time range specified by the request parameters
func TimeframeFromRequest(r *http.Request) db.Timeframe {
	opts := OptsFromRequest(r)
	now := time.Now()
	loc := now.Location()

	// if 'from' is set, but 'to' is not set, assume 'to' should be now
	if opts.From != 0 && opts.To == 0 {
		opts.To = now.Unix()
	}

	// YEAR
	if opts.Year != 0 && opts.Month == 0 && opts.Week == 0 {
		start := time.Date(opts.Year, 1, 1, 0, 0, 0, 0, loc)
		end := time.Date(opts.Year+1, 1, 1, 0, 0, 0, 0, loc).Add(-time.Second)

		opts.From = start.Unix()
		opts.To = end.Unix()
	}

	// MONTH (+ optional year)
	if opts.Month != 0 {
		year := opts.Year
		if year == 0 {
			year = now.Year()
			if int(now.Month()) < opts.Month {
				year--
			}
		}

		start := time.Date(year, time.Month(opts.Month), 1, 0, 0, 0, 0, loc)
		end := endOfMonth(year, time.Month(opts.Month), loc)

		opts.From = start.Unix()
		opts.To = end.Unix()
	}

	// WEEK (+ optional year)
	if opts.Week != 0 {
		year := opts.Year
		if year == 0 {
			year = now.Year()

			_, currentWeek := now.ISOWeek()
			if currentWeek < opts.Week {
				year--
			}
		}

		// ISO week 1 is defined as the week with Jan 4 in it
		jan4 := time.Date(year, 1, 4, 0, 0, 0, 0, loc)
		week1Start := startOfWeek(jan4)

		start := week1Start.AddDate(0, 0, (opts.Week-1)*7)
		end := endOfWeek(start)

		opts.From = start.Unix()
		opts.To = end.Unix()
	}

	return db.Timeframe{
		Period: opts.Period,
		T1u:    opts.From,
		T2u:    opts.To,
	}
}
func startOfWeek(t time.Time) time.Time {
	// ISO week: Monday = 1
	weekday := int(t.Weekday())
	if weekday == 0 { // Sunday
		weekday = 7
	}
	return time.Date(t.Year(), t.Month(), t.Day()-weekday+1, 0, 0, 0, 0, t.Location())
}
func endOfWeek(t time.Time) time.Time {
	return startOfWeek(t).AddDate(0, 0, 7).Add(-time.Second)
}
func endOfMonth(year int, month time.Month, loc *time.Location) time.Time {
	startNextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, loc)
	return startNextMonth.Add(-time.Second)
}
