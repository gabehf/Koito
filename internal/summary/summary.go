package summary

import (
	"context"
	"fmt"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/models"
)

type Summary struct {
	Title            string                          `json:"title,omitempty"`
	TopArtists       []db.RankedItem[*models.Artist] `json:"top_artists"` // ListenCount and TimeListened are overriden with stats from timeframe
	TopAlbums        []db.RankedItem[*models.Album]  `json:"top_albums"`  // ListenCount and TimeListened are overriden with stats from timeframe
	TopTracks        []db.RankedItem[*models.Track]  `json:"top_tracks"`  // ListenCount and TimeListened are overriden with stats from timeframe
	MinutesListened  int                             `json:"minutes_listened"`
	AvgMinutesPerDay int                             `json:"avg_minutes_listened_per_day"`
	Plays            int                             `json:"plays"`
	AvgPlaysPerDay   float32                         `json:"avg_plays_per_day"`
	UniqueTracks     int                             `json:"unique_tracks"`
	UniqueAlbums     int                             `json:"unique_albums"`
	UniqueArtists    int                             `json:"unique_artists"`
	NewTracks        int                             `json:"new_tracks"`
	NewAlbums        int                             `json:"new_albums"`
	NewArtists       int                             `json:"new_artists"`
}

func GenerateSummary(ctx context.Context, store db.DB, userId int32, timeframe db.Timeframe, title string) (summary *Summary, err error) {
	// l := logger.FromContext(ctx)

	summary = new(Summary)

	topArtists, err := store.GetTopArtistsPaginated(ctx, db.GetItemsOpts{Page: 1, Limit: 5, Timeframe: timeframe})
	if err != nil {
		return nil, fmt.Errorf("GenerateSummary: %w", err)
	}
	summary.TopArtists = topArtists.Items
	// replace ListenCount and TimeListened with stats from timeframe
	for i, artist := range summary.TopArtists {
		timelistened, err := store.CountTimeListenedToItem(ctx, db.TimeListenedOpts{ArtistID: artist.Item.ID, Timeframe: timeframe})
		if err != nil {
			return nil, fmt.Errorf("GenerateSummary: %w", err)
		}
		listens, err := store.CountListensToItem(ctx, db.TimeListenedOpts{ArtistID: artist.Item.ID, Timeframe: timeframe})
		if err != nil {
			return nil, fmt.Errorf("GenerateSummary: %w", err)
		}
		summary.TopArtists[i].Item.TimeListened = timelistened
		summary.TopArtists[i].Item.ListenCount = listens
	}

	topAlbums, err := store.GetTopAlbumsPaginated(ctx, db.GetItemsOpts{Page: 1, Limit: 5, Timeframe: timeframe})
	if err != nil {
		return nil, fmt.Errorf("GenerateSummary: %w", err)
	}
	summary.TopAlbums = topAlbums.Items
	// replace ListenCount and TimeListened with stats from timeframe
	for i, album := range summary.TopAlbums {
		timelistened, err := store.CountTimeListenedToItem(ctx, db.TimeListenedOpts{AlbumID: album.Item.ID, Timeframe: timeframe})
		if err != nil {
			return nil, fmt.Errorf("GenerateSummary: %w", err)
		}
		listens, err := store.CountListensToItem(ctx, db.TimeListenedOpts{AlbumID: album.Item.ID, Timeframe: timeframe})
		if err != nil {
			return nil, fmt.Errorf("GenerateSummary: %w", err)
		}
		summary.TopAlbums[i].Item.TimeListened = timelistened
		summary.TopAlbums[i].Item.ListenCount = listens
	}

	topTracks, err := store.GetTopTracksPaginated(ctx, db.GetItemsOpts{Page: 1, Limit: 5, Timeframe: timeframe})
	if err != nil {
		return nil, fmt.Errorf("GenerateSummary: %w", err)
	}
	summary.TopTracks = topTracks.Items
	// replace ListenCount and TimeListened with stats from timeframe
	for i, track := range summary.TopTracks {
		timelistened, err := store.CountTimeListenedToItem(ctx, db.TimeListenedOpts{TrackID: track.Item.ID, Timeframe: timeframe})
		if err != nil {
			return nil, fmt.Errorf("GenerateSummary: %w", err)
		}
		listens, err := store.CountListensToItem(ctx, db.TimeListenedOpts{TrackID: track.Item.ID, Timeframe: timeframe})
		if err != nil {
			return nil, fmt.Errorf("GenerateSummary: %w", err)
		}
		summary.TopTracks[i].Item.TimeListened = timelistened
		summary.TopTracks[i].Item.ListenCount = listens
	}

	t1, t2 := db.TimeframeToTimeRange(timeframe)
	daycount := int(t2.Sub(t1).Hours() / 24)
	// bandaid
	if daycount == 0 {
		daycount = 1
	}

	tmp, err := store.CountTimeListened(ctx, timeframe)
	if err != nil {
		return nil, fmt.Errorf("GenerateSummary: %w", err)
	}
	summary.MinutesListened = int(tmp) / 60
	summary.AvgMinutesPerDay = summary.MinutesListened / daycount
	tmp, err = store.CountListens(ctx, timeframe)
	if err != nil {
		return nil, fmt.Errorf("GenerateSummary: %w", err)
	}
	summary.Plays = int(tmp)
	summary.AvgPlaysPerDay = float32(summary.Plays) / float32(daycount)
	tmp, err = store.CountTracks(ctx, timeframe)
	if err != nil {
		return nil, fmt.Errorf("GenerateSummary: %w", err)
	}
	summary.UniqueTracks = int(tmp)
	tmp, err = store.CountAlbums(ctx, timeframe)
	if err != nil {
		return nil, fmt.Errorf("GenerateSummary: %w", err)
	}
	summary.UniqueAlbums = int(tmp)
	tmp, err = store.CountArtists(ctx, timeframe)
	if err != nil {
		return nil, fmt.Errorf("GenerateSummary: %w", err)
	}
	summary.UniqueArtists = int(tmp)
	tmp, err = store.CountNewTracks(ctx, timeframe)
	if err != nil {
		return nil, fmt.Errorf("GenerateSummary: %w", err)
	}
	summary.NewTracks = int(tmp)
	tmp, err = store.CountNewAlbums(ctx, timeframe)
	if err != nil {
		return nil, fmt.Errorf("GenerateSummary: %w", err)
	}
	summary.NewAlbums = int(tmp)
	tmp, err = store.CountNewArtists(ctx, timeframe)
	if err != nil {
		return nil, fmt.Errorf("GenerateSummary: %w", err)
	}
	summary.NewArtists = int(tmp)

	return summary, nil
}
