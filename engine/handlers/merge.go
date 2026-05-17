package handlers

import (
	"net/http"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
)

func MergeArtistsHandler(store db.ArtistStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logger.FromContext(r.Context())

		l.Debug().Msg("MergeArtistsHandler: Received request to merge artists")

		body, err := utils.DecodeBody[struct {
			MergeFromID  int32 `json:"merge_from_id"`
			ReplaceImage bool  `json:"replace_image"`
		}](r)
		if err != nil {
			l.Debug().AnErr("error", err).Msg("MergeArtistsHandler: request body invalid or missing")
			utils.WriteError(w, "request body is invalid or missing", http.StatusBadRequest)
			return
		} else if body.MergeFromID == 0 {
			l.Debug().AnErr("error", err).Msg("MergeArtistsHandler: required body key 'merge_from_id' invalid or missing")
			utils.WriteError(w, "merge_from_id is invalid or missing", http.StatusBadRequest)
			return
		}

		toId, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("MergeArtistsHandler: Invalid artist id")
			utils.WriteError(w, "invalid artist id", http.StatusBadRequest)
			return
		}

		l.Debug().Msgf("MergeArtistsHandler: Merging artists from ID %d to ID %d", body.MergeFromID, toId)

		err = store.MergeArtists(r.Context(), body.MergeFromID, toId, body.ReplaceImage)
		if err != nil {
			l.Err(err).Msg("MergeArtistsHandler: Failed to merge artists")
			utils.WriteError(w, "Failed to merge artists: "+err.Error(), http.StatusInternalServerError)
			return
		}

		l.Debug().Msgf("MergeArtistsHandler: Successfully merged artists from ID %d to ID %d", body.MergeFromID, toId)
		w.WriteHeader(http.StatusNoContent)
	}
}

func MergeAlbumsHandler(store db.AlbumStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logger.FromContext(r.Context())

		l.Debug().Msg("MergeAlbumsHandler: Received request to merge release groups")

		body, err := utils.DecodeBody[struct {
			MergeFromID  int32 `json:"merge_from_id"`
			ReplaceImage bool  `json:"replace_image"`
		}](r)
		if err != nil {
			l.Debug().AnErr("error", err).Msg("MergeAlbumsHandler: request body invalid or missing")
			utils.WriteError(w, "request body is invalid or missing", http.StatusBadRequest)
			return
		} else if body.MergeFromID == 0 {
			l.Debug().AnErr("error", err).Msg("MergeAlbumsHandler: required body key 'merge_from_id' invalid or missing")
			utils.WriteError(w, "merge_from_id is invalid or missing", http.StatusBadRequest)
			return
		}

		toId, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("MergeAlbumsHandler: Invalid album id")
			utils.WriteError(w, "invalid album id", http.StatusBadRequest)
			return
		}

		l.Debug().Msgf("MergeAlbumsHandler: Merging albums from ID %d to ID %d", body.MergeFromID, toId)

		err = store.MergeAlbums(r.Context(), body.MergeFromID, toId, body.ReplaceImage)
		if err != nil {
			l.Err(err).Msg("MergeAlbumsHandler: Failed to merge albums")
			utils.WriteError(w, "Failed to merge albums: "+err.Error(), http.StatusInternalServerError)
			return
		}

		l.Debug().Msgf("MergeAlbumsHandler: Successfully merged albums from ID %d to ID %d", body.MergeFromID, toId)
		w.WriteHeader(http.StatusNoContent)
	}
}

func MergeTracksHandler(store db.TrackStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logger.FromContext(r.Context())

		l.Debug().Msg("MergeTracksHandler: Received request to merge tracks")

		body, err := utils.DecodeBody[struct {
			MergeFromID int32 `json:"merge_from_id"`
		}](r)
		if err != nil {
			l.Debug().AnErr("error", err).Msg("MergeTracksHandler: request body invalid or missing")
			utils.WriteError(w, "request body is invalid or missing", http.StatusBadRequest)
			return
		} else if body.MergeFromID == 0 {
			l.Debug().AnErr("error", err).Msg("MergeTracksHandler: required body key 'merge_from_id' invalid or missing")
			utils.WriteError(w, "merge_from_id is invalid or missing", http.StatusBadRequest)
			return
		}

		toId, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("MergeTracksHandler: Invalid track id")
			utils.WriteError(w, "invalid track id", http.StatusBadRequest)
			return
		}

		l.Debug().Msgf("MergeTracksHandler: Merging tracks from ID %d to ID %d", body.MergeFromID, toId)

		err = store.MergeTracks(r.Context(), body.MergeFromID, toId)
		if err != nil {
			l.Err(err).Msg("MergeTracksHandler: Failed to merge tracks")
			utils.WriteError(w, "Failed to merge tracks: "+err.Error(), http.StatusInternalServerError)
			return
		}

		l.Debug().Msgf("MergeTracksHandler: Successfully merged tracks from ID %d to ID %d", body.MergeFromID, toId)
		w.WriteHeader(http.StatusNoContent)
	}
}
