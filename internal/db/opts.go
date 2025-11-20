package db

import (
	"time"

	"github.com/gabehf/koito/internal/models"
	"github.com/google/uuid"
)

type GetAlbumOpts struct {
	ID            int32
	MusicBrainzID uuid.UUID
	ArtistID      int32
	Title         string
	Titles        []string
	Image         uuid.UUID
}

type GetArtistOpts struct {
	ID            int32
	MusicBrainzID uuid.UUID
	Name          string
	Image         uuid.UUID
}

type GetTrackOpts struct {
	ID            int32
	MusicBrainzID uuid.UUID
	Title         string
	ArtistIDs     []int32
}

type SaveTrackOpts struct {
	Title          string
	AlbumID        int32
	ArtistIDs      []int32
	RecordingMbzID uuid.UUID
	Duration       int32
}

type SaveAlbumOpts struct {
	Title          string
	MusicBrainzID  uuid.UUID
	Type           string
	ArtistIDs      []int32
	VariousArtists bool
	Image          uuid.UUID
	ImageSrc       string
	Aliases        []string
}

type SaveArtistOpts struct {
	Name          string
	MusicBrainzID uuid.UUID
	Aliases       []string
	Image         uuid.UUID
	ImageSrc      string
}

type UpdateApiKeyLabelOpts struct {
	UserID int32
	ID     int32
	Label  string
}

type SaveUserOpts struct {
	Username string
	Password string
	Role     models.UserRole
}

type SaveApiKeyOpts struct {
	Key    string
	UserID int32
	Label  string
}

type SaveListenOpts struct {
	TrackID int32
	Time    time.Time
	UserID  int32
	Client  string
}

type UpdateTrackOpts struct {
	ID            int32
	MusicBrainzID uuid.UUID
	Duration      int32
}

type UpdateArtistOpts struct {
	ID            int32
	MusicBrainzID uuid.UUID
	Image         uuid.UUID
	ImageSrc      string
}

type UpdateAlbumOpts struct {
	ID                   int32
	MusicBrainzID        uuid.UUID
	Image                uuid.UUID
	ImageSrc             string
	VariousArtistsUpdate bool
	VariousArtistsValue  bool
}

type UpdateUserOpts struct {
	ID       int32
	Username string
	Password string
}

type AddArtistsToAlbumOpts struct {
	AlbumID   int32
	ArtistIDs []int32
}

type GetItemsOpts struct {
	Limit  int
	Period Period
	Page   int
	Week   int // 1-52
	Month  int // 1-12
	Year   int
	From   int // unix timestamp
	To     int // unix timestamp

	// Used only for getting top tracks
	ArtistID int
	AlbumID  int

	// Used for getting listens
	TrackID int
}

type ListenActivityOpts struct {
	Step     StepInterval
	Range    int
	Month    int
	Year     int
	AlbumID  int32
	ArtistID int32
	TrackID  int32
}

type TimeListenedOpts struct {
	Period   Period
	AlbumID  int32
	ArtistID int32
	TrackID  int32
}

type GetExportPageOpts struct {
	UserID     int32
	ListenedAt time.Time
	TrackID    int32
	Limit      int32
}
