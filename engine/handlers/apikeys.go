package handlers

import (
	"net/http"

	"github.com/gabehf/koito/engine/middleware"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
)

func GenerateApiKeyHandler(store db.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		user := middleware.GetUserFromContext(ctx)
		if user == nil {
			l.Debug().Msg("GenerateApiKeyHandler: Invalid user context")
			utils.WriteError(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		body, err := utils.DecodeBody[struct {
			Label string `json:"label"`
		}](r)
		if err != nil || body.Label == "" {
			l.Debug().Msg("GenerateApiKeyHandler: Invalid or missing label in request body")
			utils.WriteError(w, "label is required", http.StatusBadRequest)
			return
		}

		apiKey, err := utils.GenerateRandomString(48)
		if err != nil {
			l.Error().Err(err).Msg("GenerateApiKeyHandler: Failed to generate API key")
			utils.WriteError(w, "failed to generate api key", http.StatusInternalServerError)
			return
		}

		key, err := store.SaveApiKey(ctx, db.SaveApiKeyOpts{
			UserID: user.ID,
			Key:    apiKey,
			Label:  body.Label,
		})
		if err != nil {
			l.Error().Err(err).Msg("GenerateApiKeyHandler: Failed to save API key")
			utils.WriteError(w, "failed to save api key", http.StatusInternalServerError)
			return
		}

		l.Debug().Msgf("GenerateApiKeyHandler: Successfully generated API key ID %d", key.ID)
		utils.WriteJSON(w, http.StatusCreated, key)
	}
}

func DeleteApiKeyHandler(store db.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		user := middleware.GetUserFromContext(ctx)
		if user == nil {
			l.Debug().Msg("DeleteApiKeyHandler: Invalid user context")
			utils.WriteError(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		apiKeyID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("DeleteApiKeyHandler: Invalid API key ID")
			utils.WriteError(w, "invalid id", http.StatusBadRequest)
			return
		}

		if err := store.DeleteApiKey(ctx, apiKeyID); err != nil {
			l.Error().Err(err).Msg("DeleteApiKeyHandler: Failed to delete API key")
			utils.WriteError(w, "failed to delete api key", http.StatusInternalServerError)
			return
		}

		l.Debug().Msgf("DeleteApiKeyHandler: Successfully deleted API key ID %d", apiKeyID)
		w.WriteHeader(http.StatusNoContent)
	}
}

func GetApiKeysHandler(store db.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		user := middleware.GetUserFromContext(ctx)
		if user == nil {
			l.Debug().Msg("GetApiKeysHandler: Invalid user context")
			utils.WriteError(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		apiKeys, err := store.GetApiKeysByUserID(ctx, user.ID)
		if err != nil {
			l.Error().Err(err).Msg("GetApiKeysHandler: Failed to retrieve API keys")
			utils.WriteError(w, "failed to retrieve api keys", http.StatusInternalServerError)
			return
		}

		l.Debug().Msgf("GetApiKeysHandler: Retrieved %d API keys", len(apiKeys))
		utils.WriteJSON(w, http.StatusOK, apiKeys)
	}
}

func UpdateApiKeyLabelHandler(store db.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		user := middleware.GetUserFromContext(ctx)
		if user == nil {
			l.Debug().Msg("UpdateApiKeyLabelHandler: Invalid user context")
			utils.WriteError(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		apiKeyID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("UpdateApiKeyLabelHandler: Invalid API key ID")
			utils.WriteError(w, "invalid id", http.StatusBadRequest)
			return
		}

		body, err := utils.DecodeBody[struct {
			Label string `json:"label"`
		}](r)
		if err != nil || body.Label == "" {
			l.Debug().Msg("UpdateApiKeyLabelHandler: Invalid or missing label in request body")
			utils.WriteError(w, "label is required", http.StatusBadRequest)
			return
		}

		if err := store.UpdateApiKeyLabel(ctx, db.UpdateApiKeyLabelOpts{
			UserID: user.ID,
			ID:     apiKeyID,
			Label:  body.Label,
		}); err != nil {
			l.Error().Err(err).Msg("UpdateApiKeyLabelHandler: Failed to update API key label")
			utils.WriteError(w, "failed to update api key label", http.StatusInternalServerError)
			return
		}

		l.Debug().Msgf("UpdateApiKeyLabelHandler: Successfully updated label for API key ID %d", apiKeyID)
		w.WriteHeader(http.StatusNoContent)
	}
}
