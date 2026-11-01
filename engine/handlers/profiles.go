package handlers

import (
	"github.com/gabehf/koito/engine/middleware"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/utils"
	"net/http"
)

func ListProfilesHandler(store db.ProfileStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		profiles, err := store.ListListeningProfiles(r.Context())
		if err != nil {
			utils.WriteError(w, "failed to list profiles", 500)
			return
		}
		utils.WriteJSON(w, 200, profiles)
	}
}
func GetProfileHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.WriteJSON(w, 200, middleware.ProfileFromContext(r.Context()))
	}
}
func UpdateProfileHandler(store db.ProfileStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := middleware.ProfileFromContext(r.Context())
		body, err := utils.DecodeBody[struct {
			DisplayName *string `json:"display_name"`
			Slug        *string `json:"slug"`
		}](r)
		if err != nil || (body.DisplayName == nil && body.Slug == nil) {
			utils.WriteError(w, "display_name or slug required", 400)
			return
		}
		name, slug := p.DisplayName, p.Slug
		if body.DisplayName != nil {
			name = *body.DisplayName
		}
		if body.Slug != nil {
			slug = *body.Slug
		}
		if err := store.UpdateListeningProfile(r.Context(), p.ID, name, slug); err != nil {
			utils.WriteError(w, "invalid, reserved or duplicate profile name/slug", 400)
			return
		}
		updated, err := store.GetListeningProfileByID(r.Context(), p.ID)
		if err != nil {
			utils.WriteError(w, "profile not found", 404)
			return
		}
		utils.WriteJSON(w, 200, updated)
	}
}
