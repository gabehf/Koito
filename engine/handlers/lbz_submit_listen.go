package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gabehf/koito/engine/middleware"
	"github.com/gabehf/koito/internal/catalog"
	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/mbz"
	"github.com/gabehf/koito/internal/utils"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"golang.org/x/sync/singleflight"
)

type LbzListenType string

const (
	ListenTypeSingle     LbzListenType = "single"
	ListenTypePlayingNow LbzListenType = "playing_now"
	ListenTypeImport     LbzListenType = "import"
)

type LbzSubmitListenRequest struct {
	ListenType LbzListenType            `json:"listen_type,omitempty"`
	Payload    []LbzSubmitListenPayload `json:"payload,omitempty"`
}

type LbzSubmitListenPayload struct {
	ListenedAt int64        `json:"listened_at,omitempty"`
	TrackMeta  LbzTrackMeta `json:"track_metadata"`
}

type LbzTrackMeta struct {
	ArtistName     string            `json:"artist_name"` // required
	TrackName      string            `json:"track_name"`  // required
	ReleaseName    string            `json:"release_name,omitempty"`
	MBIDMapping    LbzMBIDMapping    `json:"mbid_mapping"`
	AdditionalInfo LbzAdditionalInfo `json:"additional_info,omitempty"`
}
type LbzArtist struct {
	ArtistMBID string `json:"artist_mbid"`
	ArtistName string `json:"artist_credit_name"`
}
type LbzMBIDMapping struct {
	ReleaseMBID   string      `json:"release_mbid"`
	RecordingMBID string      `json:"recording_mbid"`
	ArtistMBIDs   []string    `json:"artist_mbids"`
	Artists       []LbzArtist `json:"artists"`
}

type LbzAdditionalInfo struct {
	MediaPlayer             string   `json:"media_player,omitempty"`
	SubmissionClient        string   `json:"submission_client,omitempty"`
	SubmissionClientVersion string   `json:"submission_client_version,omitempty"`
	ReleaseMBID             string   `json:"release_mbid,omitempty"`
	ReleaseGroupMBID        string   `json:"release_group_mbid,omitempty"`
	ArtistMBIDs             []string `json:"artist_mbids,omitempty"`
	ArtistNames             []string `json:"artist_names,omitempty"`
	RecordingMBID           string   `json:"recording_mbid,omitempty"`
	DurationMs              int32    `json:"duration_ms,omitempty"`
	Duration                int32    `json:"duration,omitempty"`
	Tags                    []string `json:"tags,omitempty"`
	AlbumArtist             string   `json:"albumartist,omitempty"`
}

const (
	maxListensPerRequest = 1000
)

var sfGroup singleflight.Group

func LbzSubmitListenHandler(store db.DB, mbzc mbz.MusicBrainzCaller) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logger.FromContext(r.Context())

		l.Debug().Msg("LbzSubmitListenHandler: Received request to submit listens")

		var req LbzSubmitListenRequest
		requestBytes, err := io.ReadAll(r.Body)
		if err != nil {
			l.Err(err).Msg("LbzSubmitListenHandler: Failed to read request body")
			utils.WriteError(w, "failed to read request body", http.StatusBadRequest)
			return
		}
		if err := json.NewDecoder(bytes.NewBuffer(requestBytes)).Decode(&req); err != nil {
			l.Err(err).Msg("LbzSubmitListenHandler: Failed to decode request")
			utils.WriteError(w, "failed to decode request", http.StatusBadRequest)
			return
		}

		u := middleware.GetUserFromContext(r.Context())
		if u == nil {
			l.Debug().Msg("LbzSubmitListenHandler: Unauthorized request (user context is nil)")
			utils.WriteError(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		l.Debug().Any("request_body", req).Msg("LbzSubmitListenHandler: Parsed request body")

		if len(req.Payload) < 1 {
			l.Debug().Msg("LbzSubmitListenHandler: Payload is empty")
			utils.WriteError(w, "payload is nil", http.StatusBadRequest)
			return
		}

		if len(req.Payload) > maxListensPerRequest {
			l.Debug().Msgf("LbzSubmitListenHandler: Payload exceeds max listens per request (%d > %d)", len(req.Payload), maxListensPerRequest)
			utils.WriteError(w, "payload exceeds max listens per request", http.StatusBadRequest)
			return
		}

		if len(req.Payload) != 1 && req.ListenType != "import" {
			l.Debug().Msg("LbzSubmitListenHandler: Payload must only contain one listen for non-import requests")
			utils.WriteError(w, "payload must only contain one listen for non-import requests", http.StatusBadRequest)
			return
		}

		for _, payload := range req.Payload {
			if payload.TrackMeta.ArtistName == "" || payload.TrackMeta.TrackName == "" {
				l.Debug().Msg("LbzSubmitListenHandler: Artist name or track name are missing")
				utils.WriteError(w, "Artist name or track name are missing", http.StatusBadRequest)
				return
			}

			if req.ListenType != ListenTypePlayingNow && req.ListenType != ListenTypeSingle && req.ListenType != ListenTypeImport {
				l.Debug().Msg("LbzSubmitListenHandler: No listen type provided, assuming 'single'")
				req.ListenType = "single"
			}

			artistMbzIDs, err := utils.ParseUUIDSlice(payload.TrackMeta.AdditionalInfo.ArtistMBIDs)
			if err != nil {
				l.Debug().AnErr("error", err).Msg("LbzSubmitListenHandler: Failed to parse one or more UUIDs")
			}
			if len(artistMbzIDs) < 1 {
				l.Debug().AnErr("error", err).Msg("LbzSubmitListenHandler: Attempting to parse artist UUIDs from mbid_mapping")
				utils.ParseUUIDSlice(payload.TrackMeta.MBIDMapping.ArtistMBIDs)
				if err != nil {
					l.Debug().AnErr("error", err).Msg("LbzSubmitListenHandler: Failed to parse one or more UUIDs")
				}
			}
			rgMbzID, err := uuid.Parse(payload.TrackMeta.AdditionalInfo.ReleaseGroupMBID)
			if err != nil {
				rgMbzID = uuid.Nil
			}
			releaseMbzID, err := uuid.Parse(payload.TrackMeta.AdditionalInfo.ReleaseMBID)
			if err != nil {
				releaseMbzID, err = uuid.Parse(payload.TrackMeta.MBIDMapping.ReleaseMBID)
				if err != nil {
					releaseMbzID = uuid.Nil
				}
			}
			recordingMbzID, err := uuid.Parse(payload.TrackMeta.AdditionalInfo.RecordingMBID)
			if err != nil {
				recordingMbzID, err = uuid.Parse(payload.TrackMeta.MBIDMapping.RecordingMBID)
				if err != nil {
					recordingMbzID = uuid.Nil
				}
			}

			var client string
			if payload.TrackMeta.AdditionalInfo.MediaPlayer != "" {
				client = payload.TrackMeta.AdditionalInfo.MediaPlayer
			} else if payload.TrackMeta.AdditionalInfo.SubmissionClient != "" {
				client = payload.TrackMeta.AdditionalInfo.SubmissionClient
			}

			var duration int32
			if payload.TrackMeta.AdditionalInfo.Duration != 0 {
				duration = payload.TrackMeta.AdditionalInfo.Duration
			} else if payload.TrackMeta.AdditionalInfo.DurationMs != 0 {
				duration = payload.TrackMeta.AdditionalInfo.DurationMs / 1000
			}

			var listenedAt = time.Now()
			if payload.ListenedAt != 0 {
				listenedAt = time.Unix(payload.ListenedAt, 0)
			}

			var artistMbidMap []catalog.ArtistMbidMap
			for _, a := range payload.TrackMeta.MBIDMapping.Artists {
				if a.ArtistMBID == "" || a.ArtistName == "" {
					continue
				}
				mbid, err := uuid.Parse(a.ArtistMBID)
				if err != nil {
					l.Debug().AnErr("error", err).Msgf("LbzSubmitListenHandler: Failed to parse UUID for artist '%s'", a.ArtistName)
				}
				artistMbidMap = append(artistMbidMap, catalog.ArtistMbidMap{Artist: a.ArtistName, Mbid: mbid})
			}

			opts := catalog.SubmitListenOpts{
				MbzCaller:          mbzc,
				ArtistNames:        payload.TrackMeta.AdditionalInfo.ArtistNames,
				Artist:             payload.TrackMeta.ArtistName,
				ArtistMbzIDs:       artistMbzIDs,
				TrackTitle:         payload.TrackMeta.TrackName,
				RecordingMbzID:     recordingMbzID,
				ReleaseTitle:       payload.TrackMeta.ReleaseName,
				ReleaseMbzID:       releaseMbzID,
				ReleaseGroupMbzID:  rgMbzID,
				ArtistMbidMappings: artistMbidMap,
				Duration:           duration,
				Time:               listenedAt,
				UserID:             u.ID,
				Client:             client,
				IsNowPlaying:       req.ListenType == ListenTypePlayingNow,
				SkipSaveListen:     req.ListenType == ListenTypePlayingNow,
			}

			_, err, shared := sfGroup.Do(buildCaolescingKey(payload), func() (interface{}, error) {
				return 0, catalog.SubmitListen(r.Context(), store, opts)
			})
			if shared {
				l.Info().Msg("LbzSubmitListenHandler: Duplicate requests detected; results were coalesced")
			}
			if err != nil {
				l.Err(err).Msg("LbzSubmitListenHandler: Failed to submit listen")
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte("{\"status\": \"internal server error\"}"))
				return
			}
		}

		l.Debug().Msg("LbzSubmitListenHandler: Successfully processed listens")
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"status\": \"ok\"}"))

		if cfg.LbzRelayEnabled() {
			go doLbzRelay(requestBytes, l)
		}
	}
}

func doLbzRelay(requestBytes []byte, l *zerolog.Logger) {
	defer func() {
		if r := recover(); r != nil {
			l.Error().Interface("recover", r).Msg("doLbzRelay: Panic occurred")
		}
	}()
	const (
		maxRetryDuration = 3 * time.Minute
		initialBackoff   = 5 * time.Second
		maxBackoff       = 40 * time.Second
	)
	l.Debug().Msg("doLbzRelay: Building ListenBrainz relay request")
	req, err := http.NewRequest("POST", cfg.LbzRelayUrl()+"/submit-listens", bytes.NewBuffer(requestBytes))
	if err != nil {
		l.Err(err).Msg("doLbzRelay: Failed to build ListenBrainz relay request")
		return
	}
	req.Header.Add("Authorization", "Token "+cfg.LbzRelayToken())
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	var resp *http.Response
	var body []byte
	start := time.Now()
	backoff := initialBackoff

	for {
		l.Debug().Msg("doLbzRelay: Sending ListenBrainz relay request")
		resp, err = client.Do(req)
		if err != nil {
			l.Err(err).Msg("doLbzRelay: Failed to send ListenBrainz relay request")
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			l.Info().Msg("doLbzRelay: Successfully relayed ListenBrainz submission")
			return
		}

		body, _ = io.ReadAll(resp.Body)

		if resp.StatusCode >= 500 && time.Since(start)+backoff <= maxRetryDuration {
			l.Warn().
				Int("status", resp.StatusCode).
				Str("response", string(body)).
				Msg("doLbzRelay: Retryable server error from ListenBrainz relay, retrying...")
			time.Sleep(backoff)
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		l.Warn().
			Int("status", resp.StatusCode).
			Str("response", string(body)).
			Msg("doLbzRelay: Non-2XX response from ListenBrainz relay")
		return
	}
}

func buildCaolescingKey(p LbzSubmitListenPayload) string {
	// the key not including the listen_type introduces the very rare possibility of a playing_now
	// request taking precedence over a single, meaning that a listen will not be logged when it
	// should, however that would require a playing_now request to fire a few seconds before a 'single'
	// of the same track, which should never happen outside of misbehaving clients
	//
	// this could be fixed by restructuring the database inserts for idempotency, which would
	// eliminate the need to coalesce responses, however i'm not gonna do that right now
	return fmt.Sprintf("%s:%s:%s", p.TrackMeta.ArtistName, p.TrackMeta.TrackName, p.TrackMeta.ReleaseName)
}
