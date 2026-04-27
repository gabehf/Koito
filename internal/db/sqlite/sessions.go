package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gabehf/koito/internal/models"
	"github.com/google/uuid"
)

func (s *Sqlite) SaveSession(ctx context.Context, userID int32, expiresAt time.Time, persistent bool) (*models.Session, error) {
	id := uuid.New()
	now := time.Now().Unix()
	persistentInt := 0
	if persistent {
		persistentInt = 1
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO sessions (id, user_id, created_at, expires_at, persistent) VALUES (?,?,?,?,?)`,
		id.String(), userID, now, expiresAt.Unix(), persistentInt,
	)
	if err != nil {
		return nil, fmt.Errorf("SaveSession: %w", err)
	}
	return &models.Session{
		ID:         id,
		UserID:     userID,
		CreatedAt:  time.Unix(now, 0).UTC(),
		ExpiresAt:  expiresAt,
		Persistent: persistent,
	}, nil
}

func (s *Sqlite) RefreshSession(ctx context.Context, sessionId uuid.UUID, expiresAt time.Time) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE sessions SET expires_at = ? WHERE id = ?`,
		expiresAt.Unix(), sessionId.String())
	return err
}

func (s *Sqlite) DeleteSession(ctx context.Context, sessionId uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE id = ?`, sessionId.String())
	return err
}

func (s *Sqlite) GetUserBySession(ctx context.Context, sessionId uuid.UUID) (*models.User, error) {
	var u models.User
	var role string
	err := s.db.QueryRowContext(ctx, `
		SELECT u.id, u.username, u.role, u.password
		FROM users u JOIN sessions se ON u.id = se.user_id
		WHERE se.id = ? AND se.expires_at > ?
		LIMIT 1`,
		sessionId.String(), time.Now().Unix()).Scan(&u.ID, &u.Username, &role, &u.Password)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("GetUserBySession: %w", err)
	}
	u.Role = models.UserRole(role)
	return &u, nil
}
