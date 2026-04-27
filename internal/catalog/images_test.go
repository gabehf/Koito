package catalog_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gabehf/koito/internal/catalog"
	"github.com/gabehf/koito/internal/cfg"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageLifecycle(t *testing.T) {
	store := newTestDB()

	// serve yuu.jpg as test image
	imageBytes, err := os.ReadFile(filepath.Join("test_assets", "yuu.jpg"))
	require.NoError(t, err)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.WriteHeader(http.StatusOK)
		w.Write(imageBytes)
	}))
	defer server.Close()

	imgID := uuid.New()

	err = catalog.DownloadAndCacheImage(context.Background(), imgID, server.URL, catalog.ImageSizeFull)
	require.NoError(t, err)
	err = catalog.DownloadAndCacheImage(context.Background(), imgID, server.URL, catalog.ImageSizeMedium)
	require.NoError(t, err)

	// ensure download is correct

	imagePath := filepath.Join(cfg.ConfigDir(), catalog.ImageCacheDir, "full", imgID.String())
	_, err = os.Stat(imagePath)
	assert.NoError(t, err)
	imagePath = filepath.Join(cfg.ConfigDir(), catalog.ImageCacheDir, "medium", imgID.String())
	_, err = os.Stat(imagePath)
	assert.NoError(t, err)

	assert.NoError(t, catalog.DeleteImage(imgID))

	// ensure delete works

	imagePath = filepath.Join(cfg.ConfigDir(), catalog.ImageCacheDir, "full", imgID.String())
	_, err = os.Stat(imagePath)
	assert.Error(t, err)
	imagePath = filepath.Join(cfg.ConfigDir(), catalog.ImageCacheDir, "medium", imgID.String())
	_, err = os.Stat(imagePath)
	assert.Error(t, err)

	// re-download for prune

	err = catalog.DownloadAndCacheImage(context.Background(), imgID, server.URL, catalog.ImageSizeFull)
	require.NoError(t, err)
	err = catalog.DownloadAndCacheImage(context.Background(), imgID, server.URL, catalog.ImageSizeMedium)
	require.NoError(t, err)

	assert.NoError(t, catalog.PruneOrphanedImages(context.Background(), store))

	// ensure prune works

	imagePath = filepath.Join(cfg.ConfigDir(), catalog.ImageCacheDir, "full", imgID.String())
	_, err = os.Stat(imagePath)
	assert.Error(t, err)
	imagePath = filepath.Join(cfg.ConfigDir(), catalog.ImageCacheDir, "medium", imgID.String())
	_, err = os.Stat(imagePath)
	assert.Error(t, err)
}
