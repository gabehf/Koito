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

func (s *Sqlite) GetAlbum(ctx context.Context, opts db.GetAlbumOpts) (*models.Album, error) {
	if opts.MusicBrainzID != uuid.Nil {
		var id int32
		err := s.db.QueryRowContext(ctx,
			`SELECT id FROM releases WHERE musicbrainz_id = ? LIMIT 1`, opts.MusicBrainzID.String()).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("GetAlbum: by MbzID: %w", db.ErrNotFound)
		}
		if err != nil {
			return nil, fmt.Errorf("GetAlbum: by MbzID: %w", err)
		}
		opts.ID = id
	} else if opts.ArtistID != 0 && opts.Title != "" {
		var id int32
		err := s.db.QueryRowContext(ctx, `
			SELECT r.id FROM releases_with_title r
			JOIN artist_releases ar ON r.id = ar.release_id
			WHERE r.title = ? AND ar.artist_id = ? LIMIT 1`,
			opts.Title, opts.ArtistID).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("GetAlbum: by artist+title: %w", db.ErrNotFound)
		}
		if err != nil {
			return nil, fmt.Errorf("GetAlbum: by artist+title: %w", err)
		}
		opts.ID = id
	} else if opts.ArtistID != 0 && len(opts.Titles) > 0 {
		id, err := s.releaseByArtistAndTitles(ctx, opts.ArtistID, opts.Titles, false)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("GetAlbum: by artist+titles: %w", db.ErrNotFound)
		}
		if err != nil {
			return nil, fmt.Errorf("GetAlbum: by artist+titles: %w", err)
		}
		opts.ID = id
	}

	return s.getAlbumByID(ctx, opts.ID)
}

func (s *Sqlite) getAlbumByID(ctx context.Context, id int32) (*models.Album, error) {
	var ret models.Album
	var mbzID, image, imageSrc sql.NullString
	var variousArtists int
	err := s.db.QueryRowContext(ctx, `
		SELECT id, musicbrainz_id, image, image_source, various_artists, title
		FROM releases_with_title WHERE id = ? LIMIT 1`, id).
		Scan(&ret.ID, &mbzID, &image, &imageSrc, &variousArtists, &ret.Title)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("getAlbumByID: %w", db.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("getAlbumByID: %w", err)
	}
	ret.MbzID = parseNullableUUID(mbzID)
	ret.Image = parseNullableUUID(image)
	ret.VariousArtists = variousArtists == 1

	artists, err := s.artistsForRelease(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getAlbumByID: artists: %w", err)
	}
	ret.Artists = artists

	var listenCount int64
	s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM listens l JOIN tracks t ON l.track_id = t.id WHERE t.release_id = ?`,
		id).Scan(&listenCount)
	ret.ListenCount = listenCount

	var timeListened int64
	s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(t.duration), 0)
		FROM listens l JOIN tracks t ON l.track_id = t.id WHERE t.release_id = ?`,
		id).Scan(&timeListened)
	ret.TimeListened = timeListened

	var firstListenUnix int64
	err = s.db.QueryRowContext(ctx, `
		SELECT l.listened_at FROM listens l JOIN tracks t ON l.track_id = t.id
		WHERE t.release_id = ? ORDER BY l.listened_at ASC LIMIT 1`, id).Scan(&firstListenUnix)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("getAlbumByID: first listen: %w", err)
	}
	ret.FirstListen = firstListenUnix

	var rank int64
	s.db.QueryRowContext(ctx, `
		SELECT rank FROM (
			SELECT t.release_id,
			       RANK() OVER (ORDER BY COUNT(*) DESC) AS rank
			FROM listens l JOIN tracks t ON l.track_id = t.id
			GROUP BY t.release_id
		) WHERE release_id = ?`, id).Scan(&rank)
	ret.AllTimeRank = rank

	return &ret, nil
}

func (s *Sqlite) releaseByArtistAndTitles(ctx context.Context, artistID int32, titles []string, requireNoMbzID bool) (int32, error) {
	if len(titles) == 0 {
		return 0, errors.New("releaseByArtistAndTitles: no titles provided")
	}
	placeholders := strings.Repeat("?,", len(titles))
	placeholders = placeholders[:len(placeholders)-1]

	noMbz := ""
	if requireNoMbzID {
		noMbz = " AND r.musicbrainz_id IS NULL"
	}
	query := fmt.Sprintf(`
		SELECT r.id FROM releases_with_title r
		JOIN artist_releases ar ON r.id = ar.release_id
		WHERE r.title IN (%s) AND ar.artist_id = ?%s LIMIT 1`, placeholders, noMbz)

	args := make([]any, len(titles)+1)
	for i, t := range titles {
		args[i] = t
	}
	args[len(titles)] = artistID

	var id int32
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&id)
	return id, err
}

func (s *Sqlite) GetAlbumWithNoMbzIDByTitles(ctx context.Context, artistId int32, titles []string) (*models.Album, error) {
	if artistId == 0 || len(titles) == 0 {
		return nil, errors.New("GetAlbumWithNoMbzIDByTitles: insufficient information")
	}
	id, err := s.releaseByArtistAndTitles(ctx, artistId, titles, true)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("GetAlbumWithNoMbzIDByTitles: %w", db.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("GetAlbumWithNoMbzIDByTitles: %w", err)
	}
	return s.getAlbumByID(ctx, id)
}

func (s *Sqlite) SaveAlbum(ctx context.Context, opts db.SaveAlbumOpts) (*models.Album, error) {
	if len(opts.ArtistIDs) < 1 {
		return nil, errors.New("SaveAlbum: required parameter 'ArtistIDs' missing")
	}
	if slices.Contains(opts.ArtistIDs, 0) {
		return nil, errors.New("SaveAlbum: none of 'ArtistIDs' may be 0")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("SaveAlbum: BeginTx: %w", err)
	}
	defer tx.Rollback()

	variousArtistsInt := 0
	if opts.VariousArtists {
		variousArtistsInt = 1
	}
	res, err := tx.ExecContext(ctx,
		`INSERT INTO releases (musicbrainz_id, various_artists, image, image_source) VALUES (?,?,?,?)`,
		nullableUUID(&opts.MusicBrainzID), variousArtistsInt,
		nullableUUID(&opts.Image),
		sql.NullString{String: opts.ImageSrc, Valid: opts.ImageSrc != ""},
	)
	if err != nil {
		return nil, fmt.Errorf("SaveAlbum: insert: %w", err)
	}
	id64, _ := res.LastInsertId()
	id := int32(id64)

	for i, artistID := range opts.ArtistIDs {
		isPrimary := 0
		if i == 0 {
			isPrimary = 1
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO artist_releases (artist_id, release_id, is_primary) VALUES (?,?,?)`,
			artistID, id, isPrimary); err != nil {
			return nil, fmt.Errorf("SaveAlbum: associate artist: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO release_aliases (release_id, alias, source, is_primary) VALUES (?,?,?,1)`,
		id, opts.Title, "Canonical"); err != nil {
		return nil, fmt.Errorf("SaveAlbum: canonical alias: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("SaveAlbum: commit: %w", err)
	}

	if len(opts.Aliases) > 0 {
		s.SaveAlbumAliases(ctx, id, opts.Aliases, "MusicBrainz")
	}

	ret := &models.Album{
		ID:             id,
		Title:          opts.Title,
		VariousArtists: opts.VariousArtists,
	}
	if opts.MusicBrainzID != uuid.Nil {
		u := opts.MusicBrainzID
		ret.MbzID = &u
	}
	if opts.Image != uuid.Nil {
		u := opts.Image
		ret.Image = &u
	}
	return ret, nil
}

func (s *Sqlite) SaveAlbumAliases(ctx context.Context, id int32, aliases []string, source string) error {
	if id == 0 {
		return errors.New("SaveAlbumAliases: album id not specified")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("SaveAlbumAliases: BeginTx: %w", err)
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `SELECT alias FROM release_aliases WHERE release_id = ?`, id)
	if err != nil {
		return fmt.Errorf("SaveAlbumAliases: fetch existing: %w", err)
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
			return errors.New("SaveAlbumAliases: aliases cannot be blank")
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO release_aliases (release_id, alias, source, is_primary) VALUES (?,?,?,0)`,
			id, alias, source); err != nil {
			return fmt.Errorf("SaveAlbumAliases: insert: %w", err)
		}
	}
	return tx.Commit()
}

func (s *Sqlite) UpdateAlbum(ctx context.Context, opts db.UpdateAlbumOpts) error {
	if opts.ID == 0 {
		return errors.New("UpdateAlbum: missing album id")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("UpdateAlbum: BeginTx: %w", err)
	}
	defer tx.Rollback()

	if opts.MusicBrainzID != uuid.Nil {
		if _, err := tx.ExecContext(ctx,
			`UPDATE releases SET musicbrainz_id = ? WHERE id = ?`,
			opts.MusicBrainzID.String(), opts.ID); err != nil {
			return fmt.Errorf("UpdateAlbum: mbzid: %w", err)
		}
	}
	if opts.Image != uuid.Nil {
		if opts.ImageSrc == "" {
			return errors.New("UpdateAlbum: image source must be provided when updating an image")
		}
		if _, err := tx.ExecContext(ctx,
			`UPDATE releases SET image = ?, image_source = ? WHERE id = ?`,
			opts.Image.String(), opts.ImageSrc, opts.ID); err != nil {
			return fmt.Errorf("UpdateAlbum: image: %w", err)
		}
	}
	if opts.VariousArtistsUpdate {
		v := 0
		if opts.VariousArtistsValue {
			v = 1
		}
		if _, err := tx.ExecContext(ctx,
			`UPDATE releases SET various_artists = ? WHERE id = ?`, v, opts.ID); err != nil {
			return fmt.Errorf("UpdateAlbum: various_artists: %w", err)
		}
	}
	return tx.Commit()
}

func (s *Sqlite) DeleteAlbum(ctx context.Context, id int32) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM releases WHERE id = ?`, id)
	return err
}

func (s *Sqlite) DeleteAlbumAlias(ctx context.Context, id int32, alias string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM release_aliases WHERE release_id = ? AND alias = ? AND is_primary = 0`,
		id, alias)
	return err
}

func (s *Sqlite) GetAllAlbumAliases(ctx context.Context, id int32) ([]models.Alias, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT alias, source, is_primary FROM release_aliases WHERE release_id = ?`, id)
	if err != nil {
		return nil, fmt.Errorf("GetAllAlbumAliases: %w", err)
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

func (s *Sqlite) SetPrimaryAlbumAlias(ctx context.Context, id int32, alias string) error {
	if id == 0 {
		return errors.New("SetPrimaryAlbumAlias: album id not specified")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("SetPrimaryAlbumAlias: BeginTx: %w", err)
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx,
		`SELECT alias, is_primary FROM release_aliases WHERE release_id = ?`, id)
	if err != nil {
		return fmt.Errorf("SetPrimaryAlbumAlias: fetch: %w", err)
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
		return errors.New("SetPrimaryAlbumAlias: alias does not exist")
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE release_aliases SET is_primary = 1 WHERE release_id = ? AND alias = ?`, id, alias); err != nil {
		return fmt.Errorf("SetPrimaryAlbumAlias: set new: %w", err)
	}
	if primary != "" {
		if _, err := tx.ExecContext(ctx,
			`UPDATE release_aliases SET is_primary = 0 WHERE release_id = ? AND alias = ?`, id, primary); err != nil {
			return fmt.Errorf("SetPrimaryAlbumAlias: clear old: %w", err)
		}
	}
	return tx.Commit()
}

func (s *Sqlite) AddArtistsToAlbum(ctx context.Context, opts db.AddArtistsToAlbumOpts) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("AddArtistsToAlbum: BeginTx: %w", err)
	}
	defer tx.Rollback()
	for _, artistID := range opts.ArtistIDs {
		if _, err := tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO artist_releases (artist_id, release_id, is_primary) VALUES (?,?,0)`,
			artistID, opts.AlbumID); err != nil {
			return fmt.Errorf("AddArtistsToAlbum: insert: %w", err)
		}
	}
	return tx.Commit()
}

func (s *Sqlite) GetTopAlbumsPaginated(ctx context.Context, opts db.GetItemsOpts) (*db.PaginatedResponse[db.RankedItem[*models.Album]], error) {
	if opts.Limit == 0 {
		opts.Limit = defaultItemsPerPage
	}
	offset := (opts.Page - 1) * opts.Limit
	t1, t2 := db.TimeframeToTimeRange(opts.Timeframe)

	// Count first so it never competes with an open rows for the connection.
	var count int64
	if opts.ArtistID != 0 {
		s.db.QueryRowContext(ctx, `
			SELECT COUNT(DISTINCT r.id) FROM releases r
			JOIN artist_releases ar ON r.id = ar.release_id
			WHERE ar.artist_id = ?`, opts.ArtistID).Scan(&count)
	} else {
		s.db.QueryRowContext(ctx, `
			SELECT COUNT(DISTINCT t.release_id) FROM listens l JOIN tracks t ON l.track_id = t.id
			WHERE l.listened_at BETWEEN ? AND ?`, t1.Unix(), t2.Unix()).Scan(&count)
	}

	var rows *sql.Rows
	var err error
	if opts.ArtistID != 0 {
		rows, err = s.db.QueryContext(ctx, `
			SELECT x.id, x.title, x.musicbrainz_id, x.image, x.various_artists, x.listen_count,
			       RANK() OVER (ORDER BY x.listen_count DESC) AS rank
			FROM (
				SELECT r.id, r.title, r.musicbrainz_id, r.image, r.various_artists, COUNT(*) AS listen_count
				FROM listens l
				JOIN tracks t ON l.track_id = t.id
				JOIN releases_with_title r ON t.release_id = r.id
				JOIN artist_releases ar ON r.id = ar.release_id
				WHERE ar.artist_id = ? AND l.listened_at BETWEEN ? AND ?
				GROUP BY r.id, r.title, r.musicbrainz_id, r.image, r.various_artists
			) x
			ORDER BY listen_count DESC, x.id
			LIMIT ? OFFSET ?`,
			opts.ArtistID, t1.Unix(), t2.Unix(), opts.Limit, offset,
		)
	} else {
		rows, err = s.db.QueryContext(ctx, `
			SELECT x.id, x.title, x.musicbrainz_id, x.image, x.various_artists, x.listen_count,
			       RANK() OVER (ORDER BY x.listen_count DESC) AS rank
			FROM (
				SELECT r.id, r.title, r.musicbrainz_id, r.image, r.various_artists, COUNT(*) AS listen_count
				FROM listens l
				JOIN tracks t ON l.track_id = t.id
				JOIN releases_with_title r ON t.release_id = r.id
				WHERE l.listened_at BETWEEN ? AND ?
				GROUP BY r.id, r.title, r.musicbrainz_id, r.image, r.various_artists
			) x
			ORDER BY listen_count DESC, x.id
			LIMIT ? OFFSET ?`,
			t1.Unix(), t2.Unix(), opts.Limit, offset,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("GetTopAlbumsPaginated: %w", err)
	}

	type albumRow struct {
		id             int32
		title          string
		mbzID          sql.NullString
		image          sql.NullString
		variousArtists int
		listenCount    int64
		rank           int64
	}
	var raw []albumRow
	for rows.Next() {
		var r albumRow
		if err := rows.Scan(&r.id, &r.title, &r.mbzID, &r.image, &r.variousArtists, &r.listenCount, &r.rank); err != nil {
			rows.Close()
			return nil, err
		}
		raw = append(raw, r)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	albums := make([]db.RankedItem[*models.Album], 0, len(raw))
	for _, r := range raw {
		a := &models.Album{
			ID:             r.id,
			Title:          r.title,
			ListenCount:    r.listenCount,
			MbzID:          parseNullableUUID(r.mbzID),
			Image:          parseNullableUUID(r.image),
			VariousArtists: r.variousArtists == 1,
		}
		a.Artists, err = s.artistsForRelease(ctx, a.ID)
		if err != nil {
			return nil, err
		}
		albums = append(albums, db.RankedItem[*models.Album]{Item: a, Rank: r.rank})
	}

	return &db.PaginatedResponse[db.RankedItem[*models.Album]]{
		Items:        albums,
		TotalCount:   count,
		ItemsPerPage: int32(opts.Limit),
		HasNextPage:  int64(offset+len(albums)) < count,
		CurrentPage:  int32(opts.Page),
	}, nil
}

func (s *Sqlite) AlbumsWithoutImages(ctx context.Context, from int32) ([]*models.Album, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, musicbrainz_id, image, image_source, various_artists, title
		FROM releases_with_title
		WHERE image IS NULL AND id > ?
		ORDER BY id ASC LIMIT 20`,
		from)
	if err != nil {
		return nil, fmt.Errorf("AlbumsWithoutImages: %w", err)
	}
	type albumRow struct {
		album          models.Album
		variousArtists int
	}
	var raw []albumRow
	for rows.Next() {
		var r albumRow
		var mbzID, image, imageSrc sql.NullString
		if err := rows.Scan(&r.album.ID, &mbzID, &image, &imageSrc, &r.variousArtists, &r.album.Title); err != nil {
			rows.Close()
			return nil, err
		}
		r.album.MbzID = parseNullableUUID(mbzID)
		r.album.Image = parseNullableUUID(image)
		raw = append(raw, r)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	var albums []*models.Album
	for _, r := range raw {
		a := r.album
		a.VariousArtists = r.variousArtists == 1
		a.Artists, err = s.artistsForRelease(ctx, a.ID)
		if err != nil {
			return nil, err
		}
		albums = append(albums, &a)
	}
	return albums, rows.Err()
}

func (s *Sqlite) MergeAlbums(ctx context.Context, fromId, toId int32, replaceImage bool) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("MergeAlbums: BeginTx: %w", err)
	}
	defer tx.Rollback()

	// fetch artists from fromId before moving tracks (for re-association)
	fromArtists, err := s.artistsForRelease(ctx, fromId)
	if err != nil {
		return fmt.Errorf("MergeAlbums: fetch from artists: %w", err)
	}

	if replaceImage {
		var image, imageSrc sql.NullString
		tx.QueryRowContext(ctx, `SELECT image, image_source FROM releases WHERE id = ?`, fromId).
			Scan(&image, &imageSrc)
		if image.Valid {
			if _, err := tx.ExecContext(ctx,
				`UPDATE releases SET image = ?, image_source = ? WHERE id = ?`,
				image, imageSrc, toId); err != nil {
				return fmt.Errorf("MergeAlbums: update image: %w", err)
			}
		}
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE tracks SET release_id = ? WHERE release_id = ?`, toId, fromId); err != nil {
		return fmt.Errorf("MergeAlbums: move tracks: %w", err)
	}

	for _, artist := range fromArtists {
		if _, err := tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO artist_releases (artist_id, release_id, is_primary) VALUES (?,?,0)`,
			artist.ID, toId); err != nil {
			return fmt.Errorf("MergeAlbums: associate artist: %w", err)
		}
	}

	if err := cleanOrphanedEntries(ctx, tx); err != nil {
		return fmt.Errorf("MergeAlbums: clean: %w", err)
	}
	return tx.Commit()
}

func (s *Sqlite) CountAlbums(ctx context.Context, timeframe db.Timeframe) (int64, error) {
	t1, t2 := db.TimeframeToTimeRange(timeframe)
	var count int64
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT t.release_id)
		FROM listens l JOIN tracks t ON l.track_id = t.id
		WHERE l.listened_at BETWEEN ? AND ?`,
		t1.Unix(), t2.Unix()).Scan(&count)
	return count, err
}

func (s *Sqlite) CountNewAlbums(ctx context.Context, timeframe db.Timeframe) (int64, error) {
	t1, t2 := db.TimeframeToTimeRange(timeframe)
	var count int64
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM (
			SELECT t.release_id FROM listens l JOIN tracks t ON l.track_id = t.id
			GROUP BY t.release_id
			HAVING MIN(l.listened_at) BETWEEN ? AND ?
		)`,
		t1.Unix(), t2.Unix()).Scan(&count)
	return count, err
}
