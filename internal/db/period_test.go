package db_test

import (
	"testing"
	"time"

	"github.com/gabehf/koito/internal/db"
	"github.com/stretchr/testify/require"
)

func TestListenActivityOptsToTimes(t *testing.T) {

	// default range
	// opts := db.ListenActivityOpts{}
	// t1, t2 := db.ListenActivityOptsToTimes(opts)
	// t.Logf("%s to %s", t1, t2)
	// assert.WithinDuration(t, bod(time.Now().Add(-11*24*time.Hour)), t1, 5*time.Second)
	// assert.WithinDuration(t, eod(time.Now()), t2, 5*time.Second)
}

func eod(t time.Time) time.Time {
	year, month, day := t.Date()
	loc := t.Location()
	return time.Date(year, month, day, 23, 59, 59, 0, loc)
}

func TestPeriodUnset(t *testing.T) {
	var p db.Period
	require.True(t, p.IsZero())
}

func TestStartTimeFromPeriod_Rolling(t *testing.T) {
	loc := time.UTC
	now := time.Date(2026, time.September, 9, 15, 30, 45, 0, loc)

	require.Equal(t, now.AddDate(0, 0, -1), db.StartTimeFromPeriod(db.PeriodDay, now, false, loc, time.Monday))
	require.Equal(t, now.AddDate(0, 0, -7), db.StartTimeFromPeriod(db.PeriodWeek, now, false, loc, time.Monday))
	require.Equal(t, now.AddDate(0, -1, 0), db.StartTimeFromPeriod(db.PeriodMonth, now, false, loc, time.Monday))
	require.Equal(t, now.AddDate(-1, 0, 0), db.StartTimeFromPeriod(db.PeriodYear, now, false, loc, time.Monday))
	require.True(t, db.StartTimeFromPeriod(db.PeriodAllTime, now, false, loc, time.Monday).IsZero())
}

func TestStartTimeFromPeriod_CalendarAnchored_DayMonthYear(t *testing.T) {
	loc := time.UTC
	now := time.Date(2026, time.September, 9, 15, 30, 45, 0, loc)

	require.Equal(t, time.Date(2026, time.September, 9, 0, 0, 0, 0, loc), db.StartTimeFromPeriod(db.PeriodDay, now, true, loc, time.Monday))
	require.Equal(t, time.Date(2026, time.September, 1, 0, 0, 0, 0, loc), db.StartTimeFromPeriod(db.PeriodMonth, now, true, loc, time.Monday))
	require.Equal(t, time.Date(2026, time.January, 1, 0, 0, 0, 0, loc), db.StartTimeFromPeriod(db.PeriodYear, now, true, loc, time.Monday))
	require.True(t, db.StartTimeFromPeriod(db.PeriodAllTime, now, true, loc, time.Monday).IsZero())
}

func TestStartTimeFromPeriod_CalendarAnchored_Week(t *testing.T) {
	loc := time.UTC

	// Covers every "now" weekday against every configurable week-start day,
	// checked against a naive reference implementation rather than hardcoded
	// dates, since this modulo arithmetic is easy to get subtly wrong.
	weekdays := []time.Weekday{
		time.Sunday, time.Monday, time.Tuesday, time.Wednesday,
		time.Thursday, time.Friday, time.Saturday,
	}

	base := time.Date(2026, time.September, 6, 15, 30, 45, 0, loc) // a Sunday

	for offset := 0; offset < 7; offset++ {
		now := base.AddDate(0, 0, offset)
		for _, weekStart := range weekdays {
			got := db.StartTimeFromPeriod(db.PeriodWeek, now, true, loc, weekStart)
			want := naiveStartOfWeek(now, weekStart)
			require.Equalf(t, want, got, "now=%s (%s) weekStart=%s", now.Format("2006-01-02"), now.Weekday(), weekStart)
		}
	}
}

// naiveStartOfWeek walks backward day by day until it finds the most recent
// occurrence of weekStart on or before t, independent of the modulo-based
// implementation under test.
func naiveStartOfWeek(t time.Time, weekStart time.Weekday) time.Time {
	y, m, d := t.Date()
	day := time.Date(y, m, d, 0, 0, 0, 0, t.Location())
	for day.Weekday() != weekStart {
		day = day.AddDate(0, 0, -1)
	}
	return day
}

func bod(t time.Time) time.Time {
	year, month, day := t.Date()
	loc := t.Location()
	return time.Date(year, month, day, 0, 0, 0, 0, loc)
}
