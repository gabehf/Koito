package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gabehf/koito/engine/middleware"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func LoginHandler(store db.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("LoginHandler: Received request")

		if err := r.ParseForm(); err != nil {
			l.Debug().AnErr("error", err).Msg("LoginHandler: Failed to parse form")
			utils.WriteError(w, "invalid request format", http.StatusBadRequest)
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")
		if username == "" || password == "" {
			l.Debug().Msg("LoginHandler: Missing credentials")
			utils.WriteError(w, "username and password required", http.StatusBadRequest)
			return
		}

		user, err := store.GetUserByUsername(ctx, username)
		if err != nil {
			l.Error().Err(err).Msg("LoginHandler: Database error fetching user")
			utils.WriteError(w, "authentication failed", http.StatusInternalServerError)
			return
		}
		if user == nil {
			l.Debug().Msg("LoginHandler: User not found")
			utils.WriteError(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		if err := bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
			l.Debug().Msg("LoginHandler: Invalid password")
			utils.WriteError(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		expiresAt := time.Now().Add(24 * time.Hour)
		if strings.ToLower(r.FormValue("remember_me")) == "true" {
			expiresAt = time.Now().Add(30 * 24 * time.Hour)
		}

		session, err := store.SaveSession(ctx, user.ID, expiresAt, r.FormValue("remember_me") == "true")
		if err != nil {
			l.Error().Err(err).Msg("LoginHandler: Failed to create session")
			utils.WriteError(w, "authentication failed", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "koito_session",
			Value:    session.ID.String(),
			Expires:  expiresAt,
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
		})

		l.Debug().Msgf("LoginHandler: User %d authenticated", user.ID)
		w.WriteHeader(http.StatusNoContent)
	}
}

func LogoutHandler(store db.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("LogoutHandler: Received request")

		cookie, err := r.Cookie("koito_session")
		if err == nil {
			sid, err := uuid.Parse(cookie.Value)
			if err != nil {
				l.Debug().AnErr("error", err).Msg("LogoutHandler: Invalid session ID")
			} else if err := store.DeleteSession(ctx, sid); err != nil {
				l.Error().Err(err).Msg("LogoutHandler: Failed to delete session")
			}
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "koito_session",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			MaxAge:   -1,
		})

		l.Debug().Msg("LogoutHandler: Session terminated")
		w.WriteHeader(http.StatusNoContent)
	}
}

func MeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("MeHandler: Received request")

		user := middleware.GetUserFromContext(ctx)
		if user == nil {
			l.Debug().Msg("MeHandler: Unauthorized access")
			utils.WriteError(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		l.Debug().Msgf("MeHandler: Returning user data for ID %d", user.ID)
		utils.WriteJSON(w, http.StatusOK, user)
	}
}

func UpdateUserHandler(store db.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("UpdateUserHandler: Received request")

		user := middleware.GetUserFromContext(ctx)
		if user == nil {
			l.Debug().Msg("UpdateUserHandler: Unauthorized access")
			utils.WriteError(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if err := r.ParseForm(); err != nil {
			l.Error().Err(err).Msg("UpdateUserHandler: Invalid form data")
			utils.WriteError(w, "invalid request", http.StatusBadRequest)
			return
		}

		opts := db.UpdateUserOpts{ID: user.ID}
		if username := r.FormValue("username"); username != "" {
			opts.Username = username
		}
		if password := r.FormValue("password"); password != "" {
			opts.Password = password
		}

		if opts.Username == "" && opts.Password == "" {
			l.Debug().Msg("UpdateUserHandler: No update parameters provided")
			utils.WriteError(w, "no changes specified", http.StatusBadRequest)
			return
		}

		if err := store.UpdateUser(ctx, opts); err != nil {
			l.Error().Err(err).Msg("UpdateUserHandler: Update failed")
			utils.WriteError(w, "update failed", http.StatusBadRequest)
			return
		}

		l.Debug().Msgf("UpdateUserHandler: User %d updated", user.ID)
		w.WriteHeader(http.StatusNoContent)
	}
}
