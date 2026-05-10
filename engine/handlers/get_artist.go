package handlers

import (
	"net/http"
	"strconv"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
)

func GetArtistHandler(store db.ArtistStore) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("GetArtistHandler: Received request to retrieve artist")

		id, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("GetArtistHandler: Invalid album id")
			utils.WriteError(w, "invalid album id", http.StatusBadRequest)
			return
		}

		l.Debug().Msgf("GetArtistHandler: Retrieving artist with ID %d", id)

		artist, err := store.GetArtist(ctx, db.GetArtistOpts{ID: id})
		if err != nil {
			l.Err(err).Msgf("GetArtistHandler: Failed to retrieve artist with ID %d", id)
			utils.WriteError(w, "artist with specified id could not be found", http.StatusNotFound)
			return
		}

		l.Debug().Msgf("GetArtistHandler: Successfully retrieved artist with ID %d", id)
		utils.WriteJSON(w, http.StatusOK, artist)
	}
}

// GetArtistAliasesHandler retrieves all aliases for a given artist.
func GetArtistAliasesHandler(store db.ArtistStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		artistID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("GetArtistAliasesHandler: Invalid artist id")
			utils.WriteError(w, "invalid artist id", http.StatusBadRequest)
			return
		}

		aliases, err := store.GetAllArtistAliases(ctx, artistID)
		if err != nil {
			l.Err(err).Msg("GetArtistAliasesHandler: Failed to get artist aliases")
			utils.WriteError(w, "failed to retrieve aliases", http.StatusInternalServerError)
			return
		}

		utils.WriteJSON(w, http.StatusOK, aliases)
	}
}

// GetArtistInterestHandler retrieves interest data for a given artist.
func GetArtistInterestHandler(store db.ListenStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		artistID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("GetArtistInterestHandler: Invalid artist id")
			utils.WriteError(w, "invalid artist id", http.StatusBadRequest)
			return
		}

		buckets, err := strconv.Atoi(r.URL.Query().Get("buckets"))
		if err != nil {
			l.Debug().Msg("GetArtistInterestHandler: Buckets is not an integer")
			utils.WriteError(w, "parameter 'buckets' must be an integer", http.StatusBadRequest)
			return
		}

		interest, err := store.GetInterest(ctx, db.GetInterestOpts{
			ArtistID: artistID,
			Buckets:  buckets,
		})
		if err != nil {
			l.Err(err).Msg("GetArtistInterestHandler: Failed to query interest")
			utils.WriteError(w, "failed to retrieve interest: "+err.Error(), http.StatusInternalServerError)
			return
		}

		utils.WriteJSON(w, http.StatusOK, interest)
	}
}
