package handlers

import (
	"net/http"
	"strconv"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
)

func GetInterestHandler(store db.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("GetInterestHandler: Received request to retrieve interest")

		// im just using this to parse the artist/album/track id, which is bad
		parsed := OptsFromRequest(r)

		bucketCountStr := r.URL.Query().Get("buckets")
		var buckets = 0
		var err error
		if buckets, err = strconv.Atoi(bucketCountStr); err != nil {
			l.Debug().Msg("GetInterestHandler: Buckets is not an integer")
			utils.WriteError(w, "parameter 'buckets' must be an integer", http.StatusBadRequest)
			return
		}

		opts := db.GetInterestOpts{
			Buckets:  buckets,
			AlbumID:  int32(parsed.AlbumID),
			ArtistID: int32(parsed.ArtistID),
			TrackID:  int32(parsed.TrackID),
		}

		interest, err := store.GetInterest(ctx, opts)
		if err != nil {
			l.Err(err).Msg("GetInterestHandler: Failed to query interest")
			utils.WriteError(w, "Failed to retrieve interest: "+err.Error(), http.StatusInternalServerError)
			return
		}

		utils.WriteJSON(w, http.StatusOK, interest)
	}
}
