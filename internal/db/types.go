package db

import (
	"time"

	"github.com/gabehf/koito/internal/models"
	"github.com/google/uuid"
)

type InformationSource string

const (
	InformationSourceInferred     InformationSource = "Inferred"
	InformationSourceMusicBrainz  InformationSource = "MusicBrainz"
	InformationSourceUserProvided InformationSource = "User"
)

type ListenActivityItem struct {
	Start   time.Time `json:"start_time"`
	Listens int64     `json:"listens"`
}

type PaginatedResponse[T any] struct {
	Items        []T   `json:"items"`
	TotalCount   int64 `json:"total_record_count"`
	ItemsPerPage int32 `json:"items_per_page"`
	HasNextPage  bool  `json:"has_next_page"`
	CurrentPage  int32 `json:"current_page"`
}

type RankedItem[T any] struct {
	Item T     `json:"item"`
	Rank int64 `json:"rank"`
}

type ExportItem struct {
	ListenedAt         time.Time
	UserID             int32
	Client             *string
	TrackID            int32
	TrackMbid          *uuid.UUID
	TrackDuration      int32
	TrackAliases       []models.Alias
	ReleaseID          int32
	ReleaseMbid        *uuid.UUID
	ReleaseImage       *uuid.UUID
	ReleaseImageSource string
	VariousArtists     bool
	ReleaseAliases     []models.Alias
	Artists            []models.ArtistWithFullAliases
}

type InterestBucket struct {
	BucketStart time.Time `json:"bucket_start"`
	BucketEnd   time.Time `json:"bucket_end"`
	ListenCount int64     `json:"listen_count"`
}

// TrackDetail is the metadata of a track, its release and its artists, as needed
// to render a track in an external API.
type TrackDetail struct {
	TrackID      int32
	TrackTitle   string
	TrackMbid    *uuid.UUID
	Duration     int32
	ReleaseTitle string
	ReleaseMbid  *uuid.UUID
	ArtistNames  []string
	ArtistMbids  []uuid.UUID
}

// ListenDetail is a listen with the full detail of the listened track.
type ListenDetail struct {
	ListenedAt time.Time
	Client     string
	Track      TrackDetail
}

// ListenRangeOpts selects listens strictly between From and To (unix seconds).
// A zero From or To leaves that side unbounded. Results are newest first.
type ListenRangeOpts struct {
	From  int64
	To    int64
	Limit int
}

// TrackFeedback is a love (1) or hate (-1) a user gave to a track.
type TrackFeedback struct {
	CreatedAt time.Time
	Score     int
	Track     TrackDetail
}

// SaveTrackFeedbackOpts sets a user's feedback on a track. Score 0 removes it.
type SaveTrackFeedbackOpts struct {
	UserID  int32
	TrackID int32
	Score   int
}
