package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

func (s *Sqlite) ImageHasAssociation(ctx context.Context, image uuid.UUID) (bool, error) {
	var exists int
	err := s.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM artists WHERE image = ?
			UNION ALL
			SELECT 1 FROM releases WHERE image = ?
		)`, image.String(), image.String()).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("ImageHasAssociation: %w", err)
	}
	return exists == 1, nil
}

func (s *Sqlite) GetImageSource(ctx context.Context, image uuid.UUID) (string, error) {
	var src sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT image_source FROM artists WHERE image = ?
		UNION ALL
		SELECT image_source FROM releases WHERE image = ?
		LIMIT 1`,
		image.String(), image.String()).Scan(&src)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("GetImageSource: %w", err)
	}
	return src.String, nil
}

func (s *Sqlite) GetUserUploadedImageIDs(ctx context.Context) ([]uuid.UUID, error) {
	ret := make([]uuid.UUID, 0)
	rows, err := s.db.QueryContext(ctx, `
		SELECT image FROM artists WHERE image_source = 'User Upload' AND image IS NOT NULL`)
	if err != nil {
		return nil, fmt.Errorf("GetUserUploadedImageIDs: %w", err)
	}
	for rows.Next() {
		var imgidstr string
		err := rows.Scan(&imgidstr)
		if err != nil {
			return nil, fmt.Errorf("GetUserUploadedImageIDs: rows.Scan: %w", err)
		}
		imgid, err := uuid.Parse(imgidstr)
		if err != nil {
			return nil, fmt.Errorf("GetUserUploadedImageIDs: uuid.Parse: %w", err)
		}
		ret = append(ret, imgid)
	}
	rows, err = s.db.QueryContext(ctx, `
		SELECT image FROM releases WHERE image_source = 'User Upload' AND image IS NOT NULL`)
	if err != nil {
		return nil, fmt.Errorf("GetUserUploadedImageIDs: %w", err)
	}
	for rows.Next() {
		var imgidstr string
		err := rows.Scan(&imgidstr)
		if err != nil {
			return nil, fmt.Errorf("GetUserUploadedImageIDs: rows.Scan: %w", err)
		}
		imgid, err := uuid.Parse(imgidstr)
		if err != nil {
			return nil, fmt.Errorf("GetUserUploadedImageIDs: uuid.Parse: %w", err)
		}
		ret = append(ret, imgid)
	}
	return ret, nil
}
