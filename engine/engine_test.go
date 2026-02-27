package engine_test

import (
	"context"
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
	"github.com/gabehf/koito/internal/db/psql"
	"github.com/gabehf/koito/internal/utils"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var store *psql.Psql

func getTestGetenv(resource *dockertest.Resource) func(string) string {
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
		case cfg.DATABASE_URL_ENV:
			return fmt.Sprintf("postgres://postgres:secret@localhost:%s", resource.GetPort("5432/tcp"))
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
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	// uses pool to try to connect to Docker
	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.Run("postgres", "latest", []string{"POSTGRES_PASSWORD=secret"})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	getenv := getTestGetenv(resource)
	err = cfg.Load(getenv, "test")
	if err != nil {
		log.Fatalf("Could not load cfg: %s", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		var err error
		store, err = psql.New()
		if err != nil {
			log.Println("Failed to connect to test database, retrying...")
			return err
		}
		return store.Ping(context.Background())
	}); err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	go engine.Run(getenv, os.Stdout, "vTest")

	// Wait until the web server is reachable
	for i := 0; i < 20; i++ {
		url := fmt.Sprintf("http://%s/apis/web/v1/health", cfg.ListenAddr())
		client := &http.Client{
			Timeout: 2 * time.Second, // Set your desired timeout
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
			err = nil
			break
		}
		log.Printf("Unexpected status code at %s, retrying... (%d/20)", url, i+1)
		time.Sleep(1 * time.Second)
	}

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	err = os.RemoveAll(cfg.ConfigDir())
	if err != nil {
		log.Fatalf("Could not remove temporary config dir: %v", err)
	}

	os.Exit(code)
}

func host() string {
	return fmt.Sprintf("http://%s", cfg.ListenAddr())
}
func TestHealthEndpointNoAuth(t *testing.T) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("%s/apis/web/v1/health", host())
	resp, err := client.Get(url)
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHealthEndpointMethodNotAllowed(t *testing.T) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("%s/apis/web/v1/health", host())
	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		req, err := http.NewRequest(method, url, nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode, "method %s should return 405", method)
	}
}
