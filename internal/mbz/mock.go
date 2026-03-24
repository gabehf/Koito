package mbz

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"
)

// implements a mock caller

type MbzMockCaller struct {
	Artists       map[uuid.UUID]*MusicBrainzArtist
	ReleaseGroups map[uuid.UUID]*MusicBrainzReleaseGroup
	Releases      map[uuid.UUID]*MusicBrainzRelease
	Tracks        map[uuid.UUID]*MusicBrainzTrack
	SearchResults map[string]*MusicBrainzSearchResult
}

func (m *MbzMockCaller) GetReleaseGroup(ctx context.Context, id uuid.UUID) (*MusicBrainzReleaseGroup, error) {
	releaseGroup, exists := m.ReleaseGroups[id]
	if !exists {
		return nil, fmt.Errorf("release group with ID %s not found", id)
	}
	return releaseGroup, nil
}

func (m *MbzMockCaller) GetRelease(ctx context.Context, id uuid.UUID) (*MusicBrainzRelease, error) {
	release, exists := m.Releases[id]
	if !exists {
		return nil, fmt.Errorf("release group with ID %s not found", id)
	}
	return release, nil
}

func (m *MbzMockCaller) GetReleaseTitles(ctx context.Context, RGID uuid.UUID) ([]string, error) {
	rg, exists := m.ReleaseGroups[RGID]
	if !exists {
		return nil, fmt.Errorf("release with ID %s not found", RGID)
	}

	var titles []string
	for _, release := range rg.Releases {
		if !slices.Contains(titles, release.Title) {
			titles = append(titles, release.Title)
		}
	}
	return titles, nil
}

func (m *MbzMockCaller) GetTrack(ctx context.Context, id uuid.UUID) (*MusicBrainzTrack, error) {
	track, exists := m.Tracks[id]
	if !exists {
		return nil, fmt.Errorf("track with ID %s not found", id)
	}
	return track, nil
}

func (m *MbzMockCaller) GetArtistPrimaryAliases(ctx context.Context, id uuid.UUID) ([]string, error) {
	artist, exists := m.Artists[id]
	if !exists {
		return nil, fmt.Errorf("artist with ID %s not found", id)
	}
	name := artist.Name
	ss := make([]string, len(artist.Aliases)+1)
	ss[0] = name
	for i, alias := range artist.Aliases {
		ss[i+1] = alias.Name
	}
	return ss, nil
}

func (m *MbzMockCaller) SearchRecording(ctx context.Context, artist string, track string) (*MusicBrainzSearchResult, error) {
	key := artist + "\x00" + track
	if result, exists := m.SearchResults[key]; exists {
		return result, nil
	}
	return nil, nil
}

func (m *MbzMockCaller) Shutdown() {}

type MbzErrorCaller struct{}

func (m *MbzErrorCaller) GetReleaseGroup(ctx context.Context, id uuid.UUID) (*MusicBrainzReleaseGroup, error) {
	return nil, fmt.Errorf("error: GetReleaseGroup not implemented")
}

func (m *MbzErrorCaller) GetRelease(ctx context.Context, id uuid.UUID) (*MusicBrainzRelease, error) {
	return nil, fmt.Errorf("error: GetRelease not implemented")
}

func (m *MbzErrorCaller) GetReleaseTitles(ctx context.Context, RGID uuid.UUID) ([]string, error) {
	return nil, fmt.Errorf("error: GetReleaseTitles not implemented")
}

func (m *MbzErrorCaller) GetTrack(ctx context.Context, id uuid.UUID) (*MusicBrainzTrack, error) {
	return nil, fmt.Errorf("error: GetTrack not implemented")
}

func (m *MbzErrorCaller) GetArtistPrimaryAliases(ctx context.Context, id uuid.UUID) ([]string, error) {
	return nil, fmt.Errorf("error: GetArtistPrimaryAliases not implemented")
}

func (m *MbzErrorCaller) SearchRecording(ctx context.Context, artist string, track string) (*MusicBrainzSearchResult, error) {
	return nil, fmt.Errorf("error: SearchRecording not implemented")
}

func (m *MbzErrorCaller) Shutdown() {}
