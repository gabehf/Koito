package models

import "github.com/google/uuid"

type Album struct {
	ID             int32          `json:"id"`
	MbzID          *uuid.UUID     `json:"musicbrainz_id"`
	Title          string         `json:"title"`
	Image          *uuid.UUID     `json:"image"`
	Artists        []SimpleArtist `json:"artists"`
	VariousArtists bool           `json:"is_various_artists"`
	ListenCount    int64          `json:"listen_count"`
	TimeListened   int64          `json:"time_listened"`
	FirstListen    int64          `json:"first_listen"`
}

// type SimpleAlbum struct {
// 	ID             int32     `json:"id"`
// 	Title          string    `json:"title"`
// 	VariousArtists bool      `json:"is_various_artists"`
// 	Image          uuid.UUID `json:"image"`
// }
