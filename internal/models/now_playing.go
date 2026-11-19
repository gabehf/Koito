package models

import "time"

type NowPlayingData struct {
	TrackID   int32
	StartTime time.Time
}
