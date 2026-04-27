package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	"github.com/gabehf/koito/internal/models"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

// fuzzyScore returns a similarity score [0, 1] between query and target using
// trigram-based matching. Scores below ~0.2 indicate poor matches.
func fuzzyScore(query, target string) float64 {
	if query == target {
		return 1.0
	}
	// fuzzy.RankMatchNormalized returns 0 if no match, positive float for match quality
	rank := fuzzy.RankMatchNormalized(query, target)
	if rank < 0 {
		return 0
	}
	return float64(rank) / float64(len(target)+1)
}

// sliceTopN deduplicates a candidate slice by highest score and returns the top N.
// The score function extracts the comparable score from each candidate.
func sliceTopN[T any](candidates []T, scoreOf func(T) float64, n int) []T {
	sort.Slice(candidates, func(i, j int) bool {
		return scoreOf(candidates[i]) > scoreOf(candidates[j])
	})
	if len(candidates) > n {
		return candidates[:n]
	}
	return candidates
}

func (s *Sqlite) SearchArtists(ctx context.Context, q string) ([]*models.Artist, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT DISTINCT a.id, awn.name, a.musicbrainz_id, a.image
		FROM artist_aliases aa
		JOIN artists_with_name awn ON aa.artist_id = awn.id
		JOIN artists a ON a.id = awn.id
		WHERE aa.alias LIKE ? OR aa.alias LIKE ?
		LIMIT 50`,
		q+"%", "%"+q+"%",
	)
	if err != nil {
		return nil, fmt.Errorf("SearchArtists: %w", err)
	}
	defer rows.Close()

	type candidate struct {
		artist *models.Artist
		score  float64
	}
	seen := map[int32]float64{}
	var candidates []candidate

	for rows.Next() {
		var a models.Artist
		var mbzID, image sql.NullString
		if err := rows.Scan(&a.ID, &a.Name, &mbzID, &image); err != nil {
			return nil, err
		}
		a.MbzID = parseNullableUUID(mbzID)
		a.Image = parseNullableUUID(image)
		score := fuzzyScore(q, a.Name)
		if prev, ok := seen[a.ID]; !ok || score > prev {
			seen[a.ID] = score
			candidates = append(candidates, candidate{&a, score})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	top := sliceTopN(candidates, func(c candidate) float64 { return c.score }, 8)
	out := make([]*models.Artist, len(top))
	for i, c := range top {
		out[i] = c.artist
	}
	return out, nil
}

func (s *Sqlite) SearchAlbums(ctx context.Context, q string) ([]*models.Album, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT DISTINCT r.id, r.title, r.musicbrainz_id, r.image, r.various_artists
		FROM release_aliases ra
		JOIN releases_with_title r ON ra.release_id = r.id
		WHERE ra.alias LIKE ? OR ra.alias LIKE ?
		LIMIT 50`,
		q+"%", "%"+q+"%",
	)
	if err != nil {
		return nil, fmt.Errorf("SearchAlbums: %w", err)
	}
	defer rows.Close()

	type candidate struct {
		album *models.Album
		score float64
	}
	seen := map[int32]float64{}
	var candidates []candidate

	for rows.Next() {
		var a models.Album
		var mbzID, image sql.NullString
		var variousArtists int
		if err := rows.Scan(&a.ID, &a.Title, &mbzID, &image, &variousArtists); err != nil {
			return nil, err
		}
		a.MbzID = parseNullableUUID(mbzID)
		a.Image = parseNullableUUID(image)
		a.VariousArtists = variousArtists == 1
		score := fuzzyScore(q, a.Title)
		if prev, ok := seen[a.ID]; !ok || score > prev {
			seen[a.ID] = score
			candidates = append(candidates, candidate{&a, score})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	top := sliceTopN(candidates, func(c candidate) float64 { return c.score }, 8)
	out := make([]*models.Album, 0, len(top))
	for _, c := range top {
		artists, err := s.artistsForRelease(ctx, c.album.ID)
		if err != nil {
			return nil, err
		}
		c.album.Artists = artists
		out = append(out, c.album)
	}
	return out, nil
}
