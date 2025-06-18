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
