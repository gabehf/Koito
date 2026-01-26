package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/models"
	"github.com/gabehf/koito/internal/utils"
	"github.com/google/uuid"
)

type MiddlwareContextKey string

const (
	UserContextKey   MiddlwareContextKey = "user"
	apikeyContextKey MiddlwareContextKey = "apikeyID"
)

type AuthMode int

const (
	AuthModeSessionCookie AuthMode = iota
	AuthModeAPIKey
	AuthModeSessionOrAPIKey
	AuthModeLoginGate
)

func Authenticate(store db.DB, mode AuthMode) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			l := logger.FromContext(ctx)

			var user *models.User
			var err error

			switch mode {
			case AuthModeSessionCookie:
				user, err = validateSession(ctx, store, r)

			case AuthModeAPIKey:
				user, err = validateAPIKey(ctx, store, r)

			case AuthModeSessionOrAPIKey:
				user, err = validateSession(ctx, store, r)
				if err != nil || user == nil {
					user, err = validateAPIKey(ctx, store, r)
				}

			case AuthModeLoginGate:
				if cfg.LoginGate() {
					user, err = validateSession(ctx, store, r)
					if err != nil || user == nil {
						user, err = validateAPIKey(ctx, store, r)
					}
				} else {
					next.ServeHTTP(w, r)
				}
			}

			if err != nil {
				l.Err(err).Msg("authentication failed")
				utils.WriteError(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if user == nil {
				utils.WriteError(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if user != nil {
				ctx = context.WithValue(ctx, UserContextKey, user)
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func validateSession(ctx context.Context, store db.DB, r *http.Request) (*models.User, error) {
	l := logger.FromContext(r.Context())

	l.Debug().Msgf("ValidateSession: Checking user authentication via session cookie")

	cookie, err := r.Cookie("koito_session")
	var sid uuid.UUID
	if err == nil {
		sid, err = uuid.Parse(cookie.Value)
		if err != nil {
			l.Err(err).Msg("ValidateSession: Could not parse UUID from session cookie")
			return nil, errors.New("session cookie is invalid")
		}
	} else {
		l.Debug().Msgf("ValidateSession: No session cookie found; attempting API key authentication")
		return nil, errors.New("session cookie is missing")
	}

	l.Debug().Msg("ValidateSession: Retrieved login cookie from request")

	u, err := store.GetUserBySession(r.Context(), sid)
	if err != nil {
		l.Err(fmt.Errorf("ValidateSession: %w", err)).Msg("Error accessing database")
		return nil, errors.New("internal server error")
	}
	if u == nil {
		l.Debug().Msg("ValidateSession: No user with session id found")
		return nil, errors.New("no user with session id found")
	}

	ctx = context.WithValue(r.Context(), UserContextKey, u)
	r = r.WithContext(ctx)

	l.Debug().Msgf("ValidateSession: Refreshing session for user '%s'", u.Username)

	store.RefreshSession(r.Context(), sid, time.Now().Add(30*24*time.Hour))

	l.Debug().Msgf("ValidateSession: Refreshed session for user '%s'", u.Username)

	return u, nil
}

func validateAPIKey(ctx context.Context, store db.DB, r *http.Request) (*models.User, error) {
	l := logger.FromContext(ctx)

	l.Debug().Msg("ValidateApiKey: Checking if user is already authenticated")

	authH := r.Header.Get("Authorization")
	var token string
	if strings.HasPrefix(strings.ToLower(authH), "token ") {
		token = strings.TrimSpace(authH[6:]) // strip "Token "
	} else {
		l.Error().Msg("ValidateApiKey: Authorization header must be formatted 'Token {token}'")
		return nil, errors.New("authorization header is invalid")
	}

	u, err := store.GetUserByApiKey(ctx, token)
	if err != nil {
		l.Err(err).Msg("ValidateApiKey: Failed to get user from database using api key")
		return nil, errors.New("internal server error")
	}
	if u == nil {
		l.Debug().Msg("ValidateApiKey: API key does not exist")
		return nil, errors.New("authorization token is invalid")
	}

	ctx = context.WithValue(r.Context(), UserContextKey, u)
	r = r.WithContext(ctx)

	return u, nil
}

func GetUserFromContext(ctx context.Context) *models.User {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}
