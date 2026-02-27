package handlers

import (
	"context"
	"time"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/models"
	"github.com/google/uuid"
)

// Focused interfaces for handlers to depend on smaller contracts.
type ArtistStore interface {
	GetArtist(ctx context.Context, opts db.GetArtistOpts) (*models.Artist, error)
}

type AlbumStore interface {
	GetAlbum(ctx context.Context, opts db.GetAlbumOpts) (*models.Album, error)
}

type TrackStore interface {
	GetTrack(ctx context.Context, opts db.GetTrackOpts) (*models.Track, error)
}

type LoginStore interface {
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	SaveSession(ctx context.Context, userId int32, expiresAt time.Time, persistent bool) (*models.Session, error)
}

type SessionStore interface {
	DeleteSession(ctx context.Context, sessionId uuid.UUID) error
	RefreshSession(ctx context.Context, sessionId uuid.UUID, expiresAt time.Time) error
}

type UserUpdater interface {
	UpdateUser(ctx context.Context, opts db.UpdateUserOpts) error
}
