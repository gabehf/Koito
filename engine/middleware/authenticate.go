package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/models"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type MiddlwareContextKey string

const UserContextKey MiddlwareContextKey = "user"

func validateSession(ctx context.Context, store db.UserStore, r *http.Request) (*models.User, error) {
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

func GetUserFromContext(ctx context.Context) *models.User {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}
