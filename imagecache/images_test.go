package imagecache_test

// func TestImageLifecycle(t *testing.T) {
// 	store := newTestDB()

// 	// serve yuu.jpg as test image
// 	imageBytes, err := os.ReadFile(filepath.Join("test_assets", "yuu.jpg"))
// 	require.NoError(t, err)
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set("Content-Type", "image/jpeg")
// 		w.WriteHeader(http.StatusOK)
// 		w.Write(imageBytes)
// 	}))
// 	defer server.Close()

// 	imgID := uuid.New()

// 	err = catalog.DownloadAndCacheImage(context.Background(), imgID, server.URL, catalog.ImageSizeSource)
// 	require.NoError(t, err)
// 	err = catalog.DownloadAndCacheImage(context.Background(), imgID, server.URL, catalog.ImageSizeMedium)
// 	require.NoError(t, err)

// 	// ensure download is correct

// 	imagePath := catalog.BuildImagePath(imgID, catalog.ImageSizeSource)
// 	_, err = os.Stat(imagePath)
// 	assert.NoError(t, err)
// 	imagePath = catalog.BuildImagePath(imgID, catalog.ImageSizeMedium)
// 	_, err = os.Stat(imagePath)
// 	assert.NoError(t, err)

// 	assert.NoError(t, catalog.DeleteImage(imgID))

// 	// ensure delete works

// 	imagePath = catalog.BuildImagePath(imgID, catalog.ImageSizeSource)
// 	_, err = os.Stat(imagePath)
// 	assert.ErrorIs(t, err, os.ErrNotExist)
// 	imagePath = catalog.BuildImagePath(imgID, catalog.ImageSizeMedium)
// 	_, err = os.Stat(imagePath)
// 	assert.ErrorIs(t, err, os.ErrNotExist)

// 	// re-download for prune

// 	err = catalog.DownloadAndCacheImage(context.Background(), imgID, server.URL, catalog.ImageSizeSource)
// 	require.NoError(t, err)
// 	err = catalog.DownloadAndCacheImage(context.Background(), imgID, server.URL, catalog.ImageSizeMedium)
// 	require.NoError(t, err)

// 	assert.NoError(t, catalog.PruneOrphanedImages(context.Background(), store))

// 	// ensure prune works

// 	imagePath = catalog.BuildImagePath(imgID, catalog.ImageSizeSource)
// 	_, err = os.Stat(imagePath)
// 	assert.ErrorIs(t, err, os.ErrNotExist)
// 	imagePath = catalog.BuildImagePath(imgID, catalog.ImageSizeMedium)
// 	_, err = os.Stat(imagePath)
// 	assert.ErrorIs(t, err, os.ErrNotExist)
// }
