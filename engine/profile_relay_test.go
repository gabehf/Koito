package engine

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProfileRelayAdministration(t *testing.T) {
	a := newProfileTestApp(t)
	path := "/apis/web/v1/profiles/default/relay"
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		require.Equal(t, "GET", r.Method)
		require.Equal(t, "/1/validate-token", r.URL.Path)
		require.Empty(t, r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		switch r.Header.Get("Authorization") {
		case "Token external-token":
			fmt.Fprint(w, `{"valid":true,"user_name":"external-user"}`)
		case "Token redirect":
			http.Redirect(w, r, serverURLPlaceholder, http.StatusFound)
		case "Token invalid-json":
			fmt.Fprint(w, `not json`)
		default:
			fmt.Fprint(w, `{"valid":false}`)
		}
	}))
	defer server.Close()
	body := fmt.Sprintf(`{"url":%q,"token":"external-token"}`, server.URL+"/1/")
	for _, method := range []string{"GET", "PUT", "DELETE"} {
		require.Equal(t, 401, a.request(method, path, body, "first-key", false).Code)
	}
	require.Equal(t, 401, a.request("POST", path+"/test", body, "first-key", false).Code)
	require.Zero(t, calls.Load())
	w := a.request("GET", path, "", "", true)
	require.Equal(t, 200, w.Code)
	require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
	require.JSONEq(t, `{"enabled":false,"url":"","token":""}`, w.Body.String())
	w = a.request("POST", path+"/test", body, "", true)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), "external-user")
	require.NotContains(t, w.Body.String(), "external-token")
	url, _, err := a.store.GetProfileRelay(context.Background(), 1)
	require.NoError(t, err)
	require.Empty(t, url, "testing must not enable or save a relay")
	require.Equal(t, 204, a.request("PUT", path, body, "", true).Code)
	w = a.request("GET", path, "", "", true)
	require.JSONEq(t, fmt.Sprintf(`{"enabled":true,"url":%q,"token":"external-token"}`, server.URL+"/1"), w.Body.String())
	w = a.request("GET", "/apis/web/v1/profiles/second/relay", "", "", true)
	require.JSONEq(t, `{"enabled":false,"url":"","token":""}`, w.Body.String())
	for _, token := range []string{"invalid", "redirect", "invalid-json"} {
		w = a.request("POST", path+"/test", strings.Replace(body, "external-token", token, 1), "", true)
		require.GreaterOrEqual(t, w.Code, 400, token)
	}
	for _, bad := range []string{`{`, `{"url":"file:///tmp/test","token":"x"}`, `{"url":"https://example.com/1?token=x","token":"x"}`, `{"url":"https://example.com/1","token":""}`} {
		require.Equal(t, 400, a.request("PUT", path, bad, "", true).Code)
		require.Equal(t, 400, a.request("POST", path+"/test", bad, "", true).Code)
	}
	require.Equal(t, 204, a.request("DELETE", path, "", "", true).Code)
	w = a.request("GET", path, "", "", true)
	require.JSONEq(t, `{"enabled":false,"url":"","token":""}`, w.Body.String())
}

const serverURLPlaceholder = "https://example.invalid/never-forward-token"
