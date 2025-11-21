package engine

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gabehf/koito/engine/handlers"
	"github.com/gabehf/koito/engine/middleware"
	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/db"
	mbz "github.com/gabehf/koito/internal/mbz"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

func bindRoutes(
	r *chi.Mux,
	ready *atomic.Bool,
	db db.DB,
	mbz mbz.MusicBrainzCaller,
) {
	if !(len(cfg.AllowedOrigins()) == 0) && !(cfg.AllowedOrigins()[0] == "") {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins: cfg.AllowedOrigins(),
			AllowedMethods: []string{"GET", "OPTIONS", "HEAD"},
		}))
	}
	r.With(chimiddleware.RequestSize(5<<20)).
		Get("/images/{size}/{filename}", handlers.ImageHandler(db))

	r.Route("/apis/web/v1", func(r chi.Router) {
		r.Get("/config", handlers.GetCfgHandler())

		r.Group(func(r chi.Router) {
			if cfg.LoginGate() {
				r.Use(middleware.ValidateSession(db))
			}
			r.Get("/artist", handlers.GetArtistHandler(db))
			r.Get("/artists", handlers.GetArtistsForItemHandler(db))
			r.Get("/album", handlers.GetAlbumHandler(db))
			r.Get("/track", handlers.GetTrackHandler(db))
			r.Get("/top-tracks", handlers.GetTopTracksHandler(db))
			r.Get("/top-albums", handlers.GetTopAlbumsHandler(db))
			r.Get("/top-artists", handlers.GetTopArtistsHandler(db))
			r.Get("/listens", handlers.GetListensHandler(db))
			r.Get("/listen-activity", handlers.GetListenActivityHandler(db))
			r.Get("/now-playing", handlers.NowPlayingHandler(db))
			r.Get("/stats", handlers.StatsHandler(db))
			r.Get("/search", handlers.SearchHandler(db))
			r.Get("/aliases", handlers.GetAliasesHandler(db))
		})
		r.Post("/logout", handlers.LogoutHandler(db))
		if !cfg.RateLimitDisabled() {
			r.With(httprate.Limit(
				10,
				time.Minute,
				httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, `{"error":"too many requests"}`, http.StatusTooManyRequests)
				}),
			)).Post("/login", handlers.LoginHandler(db))
		} else {
			r.Post("/login", handlers.LoginHandler(db))
		}

		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			if !ready.Load() {
				http.Error(w, "not ready", http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.ValidateSession(db))
			r.Get("/export", handlers.ExportHandler(db))
			r.Post("/replace-image", handlers.ReplaceImageHandler(db))
			r.Patch("/album", handlers.UpdateAlbumHandler(db))
			r.Post("/merge/tracks", handlers.MergeTracksHandler(db))
			r.Post("/merge/albums", handlers.MergeReleaseGroupsHandler(db))
			r.Post("/merge/artists", handlers.MergeArtistsHandler(db))
			r.Delete("/artist", handlers.DeleteArtistHandler(db))
			r.Post("/artists/primary", handlers.SetPrimaryArtistHandler(db))
			r.Delete("/album", handlers.DeleteAlbumHandler(db))
			r.Delete("/track", handlers.DeleteTrackHandler(db))
			r.Post("/listen", handlers.SubmitListenWithIDHandler(db))
			r.Delete("/listen", handlers.DeleteListenHandler(db))
			r.Post("/aliases", handlers.CreateAliasHandler(db))
			r.Post("/aliases/delete", handlers.DeleteAliasHandler(db))
			r.Post("/aliases/primary", handlers.SetPrimaryAliasHandler(db))
			r.Get("/user/apikeys", handlers.GetApiKeysHandler(db))
			r.Post("/user/apikeys", handlers.GenerateApiKeyHandler(db))
			r.Patch("/user/apikeys", handlers.UpdateApiKeyLabelHandler(db))
			r.Delete("/user/apikeys", handlers.DeleteApiKeyHandler(db))
			r.Get("/user/me", handlers.MeHandler(db))
			r.Patch("/user", handlers.UpdateUserHandler(db))
		})
	})

	r.Route("/apis/listenbrainz/1", func(r chi.Router) {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedHeaders: []string{"Content-Type", "Authorization"},
		}))

		r.With(middleware.ValidateApiKey(db)).Post("/submit-listens", handlers.LbzSubmitListenHandler(db, mbz))
		r.With(middleware.ValidateApiKey(db)).Get("/validate-token", handlers.LbzValidateTokenHandler(db))
	})

	// serve react client
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "client/build/client"))
	fileServer(r, "/", filesDir)

	// serve client public files
	filesDir = http.Dir(filepath.Join(workDir, "client/public"))
	publicServer(r, "/public", filesDir)
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func fileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	// Serve static files
	fs := http.FileServer(root)
	r.Get(path+"*", func(w http.ResponseWriter, r *http.Request) {
		// Check if file exists
		filePath := filepath.Join("client/build/client", strings.TrimPrefix(r.URL.Path, path))
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// File doesn't exist, serve index.html
			http.ServeFile(w, r, filepath.Join("client/build/client", "index.html"))
			return
		}

		// Serve file normally
		fs.ServeHTTP(w, r)
	})
}

func publicServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}
	fs := http.FileServer(root)
	r.Get(path+"*", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, path)
		fs.ServeHTTP(w, r)
	})
}
