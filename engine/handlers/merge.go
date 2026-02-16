package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
)

func MergeTracksHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logger.FromContext(r.Context())

		l.Debug().Msg("MergeTracksHandler: Received request to merge tracks")

		fromidStr := r.URL.Query().Get("from_id")
		fromId, err := strconv.Atoi(fromidStr)
		if err != nil {
			l.Debug().AnErr("error", err).Msg("MergeTracksHandler: Invalid from_id parameter")
			utils.WriteError(w, "from_id is invalid", http.StatusBadRequest)
			return
		}

		toidStr := r.URL.Query().Get("to_id")
		toId, err := strconv.Atoi(toidStr)
		if err != nil {
			l.Debug().AnErr("error", err).Msg("MergeTracksHandler: Invalid to_id parameter")
			utils.WriteError(w, "to_id is invalid", http.StatusBadRequest)
			return
		}

		l.Debug().Msgf("MergeTracksHandler: Merging tracks from ID %d to ID %d", fromId, toId)

		err = store.MergeTracks(r.Context(), int32(fromId), int32(toId))
		if err != nil {
			l.Err(err).Msg("MergeTracksHandler: Failed to merge tracks")
			utils.WriteError(w, "Failed to merge tracks: "+err.Error(), http.StatusInternalServerError)
			return
		}

		l.Debug().Msgf("MergeTracksHandler: Successfully merged tracks from ID %d to ID %d", fromId, toId)
		w.WriteHeader(http.StatusNoContent)
	}
}

func MergeReleaseGroupsHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logger.FromContext(r.Context())

		l.Debug().Msg("MergeReleaseGroupsHandler: Received request to merge release groups")

		fromidStr := r.URL.Query().Get("from_id")
		fromId, err := strconv.Atoi(fromidStr)
		if err != nil {
			l.Debug().AnErr("error", err).Msg("MergeReleaseGroupsHandler: Invalid from_id parameter")
			utils.WriteError(w, "from_id is invalid", http.StatusBadRequest)
			return
		}

		toidStr := r.URL.Query().Get("to_id")
		toId, err := strconv.Atoi(toidStr)
		if err != nil {
			l.Debug().AnErr("error", err).Msg("MergeReleaseGroupsHandler: Invalid to_id parameter")
			utils.WriteError(w, "to_id is invalid", http.StatusBadRequest)
			return
		}

		var replaceImage bool
		replaceImgStr := r.URL.Query().Get("replace_image")
		if strings.ToLower(replaceImgStr) == "true" {
			l.Debug().Msg("MergeReleaseGroupsHandler: Merge will replace image")
			replaceImage = true
		}

		l.Debug().Msgf("MergeReleaseGroupsHandler: Merging release groups from ID %d to ID %d", fromId, toId)

		err = store.MergeAlbums(r.Context(), int32(fromId), int32(toId), replaceImage)
		if err != nil {
			l.Err(err).Msg("MergeReleaseGroupsHandler: Failed to merge release groups")
			utils.WriteError(w, "Failed to merge release groups: "+err.Error(), http.StatusInternalServerError)
			return
		}

		l.Debug().Msgf("MergeReleaseGroupsHandler: Successfully merged release groups from ID %d to ID %d", fromId, toId)
		w.WriteHeader(http.StatusNoContent)
	}
}

func MergeArtistsHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logger.FromContext(r.Context())

		l.Debug().Msg("MergeArtistsHandler: Received request to merge artists")

		fromidStr := r.URL.Query().Get("from_id")
		fromId, err := strconv.Atoi(fromidStr)
		if err != nil {
			l.Debug().AnErr("error", err).Msg("MergeArtistsHandler: Invalid from_id parameter")
			utils.WriteError(w, "from_id is invalid", http.StatusBadRequest)
			return
		}

		toidStr := r.URL.Query().Get("to_id")
		toId, err := strconv.Atoi(toidStr)
		if err != nil {
			l.Debug().AnErr("error", err).Msg("MergeArtistsHandler: Invalid to_id parameter")
			utils.WriteError(w, "to_id is invalid", http.StatusBadRequest)
			return
		}

		var replaceImage bool
		replaceImgStr := r.URL.Query().Get("replace_image")
		if strings.ToLower(replaceImgStr) == "true" {
			l.Debug().Msg("MergeReleaseGroupsHandler: Merge will replace image")
			replaceImage = true
		}

		l.Debug().Msgf("MergeArtistsHandler: Merging artists from ID %d to ID %d", fromId, toId)

		err = store.MergeArtists(r.Context(), int32(fromId), int32(toId), replaceImage)
		if err != nil {
			l.Err(err).Msg("MergeArtistsHandler: Failed to merge artists")
			utils.WriteError(w, "Failed to merge artists: "+err.Error(), http.StatusInternalServerError)
			return
		}

		l.Debug().Msgf("MergeArtistsHandler: Successfully merged artists from ID %d to ID %d", fromId, toId)
		w.WriteHeader(http.StatusNoContent)
	}
}
