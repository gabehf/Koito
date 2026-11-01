package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/memkv"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/db/sqlite"
	"github.com/gabehf/koito/internal/mbz"
	"github.com/gabehf/koito/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

type profileTestApp struct {
	store  *sqlite.Sqlite
	router http.Handler
	cookie string
}

func newProfileTestApp(t *testing.T) *profileTestApp {
	t.Helper()
	for _, id := range []int{1, 2, 3} {
		memkv.Store.Delete(fmt.Sprintf("now-playing:%d", id))
	}
	t.Cleanup(func() {
		for _, id := range []int{1, 2, 3} {
			memkv.Store.Delete(fmt.Sprintf("now-playing:%d", id))
		}
	})
	s, err := sqlite.NewInMemory()
	require.NoError(t, err)
	t.Cleanup(func() { s.Close(context.Background()) })
	admin, err := s.SaveUser(context.Background(), db.SaveUserOpts{Username: "administrator", Password: "test-password", Role: models.UserRoleAdmin})
	require.NoError(t, err)
	require.NoError(t, s.Exec(`INSERT INTO listening_profiles(id,slug,display_name,created_at) VALUES (2,'second','Second',123),(3,'empty','Empty',123)`))
	for _, k := range []struct {
		key string
		id  int32
	}{{"first-key", 1}, {"phone-key", 1}, {"second-key", 2}} {
		_, err := s.SaveApiKey(context.Background(), db.SaveApiKeyOpts{Key: k.key, ProfileID: k.id})
		require.NoError(t, err)
	}
	session, err := s.SaveSession(context.Background(), admin.ID, time.Now().Add(time.Hour), false)
	require.NoError(t, err)
	router := chi.NewRouter()
	var ready atomic.Bool
	ready.Store(true)
	bindRoutes(router, &ready, s, &mbz.MbzErrorCaller{})
	return &profileTestApp{s, router, session.ID.String()}
}
func (a *profileTestApp) request(method, url, body, key string, admin bool) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, "http://koito.test"+url, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	if key != "" {
		r.Header.Set("Authorization", "Token "+key)
	}
	if admin {
		r.AddCookie(&http.Cookie{Name: "koito_session", Value: a.cookie})
	}
	w := httptest.NewRecorder()
	a.router.ServeHTTP(w, r)
	return w
}
func listenBody(kind, track string, ts int64) string {
	return fmt.Sprintf(`{"listen_type":%q,"payload":[{"listened_at":%d,"track_metadata":{"artist_name":"Shared Artist","release_name":"Shared Album","track_name":%q,"additional_info":{"duration":180}}}]}`, kind, ts, track)
}
func getJSON(t *testing.T, a *profileTestApp, path string) map[string]any {
	t.Helper()
	w := a.request("GET", path, "", "", false)
	require.Equal(t, 200, w.Code, w.Body.String())
	var result map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	return result
}
func TestProfileBackendIsolationAndIngestion(t *testing.T) {
	a := newProfileTestApp(t)
	ts := time.Now().Add(-time.Hour).Unix()
	// Simultaneous identical music is two distinct scrobbles, not a coalesced request.
	var wg sync.WaitGroup
	codes := make(chan int, 2)
	for _, key := range []string{"first-key", "second-key"} {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			w := a.request("POST", "/apis/listenbrainz/1/submit-listens", listenBody("single", "Shared Track", ts), key, false)
			codes <- w.Code
		}(key)
	}
	wg.Wait()
	close(codes)
	for code := range codes {
		require.Equal(t, 200, code)
	}
	w := a.request("POST", "/apis/listenbrainz/1/submit-listens", listenBody("single", "Shared Track", ts+1), "phone-key", false)
	require.Equal(t, 200, w.Code, w.Body.String())
	w = a.request("POST", "/apis/listenbrainz/1/submit-listens", listenBody("single", "Other Track", ts+2), "second-key", false)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Equal(t, float64(2), getJSON(t, a, "/apis/web/v1/profiles/default/stats")["listen_count"])
	require.Equal(t, float64(2), getJSON(t, a, "/apis/web/v1/profiles/second/stats")["listen_count"])
	require.Equal(t, float64(1), getJSON(t, a, "/apis/web/v1/profiles/default/stats")["track_count"])
	require.Equal(t, float64(2), getJSON(t, a, "/apis/web/v1/profiles/second/stats")["track_count"])
	empty := getJSON(t, a, "/apis/web/v1/profiles/empty/stats")
	require.Equal(t, float64(0), empty["listen_count"])
	require.Equal(t, float64(0), empty["avg_daily_plays"])
	require.Equal(t, 409, a.request("GET", "/apis/web/v1/stats", "", "", false).Code)
	require.Equal(t, 404, a.request("GET", "/apis/web/v1/profiles/missing/stats", "", "", false).Code)
	for _, slug := range []string{"default", "second"} {
		prefix := "/apis/web/v1/profiles/" + slug
		for _, endpoint := range []string{"/listens", "/top/tracks", "/top/albums", "/top/artists", "/first-activity", "/listen-activity", "/summary?year=" + fmt.Sprint(time.Now().Year())} {
			w := a.request("GET", prefix+endpoint, "", "", false)
			require.Equal(t, 200, w.Code, endpoint+" "+w.Body.String())
		}
	}
	shared := getJSON(t, a, "/apis/web/v1/profiles/default/track/1")
	require.Equal(t, float64(2), shared["listen_count"])
	require.Equal(t, float64(360), shared["time_listened"])
	require.Equal(t, float64(1), getJSON(t, a, "/apis/web/v1/profiles/second/track/1")["listen_count"])
	require.Equal(t, 404, a.request("GET", "/apis/web/v1/profiles/default/track/2", "", "", false).Code)
	require.Equal(t, float64(2), getJSON(t, a, "/apis/web/v1/profiles/default/summary?period=all_time")["plays"])
	// A new playing-now track has no recorded listen yet, but must be visible only here.
	w = a.request("POST", "/apis/listenbrainz/1/submit-listens", listenBody("playing_now", "Unplayed Track", 0), "second-key", false)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Equal(t, true, getJSON(t, a, "/apis/web/v1/profiles/second/now-playing")["currently_playing"])
	require.Equal(t, false, getJSON(t, a, "/apis/web/v1/profiles/default/now-playing")["currently_playing"])
	search := getJSON(t, a, "/apis/web/v1/profiles/default/search?q=Other")
	require.Len(t, search["tracks"], 0)
	search = getJSON(t, a, "/apis/web/v1/profiles/second/search?q=Other")
	require.Len(t, search["tracks"], 1)
	// Public JSON contains metadata only, never ingest credentials.
	w = a.request("GET", "/apis/web/v1/profiles", "", "", false)
	require.Equal(t, 200, w.Code)
	require.NotContains(t, w.Body.String(), "first-key")
	// Deleting a single profile's shared listen cannot erase the other's.
	w = a.request("DELETE", fmt.Sprintf("/apis/web/v1/profiles/default/listens?track_id=1&unix=%d", ts), "", "", true)
	require.Equal(t, 204, w.Code)
	require.Equal(t, float64(1), getJSON(t, a, "/apis/web/v1/profiles/default/stats")["listen_count"])
	require.Equal(t, float64(2), getJSON(t, a, "/apis/web/v1/profiles/second/stats")["listen_count"])
}
func TestProfileBackendAdminAndSlugChanges(t *testing.T) {
	a := newProfileTestApp(t)
	// Vite preserves the browser Host so same-origin admin login works through its proxy.
	loginRequest := httptest.NewRequest("POST", "http://localhost:5173/apis/web/v1/login", strings.NewReader(`{"username":"administrator","password":"test-password"}`))
	loginRequest.Header.Set("Origin", "http://localhost:5173")
	loginResponse := httptest.NewRecorder()
	a.router.ServeHTTP(loginResponse, loginRequest)
	require.Equal(t, 204, loginResponse.Code, loginResponse.Body.String())
	require.Equal(t, http.SameSiteStrictMode, loginResponse.Result().Cookies()[0].SameSite)
	for _, path := range []string{"/apis/web/v1/data", "/apis/web/v1/artist/1", "/apis/web/v1/profiles/default/user/apikeys/1"} {
		require.Equal(t, 401, a.request("DELETE", path, "", "first-key", false).Code)
		require.Equal(t, 401, a.request("DELETE", path, "", "", false).Code)
	}
	require.Equal(t, 401, a.request("GET", "/apis/web/v1/profiles/default/export", "", "first-key", false).Code)
	req := httptest.NewRequest("PATCH", "http://koito.test/apis/web/v1/profiles/default", strings.NewReader(`{"slug":"stolen"}`))
	req.Header.Set("Origin", "https://evil.test")
	req.AddCookie(&http.Cookie{Name: "koito_session", Value: a.cookie})
	w := httptest.NewRecorder()
	a.router.ServeHTTP(w, req)
	require.Equal(t, 403, w.Code)
	for _, body := range []string{`{"slug":"admin"}`, `{"slug":"assets"}`, `{"slug":"second"}`, `{"slug":"Invalid_Name"}`} {
		require.Equal(t, 400, a.request("PATCH", "/apis/web/v1/profiles/default", body, "", true).Code)
	}
	w = a.request("PATCH", "/apis/web/v1/profiles/default", `{"slug":"everton","display_name":"Everton \"Music\""}`, "", true)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Equal(t, 404, a.request("GET", "/apis/web/v1/profiles/default", "", "", false).Code)
	p := getJSON(t, a, "/apis/web/v1/profiles/everton")
	require.Equal(t, float64(1), p["id"])
	ts := time.Now().Add(-time.Hour).Unix()
	w = a.request("POST", "/apis/listenbrainz/1/submit-listens", listenBody("single", "Shared Track", ts), "first-key", false)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Equal(t, float64(1), getJSON(t, a, "/apis/web/v1/profiles/everton/stats")["listen_count"])
	w = a.request("GET", "/apis/web/v1/profiles/everton/export", "", "", true)
	require.Equal(t, 200, w.Code, w.Body.String())
	var exported map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &exported))
	require.Equal(t, `Everton "Music"`, exported["user"])
	require.Len(t, exported["listens"], 1)
	w = a.request("GET", "/apis/web/v1/profiles/second/export", "", "", true)
	require.Equal(t, 200, w.Code)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &exported))
	require.Len(t, exported["listens"], 0)
	// A key can only be revoked through its own selected profile.
	require.Equal(t, 404, a.request("DELETE", "/apis/web/v1/profiles/second/user/apikeys/1", "", "", true).Code)
	require.Equal(t, 204, a.request("DELETE", "/apis/web/v1/profiles/everton/user/apikeys/1", "", "", true).Code)
	require.Equal(t, 401, a.request("POST", "/apis/listenbrainz/1/submit-listens", listenBody("single", "Shared Track", ts+1), "first-key", false).Code)
	require.Equal(t, float64(1), getJSON(t, a, "/apis/web/v1/profiles/everton/stats")["listen_count"])
}

func TestProfileImportAndRelay(t *testing.T) {
	a := newProfileTestApp(t)
	dir := filepath.Join(cfg.ConfigDir(), "import", "2")
	require.NoError(t, os.MkdirAll(dir, 0744))
	name := "maloja_profile_backend_test.json"
	file := filepath.Join(dir, name)
	body := fmt.Sprintf(`{"scrobbles":[{"time":%d,"track":{"artists":["Imported Artist"],"title":"Imported Track","album":{"albumtitle":"Imported Album"}}}]}`, time.Now().Add(-time.Hour).Unix())
	require.NoError(t, os.WriteFile(file, []byte(body), 0600))
	t.Cleanup(func() { os.Remove(file); os.Remove(filepath.Join(cfg.ConfigDir(), "import_complete", "2", name)) })
	reqBody := `{"filename":"` + name + `"}`
	require.Equal(t, 401, a.request("POST", "/apis/web/v1/profiles/second/import", reqBody, "second-key", false).Code)
	require.Equal(t, 400, a.request("POST", "/apis/web/v1/profiles/second/import", `{"filename":"../secret"}`, "", true).Code)
	w := a.request("POST", "/apis/web/v1/profiles/second/import", reqBody, "", true)
	require.Equal(t, 204, w.Code, w.Body.String())
	require.Equal(t, float64(0), getJSON(t, a, "/apis/web/v1/profiles/default/stats")["listen_count"])
	require.Equal(t, float64(1), getJSON(t, a, "/apis/web/v1/profiles/second/stats")["listen_count"])
	_, err := os.Stat(filepath.Join(cfg.ConfigDir(), "import_complete", "2", name))
	require.NoError(t, err)
	received := make(chan string, 4)
	relay := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- r.Header.Get("Authorization")
		w.WriteHeader(200)
	}))
	defer relay.Close()
	relayBody := fmt.Sprintf(`{"url":%q,"token":"external-token"}`, relay.URL)
	require.Equal(t, 401, a.request("PUT", "/apis/web/v1/profiles/default/relay", relayBody, "first-key", false).Code)
	require.Equal(t, 204, a.request("PUT", "/apis/web/v1/profiles/default/relay", relayBody, "", true).Code)
	ts := time.Now().Add(-time.Hour).Unix()
	require.Equal(t, 200, a.request("POST", "/apis/listenbrainz/1/submit-listens", listenBody("single", "Relay Track", ts), "first-key", false).Code)
	select {
	case token := <-received:
		require.Equal(t, "Token external-token", token)
	case <-time.After(2 * time.Second):
		t.Fatal("profile relay not called")
	}
	require.Equal(t, 200, a.request("POST", "/apis/listenbrainz/1/submit-listens", listenBody("single", "Relay Track", ts), "second-key", false).Code)
	select {
	case <-received:
		t.Fatal("another profile was forwarded to the relay")
	case <-time.After(100 * time.Millisecond):
	}
	require.Equal(t, 204, a.request("DELETE", "/apis/web/v1/profiles/default/relay", "", "", true).Code)
}
