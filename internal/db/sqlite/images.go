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
