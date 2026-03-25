package catalog

import "strings"

// TrackLookupKey builds a normalized cache key for entity resolution.
// Uses unit separator (U+001F) to avoid collisions between field values.
func TrackLookupKey(artist, track, album string) string {
	return strings.ToLower(artist) + "\x1f" + strings.ToLower(track) + "\x1f" + strings.ToLower(album)
}
