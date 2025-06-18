// package db defines the database interface
package db

import (
	"context"
	"time"

	"github.com/gabehf/koito/internal/models"
	"github.com/google/uuid"
)

type DB interface {
	// Get
	GetArtist(ctx context.Context, opts GetArtistOpts) (*models.Artist, error)
	GetAlbum(ctx context.Context, opts GetAlbumOpts) (*models.Album, error)
	GetTrack(ctx context.Context, opts GetTrackOpts) (*models.Track, error)
	GetArtistsForAlbum(ctx context.Context, id int32) ([]*models.Artist, error)
	GetArtistsForTrack(ctx context.Context, id int32) ([]*models.Artist, error)
	GetTopTracksPaginated(ctx context.Context, opts GetItemsOpts) (*PaginatedResponse[*models.Track], error)
	GetTopArtistsPaginated(ctx context.Context, opts GetItemsOpts) (*PaginatedResponse[*models.Artist], error)
	GetTopAlbumsPaginated(ctx context.Context, opts GetItemsOpts) (*PaginatedResponse[*models.Album], error)
	GetListensPaginated(ctx context.Context, opts GetItemsOpts) (*PaginatedResponse[*models.Listen], error)
	GetListenActivity(ctx context.Context, opts ListenActivityOpts) ([]ListenActivityItem, error)
	GetAllArtistAliases(ctx context.Context, id int32) ([]models.Alias, error)
	GetAllAlbumAliases(ctx context.Context, id int32) ([]models.Alias, error)
	GetAllTrackAliases(ctx context.Context, id int32) ([]models.Alias, error)
	GetApiKeysByUserID(ctx context.Context, id int32) ([]models.ApiKey, error)
	GetUserBySession(ctx context.Context, sessionId uuid.UUID) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByApiKey(ctx context.Context, key string) (*models.User, error)
	// Save
	SaveArtist(ctx context.Context, opts SaveArtistOpts) (*models.Artist, error)
	SaveArtistAliases(ctx context.Context, id int32, aliases []string, source string) error
	SaveAlbum(ctx context.Context, opts SaveAlbumOpts) (*models.Album, error)
	SaveAlbumAliases(ctx context.Context, id int32, aliases []string, source string) error
	SaveTrack(ctx context.Context, opts SaveTrackOpts) (*models.Track, error)
	SaveTrackAliases(ctx context.Context, id int32, aliases []string, source string) error
	SaveListen(ctx context.Context, opts SaveListenOpts) error
	SaveUser(ctx context.Context, opts SaveUserOpts) (*models.User, error)
	SaveApiKey(ctx context.Context, opts SaveApiKeyOpts) (*models.ApiKey, error)
	SaveSession(ctx context.Context, userId int32, expiresAt time.Time, persistent bool) (*models.Session, error)
	// Update
	UpdateArtist(ctx context.Context, opts UpdateArtistOpts) error
	UpdateTrack(ctx context.Context, opts UpdateTrackOpts) error
	UpdateAlbum(ctx context.Context, opts UpdateAlbumOpts) error
	AddArtistsToAlbum(ctx context.Context, opts AddArtistsToAlbumOpts) error
	UpdateUser(ctx context.Context, opts UpdateUserOpts) error
	UpdateApiKeyLabel(ctx context.Context, opts UpdateApiKeyLabelOpts) error
	RefreshSession(ctx context.Context, sessionId uuid.UUID, expiresAt time.Time) error
	SetPrimaryArtistAlias(ctx context.Context, id int32, alias string) error
	SetPrimaryAlbumAlias(ctx context.Context, id int32, alias string) error
	SetPrimaryTrackAlias(ctx context.Context, id int32, alias string) error
	SetPrimaryAlbumArtist(ctx context.Context, id int32, artistId int32, value bool) error
	SetPrimaryTrackArtist(ctx context.Context, id int32, artistId int32, value bool) error
	// Delete
	DeleteArtist(ctx context.Context, id int32) error
	DeleteAlbum(ctx context.Context, id int32) error
	DeleteTrack(ctx context.Context, id int32) error
	DeleteListen(ctx context.Context, trackId int32, listenedAt time.Time) error
	DeleteArtistAlias(ctx context.Context, id int32, alias string) error
	DeleteAlbumAlias(ctx context.Context, id int32, alias string) error
	DeleteTrackAlias(ctx context.Context, id int32, alias string) error
	DeleteSession(ctx context.Context, sessionId uuid.UUID) error
	DeleteApiKey(ctx context.Context, id int32) error
	// Count
	CountListens(ctx context.Context, period Period) (int64, error)
	CountTracks(ctx context.Context, period Period) (int64, error)
	CountAlbums(ctx context.Context, period Period) (int64, error)
	CountArtists(ctx context.Context, period Period) (int64, error)
	CountTimeListened(ctx context.Context, period Period) (int64, error)
	CountTimeListenedToItem(ctx context.Context, opts TimeListenedOpts) (int64, error)
	CountUsers(ctx context.Context) (int64, error)
	// Search
	SearchArtists(ctx context.Context, q string) ([]*models.Artist, error)
	SearchAlbums(ctx context.Context, q string) ([]*models.Album, error)
	SearchTracks(ctx context.Context, q string) ([]*models.Track, error)
	// Merge
	MergeTracks(ctx context.Context, fromId, toId int32) error
	MergeAlbums(ctx context.Context, fromId, toId int32, replaceImage bool) error
	MergeArtists(ctx context.Context, fromId, toId int32, replaceImage bool) error
	// Etc
	ImageHasAssociation(ctx context.Context, image uuid.UUID) (bool, error)
	GetImageSource(ctx context.Context, image uuid.UUID) (string, error)
	AlbumsWithoutImages(ctx context.Context, from int32) ([]*models.Album, error)
	GetExportPage(ctx context.Context, opts GetExportPageOpts) ([]*ExportItem, error)
	Ping(ctx context.Context) error
	Close(ctx context.Context)
}
