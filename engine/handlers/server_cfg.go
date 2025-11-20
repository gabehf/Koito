package handlers

import (
	"net/http"

	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/utils"
)

type ServerConfig struct {
	DefaultTheme string `json:"default_theme"`
}

func GetCfgHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.WriteJSON(w, http.StatusOK, ServerConfig{DefaultTheme: cfg.DefaultTheme()})
	}
}
