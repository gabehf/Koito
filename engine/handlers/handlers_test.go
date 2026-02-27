package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/gabehf/koito/engine/middleware"
	"github.com/gabehf/koito/internal/models"
	"github.com/gabehf/koito/internal/db"
	"github.com/google/uuid"
)

// Focused, small test set that avoids needing a full db.DB mock.

func TestHealthHandler_Returns200(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	HealthHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestMeHandler_Unauthorized(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)

	// MeHandler does not use the store for the unauthorized path, pass nil
	MeHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestMeHandler_Success(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)

	user := &models.User{ID: 1, Username: "testuser"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)

	MeHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "testuser") {
		t.Fatalf("expected response to contain username, got %s", rr.Body.String())
	}
}

func TestGetArtistHandler_MissingID_Returns400(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/artist", nil)

	// Handler will validate query param before touching the store; pass nil
	http.HandlerFunc(GetArtistHandler(nil)).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestGetArtistHandler_InvalidID_Returns400(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/artist?id=abc", nil)

	http.HandlerFunc(GetArtistHandler(nil)).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid id, got %d", rr.Code)
	}
}

func TestGetTrackHandler_MissingID_Returns400(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/track", nil)

	http.HandlerFunc(GetTrackHandler(nil)).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestGetTrackHandler_InvalidID_Returns400(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/track?id=xyz", nil)

	http.HandlerFunc(GetTrackHandler(nil)).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid id, got %d", rr.Code)
	}
}

func TestGetArtistHandler_Success(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/artist?id=5", nil)

	mock := artistStoreMock{}
	http.HandlerFunc(GetArtistHandler(mock)).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Test Artist") {
		t.Fatalf("expected body to contain artist name, got %s", rr.Body.String())
	}
}

func TestGetTrackHandler_Success(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/track?id=7", nil)

	mock := trackStoreMock{}
	http.HandlerFunc(GetTrackHandler(mock)).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Test Track") {
		t.Fatalf("expected body to contain track title, got %s", rr.Body.String())
	}
}

func TestLoginHandler_Success(t *testing.T) {
	// prepare hashed password
	pass := []byte("secretpass")
	hashed, _ := bcrypt.GenerateFromPassword(pass, bcrypt.DefaultCost)

	store := &loginStoreMock{user: &models.User{ID: 3, Username: "bob", Password: hashed}}

	form := "username=bob&password=secretpass"
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	LoginHandler(store).ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	// cookie set
	found := false
	for _, c := range rr.Result().Cookies() {
		if c.Name == "koito_session" && c.Value != "" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected koito_session cookie to be set")
	}
}

// --- minimal mocks ---

type artistStoreMock struct{}
func (artistStoreMock) GetArtist(ctx context.Context, opts db.GetArtistOpts) (*models.Artist, error) {
	return &models.Artist{ID: opts.ID, Name: "Test Artist"}, nil
}

type trackStoreMock struct{}
func (trackStoreMock) GetTrack(ctx context.Context, opts db.GetTrackOpts) (*models.Track, error) {
	return &models.Track{ID: opts.ID, Title: "Test Track"}, nil
}

type loginStoreMock struct{ user *models.User }
func (l *loginStoreMock) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	if l.user != nil && l.user.Username == username {
		return l.user, nil
	}
	return nil, nil
}
func (l *loginStoreMock) SaveSession(ctx context.Context, userId int32, expiresAt time.Time, persistent bool) (*models.Session, error) {
	return &models.Session{ID: uuid.New(), UserID: userId, ExpiresAt: expiresAt, Persistent: persistent}, nil
}
