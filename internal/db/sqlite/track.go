package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/models"
	"github.com/gabehf/koito/internal/utils"
	"github.com/google/uuid"
)

func (s *Sqlite) GetTrack(ctx context.Context, opts db.GetTrackOpts) (*models.Track, error) {
	if opts.MusicBrainzID != uuid.Nil {
		var id int32
		err := s.db.QueryRowContext(ctx,
			`SELECT id FROM tracks WHERE musicbrainz_id = ? LIMIT 1`, opts.MusicBrainzID.String()).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("GetTrack: by MbzID: %w", db.ErrNotFound)
		}
		if err != nil {
			return nil, fmt.Errorf("GetTrack: by MbzID: %w", err)
		}
		opts.ID = id
	} else if len(opts.ArtistIDs) > 0 && opts.ReleaseID != 0 {
		id, err := s.trackByTrackInfo(ctx, opts.Title, opts.ReleaseID, opts.ArtistIDs)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("GetTrack: by track info: %w", db.ErrNotFound)
		}
		if err != nil {
			return nil, fmt.Errorf("GetTrack: by track info: %w", err)
		}
		opts.ID = id
	}

	return s.getTrackByID(ctx, opts.ID)
}

func (s *Sqlite) getTrackByID(ctx context.Context, id int32) (*models.Track, error) {
	var track models.Track
	var mbzID, image sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT t.id, t.musicbrainz_id, t.duration, t.release_id, t.title, r.image
		FROM tracks_with_title t
		JOIN releases r ON t.release_id = r.id
		WHERE t.id = ? LIMIT 1`, id).
		Scan(&track.ID, &mbzID, &track.Duration, &track.AlbumID, &track.Title, &image)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("getTrackByID: %w", db.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("getTrackByID: %w", err)
	}
	track.MbzID = parseNullableUUID(mbzID)
	track.Image = parseNullableUUID(image)

	artists, err := s.artistsForTrack(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getTrackByID: artists: %w", err)
	}
	track.Artists = artists

	var listenCount int64
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM listens WHERE track_id = ?`, id).Scan(&listenCount)
	track.ListenCount = listenCount

	var timeListened int64
	s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(t.duration), 0) FROM listens l JOIN tracks t ON l.track_id = t.id WHERE t.id = ?`,
		id).Scan(&timeListened)
	track.TimeListened = timeListened

	var firstListenUnix int64
	err = s.db.QueryRowContext(ctx,
		`SELECT listened_at FROM listens WHERE track_id = ? ORDER BY listened_at ASC LIMIT 1`, id).
		Scan(&firstListenUnix)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("getTrackByID: first listen: %w", err)
	}
	track.FirstListen = firstListenUnix

	var rank int64
	s.db.QueryRowContext(ctx, `
		SELECT rank FROM (
			SELECT track_id, RANK() OVER (ORDER BY COUNT(*) DESC) AS rank
			FROM listens GROUP BY track_id
		) WHERE track_id = ?`, id).Scan(&rank)
	track.AllTimeRank = rank

	return &track, nil
}

// trackByTrackInfo finds a track by title + release + all artist IDs. Mirrors GetTrackByTrackInfo.
func (s *Sqlite) trackByTrackInfo(ctx context.Context, title string, releaseID int32, artistIDs []int32) (int32, error) {
	if len(artistIDs) == 0 {
		return 0, errors.New("trackByTrackInfo: no artist IDs provided")
	}
	placeholders := strings.Repeat("?,", len(artistIDs))
	placeholders = placeholders[:len(placeholders)-1]
	query := fmt.Sprintf(`
		SELECT t.id FROM tracks_with_title t
		JOIN artist_tracks at2 ON at2.track_id = t.id
		WHERE t.title = ? AND t.release_id = ? AND at2.artist_id IN (%s)
		GROUP BY t.id
		HAVING COUNT(DISTINCT at2.artist_id) = ?`, placeholders)

	args := make([]any, 2+len(artistIDs)+1)
	args[0] = title
	args[1] = releaseID
	for i, id := range artistIDs {
		args[2+i] = id
	}
	args[2+len(artistIDs)] = len(artistIDs)

	var id int32
	return id, s.db.QueryRowContext(ctx, query, args...).Scan(&id)
}

func (s *Sqlite) SaveTrack(ctx context.Context, opts db.SaveTrackOpts) (*models.Track, error) {
	if len(opts.ArtistIDs) < 1 {
		return nil, errors.New("SaveTrack: required parameter 'ArtistIDs' missing")
	}
	if slices.Contains(opts.ArtistIDs, 0) {
		return nil, errors.New("SaveTrack: none of 'ArtistIDs' may be 0")
	}
	if opts.AlbumID == 0 {
		return nil, errors.New("SaveTrack: required parameter 'AlbumID' missing")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("SaveTrack: BeginTx: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`INSERT INTO tracks (musicbrainz_id, release_id, duration) VALUES (?,?,?)`,
		nullableUUID(&opts.RecordingMbzID), opts.AlbumID, opts.Duration,
	)
	if err != nil {
		return nil, fmt.Errorf("SaveTrack: insert: %w", err)
	}
	id64, _ := res.LastInsertId()
	id := int32(id64)

	for i, artistID := range opts.ArtistIDs {
		isPrimary := 0
		if i == 0 {
			isPrimary = 1
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO artist_tracks (artist_id, track_id, is_primary) VALUES (?,?,?)`,
			artistID, id, isPrimary); err != nil {
			return nil, fmt.Errorf("SaveTrack: associate artist to track: %w", err)
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO artist_releases (artist_id, release_id, is_primary) VALUES (?,?,0)`,
			artistID, opts.AlbumID); err != nil {
			return nil, fmt.Errorf("SaveTrack: associate artist to release: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO track_aliases (track_id, alias, source, is_primary) VALUES (?,?,?,1)`,
		id, opts.Title, "Canonical"); err != nil {
		return nil, fmt.Errorf("SaveTrack: canonical alias: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("SaveTrack: commit: %w", err)
	}

	ret := &models.Track{
		ID:       id,
		Title:    opts.Title,
		Duration: opts.Duration,
		AlbumID:  opts.AlbumID,
	}
	if opts.RecordingMbzID != uuid.Nil {
		u := opts.RecordingMbzID
		ret.MbzID = &u
	}
	return ret, nil
}

func (s *Sqlite) SaveTrackAliases(ctx context.Context, id int32, aliases []string, source string) error {
	if id == 0 {
		return errors.New("SaveTrackAliases: track id not specified")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("SaveTrackAliases: BeginTx: %w", err)
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `SELECT alias FROM track_aliases WHERE track_id = ?`, id)
	if err != nil {
		return fmt.Errorf("SaveTrackAliases: fetch existing: %w", err)
	}
	for rows.Next() {
		var a string
		rows.Scan(&a)
		aliases = append(aliases, a)
	}
	rows.Close()

	utils.Unique(&aliases)
	for _, alias := range aliases {
		alias = strings.TrimSpace(alias)
		if alias == "" {
			return errors.New("SaveTrackAliases: aliases cannot be blank")
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO track_aliases (track_id, alias, source, is_primary) VALUES (?,?,?,0)`,
			id, alias, source); err != nil {
			return fmt.Errorf("SaveTrackAliases: insert: %w", err)
		}
	}
	return tx.Commit()
}

func (s *Sqlite) UpdateTrack(ctx context.Context, opts db.UpdateTrackOpts) error {
	if opts.ID == 0 {
		return errors.New("UpdateTrack: track id not specified")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("UpdateTrack: BeginTx: %w", err)
	}
	defer tx.Rollback()

	if opts.MusicBrainzID != uuid.Nil {
		if _, err := tx.ExecContext(ctx,
			`UPDATE tracks SET musicbrainz_id = ? WHERE id = ?`,
			opts.MusicBrainzID.String(), opts.ID); err != nil {
			return fmt.Errorf("UpdateTrack: mbzid: %w", err)
		}
	}
	if opts.Duration != 0 {
		if _, err := tx.ExecContext(ctx,
			`UPDATE tracks SET duration = ? WHERE id = ?`,
			opts.Duration, opts.ID); err != nil {
			return fmt.Errorf("UpdateTrack: duration: %w", err)
		}
	}
	return tx.Commit()
}

func (s *Sqlite) DeleteTrack(ctx context.Context, id int32) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("DeleteTrack: BeginTx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM tracks WHERE id = ?`, id); err != nil {
		return fmt.Errorf("DeleteTrack: delete: %w", err)
	}
	if err := cleanOrphanedEntries(ctx, tx); err != nil {
		return fmt.Errorf("DeleteTrack: clean: %w", err)
	}
	return tx.Commit()
}

func (s *Sqlite) DeleteTrackAlias(ctx context.Context, id int32, alias string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM track_aliases WHERE track_id = ? AND alias = ? AND is_primary = 0`,
		id, alias)
	return err
}

func (s *Sqlite) GetAllTrackAliases(ctx context.Context, id int32) ([]models.Alias, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT alias, source, is_primary FROM track_aliases WHERE track_id = ?`, id)
	if err != nil {
		return nil, fmt.Errorf("GetAllTrackAliases: %w", err)
	}
	defer rows.Close()
	var aliases []models.Alias
	for rows.Next() {
		var a models.Alias
		var isPrimary int
		if err := rows.Scan(&a.Alias, &a.Source, &isPrimary); err != nil {
			return nil, err
		}
		a.ID = id
		a.Primary = isPrimary == 1
		aliases = append(aliases, a)
	}
	return aliases, rows.Err()
}

func (s *Sqlite) SetPrimaryTrackAlias(ctx context.Context, id int32, alias string) error {
	if id == 0 {
		return errors.New("SetPrimaryTrackAlias: track id not specified")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("SetPrimaryTrackAlias: BeginTx: %w", err)
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx,
		`SELECT alias, is_primary FROM track_aliases WHERE track_id = ?`, id)
	if err != nil {
		return fmt.Errorf("SetPrimaryTrackAlias: fetch: %w", err)
	}
	var primary, exists string
	for rows.Next() {
		var a string
		var isPrimary int
		rows.Scan(&a, &isPrimary)
		if isPrimary == 1 {
			primary = a
		}
		if a == alias {
			exists = a
		}
	}
	rows.Close()

	if primary == alias {
		return nil
	}
	if exists == "" {
		return errors.New("SetPrimaryTrackAlias: alias does not exist")
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE track_aliases SET is_primary = 1 WHERE track_id = ? AND alias = ?`, id, alias); err != nil {
		return fmt.Errorf("SetPrimaryTrackAlias: set new: %w", err)
	}
	if primary != "" {
		if _, err := tx.ExecContext(ctx,
			`UPDATE track_aliases SET is_primary = 0 WHERE track_id = ? AND alias = ?`, id, primary); err != nil {
			return fmt.Errorf("SetPrimaryTrackAlias: clear old: %w", err)
		}
	}
	return tx.Commit()
}

func (s *Sqlite) GetTracksWithNoDurationButHaveMbzID(ctx context.Context, from int32) ([]*models.Track, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, musicbrainz_id, duration, release_id, title
		FROM tracks_with_title
		WHERE duration = 0 AND musicbrainz_id IS NOT NULL AND id > ?
		ORDER BY id ASC LIMIT 20`,
		from)
	if err != nil {
		return nil, fmt.Errorf("GetTracksWithNoDurationButHaveMbzID: %w", err)
	}
	defer rows.Close()
	var tracks []*models.Track
	for rows.Next() {
		var t models.Track
		var mbzID sql.NullString
		if err := rows.Scan(&t.ID, &mbzID, &t.Duration, &t.AlbumID, &t.Title); err != nil {
			return nil, err
		}
		t.MbzID = parseNullableUUID(mbzID)
		tracks = append(tracks, &t)
	}
	return tracks, rows.Err()
}

func (s *Sqlite) GetTopTracksPaginated(ctx context.Context, opts db.GetItemsOpts) (*db.PaginatedResponse[db.RankedItem[*models.Track]], error) {
	if opts.Limit == 0 {
		opts.Limit = defaultItemsPerPage
	}
	offset := (opts.Page - 1) * opts.Limit
	t1, t2 := db.TimeframeToTimeRange(opts.Timeframe)

	var rows *sql.Rows
	var err error

	switch {
	case opts.AlbumID > 0:
		query := `
			WITH TrackCounts AS (
				SELECT l.track_id, COUNT(*) AS listen_count
				FROM listens l
				JOIN tracks t ON l.track_id = t.id
				WHERE l.listened_at BETWEEN ? AND ? AND t.release_id = ?
				GROUP BY l.track_id
			),
			RankedTracks AS (
				SELECT track_id, listen_count,
					   RANK() OVER (ORDER BY listen_count DESC) AS rank,
					   COUNT(*) OVER () AS total_count
				FROM TrackCounts
				ORDER BY listen_count DESC, track_id
				LIMIT ? OFFSET ?
			)
			SELECT r.track_id, twt.title, twt.musicbrainz_id, twt.release_id, rls.image, r.listen_count, r.rank, r.total_count
			FROM RankedTracks r
			JOIN tracks_with_title twt ON twt.id = r.track_id
			JOIN releases rls ON twt.release_id = rls.id
			ORDER BY r.rank, r.track_id`

		rows, err = s.db.QueryContext(ctx, query, t1.Unix(), t2.Unix(), opts.AlbumID, opts.Limit, offset)

	case opts.ArtistID > 0:
		query := `
			WITH TrackCounts AS (
				SELECT l.track_id, COUNT(*) AS listen_count
				FROM listens l
				JOIN artist_tracks at2 ON l.track_id = at2.track_id
				WHERE l.listened_at BETWEEN ? AND ? AND at2.artist_id = ?
				GROUP BY l.track_id
			),
			RankedTracks AS (
				SELECT track_id, listen_count,
					   RANK() OVER (ORDER BY listen_count DESC) AS rank,
					   COUNT(*) OVER () AS total_count
				FROM TrackCounts
				ORDER BY listen_count DESC, track_id
				LIMIT ? OFFSET ?
			)
			SELECT r.track_id, twt.title, twt.musicbrainz_id, twt.release_id, rls.image, r.listen_count, r.rank, r.total_count
			FROM RankedTracks r
			JOIN tracks_with_title twt ON twt.id = r.track_id
			JOIN releases rls ON twt.release_id = rls.id
			ORDER BY r.rank, r.track_id`

		rows, err = s.db.QueryContext(ctx, query, t1.Unix(), t2.Unix(), opts.ArtistID, opts.Limit, offset)

	default:
		query := `
			WITH TrackCounts AS (
				SELECT track_id, COUNT(*) AS listen_count
				FROM listens
				WHERE listened_at BETWEEN ? AND ?
				GROUP BY track_id
			),
			RankedTracks AS (
				SELECT track_id, listen_count,
					   RANK() OVER (ORDER BY listen_count DESC) AS rank,
					   COUNT(*) OVER () AS total_count
				FROM TrackCounts
				ORDER BY listen_count DESC, track_id
				LIMIT ? OFFSET ?
			)
			SELECT r.track_id, twt.title, twt.musicbrainz_id, twt.release_id, rls.image, r.listen_count, r.rank, r.total_count
			FROM RankedTracks r
			JOIN tracks_with_title twt ON twt.id = r.track_id
			JOIN releases rls ON twt.release_id = rls.id
			ORDER BY r.rank, r.track_id`

		rows, err = s.db.QueryContext(ctx, query, t1.Unix(), t2.Unix(), opts.Limit, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("GetTopTracksPaginated: %w", err)
	}
	defer rows.Close()

	// Pre-allocate slice
	tracks := make([]db.RankedItem[*models.Track], 0, opts.Limit)
	var totalCount int64

	// Single iteration loop
	for rows.Next() {
		var t models.Track
		var mbzID, image sql.NullString
		var item db.RankedItem[*models.Track]

		// Scan totalCount directly alongside the row data
		if err := rows.Scan(&t.ID, &t.Title, &mbzID, &t.AlbumID, &image, &t.ListenCount, &item.Rank, &totalCount); err != nil {
			return nil, err
		}

		t.MbzID = parseNullableUUID(mbzID)
		t.Image = parseNullableUUID(image)

		// N+1 Query (acceptable if volume is low, otherwise consider batching)
		t.Artists, err = s.artistsForTrack(ctx, t.ID)
		if err != nil {
			return nil, err
		}

		item.Item = &t
		tracks = append(tracks, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &db.PaginatedResponse[db.RankedItem[*models.Track]]{
		Items:        tracks,
		TotalCount:   totalCount,
		ItemsPerPage: int32(opts.Limit),
		HasNextPage:  int64(offset+len(tracks)) < totalCount,
		CurrentPage:  int32(opts.Page),
	}, nil
}

func (s *Sqlite) MergeTracks(ctx context.Context, fromId, toId int32) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("MergeTracks: BeginTx: %w", err)
	}
	defer tx.Rollback()

	// check if tracks are in different releases
	var fromRelease, toRelease int32
	tx.QueryRowContext(ctx, `SELECT release_id FROM tracks WHERE id = ?`, fromId).Scan(&fromRelease)
	tx.QueryRowContext(ctx, `SELECT release_id FROM tracks WHERE id = ?`, toId).Scan(&toRelease)

	// redirect all listens (ignore conflicts — same timestamp already exists for toId)
	if _, err := tx.ExecContext(ctx,
		`UPDATE OR IGNORE listens SET track_id = ? WHERE track_id = ?`, toId, fromId); err != nil {
		return fmt.Errorf("MergeTracks: redirect listens: %w", err)
	}

	if fromRelease != toRelease {
		// associate fromId's artists with toId's release
		rows, err := tx.QueryContext(ctx,
			`SELECT artist_id FROM artist_tracks WHERE track_id = ?`, fromId)
		if err != nil {
			return fmt.Errorf("MergeTracks: fetch artists: %w", err)
		}
		var artistIDs []int32
		for rows.Next() {
			var aid int32
			rows.Scan(&aid)
			artistIDs = append(artistIDs, aid)
		}
		rows.Close()
		for _, aid := range artistIDs {
			if _, err := tx.ExecContext(ctx,
				`INSERT OR IGNORE INTO artist_releases (artist_id, release_id, is_primary) VALUES (?,?,0)`,
				aid, toRelease); err != nil {
				return fmt.Errorf("MergeTracks: associate artist to release: %w", err)
			}
		}
	}

	if err := cleanOrphanedEntries(ctx, tx); err != nil {
		return fmt.Errorf("MergeTracks: clean: %w", err)
	}
	return tx.Commit()
}

func (s *Sqlite) CountTracks(ctx context.Context, timeframe db.Timeframe) (int64, error) {
	t1, t2 := db.TimeframeToTimeRange(timeframe)
	var count int64
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(DISTINCT track_id) FROM listens WHERE listened_at BETWEEN ? AND ?`,
		t1.Unix(), t2.Unix()).Scan(&count)
	return count, err
}

func (s *Sqlite) CountNewTracks(ctx context.Context, timeframe db.Timeframe) (int64, error) {
	t1, t2 := db.TimeframeToTimeRange(timeframe)
	var count int64
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM (
			SELECT track_id FROM listens
			GROUP BY track_id
			HAVING MIN(listened_at) BETWEEN ? AND ?
		)`,
		t1.Unix(), t2.Unix()).Scan(&count)
	return count, err
}

func (s *Sqlite) SearchTracks(ctx context.Context, q string) ([]*models.Track, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT DISTINCT t.id, t.title, t.musicbrainz_id, t.release_id, r.image
		FROM track_aliases ta
		JOIN tracks_with_title t ON ta.track_id = t.id
		JOIN releases r ON t.release_id = r.id
		WHERE ta.alias LIKE ? OR ta.alias LIKE ?
		LIMIT 50`,
		q+"%", "%"+q+"%",
	)
	if err != nil {
		return nil, fmt.Errorf("SearchTracks: %w", err)
	}
	defer rows.Close()

	type candidate struct {
		track *models.Track
		score float64
	}
	seen := map[int32]float64{}
	var candidates []candidate

	for rows.Next() {
		var t models.Track
		var mbzID, image sql.NullString
		if err := rows.Scan(&t.ID, &t.Title, &mbzID, &t.AlbumID, &image); err != nil {
			return nil, err
		}
		t.MbzID = parseNullableUUID(mbzID)
		t.Image = parseNullableUUID(image)
		score := fuzzyScore(q, t.Title)
		if prev, ok := seen[t.ID]; !ok || score > prev {
			seen[t.ID] = score
			candidates = append(candidates, candidate{&t, score})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var out []*models.Track
	for _, c := range sliceTopN(candidates, func(c candidate) float64 { return c.score }, 8) {
		artists, err := s.artistsForTrack(ctx, c.track.ID)
		if err != nil {
			return nil, err
		}
		c.track.Artists = artists
		out = append(out, c.track)
	}
	return out, nil
}
