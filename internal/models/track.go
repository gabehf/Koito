package models

import "github.com/google/uuid"

type Track struct {
	ID           int32          `json:"id"`
	Title        string         `json:"title"`
	Artists      []SimpleArtist `json:"artists"`
	MbzID        *uuid.UUID     `json:"musicbrainz_id"`
	ListenCount  int64          `json:"listen_count"`
	Duration     int32          `json:"duration"`
	Image        *uuid.UUID     `json:"image"`
	AlbumID      int32          `json:"album_id"`
	TimeListened int64          `json:"time_listened"`
	FirstListen  int64          `json:"first_listen"`
}
