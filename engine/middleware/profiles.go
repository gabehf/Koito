package middleware

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/models"
	"github.com/gabehf/koito/internal/utils"
	"github.com/go-chi/chi/v5"
)

type profileKey struct{}

func ProfileFromContext(ctx context.Context) *models.ListeningProfile {
	p, _ := ctx.Value(profileKey{}).(*models.ListeningProfile)
	return p
}
func withProfile(ctx context.Context, p *models.ListeningProfile) context.Context {
	return db.WithProfile(context.WithValue(ctx, profileKey{}, p), p.ID)
}
func ProfileScope(store db.ProfileStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slug := chi.URLParam(r, "slug")
			var p *models.ListeningProfile
			if slug != "" {
				var err error
				p, err = store.GetListeningProfile(r.Context(), slug)
				if err != nil {
					utils.WriteError(w, "profile not found", 404)
					return
				}
			} else {
				profiles, err := store.ListListeningProfiles(r.Context())
				if err != nil {
					utils.WriteError(w, "failed to resolve profile", 500)
					return
				}
				if len(profiles) != 1 {
					utils.WriteError(w, "select an explicit profile", 409)
					return
				}
				p = &profiles[0]
			}
			next.ServeHTTP(w, r.WithContext(withProfile(r.Context(), p)))
		})
	}
}
func IngestionKey(store db.ProfileStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if len(header) < 7 || !strings.EqualFold(header[:6], "Token ") {
				utils.WriteError(w, "invalid API key", 401)
				return
			}
			p, err := store.GetProfileByAPIKey(r.Context(), strings.TrimSpace(header[6:]))
			if err != nil {
				utils.WriteError(w, "invalid API key", 401)
				return
			}
			next.ServeHTTP(w, r.WithContext(withProfile(r.Context(), p)))
		})
	}
}
func RequireAdmin(store db.UserStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, err := validateSession(r.Context(), store, r)
			if err != nil || u == nil {
				utils.WriteError(w, "admin session required", 401)
				return
			}
			if u.Role != models.UserRoleAdmin {
				utils.WriteError(w, "administrator required", 403)
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), UserContextKey, u)))
		})
	}
}

// Cookie-authenticated web mutations must not be triggered by another origin.
func SameOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "HEAD" && r.Method != "OPTIONS" {
			scheme := "http"
			if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
				scheme = "https"
			}
			origin := r.Header.Get("Origin")
			if origin != "" {
				parsed, err := url.Parse(origin)
				if err != nil || parsed.Scheme != scheme || !strings.EqualFold(parsed.Host, r.Host) {
					utils.WriteError(w, "cross-origin mutation forbidden", 403)
					return
				}
			}
			if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
				utils.WriteError(w, "cross-site mutation forbidden", 403)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
