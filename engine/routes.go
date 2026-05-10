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

			r.Delete("/artist/{id}", handlers.DeleteArtistHandler(db))                         // done
			r.Delete("/artist/{id}/aliases", handlers.DeleteArtistAliasHandler(db))            // done
			r.Post("/artist/{id}/merge", handlers.MergeArtistsHandler(db))                     // done
			r.Post("/artist/{id}/aliases", handlers.CreateArtistAliasHandler(db))              // done
			r.Patch("/artist/{id}", handlers.UpdateArtistHandler(db))                          // done
			r.Patch("/artist/{id}/image", handlers.ReplaceArtistImageHandler(db))              // done
			r.Patch("/artist/{id}/aliases/primary", handlers.SetPrimaryArtistAliasHandler(db)) // done

			r.Delete("/album/{id}", handlers.DeleteAlbumHandler(db))                             // done
			r.Delete("/album/{id}/aliases", handlers.DeleteAlbumAliasHandler(db))                // done
			r.Post("/album/{id}/merge", handlers.MergeAlbumsHandler(db))                         // done
			r.Post("/album/{id}/aliases", handlers.CreateAlbumAliasHandler(db))                  // done
			r.Patch("/album/{id}", handlers.UpdateAlbumHandler(db))                              // done
			r.Patch("/album/{id}/image", handlers.ReplaceAlbumImageHandler(db))                  // done
			r.Patch("/album/{id}/aliases/primary", handlers.SetPrimaryAlbumAliasHandler(db))     // done
			r.Patch("/album/{id}/artist/{artist_id}", handlers.SetPrimaryAlbumArtistHandler(db)) // done

			r.Delete("/track/{id}", handlers.DeleteTrackHandler(db))                              // done
			r.Delete("/track/{id}/aliases", handlers.DeleteTrackAliasHandler(db))                 // done
			r.Delete("/track/{id}/artists/{artist_id}", handlers.DeleteTrackArtistHandler(db))    // done
			r.Post("/track/{id}/merge", handlers.MergeTracksHandler(db))                          // done
			r.Post("/track/{id}/aliases", handlers.CreateTrackAliasHandler(db))                   // done
			r.Post("/track/{id}/artists", handlers.AddTrackArtistsHandler(db))                    // done
			r.Patch("/track/{id}", handlers.UpdateTrackHandler(db))                               // done
			r.Patch("/track/{id}/aliases/primary", handlers.SetPrimaryTrackAliasHandler(db))      // done
			r.Patch("/track/{id}/artists/{artist_id}", handlers.SetPrimaryTrackArtistHandler(db)) // done

			r.Post("/listen", handlers.SubmitListenWithIDHandler(db))
			r.Delete("/listen", handlers.DeleteListenHandler(db))

			r.Get("/user/apikeys", handlers.GetApiKeysHandler(db))
			r.Post("/user/apikeys", handlers.GenerateApiKeyHandler(db))
			r.Patch("/user/apikeys", handlers.UpdateApiKeyLabelHandler(db))
			r.Delete("/user/apikeys", handlers.DeleteApiKeyHandler(db))

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
