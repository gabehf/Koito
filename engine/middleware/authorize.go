package middleware

import (
	"net/http"

	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/models"
	"github.com/gabehf/koito/internal/utils"
)

// ensures the authenticated user has admin role
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := logger.FromContext(r.Context())
		user := GetUserFromContext(r.Context())
		if user == nil {
			l.Debug().Msg("RequireAdmin: Unauthorized access (no user)")
			utils.WriteError(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if user.Role != models.UserRoleAdmin {
			l.Debug().Msgf("RequireAdmin: Forbidden - user %d is not admin", user.ID)
			utils.WriteError(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
