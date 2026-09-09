package db

import (
	"time"
)

// should this be in db package ???

type Period string

const (
	PeriodDay     Period = "day"
	PeriodWeek    Period = "week"
	PeriodMonth   Period = "month"
	PeriodYear    Period = "year"
	PeriodAllTime Period = "all_time"
)

func (p Period) IsZero() bool {
	return p == ""
}

// StartTimeFromPeriod returns the start time for the given period, as of now.
//
// When calendarAnchored is false, it returns a rolling window relative to now
// (e.g. "week" means the last 7 days).
//
// When calendarAnchored is true, day/week/month/year resolve to calendar
// boundaries computed in loc, with week boundaries starting on weekStart.
func StartTimeFromPeriod(p Period, now time.Time, calendarAnchored bool, loc *time.Location, weekStart time.Weekday) time.Time {
	if !calendarAnchored {
		switch p {
		case PeriodDay:
			return now.AddDate(0, 0, -1)
		case PeriodWeek:
			return now.AddDate(0, 0, -7)
		case PeriodMonth:
			return now.AddDate(0, -1, 0)
		case PeriodYear:
			return now.AddDate(-1, 0, 0)
		case PeriodAllTime:
			return time.Time{}
		default:
			// default 1 day
			return now.AddDate(0, 0, -1)
		}
	}

	local := now.In(loc)
	switch p {
	case PeriodDay:
		return calendarStartOfDay(local)
	case PeriodWeek:
		return calendarStartOfWeek(local, weekStart)
	case PeriodMonth:
		return calendarStartOfMonth(local)
	case PeriodYear:
		return calendarStartOfYear(local)
	case PeriodAllTime:
		return time.Time{}
	default:
		return calendarStartOfDay(local)
	}
}

func calendarStartOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// calendarStartOfWeek returns 00:00 on the most recent occurrence of weekStart
// on or before t's calendar day. This is distinct from startOfWeek/endOfWeek
// in timeframe.go, which are fixed to ISO Monday and used only for the explicit
// ?week=N ISO-week-number path - a different concept from a configurable
// week-start day.
func calendarStartOfWeek(t time.Time, weekStart time.Weekday) time.Time {
	day := calendarStartOfDay(t)
	diff := int(day.Weekday() - weekStart)
	if diff < 0 {
		diff += 7
	}
	return day.AddDate(0, 0, -diff)
}

func calendarStartOfMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, t.Location())
}

func calendarStartOfYear(t time.Time) time.Time {
	y, _, _ := t.Date()
	return time.Date(y, 1, 1, 0, 0, 0, 0, t.Location())
}

type StepInterval string

const (
	StepDay     StepInterval = "day"
	StepWeek    StepInterval = "week"
	StepMonth   StepInterval = "month"
	StepYear    StepInterval = "year"
	StepDefault StepInterval = "day"

	DefaultRange int = 12
)

// start is the time of 00:00 at the beginning of opts.Range opts.Steps ago,
// end is the end time of the current opts.Step.
// E.g. if step is StepWeek and range is 4, start will be the time 00:00 on Sunday on the 4th week ago,
// and end will be 23:59:59 on Saturday at the end of the current week.
// If opts.Year (or opts.Year + opts.Month) is provided, start and end will simply by the start and end times of that year/month.
func ListenActivityOptsToTimes(opts ListenActivityOpts) (start, end time.Time) {
	loc := opts.Timezone
	if loc == nil {
		loc, _ = time.LoadLocation("UTC")
	}
	now := time.Now().In(loc)

	// If Year (and optionally Month) are specified, use calendar boundaries
	if opts.Year != 0 {
		if opts.Month != 0 {
			// Specific month of a specific year
			start = time.Date(opts.Year, time.Month(opts.Month), 1, 0, 0, 0, 0, loc)
			end = start.AddDate(0, 1, 0).Add(-time.Nanosecond)
		} else {
			// Whole year
			start = time.Date(opts.Year, 1, 1, 0, 0, 0, 0, loc)
			end = start.AddDate(1, 0, 0).Add(-time.Nanosecond)
		}
		return start, end
	}

	// X days ago + today = range
	opts.Range = opts.Range - 1

	// Determine step and align accordingly
	switch opts.Step {
	case StepDay:
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		start = today.AddDate(0, 0, -opts.Range)
		end = today.AddDate(0, 0, 1).Add(-time.Nanosecond)

	case StepWeek:
		// Align to most recent Sunday
		weekday := int(now.Weekday()) // Sunday = 0
		startOfThisWeek := time.Date(now.Year(), now.Month(), now.Day()-weekday, 0, 0, 0, 0, loc)
		// need to subtract 1 from range for week because we are going back from the beginning of this
		// week, so we sort of already went back a week
		start = startOfThisWeek.AddDate(0, 0, -7*(opts.Range-1))
		end = startOfThisWeek.AddDate(0, 0, 7).Add(-time.Nanosecond)

	case StepMonth:
		firstOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		start = firstOfThisMonth.AddDate(0, -opts.Range, 0)
		end = firstOfThisMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)

	case StepYear:
		firstOfThisYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, loc)
		start = firstOfThisYear.AddDate(-opts.Range, 0, 0)
		end = firstOfThisYear.AddDate(1, 0, 0).Add(-time.Nanosecond)

	default:
		// Default to daily
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		start = today.AddDate(0, 0, -opts.Range)
		end = today.AddDate(0, 0, 1).Add(-time.Nanosecond)
	}

	return start, end
}
