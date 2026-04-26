package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/models"
	"github.com/gabehf/koito/internal/utils"
	"github.com/google/uuid"
)

func (s *Sqlite) GetArtist(ctx context.Context, opts db.GetArtistOpts) (*models.Artist, error) {
	if opts.MusicBrainzID != uuid.Nil {
		var id int32
		err := s.db.QueryRowContext(ctx,
			`SELECT id FROM artists WHERE musicbrainz_id = ? LIMIT 1`, opts.MusicBrainzID.String()).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("GetArtist: by MbzID: %w", db.ErrNotFound)
		}
		if err != nil {
			return nil, fmt.Errorf("GetArtist: by MbzID: %w", err)
		}
		opts.ID = id
	} else if opts.Name != "" {
		var id int32
		err := s.db.QueryRowContext(ctx,
			`SELECT artist_id FROM artist_aliases WHERE alias = ? LIMIT 1`, opts.Name).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("GetArtist: by name: %w", db.ErrNotFound)
		}
		if err != nil {
			return nil, fmt.Errorf("GetArtist: by name: %w", err)
		}
		opts.ID = id
	}

	var mbzID, image, imageSrc sql.NullString
	var name, aliasesConcat sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT a.id, a.musicbrainz_id, a.image, a.image_source, awn.name,
		       group_concat(aa.alias, '||') AS aliases
		FROM artists_with_name awn
		JOIN artists a ON a.id = awn.id
		LEFT JOIN artist_aliases aa ON aa.artist_id = a.id
		WHERE a.id = ?
		GROUP BY a.id`,
		opts.ID,
	).Scan(&opts.ID, &mbzID, &image, &imageSrc, &name, &aliasesConcat)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("GetArtist: %w", db.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("GetArtist: %w", err)
	}

	var aliases []string
	if aliasesConcat.Valid && aliasesConcat.String != "" {
		aliases = strings.Split(aliasesConcat.String, "||")
		utils.Unique(&aliases)
	}

	var listenCount int64
	s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM listens l JOIN artist_tracks at2 ON l.track_id = at2.track_id WHERE at2.artist_id = ?`,
		opts.ID).Scan(&listenCount)

	var timeListened int64
	s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(t.duration), 0)
		FROM listens l JOIN tracks t ON l.track_id = t.id
		JOIN artist_tracks at2 ON t.id = at2.track_id
		WHERE at2.artist_id = ?`,
		opts.ID).Scan(&timeListened)

	var firstListenUnix int64
	err = s.db.QueryRowContext(ctx, `
		SELECT l.listened_at FROM listens l
		JOIN artist_tracks at2 ON l.track_id = at2.track_id
		WHERE at2.artist_id = ? ORDER BY l.listened_at ASC LIMIT 1`,
		opts.ID).Scan(&firstListenUnix)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("GetArtist: first listen: %w", err)
	}

	var rank int64
	s.db.QueryRowContext(ctx, `
		SELECT rank FROM (
			SELECT at2.artist_id,
			       RANK() OVER (ORDER BY COUNT(*) DESC) AS rank
			FROM listens l JOIN artist_tracks at2 ON l.track_id = at2.track_id
			GROUP BY at2.artist_id
		) WHERE artist_id = ?`,
		opts.ID).Scan(&rank)

	return &models.Artist{
		ID:           opts.ID,
		MbzID:        parseNullableUUID(mbzID),
		Name:         name.String,
		Aliases:      aliases,
		Image:        parseNullableUUID(image),
		ListenCount:  listenCount,
		TimeListened: timeListened,
		FirstListen:  firstListenUnix,
		AllTimeRank:  rank,
	}, nil
}

func (s *Sqlite) SaveArtist(ctx context.Context, opts db.SaveArtistOpts) (*models.Artist, error) {
	opts.Name = strings.TrimSpace(opts.Name)
	if opts.Name == "" {
		return nil, errors.New("SaveArtist: name must not be blank")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("SaveArtist: BeginTx: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`INSERT INTO artists (musicbrainz_id, image, image_source) VALUES (?,?,?)`,
		nullableUUID(&opts.MusicBrainzID), nullableUUID(&opts.Image),
		sql.NullString{String: opts.ImageSrc, Valid: opts.ImageSrc != ""},
	)
	if err != nil {
		return nil, fmt.Errorf("SaveArtist: insert: %w", err)
	}
	id64, _ := res.LastInsertId()
	id := int32(id64)

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO artist_aliases (artist_id, alias, source, is_primary) VALUES (?,?,?,1)`,
		id, opts.Name, "Canonical"); err != nil {
		return nil, fmt.Errorf("SaveArtist: canonical alias: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("SaveArtist: commit: %w", err)
	}

	artist := &models.Artist{
		ID:      id,
		Name:    opts.Name,
		Aliases: []string{opts.Name},
	}
	if opts.MusicBrainzID != uuid.Nil {
		u := opts.MusicBrainzID
		artist.MbzID = &u
	}

	if len(opts.Aliases) > 0 {
		if err := s.SaveArtistAliases(ctx, id, opts.Aliases, "MusicBrainz"); err != nil {
			return nil, fmt.Errorf("SaveArtist: SaveArtistAliases: %w", err)
		}
		artist.Aliases = opts.Aliases
	}
	return artist, nil
}

func (s *Sqlite) SaveArtistAliases(ctx context.Context, id int32, aliases []string, source string) error {
	if id == 0 {
		return errors.New("SaveArtistAliases: artist id not specified")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("SaveArtistAliases: BeginTx: %w", err)
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `SELECT alias FROM artist_aliases WHERE artist_id = ?`, id)
	if err != nil {
		return fmt.Errorf("SaveArtistAliases: fetch existing: %w", err)
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
			return errors.New("SaveArtistAliases: aliases cannot be blank")
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO artist_aliases (artist_id, alias, source, is_primary) VALUES (?,?,?,0)`,
			id, alias, source); err != nil {
			return fmt.Errorf("SaveArtistAliases: insert: %w", err)
		}
	}
	return tx.Commit()
}

func (s *Sqlite) UpdateArtist(ctx context.Context, opts db.UpdateArtistOpts) error {
	if opts.ID == 0 {
		return errors.New("UpdateArtist: artist id not specified")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("UpdateArtist: BeginTx: %w", err)
	}
	defer tx.Rollback()

	if opts.MusicBrainzID != uuid.Nil {
		if _, err := tx.ExecContext(ctx,
			`UPDATE artists SET musicbrainz_id = ? WHERE id = ?`,
			opts.MusicBrainzID.String(), opts.ID); err != nil {
			return fmt.Errorf("UpdateArtist: mbzid: %w", err)
		}
	}
	if opts.Image != uuid.Nil {
		if opts.ImageSrc == "" {
			return errors.New("UpdateArtist: image source must be provided when updating an image")
		}
		if _, err := tx.ExecContext(ctx,
			`UPDATE artists SET image = ?, image_source = ? WHERE id = ?`,
			opts.Image.String(), opts.ImageSrc, opts.ID); err != nil {
			return fmt.Errorf("UpdateArtist: image: %w", err)
		}
	}
	return tx.Commit()
}

func (s *Sqlite) DeleteArtist(ctx context.Context, id int32) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM artists WHERE id = ?`, id)
	return err
}

func (s *Sqlite) DeleteArtistAlias(ctx context.Context, id int32, alias string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM artist_aliases WHERE artist_id = ? AND alias = ? AND is_primary = 0`,
		id, alias)
	return err
}

func (s *Sqlite) GetAllArtistAliases(ctx context.Context, id int32) ([]models.Alias, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT alias, source, is_primary FROM artist_aliases WHERE artist_id = ?`, id)
	if err != nil {
		return nil, fmt.Errorf("GetAllArtistAliases: %w", err)
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

func (s *Sqlite) SetPrimaryArtistAlias(ctx context.Context, id int32, alias string) error {
	if id == 0 {
		return errors.New("SetPrimaryArtistAlias: artist id not specified")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("SetPrimaryArtistAlias: BeginTx: %w", err)
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx,
		`SELECT alias, is_primary FROM artist_aliases WHERE artist_id = ?`, id)
	if err != nil {
		return fmt.Errorf("SetPrimaryArtistAlias: fetch: %w", err)
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
		return errors.New("SetPrimaryArtistAlias: alias does not exist")
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE artist_aliases SET is_primary = 1 WHERE artist_id = ? AND alias = ?`, id, alias); err != nil {
		return fmt.Errorf("SetPrimaryArtistAlias: set new: %w", err)
	}
	if primary != "" {
		if _, err := tx.ExecContext(ctx,
			`UPDATE artist_aliases SET is_primary = 0 WHERE artist_id = ? AND alias = ?`, id, primary); err != nil {
			return fmt.Errorf("SetPrimaryArtistAlias: clear old: %w", err)
		}
	}
	return tx.Commit()
}

func (s *Sqlite) GetArtistsForAlbum(ctx context.Context, id int32) ([]*models.Artist, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT a.id, awn.name, a.musicbrainz_id, a.image, ar.is_primary
		FROM artist_releases ar
		JOIN artists_with_name awn ON awn.id = ar.artist_id
		JOIN artists a ON a.id = ar.artist_id
		WHERE ar.release_id = ?
		ORDER BY ar.is_primary DESC, awn.name`,
		id)
	if err != nil {
		return nil, fmt.Errorf("GetArtistsForAlbum: %w", err)
	}
	defer rows.Close()
	var artists []*models.Artist
	for rows.Next() {
		var a models.Artist
		var mbzID, image sql.NullString
		var isPrimary int
		if err := rows.Scan(&a.ID, &a.Name, &mbzID, &image, &isPrimary); err != nil {
			return nil, err
		}
		a.MbzID = parseNullableUUID(mbzID)
		a.Image = parseNullableUUID(image)
		a.IsPrimary = isPrimary == 1
		artists = append(artists, &a)
	}
	return artists, rows.Err()
}

func (s *Sqlite) GetArtistsForTrack(ctx context.Context, id int32) ([]*models.Artist, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT a.id, awn.name, a.musicbrainz_id, a.image, at2.is_primary
		FROM artist_tracks at2
		JOIN artists_with_name awn ON awn.id = at2.artist_id
		JOIN artists a ON a.id = at2.artist_id
		WHERE at2.track_id = ?
		ORDER BY at2.is_primary DESC, awn.name`,
		id)
	if err != nil {
		return nil, fmt.Errorf("GetArtistsForTrack: %w", err)
	}
	defer rows.Close()
	var artists []*models.Artist
	for rows.Next() {
		var a models.Artist
		var mbzID, image sql.NullString
		var isPrimary int
		if err := rows.Scan(&a.ID, &a.Name, &mbzID, &image, &isPrimary); err != nil {
			return nil, err
		}
		a.MbzID = parseNullableUUID(mbzID)
		a.Image = parseNullableUUID(image)
		a.IsPrimary = isPrimary == 1
		artists = append(artists, &a)
	}
	return artists, rows.Err()
}

func (s *Sqlite) GetTopArtistsPaginated(ctx context.Context, opts db.GetItemsOpts) (*db.PaginatedResponse[db.RankedItem[*models.Artist]], error) {
	if opts.Limit == 0 {
		opts.Limit = defaultItemsPerPage
	}
	offset := (opts.Page - 1) * opts.Limit
	t1, t2 := db.TimeframeToTimeRange(opts.Timeframe)

	rows, err := s.db.QueryContext(ctx, `
		SELECT x.id, x.name, x.musicbrainz_id, x.image, x.listen_count,
		       RANK() OVER (ORDER BY x.listen_count DESC) AS rank
		FROM (
			SELECT a.id, awn.name, a.musicbrainz_id, a.image, COUNT(*) AS listen_count
			FROM listens l
			JOIN tracks t ON l.track_id = t.id
			JOIN artist_tracks at2 ON at2.track_id = t.id
			JOIN artists_with_name awn ON awn.id = at2.artist_id
			JOIN artists a ON a.id = at2.artist_id
			WHERE l.listened_at BETWEEN ? AND ?
			GROUP BY a.id, awn.name, a.musicbrainz_id, a.image
		) x
		ORDER BY x.listen_count DESC, x.id
		LIMIT ? OFFSET ?`,
		t1.Unix(), t2.Unix(), opts.Limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("GetTopArtistsPaginated: %w", err)
	}
	defer rows.Close()

	var artists []db.RankedItem[*models.Artist]
	for rows.Next() {
		var a models.Artist
		var mbzID, image sql.NullString
		var item db.RankedItem[*models.Artist]
		if err := rows.Scan(&a.ID, &a.Name, &mbzID, &image, &a.ListenCount, &item.Rank); err != nil {
			return nil, err
		}
		a.MbzID = parseNullableUUID(mbzID)
		a.Image = parseNullableUUID(image)
		item.Item = &a
		artists = append(artists, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var count int64
	s.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT at2.artist_id)
		FROM listens l JOIN artist_tracks at2 ON l.track_id = at2.track_id
		WHERE l.listened_at BETWEEN ? AND ?`,
		t1.Unix(), t2.Unix()).Scan(&count)

	if artists == nil {
		artists = []db.RankedItem[*models.Artist]{}
	}
	return &db.PaginatedResponse[db.RankedItem[*models.Artist]]{
		Items:        artists,
		TotalCount:   count,
		ItemsPerPage: int32(opts.Limit),
		HasNextPage:  int64(offset+len(artists)) < count,
		CurrentPage:  int32(opts.Page),
	}, nil
}

func (s *Sqlite) ArtistsWithoutImages(ctx context.Context, from int32) ([]*models.Artist, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, musicbrainz_id, image, image_source, name
		FROM artists_with_name
		WHERE image IS NULL AND id > ?
		ORDER BY id ASC LIMIT 20`,
		from)
	if err != nil {
		return nil, fmt.Errorf("ArtistsWithoutImages: %w", err)
	}
	defer rows.Close()
	var artists []*models.Artist
	for rows.Next() {
		var a models.Artist
		var mbzID, image, imageSrc sql.NullString
		if err := rows.Scan(&a.ID, &mbzID, &image, &imageSrc, &a.Name); err != nil {
			return nil, err
		}
		a.MbzID = parseNullableUUID(mbzID)
		a.Image = parseNullableUUID(image)
		artists = append(artists, &a)
	}
	return artists, rows.Err()
}

func (s *Sqlite) SetPrimaryAlbumArtist(ctx context.Context, id int32, artistId int32, value bool) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("SetPrimaryAlbumArtist: BeginTx: %w", err)
	}
	defer tx.Rollback()

	var primary int32
	tx.QueryRowContext(ctx,
		`SELECT artist_id FROM artist_releases WHERE release_id = ? AND is_primary = 1 LIMIT 1`, id).
		Scan(&primary)

	isPrimaryInt := 0
	if value {
		isPrimaryInt = 1
	}
	if value && primary == artistId {
		return nil
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE artist_releases SET is_primary = ? WHERE artist_id = ? AND release_id = ?`,
		isPrimaryInt, artistId, id); err != nil {
		return fmt.Errorf("SetPrimaryAlbumArtist: update: %w", err)
	}
	if value && primary != 0 && primary != artistId {
		if _, err := tx.ExecContext(ctx,
			`UPDATE artist_releases SET is_primary = 0 WHERE artist_id = ? AND release_id = ?`,
			primary, id); err != nil {
			return fmt.Errorf("SetPrimaryAlbumArtist: clear old: %w", err)
		}
	}
	return tx.Commit()
}

func (s *Sqlite) SetPrimaryTrackArtist(ctx context.Context, id int32, artistId int32, value bool) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("SetPrimaryTrackArtist: BeginTx: %w", err)
	}
	defer tx.Rollback()

	var primary int32
	tx.QueryRowContext(ctx,
		`SELECT artist_id FROM artist_tracks WHERE track_id = ? AND is_primary = 1 LIMIT 1`, id).
		Scan(&primary)

	isPrimaryInt := 0
	if value {
		isPrimaryInt = 1
	}
	if value && primary == artistId {
		return nil
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE artist_tracks SET is_primary = ? WHERE artist_id = ? AND track_id = ?`,
		isPrimaryInt, artistId, id); err != nil {
		return fmt.Errorf("SetPrimaryTrackArtist: update: %w", err)
	}
	if value && primary != 0 && primary != artistId {
		if _, err := tx.ExecContext(ctx,
			`UPDATE artist_tracks SET is_primary = 0 WHERE artist_id = ? AND track_id = ?`,
			primary, id); err != nil {
			return fmt.Errorf("SetPrimaryTrackArtist: clear old: %w", err)
		}
	}
	return tx.Commit()
}

func (s *Sqlite) CountArtists(ctx context.Context, timeframe db.Timeframe) (int64, error) {
	t1, t2 := db.TimeframeToTimeRange(timeframe)
	var count int64
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT at2.artist_id)
		FROM listens l JOIN artist_tracks at2 ON l.track_id = at2.track_id
		WHERE l.listened_at BETWEEN ? AND ?`,
		t1.Unix(), t2.Unix()).Scan(&count)
	return count, err
}

func (s *Sqlite) CountNewArtists(ctx context.Context, timeframe db.Timeframe) (int64, error) {
	t1, t2 := db.TimeframeToTimeRange(timeframe)
	var count int64
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM (
			SELECT at2.artist_id
			FROM listens l
			JOIN tracks t ON l.track_id = t.id
			JOIN artist_tracks at2 ON t.id = at2.track_id
			GROUP BY at2.artist_id
			HAVING MIN(l.listened_at) BETWEEN ? AND ?
		)`,
		t1.Unix(), t2.Unix()).Scan(&count)
	return count, err
}

func (s *Sqlite) MergeArtists(ctx context.Context, fromId, toId int32, replaceImage bool) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("MergeArtists: BeginTx: %w", err)
	}
	defer tx.Rollback()

	if replaceImage {
		var image, imageSrc sql.NullString
		tx.QueryRowContext(ctx, `SELECT image, image_source FROM artists WHERE id = ?`, fromId).
			Scan(&image, &imageSrc)
		if image.Valid {
			if _, err := tx.ExecContext(ctx,
				`UPDATE artists SET image = ?, image_source = ? WHERE id = ?`,
				image, imageSrc, toId); err != nil {
				return fmt.Errorf("MergeArtists: update image: %w", err)
			}
		}
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM artist_tracks WHERE artist_id = ? AND track_id IN (
			SELECT track_id FROM artist_tracks WHERE artist_id = ?
		)`, fromId, toId); err != nil {
		return fmt.Errorf("MergeArtists: delete conflicting tracks: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM artist_releases WHERE artist_id = ? AND release_id IN (
			SELECT release_id FROM artist_releases WHERE artist_id = ?
		)`, fromId, toId); err != nil {
		return fmt.Errorf("MergeArtists: delete conflicting releases: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE artist_tracks SET artist_id = ? WHERE artist_id = ?`, toId, fromId); err != nil {
		return fmt.Errorf("MergeArtists: update tracks: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE artist_releases SET artist_id = ? WHERE artist_id = ?`, toId, fromId); err != nil {
		return fmt.Errorf("MergeArtists: update releases: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM artists WHERE id = ?`, fromId); err != nil {
		return fmt.Errorf("MergeArtists: delete from: %w", err)
	}
	if err := cleanOrphanedEntries(ctx, tx); err != nil {
		return fmt.Errorf("MergeArtists: clean: %w", err)
	}
	return tx.Commit()
}
