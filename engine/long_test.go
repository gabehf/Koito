package engine_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gabehf/koito/engine/handlers"
	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var session string
var apikey string
var loginOnce sync.Once
var apikeyOnce sync.Once

func login(t *testing.T) {
	loginOnce.Do(func() {
		formdata := url.Values{}
		formdata.Set("username", cfg.DefaultUsername())
		formdata.Set("password", cfg.DefaultPassword())
		encoded := formdata.Encode()
		resp, err := http.DefaultClient.Post(host()+"/apis/web/v1/login", "application/x-www-form-urlencoded", strings.NewReader(encoded))
		respBytes, _ := io.ReadAll(resp.Body)
		t.Logf("Login request response: %s - %s", resp.Status, respBytes)
		require.NoError(t, err)
		require.Len(t, resp.Cookies(), 1)
		session = resp.Cookies()[0].Value
		require.NotEmpty(t, session)
	})
}

func makeAuthRequest(t *testing.T, session, method, endpoint string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, host()+endpoint, body)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{
		Name:  "koito_session",
		Value: session,
	})
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	t.Logf("Making request to %s with session: %s", endpoint, session)
	return http.DefaultClient.Do(req)
}

// Expects a valid session
func getApiKey(t *testing.T, session string) {
	apikeyOnce.Do(func() {
		resp, err := makeAuthRequest(t, session, "GET", "/apis/web/v1/user/apikeys", nil)
		require.NoError(t, err)
		var keys []models.ApiKey
		err = json.NewDecoder(resp.Body).Decode(&keys)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(keys), 1)
		apikey = keys[0].Key
	})
}

func truncateTestData(t *testing.T) {
	err := store.Exec(context.Background(),
		`TRUNCATE 
		artists, 
		artist_aliases,
		tracks, 
		artist_tracks, 
		releases, 
		artist_releases, 
		release_aliases, 
		listens 
		RESTART IDENTITY CASCADE`)
	require.NoError(t, err)
}

func doSubmitListens(t *testing.T) {
	login(t)
	getApiKey(t, session)
	truncateTestData(t)
	bodies := []string{fmt.Sprintf(`{
		"listen_type": "single",
		"payload": [
			{
				"listened_at": %d,
				"track_metadata": {
					"additional_info": {
						"artist_mbids": [
							"efc787f0-046f-4a60-beff-77b398c8cdf4"
						],
						"artist_names": [
							"さユり"
						],
						"duration_ms": 275960,
						"recording_mbid": "21524d55-b1f8-45d1-b172-976cba447199",
						"release_group_mbid": "3281e0d9-fa44-4337-a8ce-6f264beeae16",
						"release_mbid": "eb790e90-0065-4852-b47d-bbeede4aa9fc",
						"submission_client": "navidrome",
						"submission_client_version": "0.56.1 (fa2cf362)"
					},
					"artist_name": "さユり",
					"release_name": "酸欠少女",
					"track_name": "花の塔"
				}
			}
		]
	}`, time.Now().Add(-2*time.Hour).Unix()), // yesterday
		fmt.Sprintf(`{
		"listen_type": "single",
		"payload": [
			{
				"listened_at": %d,
				"track_metadata": {
					"additional_info": {
						"artist_mbids": [
							"80b3cb83-b7a3-4f79-ad42-8325cefb3626"
						],
						"artist_names": [
							"キタニタツヤ"
						],
						"duration_ms": 197270,
						"recording_mbid": "4e909c21-e7a8-404d-b75a-0c8c2926efb0",
						"release_group_mbid": "89069d92-e495-462c-b189-3431551868ed",
						"release_mbid": "e16a49d6-77f3-4d73-b93c-cac855ce6ad5",
						"submission_client": "navidrome",
						"submission_client_version": "0.56.1 (fa2cf362)"
					},
					"artist_name": "キタニタツヤ",
					"release_name": "Where Our Blue Is",
					"track_name": "Where Our Blue Is"
				}
			}
		]
	}`, time.Now().Unix()),
		fmt.Sprintf(`{
		"listen_type": "single",
		"payload": [
			{
				"listened_at": %d,
				"track_metadata": {
					"additional_info": {
						"artist_mbids": [
							"1262ab85-308b-46e7-b0b5-91fef8e46b62"
						],
						"artist_names": [
							"ネクライトーキー"
						],
						"duration_ms": 241560,
						"recording_mbid": "8eec4f3f-a059-4217-aad1-fbf82e33e756",
						"release_group_mbid": "14f1aff0-dd19-4b42-82dd-720386b6d4c1",
						"release_mbid": "7762d7af-7b6c-454f-977e-1b261743e265",
						"submission_client": "navidrome",
						"submission_client_version": "0.56.1 (fa2cf362)"
					},
					"artist_name": "ネクライトーキー",
					"release_name": "ONE!",
					"track_name": "こんがらがった！"
				}
			}
		]
	}`, time.Now().Add(-1*time.Hour).Unix())}
	for _, body := range bodies {
		req, err := http.NewRequest("POST", host()+"/apis/listenbrainz/1/submit-listens", strings.NewReader(body))
		require.NoError(t, err)
		req.Header.Add("Authorization", fmt.Sprintf("Token %s", apikey))
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		respBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, `{"status": "ok"}`, string(respBytes))
	}
}

func TestGetters(t *testing.T) {
	t.Run("Submit Listens", doSubmitListens)
	// Artist was saved
	resp, err := http.DefaultClient.Get(host() + "/apis/web/v1/artist?id=1")
	assert.NoError(t, err)
	var artist models.Artist
	err = json.NewDecoder(resp.Body).Decode(&artist)
	require.NoError(t, err)
	assert.Equal(t, "さユり", artist.Name)

	// Album was saved
	resp, err = http.DefaultClient.Get(host() + "/apis/web/v1/album?id=1")
	assert.NoError(t, err)
	var album models.Album
	err = json.NewDecoder(resp.Body).Decode(&album)
	require.NoError(t, err)
	assert.Equal(t, "酸欠少女", album.Title)

	// Track was saved
	resp, err = http.DefaultClient.Get(host() + "/apis/web/v1/track?id=1")
	assert.NoError(t, err)
	var track models.Track
	err = json.NewDecoder(resp.Body).Decode(&track)
	require.NoError(t, err)
	assert.Equal(t, "花の塔", track.Title)

	// Listen was saved
	resp, err = http.DefaultClient.Get(host() + "/apis/web/v1/listens")
	assert.NoError(t, err)
	var listens db.PaginatedResponse[models.Listen]
	err = json.NewDecoder(resp.Body).Decode(&listens)
	require.NoError(t, err)
	require.Len(t, listens.Items, 3)
	assert.EqualValues(t, 2, listens.Items[0].Track.ID)
	assert.Equal(t, "Where Our Blue Is", listens.Items[0].Track.Title)

	resp, err = http.DefaultClient.Get(host() + "/apis/web/v1/top-artists")
	assert.NoError(t, err)
	var artists db.PaginatedResponse[models.Artist]
	err = json.NewDecoder(resp.Body).Decode(&artists)
	require.NoError(t, err)
	require.Len(t, artists.Items, 3)

	resp, err = http.DefaultClient.Get(host() + "/apis/web/v1/top-albums")
	assert.NoError(t, err)
	var albums db.PaginatedResponse[models.Album]
	err = json.NewDecoder(resp.Body).Decode(&albums)
	require.NoError(t, err)
	require.Len(t, albums.Items, 3)

	resp, err = http.DefaultClient.Get(host() + "/apis/web/v1/top-tracks")
	assert.NoError(t, err)
	var tracks db.PaginatedResponse[models.Track]
	err = json.NewDecoder(resp.Body).Decode(&tracks)
	require.NoError(t, err)
	require.Len(t, tracks.Items, 3)

	truncateTestData(t)
}

func TestMerge(t *testing.T) {

	t.Run("Submit Listens", doSubmitListens)

	resp, err := makeAuthRequest(t, session, "POST", "/apis/web/v1/merge/tracks?from_id=1&to_id=2", nil)
	require.NoError(t, err)
	require.Equal(t, 204, resp.StatusCode)

	resp, err = http.DefaultClient.Get(host() + "/apis/web/v1/track?id=2")
	require.NoError(t, err)
	var track models.Track
	err = json.NewDecoder(resp.Body).Decode(&track)
	require.NoError(t, err)
	assert.EqualValues(t, 2, track.ListenCount)

	truncateTestData(t)

	t.Run("Submit Listens", doSubmitListens)

	resp, err = makeAuthRequest(t, session, "POST", "/apis/web/v1/merge/artists?from_id=1&to_id=2", nil)
	require.NoError(t, err)
	require.Equal(t, 204, resp.StatusCode)

	resp, err = http.DefaultClient.Get(host() + "/apis/web/v1/artist?id=2")
	require.NoError(t, err)
	var artist models.Artist
	err = json.NewDecoder(resp.Body).Decode(&artist)
	require.NoError(t, err)
	assert.EqualValues(t, 2, artist.ListenCount)

	truncateTestData(t)

	t.Run("Submit Listens", doSubmitListens)

	resp, err = makeAuthRequest(t, session, "POST", "/apis/web/v1/merge/albums?from_id=1&to_id=2", nil)
	require.NoError(t, err)
	require.Equal(t, 204, resp.StatusCode)

	resp, err = http.DefaultClient.Get(host() + "/apis/web/v1/album?id=2")
	require.NoError(t, err)
	var album models.Album
	err = json.NewDecoder(resp.Body).Decode(&album)
	require.NoError(t, err)
	assert.EqualValues(t, 2, album.ListenCount)

	truncateTestData(t)
}

func TestValidateToken(t *testing.T) {
	login(t)
	getApiKey(t, session)

	req, err := http.NewRequest("GET", host()+"/apis/listenbrainz/1/validate-token", nil)
	require.NoError(t, err)
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", apikey))
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	var actual handlers.LbzValidateResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&actual))
	t.Log(actual)
	var expected handlers.LbzValidateResponse
	expected.Code = 200
	expected.Message = "Token valid."
	expected.Valid = true
	expected.UserName = "test"
	assert.True(t, assert.ObjectsAreEqual(expected, actual))

	req, err = http.NewRequest("GET", host()+"/apis/listenbrainz/1/validate-token", nil)
	require.NoError(t, err)
	req.Header.Add("Authorization", "Token thisisasuperinvalidtoken")
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)

	req, err = http.NewRequest("GET", host()+"/apis/listenbrainz/1/validate-token", nil)
	require.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestDelete(t *testing.T) {

	t.Run("Submit Listens", doSubmitListens)

	resp, err := makeAuthRequest(t, session, "DELETE", "/apis/web/v1/artist?id=1", nil)
	require.NoError(t, err)
	require.Equal(t, 204, resp.StatusCode)

	resp, err = http.DefaultClient.Get(host() + "/apis/web/v1/artist?id=1")
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)

	resp, err = makeAuthRequest(t, session, "DELETE", "/apis/web/v1/album?id=1", nil)
	require.NoError(t, err)
	require.Equal(t, 204, resp.StatusCode)

	resp, err = http.DefaultClient.Get(host() + "/apis/web/v1/album?id=1")
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)

	resp, err = makeAuthRequest(t, session, "DELETE", "/apis/web/v1/track?id=1", nil)
	require.NoError(t, err)
	require.Equal(t, 204, resp.StatusCode)

	resp, err = http.DefaultClient.Get(host() + "/apis/web/v1/track?id=1")
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)

	truncateTestData(t)
}

func TestAliasesAndSearch(t *testing.T) {

	t.Run("Submit Listens", doSubmitListens)

	resp, err := makeAuthRequest(t, session, "POST", "/apis/web/v1/aliases?artist_id=1&alias=Sayuri", nil)
	require.NoError(t, err)
	require.Equal(t, 201, resp.StatusCode)

	resp, err = http.DefaultClient.Get(host() + "/apis/web/v1/aliases?artist_id=1")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	var actual []models.Alias
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&actual))
	require.Len(t, actual, 2)
	assert.Equal(t, actual[0].Alias, "さユり")
	assert.Equal(t, actual[0].Source, "Canonical")
	assert.Equal(t, actual[1].Alias, "Sayuri")
	assert.Equal(t, actual[1].Source, "Manual")

	resp, err = makeAuthRequest(t, session, "POST", "/apis/web/v1/aliases?album_id=1&alias=Sanketsu+Girl", nil)
	require.NoError(t, err)
	require.Equal(t, 201, resp.StatusCode)

	resp, err = http.DefaultClient.Get(host() + "/apis/web/v1/aliases?album_id=1")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	actual = nil
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&actual))
	require.Len(t, actual, 2)
	assert.Equal(t, actual[0].Alias, "酸欠少女")
	assert.Equal(t, actual[0].Source, "Canonical")
	assert.Equal(t, actual[1].Alias, "Sanketsu Girl")
	assert.Equal(t, actual[1].Source, "Manual")

	resp, err = makeAuthRequest(t, session, "POST", "/apis/web/v1/aliases?track_id=1&alias=Tower+of+Flower", nil)
	require.NoError(t, err)
	require.Equal(t, 201, resp.StatusCode)

	resp, err = makeAuthRequest(t, session, "POST", "/apis/web/v1/aliases/primary?track_id=1&alias=Tower+of+Flower", nil)
	require.NoError(t, err)
	require.Equal(t, 204, resp.StatusCode)

	resp, err = http.DefaultClient.Get(host() + "/apis/web/v1/track?id=1")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	var track models.Track
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&track))
	require.Len(t, actual, 2)
	assert.Equal(t, track.Title, "Tower of Flower")

	resp, err = makeAuthRequest(t, session, "POST", "/apis/web/v1/aliases/primary?artist_id=1&alias=Sayuri", nil)
	require.NoError(t, err)
	require.Equal(t, 204, resp.StatusCode)

	// make sure searching works with aliases

	resp, err = http.DefaultClient.Get(host() + "/apis/web/v1/search?q=Sanketsu")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	var results handlers.SearchResults
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&results))
	require.Len(t, results.Albums, 1)
	assert.Equal(t, results.Albums[0].Title, "酸欠少女")

	resp, err = http.DefaultClient.Get(host() + "/apis/web/v1/search?q=Sayuri")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	results = handlers.SearchResults{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&results))
	require.Len(t, results.Artists, 1)
	assert.Equal(t, results.Artists[0].Name, "Sayuri") // reflects the new primary alias

	truncateTestData(t)
}

func TestStats(t *testing.T) {
	// zeroes
	resp, err := http.DefaultClient.Get(host() + "/apis/web/v1/stats")
	t.Log(resp)
	require.NoError(t, err)

	t.Run("Submit Listens", doSubmitListens)

	resp, err = http.DefaultClient.Get(host() + "/apis/web/v1/stats")
	t.Log(resp)
	require.NoError(t, err)
	var actual handlers.StatsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&actual))
	assert.EqualValues(t, 3, actual.ListenCount)
	assert.EqualValues(t, 3, actual.TrackCount)
	assert.EqualValues(t, 3, actual.AlbumCount)
	assert.EqualValues(t, 3, actual.ArtistCount)
	assert.EqualValues(t, 11, actual.MinutesListened)
}

func TestListenActivity(t *testing.T) {

	// this test fails when run a bit after midnight
	// i'll figure out a better test later

	// t.Run("Submit Listens", doSubmitListens)

	// resp, err := http.DefaultClient.Get(host() + "/apis/web/v1/listen-activity?range=3")
	// t.Log(resp)
	// require.NoError(t, err)
	// var actual []db.ListenActivityItem
	// require.NoError(t, json.NewDecoder(resp.Body).Decode(&actual))
	// t.Log(actual)
	// require.Len(t, actual, 3)
	// assert.EqualValues(t, 3, actual[2].Listens)
}

func TestAuth(t *testing.T) {
	// logs in a new session
	formdata := url.Values{}
	formdata.Set("username", cfg.DefaultUsername())
	formdata.Set("password", cfg.DefaultPassword())
	encoded := formdata.Encode()
	resp, err := http.DefaultClient.Post(host()+"/apis/web/v1/login", "application/x-www-form-urlencoded", strings.NewReader(encoded))
	respBytes, _ := io.ReadAll(resp.Body)
	t.Logf("Login request response: %s - %s", resp.Status, respBytes)
	require.NoError(t, err)
	require.Len(t, resp.Cookies(), 1)
	s := resp.Cookies()[0].Value
	require.NotEmpty(t, s)

	// test update user
	req, err := http.NewRequest("PATCH", host()+"/apis/web/v1/user?username=new&password=supersecret", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{
		Name:  "koito_session",
		Value: s,
	})
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, 204, resp.StatusCode)

	// test /me with updated info
	req, err = http.NewRequest("GET", host()+"/apis/web/v1/user/me", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{
		Name:  "koito_session",
		Value: s,
	})
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	var me models.User
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&me))
	require.Equal(t, "new", me.Username)

	// login with old password fails
	formdata = url.Values{}
	formdata.Set("username", cfg.DefaultUsername())
	formdata.Set("password", cfg.DefaultPassword())
	encoded = formdata.Encode()
	resp, err = http.DefaultClient.Post(host()+"/apis/web/v1/login", "application/x-www-form-urlencoded", strings.NewReader(encoded))
	require.NoError(t, err)
	require.Equal(t, 401, resp.StatusCode)

	// reset update so other tests dont fail
	req, err = http.NewRequest("PATCH", host()+fmt.Sprintf("/apis/web/v1/user?username=%s&password=%s", cfg.DefaultUsername(), cfg.DefaultPassword()), nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{
		Name:  "koito_session",
		Value: s,
	})
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, 204, resp.StatusCode)

	// creates api key
	req, err = http.NewRequest("POST", host()+"/apis/web/v1/user/apikeys?label=testing", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{
		Name:  "koito_session",
		Value: s,
	})
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, 201, resp.StatusCode)
	var response struct {
		Key string `json:"key"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response))
	require.NotEmpty(t, response.Key)

	// validates api key
	req, err = http.NewRequest("GET", host()+"/apis/listenbrainz/1/validate-token", nil)
	require.NoError(t, err)
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", response.Key))
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	var actual handlers.LbzValidateResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&actual))
	var expected handlers.LbzValidateResponse
	expected.Code = 200
	expected.Message = "Token valid."
	expected.Valid = true
	expected.UserName = "test"
	assert.True(t, assert.ObjectsAreEqual(expected, actual))

	// changes api key label
	login(t) // i dont care about using the new session anymore
	resp, err = makeAuthRequest(t, s, "PATCH", "/apis/web/v1/user/apikeys?id=2&label=well+tested", nil)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	resp, err = makeAuthRequest(t, s, "GET", "/apis/web/v1/user/apikeys", nil)
	require.NoError(t, err)
	var keys []models.ApiKey
	err = json.NewDecoder(resp.Body).Decode(&keys)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(keys), 2)
	require.NotNil(t, keys[1].Label)
	assert.Equal(t, "well tested", keys[1].Label)

	// logs out
	req, err = http.NewRequest("POST", host()+"/apis/web/v1/logout", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{
		Name:  "koito_session",
		Value: s,
	})
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, 204, resp.StatusCode)

	// attempts to create an api key - unauthorized
	formdata = url.Values{}
	formdata.Set("label", "testing")
	encoded = formdata.Encode()
	req, err = http.NewRequest("POST", host()+"/apis/web/v1/user/apikeys", strings.NewReader(encoded))
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{
		Name:  "koito_session",
		Value: s,
	})
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, 401, resp.StatusCode)
}

func TestDeleteListen(t *testing.T) {
	login(t)
	getApiKey(t, session)

	truncateTestData(t)

	body := `{
		"listen_type": "single",
		"payload": [
			{
				"listened_at": 1749475719,
				"track_metadata": {
					"additional_info": {
						"artist_mbids": [
							"80b3cb83-b7a3-4f79-ad42-8325cefb3626"
						],
						"artist_names": [
							"キタニタツヤ"
						],
						"duration_ms": 197270,
						"recording_mbid": "4e909c21-e7a8-404d-b75a-0c8c2926efb0",
						"release_group_mbid": "89069d92-e495-462c-b189-3431551868ed",
						"release_mbid": "e16a49d6-77f3-4d73-b93c-cac855ce6ad5",
						"submission_client": "navidrome",
						"submission_client_version": "0.56.1 (fa2cf362)"
					},
					"artist_name": "キタニタツヤ",
					"release_name": "Where Our Blue Is",
					"track_name": "Where Our Blue Is"
				}
			}
		]
	}`

	req, err := http.NewRequest("POST", host()+"/apis/listenbrainz/1/submit-listens", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", apikey))
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, `{"status": "ok"}`, string(respBytes))

	resp, err = makeAuthRequest(t, session, "DELETE", "/apis/web/v1/listen?track_id=1&unix=1749475719", nil)
	require.NoError(t, err)
	require.Equal(t, 204, resp.StatusCode)

	// deletes are idempotent
	resp, err = makeAuthRequest(t, session, "DELETE", "/apis/web/v1/listen?track_id=1&unix=1749475719", nil)
	require.NoError(t, err)
	require.Equal(t, 204, resp.StatusCode)

	// listen is deleted
	resp, err = http.DefaultClient.Get(host() + "/apis/web/v1/track?id=1")
	require.NoError(t, err)
	var track models.Track
	err = json.NewDecoder(resp.Body).Decode(&track)
	require.NoError(t, err)
	assert.EqualValues(t, 0, track.ListenCount)
}

func TestArtistReplaceImage(t *testing.T) {

	t.Run("Submit Listens", doSubmitListens)

	buf := &bytes.Buffer{}
	mpw := multipart.NewWriter(buf)
	mpw.WriteField("artist_id", "1")
	w, err := mpw.CreateFormFile("image", path.Join("..", "test_assets", "yuu.jpg"))
	require.NoError(t, err)
	f, err := os.Open(path.Join("..", "test_assets", "yuu.jpg"))
	require.NoError(t, err)
	defer f.Close()
	_, err = io.Copy(w, f)
	require.NoError(t, err)
	require.NoError(t, mpw.Close())

	req, err := http.NewRequest("POST", host()+"/apis/web/v1/replace-image", buf)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{
		Name:  "koito_session",
		Value: session,
	})
	req.Header.Add("Content-Type", mpw.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	response := new(handlers.ReplaceImageResponse)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(response))
	require.NotEmpty(t, response.Image)
	newid, err := uuid.Parse(response.Image)
	require.NoError(t, err)

	a, err := store.GetArtist(context.Background(), db.GetArtistOpts{ID: 1})
	require.NoError(t, err)
	assert.NotNil(t, a.Image)
	assert.Equal(t, newid, *a.Image)
}

func TestAlbumReplaceImage(t *testing.T) {

	t.Run("Submit Listens", doSubmitListens)

	buf := &bytes.Buffer{}
	mpw := multipart.NewWriter(buf)
	mpw.WriteField("album_id", "1")
	w, err := mpw.CreateFormFile("image", path.Join("..", "test_assets", "yuu.jpg"))
	require.NoError(t, err)
	f, err := os.Open(path.Join("..", "test_assets", "yuu.jpg"))
	require.NoError(t, err)
	defer f.Close()
	_, err = io.Copy(w, f)
	require.NoError(t, err)
	require.NoError(t, mpw.Close())

	req, err := http.NewRequest("POST", host()+"/apis/web/v1/replace-image", buf)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{
		Name:  "koito_session",
		Value: session,
	})
	req.Header.Add("Content-Type", mpw.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	response := new(handlers.ReplaceImageResponse)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(response))
	require.NotEmpty(t, response.Image)
	newid, err := uuid.Parse(response.Image)
	require.NoError(t, err)

	a, err := store.GetAlbum(context.Background(), db.GetAlbumOpts{ID: 1})
	require.NoError(t, err)
	assert.NotNil(t, a.Image)
	assert.Equal(t, newid, *a.Image)
}

func TestSetPrimaryArtist(t *testing.T) {

	t.Run("Submit Listens", doSubmitListens)

	ctx := context.Background()

	// set and unset track primary artist

	formdata := url.Values{}
	formdata.Set("artist_id", "1")
	formdata.Set("track_id", "1")
	formdata.Set("is_primary", "false")
	body := formdata.Encode()
	resp, err := makeAuthRequest(t, session, "POST", "/apis/web/v1/artists/primary", strings.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, 204, resp.StatusCode)

	exists, err := store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM artist_tracks
      WHERE track_id = $1 AND artist_id = $2 AND is_primary = $3
    )`, 1, 1, false)
	require.NoError(t, err)
	assert.True(t, exists, "expected artist is_primary to be false")

	formdata = url.Values{}
	formdata.Set("artist_id", "1")
	formdata.Set("track_id", "1")
	formdata.Set("is_primary", "true")
	body = formdata.Encode()
	resp, err = makeAuthRequest(t, session, "POST", "/apis/web/v1/artists/primary", strings.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, 204, resp.StatusCode)

	exists, err = store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM artist_tracks
      WHERE track_id = $1 AND artist_id = $2 AND is_primary = $3
    )`, 1, 1, true)
	require.NoError(t, err)
	assert.True(t, exists, "expected artist is_primary to be true")

	// set and unset album primary artist

	formdata = url.Values{}
	formdata.Set("artist_id", "1")
	formdata.Set("album_id", "1")
	formdata.Set("is_primary", "false")
	body = formdata.Encode()
	resp, err = makeAuthRequest(t, session, "POST", "/apis/web/v1/artists/primary", strings.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, 204, resp.StatusCode)

	exists, err = store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM artist_releases
      WHERE release_id = $1 AND artist_id = $2 AND is_primary = $3
    )`, 1, 1, false)
	require.NoError(t, err)
	assert.True(t, exists, "expected artist is_primary to be false")

	formdata = url.Values{}
	formdata.Set("artist_id", "1")
	formdata.Set("album_id", "1")
	formdata.Set("is_primary", "true")
	body = formdata.Encode()
	resp, err = makeAuthRequest(t, session, "POST", "/apis/web/v1/artists/primary", strings.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, 204, resp.StatusCode)

	exists, err = store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM artist_releases
      WHERE release_id = $1 AND artist_id = $2 AND is_primary = $3
    )`, 1, 1, true)
	require.NoError(t, err)
	assert.True(t, exists, "expected artist is_primary to be true")

	// create a new track with multiple artists to make sure only one is primary at a time

	listenBody := `{
		"listen_type": "single",
		"payload": [
			{
				"listened_at": 1749475719,
				"track_metadata": {
					"additional_info": {
						"artist_names": [
							"Rat Tally",
							"Madeline Kenney"
						],
						"duration_ms": 197270,
						"submission_client": "navidrome",
						"submission_client_version": "0.56.1 (fa2cf362)"
					},
					"artist_name": "Rat Tally feat. Madeline Kenney",
					"release_name": "In My Car",
					"track_name": "In My Car"
				}
			}
		]
	}`

	req, err := http.NewRequest("POST", host()+"/apis/listenbrainz/1/submit-listens", strings.NewReader(listenBody))
	require.NoError(t, err)
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", apikey))
	req.Header.Add("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, `{"status": "ok"}`, string(respBytes))

	// set both artists as primary

	formdata = url.Values{}
	formdata.Set("artist_id", "4")
	formdata.Set("album_id", "4")
	formdata.Set("is_primary", "true")
	body = formdata.Encode()
	resp, err = makeAuthRequest(t, session, "POST", "/apis/web/v1/artists/primary", strings.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, 204, resp.StatusCode)
	formdata = url.Values{}
	formdata.Set("artist_id", "5")
	formdata.Set("album_id", "4")
	formdata.Set("is_primary", "true")
	body = formdata.Encode()
	resp, err = makeAuthRequest(t, session, "POST", "/apis/web/v1/artists/primary", strings.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, 204, resp.StatusCode)

	formdata = url.Values{}
	formdata.Set("artist_id", "4")
	formdata.Set("track_id", "4")
	formdata.Set("is_primary", "true")
	body = formdata.Encode()
	resp, err = makeAuthRequest(t, session, "POST", "/apis/web/v1/artists/primary", strings.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, 204, resp.StatusCode)
	formdata = url.Values{}
	formdata.Set("artist_id", "5")
	formdata.Set("track_id", "4")
	formdata.Set("is_primary", "true")
	body = formdata.Encode()
	resp, err = makeAuthRequest(t, session, "POST", "/apis/web/v1/artists/primary", strings.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, 204, resp.StatusCode)

	count, err := store.Count(ctx, `SELECT COUNT(*) FROM artist_releases WHERE release_id = $1 AND is_primary = $2`, 4, true)
	require.NoError(t, err)
	assert.EqualValues(t, 1, count, "expected only one primary artist for release")
	count, err = store.Count(ctx, `SELECT COUNT(*) FROM artist_tracks WHERE track_id = $1 AND is_primary = $2`, 4, true)
	require.NoError(t, err)
	assert.EqualValues(t, 1, count, "expected only one primary artist for track")
}

func TestManualListen(t *testing.T) {

	t.Run("Submit Listens", doSubmitListens)

	ctx := context.Background()

	// happy
	formdata := url.Values{}
	formdata.Set("track_id", "1")
	formdata.Set("unix", strconv.FormatInt(time.Now().Unix()-60, 10))
	body := formdata.Encode()
	resp, err := makeAuthRequest(t, session, "POST", "/apis/web/v1/listen", strings.NewReader(body))
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	count, _ := store.Count(ctx, `SELECT COUNT(*) FROM listens WHERE track_id = $1`, 1)
	assert.Equal(t, 2, count)

	// 400
	formdata.Set("track_id", "1")
	formdata.Set("unix", strconv.FormatInt(time.Now().Unix()+60, 10))
	body = formdata.Encode()
	resp, err = makeAuthRequest(t, session, "POST", "/apis/web/v1/listen", strings.NewReader(body))
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestNowPlaying(t *testing.T) {

	t.Run("Submit Listens", doSubmitListens)

	// no playing
	resp, err := http.DefaultClient.Get(host() + "/apis/web/v1/now-playing")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var result handlers.NowPlayingResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	require.False(t, result.CurrentlyPlaying)

	body := `{
		"listen_type": "playing_now",
		"payload": [
			{
				"track_metadata": {
					"additional_info": {
						"artist_mbids": [
							"efc787f0-046f-4a60-beff-77b398c8cdf4"
						],
						"artist_names": [
							"さユり"
						],
						"duration_ms": 275960,
						"recording_mbid": "21524d55-b1f8-45d1-b172-976cba447199",
						"release_group_mbid": "3281e0d9-fa44-4337-a8ce-6f264beeae16",
						"release_mbid": "eb790e90-0065-4852-b47d-bbeede4aa9fc",
						"submission_client": "navidrome",
						"submission_client_version": "0.56.1 (fa2cf362)"
					},
					"artist_name": "さユり",
					"release_name": "酸欠少女",
					"track_name": "花の塔"
				}
			}
		]
	}`

	req, err := http.NewRequest("POST", host()+"/apis/listenbrainz/1/submit-listens", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", apikey))
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, `{"status": "ok"}`, string(respBytes))

	// yes playing
	resp, err = http.DefaultClient.Get(host() + "/apis/web/v1/now-playing")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	require.True(t, result.CurrentlyPlaying)
	require.Equal(t, "花の塔", result.Track.Title)
}
