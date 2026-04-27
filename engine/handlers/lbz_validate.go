package handlers

import (
	"net/http"

	"github.com/gabehf/koito/engine/middleware"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
)

type LbzValidateResponse struct {
	Code     int    `json:"code"`
	Error    string `json:"error,omitempty"`
	Message  string `json:"message,omitempty"`
	Valid    bool   `json:"valid,omitempty"`
	UserName string `json:"user_name,omitempty"`
}

func LbzValidateTokenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("LbzValidateTokenHandler: Validating user token")

		u := middleware.GetUserFromContext(ctx)
		var response LbzValidateResponse
		if u == nil {
			l.Debug().Msg("LbzValidateTokenHandler: Invalid token, user not found in context")
			response.Code = http.StatusUnauthorized
			response.Error = "Incorrect Authorization"
			w.WriteHeader(http.StatusUnauthorized)
			utils.WriteJSON(w, http.StatusOK, response)
		} else {
			l.Debug().Msgf("LbzValidateTokenHandler: Token valid for user '%s'", u.Username)
			response.Code = http.StatusOK
			response.Message = "Token valid."
			response.Valid = true
			response.UserName = u.Username
			utils.WriteJSON(w, http.StatusOK, response)
		}
	}
}
