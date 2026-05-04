// package db defines the database interface
package db

import (
	"context"
	"time"

	"github.com/gabehf/koito/internal/models"
	"github.com/google/uuid"
)

type ArtistStore interface {
	GetArtist(ctx context.Context, opts GetArtistOpts) (*models.Artist, error)
	GetArtistsForAlbum(ctx context.Context, id int32) ([]*models.Artist, error)
	GetArtistsForTrack(ctx context.Context, id int32) ([]*models.Artist, error)
	GetTopArtistsPaginated(ctx context.Context, opts GetItemsOpts) (*PaginatedResponse[RankedItem[*models.Artist]], error)
	GetAllArtistAliases(ctx context.Context, id int32) ([]models.Alias, error)
	SaveArtist(ctx context.Context, opts SaveArtistOpts) (*models.Artist, error)
	SaveArtistAliases(ctx context.Context, id int32, aliases []string, source string) error
	UpdateArtist(ctx context.Context, opts UpdateArtistOpts) error
	SetPrimaryArtistAlias(ctx context.Context, id int32, alias string) error
	SetPrimaryAlbumArtist(ctx context.Context, id int32, artistId int32, value bool) error
	SetPrimaryTrackArtist(ctx context.Context, id int32, artistId int32, value bool) error
	DeleteArtist(ctx context.Context, id int32) error
	DeleteArtistAlias(ctx context.Context, id int32, alias string) error
	MergeArtists(ctx context.Context, fromId, toId int32, replaceImage bool) error
	SearchArtists(ctx context.Context, q string) ([]*models.Artist, error)
	CountArtists(ctx context.Context, timeframe Timeframe) (int64, error)
	CountNewArtists(ctx context.Context, timeframe Timeframe) (int64, error)
	ArtistsWithoutImages(ctx context.Context, from int32) ([]*models.Artist, error)
}

type AlbumStore interface {
	GetAlbum(ctx context.Context, opts GetAlbumOpts) (*models.Album, error)
	GetAlbumWithNoMbzIDByTitles(ctx context.Context, artistId int32, titles []string) (*models.Album, error)
	GetTopAlbumsPaginated(ctx context.Context, opts GetItemsOpts) (*PaginatedResponse[RankedItem[*models.Album]], error)
	GetAllAlbumAliases(ctx context.Context, id int32) ([]models.Alias, error)
	SaveAlbum(ctx context.Context, opts SaveAlbumOpts) (*models.Album, error)
	SaveAlbumAliases(ctx context.Context, id int32, aliases []string, source string) error
	UpdateAlbum(ctx context.Context, opts UpdateAlbumOpts) error
	SetPrimaryAlbumAlias(ctx context.Context, id int32, alias string) error
	AddArtistsToAlbum(ctx context.Context, opts AddArtistsToAlbumOpts) error
	DeleteAlbum(ctx context.Context, id int32) error
	DeleteAlbumAlias(ctx context.Context, id int32, alias string) error
	MergeAlbums(ctx context.Context, fromId, toId int32, replaceImage bool) error
	SearchAlbums(ctx context.Context, q string) ([]*models.Album, error)
	CountAlbums(ctx context.Context, timeframe Timeframe) (int64, error)
	CountNewAlbums(ctx context.Context, timeframe Timeframe) (int64, error)
	AlbumsWithoutImages(ctx context.Context, from int32) ([]*models.Album, error)
}

type TrackStore interface {
	GetTrack(ctx context.Context, opts GetTrackOpts) (*models.Track, error)
	GetTracksWithNoDurationButHaveMbzID(ctx context.Context, from int32) ([]*models.Track, error)
	GetTopTracksPaginated(ctx context.Context, opts GetItemsOpts) (*PaginatedResponse[RankedItem[*models.Track]], error)
	GetAllTrackAliases(ctx context.Context, id int32) ([]models.Alias, error)
	SaveTrack(ctx context.Context, opts SaveTrackOpts) (*models.Track, error)
	SaveTrackAliases(ctx context.Context, id int32, aliases []string, source string) error
	UpdateTrack(ctx context.Context, opts UpdateTrackOpts) error
	SetPrimaryTrackAlias(ctx context.Context, id int32, alias string) error
	DeleteTrack(ctx context.Context, id int32) error
	DeleteTrackAlias(ctx context.Context, id int32, alias string) error
	MergeTracks(ctx context.Context, fromId, toId int32) error
	SearchTracks(ctx context.Context, q string) ([]*models.Track, error)
	CountTracks(ctx context.Context, timeframe Timeframe) (int64, error)
	CountNewTracks(ctx context.Context, timeframe Timeframe) (int64, error)
	AddArtistsToAlbum(ctx context.Context, opts AddArtistsToAlbumOpts) error
}

type ListenStore interface {
	GetListensPaginated(ctx context.Context, opts GetItemsOpts) (*PaginatedResponse[*models.Listen], error)
	GetListenActivity(ctx context.Context, opts ListenActivityOpts) ([]ListenActivityItem, error)
	GetInterest(ctx context.Context, opts GetInterestOpts) ([]InterestBucket, error)
	SaveListen(ctx context.Context, opts SaveListenOpts) error
	DeleteListen(ctx context.Context, trackId int32, listenedAt time.Time) error
	CountListens(ctx context.Context, timeframe Timeframe) (int64, error)
	CountListensToItem(ctx context.Context, opts TimeListenedOpts) (int64, error)
	CountTimeListened(ctx context.Context, timeframe Timeframe) (int64, error)
	CountTimeListenedToItem(ctx context.Context, opts TimeListenedOpts) (int64, error)
}

type UserStore interface {
	GetUserBySession(ctx context.Context, sessionId uuid.UUID) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByApiKey(ctx context.Context, key string) (*models.User, error)
	GetAdminUser(ctx context.Context) (*models.User, error)
	GetApiKeysByUserID(ctx context.Context, id int32) ([]models.ApiKey, error)
	SaveUser(ctx context.Context, opts SaveUserOpts) (*models.User, error)
	SaveApiKey(ctx context.Context, opts SaveApiKeyOpts) (*models.ApiKey, error)
	SaveSession(ctx context.Context, userId int32, expiresAt time.Time, persistent bool) (*models.Session, error)
	UpdateUser(ctx context.Context, opts UpdateUserOpts) error
	UpdateApiKeyLabel(ctx context.Context, opts UpdateApiKeyLabelOpts) error
	RefreshSession(ctx context.Context, sessionId uuid.UUID, expiresAt time.Time) error
	DeleteSession(ctx context.Context, sessionId uuid.UUID) error
	DeleteApiKey(ctx context.Context, id int32) error
	CountUsers(ctx context.Context) (int64, error)
}

type ImageStore interface {
	ImageHasAssociation(ctx context.Context, image uuid.UUID) (bool, error)
	GetImageSource(ctx context.Context, image uuid.UUID) (string, error)
}

type ExportStore interface {
	GetExportPage(ctx context.Context, opts GetExportPageOpts) ([]*ExportItem, error)
}

type DB interface {
	ArtistStore
	AlbumStore
	TrackStore
	ListenStore
	UserStore
	ImageStore
	ExportStore
	Ping(ctx context.Context) error
	Close(ctx context.Context)
}
