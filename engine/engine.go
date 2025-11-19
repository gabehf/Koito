package engine

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gabehf/koito/engine/middleware"
	"github.com/gabehf/koito/internal/catalog"
	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/db/psql"
	"github.com/gabehf/koito/internal/images"
	"github.com/gabehf/koito/internal/importer"
	"github.com/gabehf/koito/internal/logger"
	mbz "github.com/gabehf/koito/internal/mbz"
	"github.com/gabehf/koito/internal/models"
	"github.com/gabehf/koito/internal/utils"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

func Run(
	getenv func(string) string,
	w io.Writer,
	version string,
) error {
	err := cfg.Load(getenv, version)
	if err != nil {
		panic("Engine: Failed to load configuration")
	}

	l := logger.Get()

	l.Debug().Msg("Engine: Starting application initialization")

	if cfg.StructuredLogging() {
		l.Debug().Msg("Engine: Enabling structured logging")
		*l = l.Output(w)
	} else {
		l.Debug().Msg("Engine: Enabling console logging")
		*l = l.Output(zerolog.ConsoleWriter{
			Out:        w,
			TimeFormat: time.RFC3339,
			FormatMessage: func(i interface{}) string {
				return fmt.Sprintf("\u001b[30;1m>\u001b[0m %s |", i)
			},
		})
	}

	ctx := logger.NewContext(l)

	l.Info().Msgf("Koito %s", version)

	l.Debug().Msgf("Engine: Checking config directory: %s", cfg.ConfigDir())
	_, err = os.Stat(cfg.ConfigDir())
	if err != nil {
		l.Info().Msgf("Engine: Creating config directory: %s", cfg.ConfigDir())
		err = os.MkdirAll(cfg.ConfigDir(), 0744)
		if err != nil {
			l.Fatal().Err(err).Msg("Engine: Failed to create config directory")
			return err
		}
	}
	l.Info().Msgf("Engine: Using config directory: %s", cfg.ConfigDir())

	l.Debug().Msgf("Engine: Checking import directory: %s", path.Join(cfg.ConfigDir(), "import"))
	_, err = os.Stat(path.Join(cfg.ConfigDir(), "import"))
	if err != nil {
		l.Info().Msgf("Engine: Creating import directory: %s", path.Join(cfg.ConfigDir(), "import"))
		err = os.Mkdir(path.Join(cfg.ConfigDir(), "import"), 0744)
		if err != nil {
			l.Fatal().Err(err).Msg("Engine: Failed to create import directory")
			return err
		}
	}

	l.Debug().Msg("Engine: Initializing database connection")
	var store *psql.Psql
	store, err = psql.New()
	for err != nil {
		l.Error().Err(err).Msg("Engine: Failed to connect to database; retrying in 5 seconds")
		time.Sleep(5 * time.Second)
		store, err = psql.New()
	}
	defer store.Close(ctx)
	l.Info().Msg("Engine: Database connection established")

	l.Debug().Msg("Engine: Initializing MusicBrainz client")
	var mbzC mbz.MusicBrainzCaller
	if !cfg.MusicBrainzDisabled() {
		mbzC = mbz.NewMusicBrainzClient()
		l.Info().Msg("Engine: MusicBrainz client initialized")
	} else {
		mbzC = &mbz.MbzErrorCaller{}
		l.Warn().Msg("Engine: MusicBrainz client disabled")
	}

	l.Debug().Msg("Engine: Initializing image sources")
	images.Initialize(images.ImageSourceOpts{
		UserAgent:      cfg.UserAgent(),
		EnableCAA:      !cfg.CoverArtArchiveDisabled(),
		EnableDeezer:   !cfg.DeezerDisabled(),
		EnableSubsonic: cfg.SubsonicEnabled(),
	})
	l.Info().Msg("Engine: Image sources initialized")

	l.Debug().Msg("Engine: Checking for default user")
	userCount, _ := store.CountUsers(ctx)
	if userCount < 1 {
		l.Info().Msg("Engine: Creating default user")
		user, err := store.SaveUser(ctx, db.SaveUserOpts{
			Username: cfg.DefaultUsername(),
			Password: cfg.DefaultPassword(),
			Role:     models.UserRoleAdmin,
		})
		if err != nil {
			l.Fatal().Err(err).Msg("Engine: Failed to save default user in database")
		}
		apikey, err := utils.GenerateRandomString(48)
		if err != nil {
			l.Fatal().Err(err).Msg("Engine: Failed to generate default API key")
		}
		label := "Default"
		_, err = store.SaveApiKey(ctx, db.SaveApiKeyOpts{
			Key:    apikey,
			UserID: user.ID,
			Label:  label,
		})
		if err != nil {
			l.Fatal().Err(err).Msg("Engine: Failed to save default API key in database")
		}
		l.Info().Msgf("Engine: Default user created. Login: %s : %s", cfg.DefaultUsername(), cfg.DefaultPassword())
	}

	l.Debug().Msg("Engine: Checking allowed hosts configuration")
	if cfg.AllowAllHosts() {
		l.Warn().Msg("Engine: Configuration allows requests from all hosts. This is a potential security risk!")
	} else if len(cfg.AllowedHosts()) == 0 || cfg.AllowedHosts()[0] == "" {
		l.Warn().Msgf("Engine: No hosts allowed! Did you forget to set the %s variable?", cfg.ALLOWED_HOSTS_ENV)
	} else {
		l.Info().Msgf("Engine: Allowing hosts: %v", cfg.AllowedHosts())
	}

	if len(cfg.AllowedOrigins()) == 0 || cfg.AllowedOrigins()[0] == "" {
		l.Info().Msgf("Engine: Using default CORS policy")
	} else {
		l.Info().Msgf("Engine: CORS policy: Allowing origins: %v", cfg.AllowedOrigins())
	}

	if cfg.LbzRelayEnabled() && (cfg.LbzRelayUrl() == "" || cfg.LbzRelayToken() == "") {
		l.Warn().Msg("You have enabled ListenBrainz relay, but either the URL or token is missing. Double check your configuration to make sure it is correct!")
	}

	l.Debug().Msg("Engine: Setting up HTTP server")
	var ready atomic.Bool
	mux := chi.NewRouter()
	mux.Use(middleware.WithRequestID)
	mux.Use(middleware.Logger(l))
	mux.Use(chimiddleware.Recoverer)
	mux.Use(chimiddleware.RealIP)
	mux.Use(middleware.AllowedHosts)
	bindRoutes(mux, &ready, store, mbzC)

	httpServer := &http.Server{
		Addr:    cfg.ListenAddr(),
		Handler: mux,
	}

	go func() {
		ready.Store(true)
		l.Info().Msgf("Engine: Listening on %s", cfg.ListenAddr())
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Fatal().Err(err).Msg("Engine: Error when running ListenAndServe")
		}
	}()

	l.Debug().Msg("Engine: Checking import configuration")
	if !cfg.SkipImport() {
		go func() {
			RunImporter(l, store, mbzC)
		}()
	}

	// l.Info().Msg("Creating test export file")
	// go func() {
	// 	err := export.ExportData(ctx, "koito", store)
	// 	if err != nil {
	// 		l.Err(err).Msg("Failed to generate export file")
	// 	}
	// }()

	l.Info().Msg("Engine: Pruning orphaned images")
	go catalog.PruneOrphanedImages(logger.NewContext(l), store)

	l.Info().Msg("Engine: Initialization finished")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	l.Info().Msg("Engine: Received server shutdown notice")

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	l.Info().Msg("Engine: Waiting for all processes to finish")
	mbzC.Shutdown()
	if err := httpServer.Shutdown(ctx); err != nil {
		l.Fatal().Err(err).Msg("Engine: Error during server shutdown")
		return err
	}
	l.Info().Msg("Engine: Shutdown successful")
	return nil
}

func RunImporter(l *zerolog.Logger, store db.DB, mbzc mbz.MusicBrainzCaller) {
	l.Debug().Msg("Checking for import files...")
	files, err := os.ReadDir(path.Join(cfg.ConfigDir(), "import"))
	if err != nil {
		l.Err(err).Msg("Failed to read files from import dir")
	}
	if len(files) > 0 {
		l.Info().Msg("Files found in import directory. Attempting to import...")
	} else {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			l.Error().Interface("recover", r).Msg("Panic when importing files")
		}
	}()
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.Contains(file.Name(), "Streaming_History_Audio") {
			l.Info().Msgf("Import file %s detecting as being Spotify export", file.Name())
			err := importer.ImportSpotifyFile(logger.NewContext(l), store, file.Name())
			if err != nil {
				l.Err(err).Msgf("Failed to import file: %s", file.Name())
			}
		} else if strings.Contains(file.Name(), "maloja") {
			l.Info().Msgf("Import file %s detecting as being Maloja export", file.Name())
			err := importer.ImportMalojaFile(logger.NewContext(l), store, file.Name())
			if err != nil {
				l.Err(err).Msgf("Failed to import file: %s", file.Name())
			}
		} else if strings.Contains(file.Name(), "recenttracks") {
			l.Info().Msgf("Import file %s detecting as being ghan.nl LastFM export", file.Name())
			err := importer.ImportLastFMFile(logger.NewContext(l), store, mbzc, file.Name())
			if err != nil {
				l.Err(err).Msgf("Failed to import file: %s", file.Name())
			}
		} else if strings.Contains(file.Name(), "listenbrainz") {
			l.Info().Msgf("Import file %s detecting as being ListenBrainz export", file.Name())
			err := importer.ImportListenBrainzExport(logger.NewContext(l), store, mbzc, file.Name())
			if err != nil {
				l.Err(err).Msgf("Failed to import file: %s", file.Name())
			}
		} else if strings.Contains(file.Name(), "koito") {
			l.Info().Msgf("Import file %s detecting as being Koito export", file.Name())
			err := importer.ImportKoitoFile(logger.NewContext(l), store, file.Name())
			if err != nil {
				l.Err(err).Msgf("Failed to import file: %s", file.Name())
			}
		} else {
			l.Warn().Msgf("File %s not recognized as a valid import file; make sure it is valid and named correctly", file.Name())
		}
	}
}
