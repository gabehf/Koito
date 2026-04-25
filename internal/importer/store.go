package importer

import "github.com/gabehf/koito/internal/db"

// importStore is the minimal interface required by all importer functions.
type importStore interface {
	db.ArtistStore
	db.AlbumStore
	db.TrackStore
	db.ListenStore
}
