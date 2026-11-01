package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/utils"
)

type profileRelay struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url"`
	Token   string `json:"token"`
}

func decodeProfileRelay(w http.ResponseWriter, r *http.Request) (profileRelay, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, 8192)
	body, err := utils.DecodeBody[profileRelay](r)
	body.URL = strings.TrimRight(strings.TrimSpace(body.URL), "/")
	body.Token = strings.TrimSpace(body.Token)
	parsed, parseErr := url.Parse(body.URL)
	if err != nil || parseErr != nil || parsed.Host == "" || parsed.User != nil ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.RawQuery != "" ||
		parsed.Fragment != "" || body.Token == "" || strings.ContainsAny(body.Token, "\r\n") {
		utils.WriteError(w, "valid relay base URL and token required", http.StatusBadRequest)
		return profileRelay{}, false
	}
	return body, true
}

// Only mounted behind RequireAdmin: the administrator explicitly edits the token.
func GetProfileRelayHandler(store db.ProfileStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		url, token, err := store.GetProfileRelay(r.Context(), db.ProfileID(r.Context()))
		if err != nil {
			utils.WriteError(w, "failed to load relay", 500)
			return
		}
		utils.WriteJSON(w, 200, profileRelay{Enabled: url != "" && token != "", URL: url, Token: token})
	}
}

func SetProfileRelayHandler(store db.ProfileStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := db.ProfileID(r.Context())
		if r.Method == "DELETE" {
			if err := store.SetProfileRelay(r.Context(), id, "", ""); err != nil {
				utils.WriteError(w, "failed to remove relay", 500)
				return
			}
			w.WriteHeader(204)
			return
		}
		body, ok := decodeProfileRelay(w, r)
		if !ok {
			return
		}
		if err := store.SetProfileRelay(r.Context(), id, body.URL, body.Token); err != nil {
			utils.WriteError(w, "failed to configure relay", 500)
			return
		}
		w.WriteHeader(204)
	}
}

// Validate the draft credentials without saving them or submitting a listen.
func TestProfileRelayHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		body, ok := decodeProfileRelay(w, r)
		if !ok {
			return
		}
		req, err := http.NewRequestWithContext(r.Context(), "GET", body.URL+"/validate-token", nil)
		if err != nil {
			utils.WriteError(w, "invalid relay URL", 400)
			return
		}
		req.Header.Set("Authorization", "Token "+body.Token)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", cfg.UserAgent())
		client := &http.Client{Timeout: 10 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		}}
		resp, err := client.Do(req)
		if err != nil {
			utils.WriteError(w, "could not connect to relay (check the URL, network and TLS certificate)", 502)
			return
		}
		defer resp.Body.Close()
		var result struct {
			Valid    bool   `json:"valid"`
			UserName string `json:"user_name"`
		}
		if resp.StatusCode != 200 {
			utils.WriteError(w, "relay rejected validation; check the base URL and token (redirects are not followed)", 502)
			return
		}
		if err := json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&result); err != nil {
			utils.WriteError(w, "relay returned an invalid validation response", 502)
			return
		}
		if !result.Valid {
			utils.WriteError(w, "relay token is invalid", 400)
			return
		}
		utils.WriteJSON(w, 200, result)
	}
}
