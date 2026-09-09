package handlers

import (
	"net/http"

	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/utils"
)

type ServerConfig struct {
	DefaultTheme string `json:"default_theme"`
	DateFormat   string `json:"date_format"`
	ClockFormat  string `json:"clock_format"`
	WeekStart    string `json:"week_start"`
}

func GetCfgHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.WriteJSON(w, http.StatusOK, ServerConfig{
			DefaultTheme: cfg.DefaultTheme(),
			DateFormat:   cfg.DateFormat(),
			ClockFormat:  cfg.ClockFormat(),
			WeekStart:    cfg.WeekStart(),
		})
	}
}
