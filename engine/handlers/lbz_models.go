package handlers

import (
	"strings"
	"time"

	"github.com/gabehf/koito/internal/db"
	"github.com/google/uuid"
)

// Response shapes for the ListenBrainz-compatible GET endpoints. Every field is
// emitted, even when null, because clients (e.g. pano-scrobbler) decode them as
// required fields.

type LbzOutMbidMapping struct {
	ArtistMbids   []string `json:"artist_mbids"`
	RecordingMbid *string  `json:"recording_mbid"`
	ReleaseMbid   *string  `json:"release_mbid"`
}

type LbzListenAdditionalInfo struct {
	ArtistMbids             []string `json:"artist_mbids"`
	DurationMs              *int64   `json:"duration_ms"`
	RecordingMbid           *string  `json:"recording_mbid"`
	ReleaseMbid             *string  `json:"release_mbid"`
	ReleaseGroupMbid        *string  `json:"release_group_mbid"`
	SubmissionClient        *string  `json:"submission_client"`
	SubmissionClientVersion *string  `json:"submission_client_version"`
	MediaPlayer             *string  `json:"media_player"`
}

type LbzTrackMetadata struct {
	ArtistName     string                  `json:"artist_name"`
	TrackName      string                  `json:"track_name"`
	ReleaseName    *string                 `json:"release_name"`
	AdditionalInfo LbzListenAdditionalInfo `json:"additional_info"`
	MbidMapping    LbzOutMbidMapping       `json:"mbid_mapping"`
}

type LbzListen struct {
	InsertedAt    *int64           `json:"inserted_at"`
	ListenedAt    *int64           `json:"listened_at"`
	RecordingMsid *string          `json:"recording_msid"`
	PlayingNow    *bool            `json:"playing_now"`
	UserName      string           `json:"user_name,omitempty"`
	TrackMetadata LbzTrackMetadata `json:"track_metadata"`
}

type LbzListensPayload struct {
	Count          int         `json:"count"`
	Listens        []LbzListen `json:"listens"`
	UserID         string      `json:"user_id,omitempty"`
	LatestListenTs *int64      `json:"latest_listen_ts"`
	OldestListenTs *int64      `json:"oldest_listen_ts"`
}

type lbzListensResponse struct {
	Payload LbzListensPayload `json:"payload"`
}

type lbzPlayingNowPayload struct {
	Count          int         `json:"count"`
	Listens        []LbzListen `json:"listens"`
	PlayingNow     bool        `json:"playing_now"`
	UserID         string      `json:"user_id"`
	LatestListenTs *int64      `json:"latest_listen_ts"`
	OldestListenTs *int64      `json:"oldest_listen_ts"`
}

type lbzPlayingNowResponse struct {
	Payload lbzPlayingNowPayload `json:"payload"`
}

type lbzListenCountResponse struct {
	Payload struct {
		Count int64 `json:"count"`
	} `json:"payload"`
}

type lbzFollowingResponse struct {
	Following []string `json:"following"`
	User      string   `json:"user"`
}

type LbzStatsEntry struct {
	ArtistMbids   []string `json:"artist_mbids"`
	ArtistName    string   `json:"artist_name"`
	RecordingMbid *string  `json:"recording_mbid"`
	ReleaseMbid   *string  `json:"release_mbid"`
	ReleaseName   *string  `json:"release_name"`
	TrackName     *string  `json:"track_name"`
	ListenCount   int64    `json:"listen_count"`
}

type LbzStatsPayload struct {
	Artists             []LbzStatsEntry `json:"artists,omitempty"`
	Releases            []LbzStatsEntry `json:"releases,omitempty"`
	Recordings          []LbzStatsEntry `json:"recordings,omitempty"`
	Count               int             `json:"count"`
	Offset              int             `json:"offset"`
	Range               string          `json:"range"`
	FromTs              int64           `json:"from_ts"`
	ToTs                int64           `json:"to_ts"`
	LastUpdated         int64           `json:"last_updated"`
	TotalArtistCount    *int64          `json:"total_artist_count,omitempty"`
	TotalReleaseCount   *int64          `json:"total_release_count,omitempty"`
	TotalRecordingCount *int64          `json:"total_recording_count,omitempty"`
}

type LbzListeningActivity struct {
	FromTs      int64  `json:"from_ts"`
	ToTs        int64  `json:"to_ts"`
	ListenCount int64  `json:"listen_count"`
	TimeRange   string `json:"time_range"`
}

type LbzFeedbackPayload struct {
	Created       int64             `json:"created"`
	RecordingMbid *string           `json:"recording_mbid"`
	RecordingMsid *string           `json:"recording_msid"`
	Score         int               `json:"score"`
	TrackMetadata *LbzTrackMetadata `json:"track_metadata"`
}

// lbzTrackMetadataFromDetail builds the ListenBrainz track_metadata for a track.
func lbzTrackMetadataFromDetail(d *db.TrackDetail) LbzTrackMetadata {
	artistName := strings.Join(d.ArtistNames, ", ")
	releaseName := nonEmptyPtr(d.ReleaseTitle)

	artistMbids := uuidsToStrings(d.ArtistMbids)
	var durationMs *int64
	if d.Duration > 0 {
		ms := int64(d.Duration) * 1000
		durationMs = &ms
	}
	releaseMbid := uuidPtrToString(d.ReleaseMbid)
	recordingMbid := uuidPtrToString(d.TrackMbid)

	return LbzTrackMetadata{
		ArtistName:  artistName,
		TrackName:   d.TrackTitle,
		ReleaseName: releaseName,
		AdditionalInfo: LbzListenAdditionalInfo{
			ArtistMbids:   artistMbids,
			DurationMs:    durationMs,
			RecordingMbid: recordingMbid,
			ReleaseMbid:   releaseMbid,
		},
		MbidMapping: LbzOutMbidMapping{
			ArtistMbids:   artistMbids,
			RecordingMbid: recordingMbid,
			ReleaseMbid:   releaseMbid,
		},
	}
}

func uuidsToStrings(ids []uuid.UUID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	return out
}

func uuidPtrToString(id *uuid.UUID) *string {
	if id == nil || *id == uuid.Nil {
		return nil
	}
	s := id.String()
	return &s
}

func unixPtr(t time.Time) *int64 {
	u := t.Unix()
	return &u
}
