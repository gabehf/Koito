package mbz

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type MusicBrainzTrack struct {
	Title    string `json:"title"`
	LengthMs int    `json:"length"`
}

const recordingFmtStr = "%s/ws/2/recording/%s"

// Returns the artist name at index 0, and all primary aliases after.
func (c *MusicBrainzClient) GetTrack(ctx context.Context, id uuid.UUID) (*MusicBrainzTrack, error) {
	track := new(MusicBrainzTrack)
	err := c.getEntity(ctx, recordingFmtStr, id, track)
	if err != nil {
		return nil, fmt.Errorf("GetTrack: %w", err)
	}
	return track, nil
}
