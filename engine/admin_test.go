package engine_test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/models"
	"github.com/stretchr/testify/require"
)

// Minimal test: non-admin gets 403 on delete artist; admin succeeds.
func TestAdminProtectedDeleteArtist(t *testing.T) {
	// create non-admin user
	_, err := store.SaveUser(context.Background(), db.SaveUserOpts{
		Username: "regular_user",
		Password: "password123",
		Role:     models.UserRoleUser,
	})
	require.NoError(t, err)

	// login as non-admin
	form := url.Values{}
	form.Set("username", "regular_user")
	form.Set("password", "password123")
	resp, err := http.DefaultClient.Post(host()+"/apis/web/v1/login", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	require.NoError(t, err)
	require.Len(t, resp.Cookies(), 1)
	session := resp.Cookies()[0].Value

	// create an artist to delete
	artist, err := store.SaveArtist(context.Background(), db.SaveArtistOpts{Name: "ToBeDeleted"})
	require.NoError(t, err)

	// attempt delete as non-admin, expect Forbidden
	req, err := http.NewRequest("DELETE", host()+fmt.Sprintf("/apis/web/v1/artist?id=%d", artist.ID), nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: "koito_session", Value: session})
	resp2, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp2.Body.Close()
	require.Equal(t, http.StatusForbidden, resp2.StatusCode)

	// login as admin (default user 'test')
	form2 := url.Values{}
	form2.Set("username", "test")
	form2.Set("password", "testuser123")
	resp3, err := http.DefaultClient.Post(host()+"/apis/web/v1/login", "application/x-www-form-urlencoded", strings.NewReader(form2.Encode()))
	require.NoError(t, err)
	require.Len(t, resp3.Cookies(), 1)
	adminSession := resp3.Cookies()[0].Value

	// attempt delete as admin - expect NoContent
	req2, err := http.NewRequest("DELETE", host()+fmt.Sprintf("/apis/web/v1/artist?id=%d", artist.ID), nil)
	require.NoError(t, err)
	req2.AddCookie(&http.Cookie{Name: "koito_session", Value: adminSession})
	resp4, err := http.DefaultClient.Do(req2)
	require.NoError(t, err)
	defer resp4.Body.Close()
	require.Equal(t, http.StatusNoContent, resp4.StatusCode)
}
