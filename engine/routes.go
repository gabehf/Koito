package engine

import (
	"mime"
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
	r.Use(chimiddleware.GetHead)

	r.With(chimiddleware.RequestSize(5<<20)).
		Get("/image/{image_id}/{filename}", handlers.ImageHandler(db))

	r.Route("/apis/web/v1", func(r chi.Router) {
		r.Get("/config", handlers.GetCfgHandler())

		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(db, middleware.AuthModeLoginGate))
			r.Get("/artist/{id}", handlers.GetArtistHandler(db))                  // done
			r.Get("/artist/{id}/aliases", handlers.GetArtistAliasesHandler(db))   // done
			r.Get("/artist/{id}/interest", handlers.GetArtistInterestHandler(db)) // done

			r.Get("/album/{id}", handlers.GetAlbumHandler(db))                   // done
			r.Get("/album/{id}/artists", handlers.GetArtistsForAlbumHandler(db)) // done
			r.Get("/album/{id}/aliases", handlers.GetAlbumAliasesHandler(db))    // done
			r.Get("/album/{id}/interest", handlers.GetAlbumInterestHandler(db))  // done

			r.Get("/track/{id}", handlers.GetTrackHandler(db))                   // done
			r.Get("/track/{id}/artists", handlers.GetArtistsForTrackHandler(db)) // done
			r.Get("/track/{id}/aliases", handlers.GetTrackAliasesHandler(db))    // done
			r.Get("/track/{id}/interest", handlers.GetTrackInterestHandler(db))  // done

			r.Get("/top/tracks", handlers.GetTopTracksHandler(db))
			r.Get("/top/albums", handlers.GetTopAlbumsHandler(db))
			r.Get("/top/artists", handlers.GetTopArtistsHandler(db))

			r.Get("/listens", handlers.GetListensHandler(db))
			r.Get("/listen-activity", handlers.GetListenActivityHandler(db))
			r.Get("/first-activity", handlers.FirstActivityHandler(db))
			r.Get("/now-playing", handlers.NowPlayingHandler(db))
			r.Get("/stats", handlers.StatsHandler(db))
			r.Get("/search", handlers.SearchHandler(db))
			r.Get("/summary", handlers.SummaryHandler(db))
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
			r.Use(middleware.Authenticate(db, middleware.AuthModeSessionOrAPIKey))

			r.Delete("/artist/{id}", handlers.DeleteArtistHandler(db))
			r.Delete("/artist/{id}/aliases", handlers.DeleteArtistAliasHandler(db))
			r.Post("/artist/{id}/merge", handlers.MergeArtistsHandler(db))
			r.Post("/artist/{id}/aliases", handlers.CreateArtistAliasHandler(db))
			r.Patch("/artist/{id}", handlers.UpdateArtistHandler(db))
			r.Patch("/artist/{id}/image", handlers.ReplaceArtistImageHandler(db))
			r.Patch("/artist/{id}/aliases/primary", handlers.SetPrimaryArtistAliasHandler(db))

			r.Delete("/album/{id}", handlers.DeleteAlbumHandler(db))
			r.Delete("/album/{id}/aliases", handlers.DeleteAlbumAliasHandler(db))
			r.Post("/album/{id}/merge", handlers.MergeAlbumsHandler(db))
			r.Post("/album/{id}/aliases", handlers.CreateAlbumAliasHandler(db))
			r.Patch("/album/{id}", handlers.UpdateAlbumHandler(db))
			r.Patch("/album/{id}/image", handlers.ReplaceAlbumImageHandler(db))
			r.Patch("/album/{id}/aliases/primary", handlers.SetPrimaryAlbumAliasHandler(db))
			r.Patch("/album/{id}/artists/{artist_id}", handlers.SetPrimaryAlbumArtistHandler(db))

			r.Delete("/track/{id}", handlers.DeleteTrackHandler(db))
			r.Delete("/track/{id}/aliases", handlers.DeleteTrackAliasHandler(db))
			r.Delete("/track/{id}/artists/{artist_id}", handlers.DeleteTrackArtistHandler(db))
			r.Post("/track/{id}/merge", handlers.MergeTracksHandler(db))
			r.Post("/track/{id}/aliases", handlers.CreateTrackAliasHandler(db))
			r.Post("/track/{id}/artists", handlers.AddTrackArtistsHandler(db))
			r.Patch("/track/{id}", handlers.UpdateTrackHandler(db))
			r.Patch("/track/{id}/aliases/primary", handlers.SetPrimaryTrackAliasHandler(db))
			r.Patch("/track/{id}/artists/{artist_id}", handlers.SetPrimaryTrackArtistHandler(db))

			r.Post("/listens", handlers.SubmitListenWithIDHandler(db))
			r.Delete("/listens", handlers.DeleteListenHandler(db))

			r.Get("/user/apikeys", handlers.GetApiKeysHandler(db))
			r.Post("/user/apikeys", handlers.GenerateApiKeyHandler(db))
			r.Patch("/user/apikeys/{id}", handlers.UpdateApiKeyLabelHandler(db))
			r.Delete("/user/apikeys/{id}", handlers.DeleteApiKeyHandler(db))

			r.Get("/user", handlers.MeHandler())
			r.Patch("/user", handlers.UpdateUserHandler(db))

			r.Get("/export", handlers.ExportHandler(db))
			r.Delete("/data", handlers.PurgeAllDataHandler(db))
		})
	})

	r.Route("/apis/listenbrainz/1", func(r chi.Router) {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedHeaders: []string{"Content-Type", "Authorization"},
		}))

		r.With(middleware.Authenticate(db, middleware.AuthModeAPIKey)).
			Post("/submit-listens", handlers.LbzSubmitListenHandler(db, mbz))
		r.With(middleware.Authenticate(db, middleware.AuthModeAPIKey)).
			Get("/validate-token", handlers.LbzValidateTokenHandler())
	})

	// serve react client
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "client/build/client"))
	fileServer(r, "/", filesDir)

	// serve client public files
	filesDir = http.Dir(filepath.Join(workDir, "client/public"))
	publicServer(r, "/public", filesDir)
}

func fileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	fs := http.FileServer(root)

	r.Get(path+"*", func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Join("client/build/client", strings.TrimPrefix(r.URL.Path, path))

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			http.ServeFile(w, r, filepath.Join("client/build/client", "index.html"))
			return
		}

		ext := strings.ToLower(filepath.Ext(filePath))
		switch ext {
		case ".js", ".css", ".woff", ".woff2", ".ttf", ".eot",
			".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg", ".ico":
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		default:
			w.Header().Set("Cache-Control", "no-cache")
		}

		// Serve pre-compressed file if the client accepts gzip and it exists
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gzPath := filePath + ".gz"
			if _, err := os.Stat(gzPath); err == nil {
				w.Header().Set("Content-Encoding", "gzip")
				w.Header().Set("Vary", "Accept-Encoding")
				// Set the correct Content-Type for the original file, not .gz
				w.Header().Set("Content-Type", mime.TypeByExtension(ext))
				http.ServeFile(w, r, gzPath)
				return
			}
		}

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
