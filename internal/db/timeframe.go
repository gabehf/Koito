package db

import (
	"time"
)

type Timeframe struct {
	Period   Period
	Year     int
	Month    int
	Week     int
	FromUnix int64
	ToUnix   int64
	From     time.Time
	To       time.Time
	Timezone *time.Location
}

func TimeframeToTimeRange(tf Timeframe) (t1, t2 time.Time) {
	now := time.Now()
	loc := tf.Timezone
	if loc == nil {
		loc, _ = time.LoadLocation("UTC")
	}

	// ---------------------------------------------------------------------
	// 1. Explicit From / To (time.Time) — highest precedence
	// ---------------------------------------------------------------------
	if !tf.From.IsZero() {
		if tf.To.IsZero() {
			return tf.From, now
		}
		return tf.From, tf.To
	}

	// ---------------------------------------------------------------------
	// 2. Unix timestamps
	// ---------------------------------------------------------------------
	if tf.FromUnix != 0 {
		t1 = time.Unix(tf.FromUnix, 0).In(loc)
		if tf.ToUnix == 0 {
			return t1, now
		}
		t2 = time.Unix(tf.ToUnix, 0).In(loc)
		return t1, t2
	}

	// ---------------------------------------------------------------------
	// 3. Derived ranges (Year / Month / Week)
	// ---------------------------------------------------------------------

	// YEAR only
	if tf.Year != 0 && tf.Month == 0 && tf.Week == 0 {
		start := time.Date(tf.Year, 1, 1, 0, 0, 0, 0, loc)
		end := time.Date(tf.Year+1, 1, 1, 0, 0, 0, 0, loc).Add(-time.Second)
		return start, end
	}

	// MONTH (+ optional year)
	if tf.Month != 0 {
		year := tf.Year
		if year == 0 {
			year = now.Year()
			if int(now.Month()) < tf.Month {
				year--
			}
		}

		start := time.Date(year, time.Month(tf.Month), 1, 0, 0, 0, 0, loc)
		end := endOfMonth(year, time.Month(tf.Month), loc)
		return start, end
	}

	// WEEK (+ optional year)
	if tf.Week != 0 {
		year := tf.Year
		if year == 0 {
			year = now.Year()
			_, currentWeek := now.ISOWeek()
			if currentWeek < tf.Week {
				year--
			}
		}

		// ISO week 1 contains Jan 4
		jan4 := time.Date(year, 1, 4, 0, 0, 0, 0, loc)
		week1Start := startOfWeek(jan4)

		start := week1Start.AddDate(0, 0, (tf.Week-1)*7)
		end := endOfWeek(start)
		return start, end
	}

	// ---------------------------------------------------------------------
	// 4. Period
	// ---------------------------------------------------------------------

	if !tf.Period.IsZero() {
		return StartTimeFromPeriod(tf.Period), now
	}

	// ---------------------------------------------------------------------
	// 5. Fallback: empty timeframe → zero values
	// ---------------------------------------------------------------------
	return time.Time{}, time.Time{}
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
