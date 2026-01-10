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

func StartTimeFromPeriod(p Period) time.Time {
	now := time.Now()
	switch p {
	case "day":
		return now.AddDate(0, 0, -1)
	case "week":
		return now.AddDate(0, 0, -7)
	case "month":
		return now.AddDate(0, -1, 0)
	case "year":
		return now.AddDate(-1, 0, 0)
	case "all_time":
		return time.Time{}
	default:
		// default 1 day
		return now.AddDate(0, 0, -1)
	}
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
	now := time.Now()
	loc := opts.Timezone

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
		start = startOfThisWeek.AddDate(0, 0, -7*opts.Range)
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
