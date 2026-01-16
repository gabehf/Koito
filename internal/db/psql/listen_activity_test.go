package psql_test

import (
	"context"
	"testing"

	"github.com/gabehf/koito/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func flattenListenCounts(items []db.ListenActivityItem) []int64 {
	ret := make([]int64, len(items))
	for i, v := range items {
		ret[i] = v.Listens
	}
	return ret
}

// TODO: This test has some inherent flakiness due to local time possibly crossing day boundaries with UTC
func TestListenActivity(t *testing.T) {
	truncateTestData(t)

	err := store.Exec(context.Background(),
		`INSERT INTO artists (musicbrainz_id)
			VALUES ('00000000-0000-0000-0000-000000000001'),
				   ('00000000-0000-0000-0000-000000000002')`)
	require.NoError(t, err)

	// Move artist names into artist_aliases
	err = store.Exec(context.Background(),
		`INSERT INTO artist_aliases (artist_id, alias, source, is_primary)
			VALUES (1, 'Artist One', 'Testing', true),
				   (2, 'Artist Two', 'Testing', true)`)
	require.NoError(t, err)

	// Insert release groups
	err = store.Exec(context.Background(),
		`INSERT INTO releases (musicbrainz_id)
			VALUES ('00000000-0000-0000-0000-000000000011'),
				   ('00000000-0000-0000-0000-000000000022')`)
	require.NoError(t, err)

	// Move release titles into release_aliases
	err = store.Exec(context.Background(),
		`INSERT INTO release_aliases (release_id, alias, source, is_primary)
			VALUES (1, 'Release One', 'Testing', true),
				   (2, 'Release Two', 'Testing', true)`)
	require.NoError(t, err)

	// Insert tracks
	err = store.Exec(context.Background(),
		`INSERT INTO tracks (musicbrainz_id, release_id)
			VALUES ('11111111-1111-1111-1111-111111111111', 1),
				   ('22222222-2222-2222-2222-222222222222', 2)`)
	require.NoError(t, err)

	// Move track titles into track_aliases
	err = store.Exec(context.Background(),
		`INSERT INTO track_aliases (track_id, alias, source, is_primary)
			VALUES (1, 'Track One', 'Testing', true),
				   (2, 'Track Two', 'Testing', true)`)
	require.NoError(t, err)

	// Associate tracks with artists
	err = store.Exec(context.Background(),
		`INSERT INTO artist_tracks (artist_id, track_id)
			VALUES (1, 1), (2, 2)`)
	require.NoError(t, err)

	// Insert listens
	err = store.Exec(context.Background(),
		`INSERT INTO listens (user_id, track_id, listened_at)
			VALUES (1, 1, NOW() - INTERVAL '1 day'),
				   (1, 1, NOW() - INTERVAL '2 days'),
				   (1, 1, NOW() - INTERVAL '1 week 1 day'),
				   (1, 1, NOW() - INTERVAL '1 month 1 day'),
				   (1, 1, NOW() - INTERVAL '1 year 1 day'),
				   (1, 2, NOW() - INTERVAL '1 day'),
				   (1, 2, NOW() - INTERVAL '2 days'),
				   (1, 2, NOW() - INTERVAL '1 week 1 day'),
				   (1, 2, NOW() - INTERVAL '1 month 1 day'),
				   (1, 2, NOW() - INTERVAL '1 year 1 day')`)
	require.NoError(t, err)

	ctx := context.Background()

	// Test for opts.Step = db.StepDay
	activity, err := store.GetListenActivity(ctx, db.ListenActivityOpts{Step: db.StepDay})
	require.NoError(t, err)
	require.Len(t, activity, 3)
	assert.Equal(t, []int64{2, 2, 2}, flattenListenCounts(activity))

	// Truncate listens table and insert specific dates for testing opts.Step = db.StepMonth
	err = store.Exec(context.Background(), `TRUNCATE TABLE listens`)
	require.NoError(t, err)

	err = store.Exec(context.Background(),
		`INSERT INTO listens (user_id, track_id, listened_at)
			VALUES (1, 1, NOW() - INTERVAL '1 month 1 day'),
				   (1, 1, NOW() - INTERVAL '2 months 1 day'),
				   (1, 1, NOW() - INTERVAL '3 months 1 day'),
				   (1, 2, NOW() - INTERVAL '1 month 1 day'),
				   (1, 2, NOW() - INTERVAL '1 second'),
				   (1, 2, NOW() - INTERVAL '2 seconds'),
				   (1, 2, NOW() - INTERVAL '2 months 1 day')`)
	require.NoError(t, err)

	activity, err = store.GetListenActivity(ctx, db.ListenActivityOpts{Step: db.StepMonth, Range: 8})
	require.NoError(t, err)
	require.Len(t, activity, 4)
	assert.Equal(t, []int64{1, 2, 2, 2}, flattenListenCounts(activity))

	// Truncate listens table and insert specific dates for testing opts.Step = db.StepYear
	err = store.Exec(context.Background(), `TRUNCATE TABLE listens RESTART IDENTITY`)
	require.NoError(t, err)

	err = store.Exec(context.Background(),
		`INSERT INTO listens (user_id, track_id, listened_at)
			VALUES (1, 1, NOW() - INTERVAL '1 year'),
				   (1, 1, NOW() - INTERVAL '2 years'),
				   (1, 2, NOW() - INTERVAL '1 year'),
				   (1, 2, NOW() - INTERVAL '3 years')`)
	require.NoError(t, err)

	activity, err = store.GetListenActivity(ctx, db.ListenActivityOpts{Step: db.StepYear})
	require.NoError(t, err)
	require.Len(t, activity, 3)
	assert.Equal(t, []int64{1, 1, 2}, flattenListenCounts(activity))
	// Truncate and insert data for a specific month/year
	err = store.Exec(context.Background(), `TRUNCATE TABLE listens RESTART IDENTITY`)
	require.NoError(t, err)

	err = store.Exec(context.Background(), `
	INSERT INTO listens (user_id, track_id, listened_at)
	VALUES (1, 1, TIMESTAMP WITH TIME ZONE '2024-03-10T12:00:00Z'),
	       (1, 2, TIMESTAMP WITH TIME ZONE '2024-03-20T12:00:00Z')`)
	require.NoError(t, err)

	activity, err = store.GetListenActivity(ctx, db.ListenActivityOpts{
		Step:  db.StepDay,
		Month: 3,
		Year:  2024,
	})
	require.NoError(t, err)
	require.Len(t, activity, 2) // number of days in march
	t.Log(activity)
	assert.EqualValues(t, 1, activity[0].Listens)
	assert.EqualValues(t, 1, activity[1].Listens)

	// Truncate and insert listens associated with two different albums
	err = store.Exec(context.Background(), `TRUNCATE TABLE listens RESTART IDENTITY`)
	require.NoError(t, err)

	err = store.Exec(context.Background(), `
	INSERT INTO listens (user_id, track_id, listened_at)
	VALUES (1, 1, NOW() - INTERVAL '1 day'), (1, 1, NOW() - INTERVAL '2 days'),
	       (1, 2, NOW() - INTERVAL '1 day')`)
	require.NoError(t, err)

	activity, err = store.GetListenActivity(ctx, db.ListenActivityOpts{
		Step:    db.StepDay,
		AlbumID: 1, // Track 1 only
	})
	require.NoError(t, err)
	require.Len(t, activity, 2)
	assert.Equal(t, []int64{1, 1}, flattenListenCounts(activity))

	activity, err = store.GetListenActivity(ctx, db.ListenActivityOpts{
		Step:    db.StepDay,
		TrackID: 1, // Track 1 only
	})
	require.NoError(t, err)
	require.Len(t, activity, 2)
	assert.Equal(t, []int64{1, 1}, flattenListenCounts(activity))

	activity, err = store.GetListenActivity(ctx, db.ListenActivityOpts{
		Step:     db.StepDay,
		ArtistID: 2, // Should only include listens to Track 2
	})
	require.NoError(t, err)
	require.Len(t, activity, 1)
	assert.Equal(t, []int64{1}, flattenListenCounts(activity))

	// month without year is disallowed
	_, err = store.GetListenActivity(ctx, db.ListenActivityOpts{
		Step:  db.StepDay,
		Month: 5,
	})
	assert.Error(t, err)
}
