package catalog

import "strings"

// TrackLookupKey builds a normalized cache key for entity resolution.
// Uses null-byte separators to avoid collisions between field values.
func TrackLookupKey(artist, track, album string) string {
	return strings.ToLower(artist) + "\x00" + strings.ToLower(track) + "\x00" + strings.ToLower(album)
}
