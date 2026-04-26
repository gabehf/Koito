package engine_test

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gabehf/koito/engine"
	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/db/sqlite"
	"github.com/gabehf/koito/internal/utils"
)

var store *sqlite.Sqlite

func getTestGetenv() func(string) string {
	dir, err := utils.GenerateRandomString(8)
	if err != nil {
		panic(err)
	}

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(fmt.Errorf("failed to get an open port: %w", err))
	}
	defer listener.Close()

	port := strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)

	return func(env string) string {
		switch env {
		case cfg.ENABLE_STRUCTURED_LOGGING_ENV:
			return "true"
		case cfg.LOG_LEVEL_ENV:
			return "debug"
		case cfg.SQLITE_ENABLED:
			return "true"
		case cfg.DATABASE_URL_ENV:
			return "bleh"
		case cfg.DEFAULT_PASSWORD_ENV:
			return "testuser123"
		case cfg.DEFAULT_USERNAME_ENV:
			return "test"
		case cfg.CONFIG_DIR_ENV:
			return dir
		case cfg.LISTEN_PORT_ENV:
			return port
		case cfg.ALLOWED_HOSTS_ENV:
			return "*"
		case cfg.DISABLE_DEEZER_ENV, cfg.DISABLE_COVER_ART_ARCHIVE_ENV, cfg.DISABLE_MUSICBRAINZ_ENV, cfg.SKIP_IMPORT_ENV:
			return "true"
		default:
			return ""
		}
	}
}

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	// pool, err := dockertest.NewPool("")
	// if err != nil {
	// 	log.Fatalf("Could not construct pool: %s", err)
	// }

	// // uses pool to try to connect to Docker
	// err = pool.Client.Ping()
	// if err != nil {
	// 	log.Fatalf("Could not connect to Docker: %s", err)
	// }

	// // pulls an image, creates a container based on it and runs it
	// resource, err := pool.Run("postgres", "latest", []string{"POSTGRES_PASSWORD=secret"})
	// if err != nil {
	// 	log.Fatalf("Could not start resource: %s", err)
	// }

	getenv := getTestGetenv()
	err := cfg.Load(getenv, "test")
	if err != nil {
		log.Fatalf("Could not load cfg: %s", err)
	}

	go engine.Run(getenv, os.Stdout, "vTest")

	// Wait until the web server is reachable
	for i := range 20 {
		url := fmt.Sprintf("http://%s/apis/web/v1/health", cfg.ListenAddr())
		client := &http.Client{
			Timeout: 2 * time.Second,
		}
		resp, err := client.Get(url)
		if err != nil {
			if i >= 19 {
				log.Fatalf("Web server is not reachable: %s", err)
			}
			log.Printf("Failed to connect to web server at %s, retrying... (%d/20)", url, i+1)
			time.Sleep(1 * time.Second)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			break
		}
		log.Printf("Unexpected status code at %s, retrying... (%d/20)", url, i+1)
		time.Sleep(1 * time.Second)
	}

	// Open a second connection to the same file the server uses so that direct
	// DB assertions in tests see the server's actual state. sqlite.New() uses
	// cfg.ConfigDir() (set by getTestGetenv) and runs idempotent migrations.
	var storeErr error
	store, storeErr = sqlite.New()
	if storeErr != nil {
		log.Fatalf("Could not open server database for test verification: %s", storeErr)
	}

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	// if err := pool.Purge(resource); err != nil {
	// 	log.Fatalf("Could not purge resource: %s", err)
	// }

	err = os.RemoveAll(cfg.ConfigDir())
	if err != nil {
		log.Fatalf("Could not remove temporary config dir: %v", err)
	}

	os.Exit(code)
}

func host() string {
	return fmt.Sprintf("http://%s", cfg.ListenAddr())
}
