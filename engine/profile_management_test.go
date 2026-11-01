package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gabehf/koito/imagecache"
	"github.com/gabehf/koito/internal/catalog"
	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func (a *profileTestApp) upload(t *testing.T, method, url string, fields map[string]string, field, name string, data []byte, admin bool) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for k, v := range fields {
		require.NoError(t, writer.WriteField(k, v))
	}
	if field != "" {
		part, err := writer.CreateFormFile(field, name)
		require.NoError(t, err)
		_, err = part.Write(data)
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())
	r := httptest.NewRequest(method, "http://koito.test"+url, &body)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	if admin {
		r.AddCookie(&http.Cookie{Name: "koito_session", Value: a.cookie})
	}
	w := httptest.NewRecorder()
	a.router.ServeHTTP(w, r)
	return w
}
func TestProfileManagementLifecycle(t *testing.T) {
	a := newProfileTestApp(t)
	ctx := context.Background()
	body := `{"display_name":"New Listener","slug":"new-listener","avatar_icon":"headphones"}`
	for _, key := range []string{"", "first-key"} {
		require.Equal(t, 401, a.request("POST", "/apis/web/v1/profiles", body, key, false).Code)
	}
	w := a.request("POST", "/apis/web/v1/profiles", body, "", true)
	require.Equal(t, 201, w.Code, w.Body.String())
	require.NotContains(t, w.Body.String(), `"key"`)
	var profile models.ListeningProfile
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &profile))
	require.Equal(t, "headphones", *profile.AvatarIcon)
	keys, err := a.store.GetApiKeysByProfileID(ctx, profile.ID)
	require.NoError(t, err)
	require.Len(t, keys, 1)
	require.Len(t, keys[0].Key, 48)
	require.Equal(t, 400, a.request("POST", "/apis/web/v1/profiles", body, "", true).Code)
	for _, slug := range []string{"admin", "assets", "user--name", "Bad", "apis"} {
		require.Equal(t, 400, a.request("POST", "/apis/web/v1/profiles", fmt.Sprintf(`{"display_name":"Test","slug":%q}`, slug), "", true).Code)
	}
	require.Equal(t, 400, a.request("POST", "/apis/web/v1/profiles", `{"display_name":"Test","slug":"bad-icon","avatar_icon":"unknown"}`, "", true).Code)
	profiles, err := a.store.ListListeningProfiles(ctx)
	require.NoError(t, err)
	require.Len(t, profiles, 4)
	ts := time.Now().Add(-time.Hour).Unix()
	for _, key := range []string{keys[0].Key, "second-key"} {
		require.Equal(t, 200, a.request("POST", "/apis/listenbrainz/1/submit-listens", listenBody("single", "Shared", ts), key, false).Code)
	}
	require.Equal(t, 200, a.request("POST", "/apis/listenbrainz/1/submit-listens", listenBody("playing_now", "Shared", 0), keys[0].Key, false).Code)
	prefix := "/apis/web/v1/profiles/new-listener"
	for _, method := range []string{"GET", "DELETE"} {
		endpoint := prefix
		if method == "GET" {
			endpoint += "/management"
		}
		require.Equal(t, 401, a.request(method, endpoint, `{"display_name":"New Listener","listen_count":1}`, keys[0].Key, false).Code)
	}
	require.Equal(t, float64(1), func() float64 {
		w := a.request("GET", prefix+"/management", "", "", true)
		require.Equal(t, 200, w.Code)
		var result map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
		return result["listen_count"].(float64)
	}())
	require.Equal(t, 409, a.request("DELETE", prefix, `{"display_name":"New Listener","listen_count":0}`, "", true).Code)
	require.Equal(t, 400, a.request("DELETE", prefix, `{"display_name":"Wrong","listen_count":1}`, "", true).Code)
	require.Equal(t, 204, a.request("DELETE", prefix, `{"display_name":"New Listener","listen_count":1}`, "", true).Code)
	require.Equal(t, 404, a.request("GET", prefix, "", "", false).Code)
	require.Equal(t, 401, a.request("GET", "/apis/listenbrainz/1/validate-token", "", keys[0].Key, false).Code)
	require.Equal(t, float64(1), getJSON(t, a, "/apis/web/v1/profiles/second/stats")["listen_count"])
	// Shared catalog survives profile deletion.
	_, err = a.store.GetTrack(ctx, db.GetTrackOpts{ID: 1})
	require.NoError(t, err)
	// Deleting the last profile leaves the administrator and session available.
	profiles, err = a.store.ListListeningProfiles(ctx)
	require.NoError(t, err)
	for _, p := range profiles {
		count, err := a.store.ProfileListenCount(ctx, p.ID)
		require.NoError(t, err)
		require.Equal(t, 204, a.request("DELETE", "/apis/web/v1/profiles/"+p.Slug, fmt.Sprintf(`{"display_name":%q,"listen_count":%d}`, p.DisplayName, count), "", true).Code)
	}
	require.Equal(t, 200, a.request("GET", "/apis/web/v1/user", "", "", true).Code)
	require.JSONEq(t, `[]`, a.request("GET", "/apis/web/v1/profiles", "", "", false).Body.String())
	require.Equal(t, 201, a.request("POST", "/apis/web/v1/profiles", body, "", true).Code)
}
func TestProfileManagementAtomicInitialKey(t *testing.T) {
	a := newProfileTestApp(t)
	// Force the second insert to fail: the profile must roll back too.
	_, err := a.store.CreateListeningProfile(context.Background(), "Rollback", "rollback", "", "first-key")
	require.Error(t, err)
	_, err = a.store.GetListeningProfile(context.Background(), "rollback")
	require.ErrorIs(t, err, db.ErrNotFound)
}
func TestProfileAvatarLifecycle(t *testing.T) {
	a := newProfileTestApp(t)
	var pngData bytes.Buffer
	require.NoError(t, png.Encode(&pngData, image.NewRGBA(image.Rect(0, 0, 32, 32))))
	prefix := "/apis/web/v1/profiles/second"
	require.Equal(t, 401, a.upload(t, "PUT", prefix+"/avatar", nil, "image", "photo.png", pngData.Bytes(), false).Code)
	require.Equal(t, 400, a.upload(t, "PUT", prefix+"/avatar", nil, "image", "fake.png", []byte("not an image"), true).Code)
	require.Equal(t, 400, a.upload(t, "PUT", prefix+"/avatar", map[string]string{"icon": "invalid"}, "", "", nil, true).Code)
	w := a.upload(t, "PUT", prefix+"/avatar", map[string]string{"icon": "music"}, "image", "photo.png", pngData.Bytes(), true)
	require.Equal(t, 200, w.Code, w.Body.String())
	p, err := a.store.GetListeningProfile(context.Background(), "second")
	require.NoError(t, err)
	require.NotNil(t, p.AvatarImageID)
	id := uuid.MustParse(*p.AvatarImageID)
	require.FileExists(t, imagecache.BuildImagePath(id, imagecache.ImageSizeSource))
	require.NoError(t, catalog.PruneOrphanedImages(context.Background(), a.store))
	require.FileExists(t, imagecache.BuildImagePath(id, imagecache.ImageSizeSource))
	imageRequest := httptest.NewRequest("GET", "http://koito.test/image/"+id.String()+"/128x128.webp", nil)
	imageResponse := httptest.NewRecorder()
	a.router.ServeHTTP(imageResponse, imageRequest)
	require.Equal(t, 200, imageResponse.Code)
	// Keep a reference from another profile: removal cannot delete a shared file.
	require.NoError(t, a.store.SetProfileAvatar(context.Background(), 3, p.AvatarImageID, nil))
	require.Equal(t, 200, a.request("DELETE", prefix+"/avatar", "", "", true).Code)
	require.FileExists(t, imagecache.BuildImagePath(id, imagecache.ImageSizeSource))
	require.Equal(t, 204, a.request("DELETE", "/apis/web/v1/profiles/empty", `{"display_name":"Empty","listen_count":0}`, "", true).Code)
	_, err = os.Stat(imagecache.BuildImagePath(id, imagecache.ImageSizeSource))
	require.True(t, os.IsNotExist(err))
	// Icon-only avatar is public metadata; arbitrary image IDs cannot be assigned.
	require.Equal(t, 200, a.upload(t, "PUT", prefix+"/avatar", map[string]string{"icon": "radio"}, "", "", nil, true).Code)
	result := getJSON(t, a, prefix)
	require.Equal(t, "radio", result["avatar_icon"])
	require.NotContains(t, result, "avatar_image_id")
}
func TestProfileImportUpload(t *testing.T) {
	a := newProfileTestApp(t)
	data := []byte(`{"scrobbles":[{"time":1700000000,"track":{"artists":["Uploaded Artist"],"title":"Uploaded Track","album":{"albumtitle":"Uploaded Album"}}}]}`)
	prefix := "/apis/web/v1/profiles/second/import"
	require.Equal(t, 401, a.upload(t, "POST", prefix, map[string]string{"format": "maloja"}, "file", "../../evil.json", data, false).Code)
	require.Equal(t, 400, a.upload(t, "POST", prefix, map[string]string{"format": "wrong"}, "file", "test.json", data, true).Code)
	w := a.upload(t, "POST", prefix, map[string]string{"format": "maloja"}, "file", "../../evil.json", data, true)
	require.Equal(t, 204, w.Code, w.Body.String())
	require.Equal(t, float64(1), getJSON(t, a, "/apis/web/v1/profiles/second/stats")["listen_count"])
	require.Equal(t, float64(0), getJSON(t, a, "/apis/web/v1/profiles/default/stats")["listen_count"])
	completeDir := filepath.Join(cfg.ConfigDir(), "import_complete", "2")
	entries, err := os.ReadDir(completeDir)
	require.NoError(t, err)
	found := false
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), "-maloja.json") {
			found = true
			t.Cleanup(func() { _ = os.Remove(filepath.Join(completeDir, entry.Name())) })
		}
	}
	require.True(t, found)
	// An uploaded Koito file with an empty artist list returns 400, not a panic.
	w = a.upload(t, "POST", prefix, map[string]string{"format": "koito"}, "file", "bad.json", []byte(`{"version":"1","listens":[{"artists":[]}]}`), true)
	require.Equal(t, 400, w.Code, w.Body.String())
	pending, err := os.ReadDir(filepath.Join(cfg.ConfigDir(), "import", "2"))
	require.NoError(t, err)
	for _, entry := range pending {
		require.False(t, strings.HasSuffix(entry.Name(), "-koito.json"))
	}
	// Other-origin cookie requests cannot upload or create profiles.
	req := httptest.NewRequest("POST", "http://koito.test/apis/web/v1/profiles", io.NopCloser(strings.NewReader(`{}`)))
	req.Header.Set("Origin", "http://external.test")
	req.AddCookie(&http.Cookie{Name: "koito_session", Value: a.cookie})
	res := httptest.NewRecorder()
	a.router.ServeHTTP(res, req)
	require.Equal(t, 403, res.Code)
}

// The account and the listening profile have separate names and lifecycles.
// Verify login with the updated credentials, not only a successful PATCH.
func TestAdminAccountChangesKeepListeningProfile(t *testing.T) {
	a := newProfileTestApp(t)
	prefix := "/apis/web/v1"
	before := getJSON(t, a, prefix+"/profiles/default")
	require.Equal(t, 204, a.request("PATCH", prefix+"/user", `{"username":"renamed-admin"}`, "", true).Code)
	me := a.request("GET", prefix+"/user", "", "", true)
	require.Equal(t, 200, me.Code)
	require.Contains(t, me.Body.String(), `"username":"renamed-admin"`)
	require.Equal(t, 204, a.request("POST", prefix+"/login", `{"username":"renamed-admin","password":"test-password"}`, "", false).Code)
	require.Equal(t, 401, a.request("POST", prefix+"/login", `{"username":"administrator","password":"test-password"}`, "", false).Code)
	require.Equal(t, 204, a.request("PATCH", prefix+"/user", `{"password":"changed-password"}`, "", true).Code)
	require.Equal(t, 401, a.request("POST", prefix+"/login", `{"username":"renamed-admin","password":"test-password"}`, "", false).Code)
	require.Equal(t, 204, a.request("POST", prefix+"/login", `{"username":"renamed-admin","password":"changed-password"}`, "", false).Code)
	require.Equal(t, before, getJSON(t, a, prefix+"/profiles/default"))
	require.Equal(t, 200, a.request("GET", "/apis/listenbrainz/1/validate-token", "", "first-key", false).Code)
	// Rename the listening identity independently of the login name.
	require.Equal(t, 200, a.request("PATCH", prefix+"/profiles/default", `{"display_name":"Public listener","slug":"listener"}`, "", true).Code)
	require.Equal(t, 404, a.request("GET", prefix+"/profiles/default", "", "", false).Code)
	require.Equal(t, 204, a.request("POST", prefix+"/login", `{"username":"renamed-admin","password":"changed-password"}`, "", false).Code)
	require.Equal(t, before["id"], getJSON(t, a, prefix+"/profiles/listener")["id"])
}
