package importer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gabehf/koito/internal/catalog"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/mbz"
)

// BulkSubmitter is a reusable import accelerator. It pre-deduplicates scrobbles
// in memory, resolves entities via the track_lookup cache (falling back to
// SubmitListen on cache miss with a worker pool for parallelism), and batch-inserts
// listens via SaveListensBatch.
type BulkSubmitter struct {
	store   db.DB
	mbzc    mbz.MusicBrainzCaller
	ctx     context.Context
	buffer  []catalog.SubmitListenOpts
	workers int
}

type BulkSubmitterOpts struct {
	Store   db.DB
	Mbzc    mbz.MusicBrainzCaller
	Workers int // default 4
}

func NewBulkSubmitter(ctx context.Context, opts BulkSubmitterOpts) *BulkSubmitter {
	workers := opts.Workers
	if workers <= 0 {
		workers = 4
	}
	return &BulkSubmitter{
		store:   opts.Store,
		mbzc:    opts.Mbzc,
		ctx:     ctx,
		workers: workers,
	}
}

// Accept buffers a scrobble for later batch processing.
func (bs *BulkSubmitter) Accept(opts catalog.SubmitListenOpts) {
	bs.buffer = append(bs.buffer, opts)
}

// Flush processes all buffered scrobbles: deduplicates, resolves entities, and batch-inserts listens.
// Returns the number of listens successfully inserted.
func (bs *BulkSubmitter) Flush() (int, error) {
	l := logger.FromContext(bs.ctx)
	if len(bs.buffer) == 0 {
		return 0, nil
	}

	l.Info().Msgf("BulkSubmitter: Processing %d scrobbles", len(bs.buffer))

	// Phase A: Deduplicate — find unique (artist, track, album) tuples
	unique := make(map[string]catalog.SubmitListenOpts)
	for _, opts := range bs.buffer {
		key := catalog.TrackLookupKey(opts.Artist, opts.TrackTitle, opts.ReleaseTitle)
		if _, exists := unique[key]; !exists {
			unique[key] = opts
		}
	}
	l.Info().Msgf("BulkSubmitter: %d unique entity combos from %d scrobbles", len(unique), len(bs.buffer))

	// Phase B: Resolve entities — check cache, create on miss
	resolved := make(map[string]int32) // key → trackID
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, bs.workers)
	cacheHits := 0

	for key, opts := range unique {
		// Check track_lookup cache first
		cached, err := bs.store.GetTrackLookup(bs.ctx, key)
		if err == nil && cached != nil {
			mu.Lock()
			resolved[key] = cached.TrackID
			cacheHits++
			mu.Unlock()
			continue
		}

		// Cache miss — create entities via SubmitListen (with worker pool)
		wg.Add(1)
		sem <- struct{}{} // acquire worker slot
		go func(k string, o catalog.SubmitListenOpts) {
			defer wg.Done()
			defer func() { <-sem }() // release worker slot

			o.SkipSaveListen = true
			o.SkipCacheImage = true
			err := catalog.SubmitListen(bs.ctx, bs.store, o)
			if err != nil {
				l.Err(err).Msgf("BulkSubmitter: Failed to create entities for '%s' by '%s'", o.TrackTitle, o.Artist)
				return
			}

			// Re-check cache (SubmitListen populates it via Phase 1's integration)
			cached, err := bs.store.GetTrackLookup(bs.ctx, k)
			if err == nil && cached != nil {
				mu.Lock()
				resolved[k] = cached.TrackID
				mu.Unlock()
			}
		}(key, opts)
	}
	wg.Wait()

	l.Info().Msgf("BulkSubmitter: Resolved %d/%d entity combos (%d cache hits)",
		len(resolved), len(unique), cacheHits)

	// Phase C: Build listen batch
	batch := make([]db.SaveListenOpts, 0, len(bs.buffer))
	skipped := 0
	for _, opts := range bs.buffer {
		key := catalog.TrackLookupKey(opts.Artist, opts.TrackTitle, opts.ReleaseTitle)
		trackID, ok := resolved[key]
		if !ok {
			skipped++
			continue
		}
		batch = append(batch, db.SaveListenOpts{
			TrackID: trackID,
			Time:    opts.Time.Truncate(time.Second),
			UserID:  opts.UserID,
			Client:  opts.Client,
		})
	}
	if skipped > 0 {
		l.Warn().Msgf("BulkSubmitter: Skipped %d scrobbles with unresolved entities", skipped)
	}

	// Phase D: Batch insert listens (in chunks to avoid huge transactions)
	const chunkSize = 5000
	var totalInserted int64
	for i := 0; i < len(batch); i += chunkSize {
		end := i + chunkSize
		if end > len(batch) {
			end = len(batch)
		}
		inserted, err := bs.store.SaveListensBatch(bs.ctx, batch[i:end])
		if err != nil {
			return int(totalInserted), fmt.Errorf("BulkSubmitter: SaveListensBatch: %w", err)
		}
		totalInserted += inserted
	}

	l.Info().Msgf("BulkSubmitter: Inserted %d listens (%d duplicates skipped)",
		totalInserted, int64(len(batch))-totalInserted)
	return int(totalInserted), nil
}
