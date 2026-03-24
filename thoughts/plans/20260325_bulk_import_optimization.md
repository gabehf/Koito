# Bulk Import Optimization — Implementation Plan

## Overview

Optimize Koito's import pipeline from ~20 listens/min to thousands/min by adopting ListenBrainz-inspired patterns: a persistent entity lookup cache, batch DB writes, pre-deduplication, and deferred enrichment. The core insight from the ecosystem research is **write raw first, enrich async** — and the persistent lookup table benefits all scrobbles (live + import), not just bulk imports.

## Current State Analysis

### The Problem

Importing 49,050 Maloja scrobbles takes ~24 hours. The import is stable (our PR 1 fixes eliminated panics) but each scrobble runs through the full `SubmitListen` path:

- **`GetArtist`** issues 6 DB queries per lookup (including rank computation via window functions)
- **`GetAlbum`** issues 6 DB queries per lookup
- **`GetTrack`** issues 5-6 DB queries per lookup
- **`GetArtistImage` / `GetAlbumImage`** makes HTTP calls even when all image providers are disabled
- **`SaveListen`** is a single INSERT — the only fast part

Per unique scrobble: ~18 DB round-trips + 2 image lookups. Per repeated scrobble: ~18 DB round-trips (no caching). With 5,589 unique artists, 2,628 unique albums, and 49,050 total scrobbles, this is massively redundant.

### Key Discoveries

- `SubmitListenOpts.SkipSaveListen` (`catalog.go:43`) can be used to create entities without recording a listen — useful for entity pre-creation
- `SubmitListenOpts.SkipCacheImage` (`catalog.go:46`) controls image download but NOT image URL resolution — the HTTP calls still happen
- The Koito native importer (`importer/koito.go`) already bypasses `SubmitListen` and does direct DB calls — a precedent for a faster import path
- `pgxpool.Pool` is goroutine-safe — concurrent DB operations are safe at the pool level
- `SaveListen` SQL uses `ON CONFLICT DO NOTHING` — re-importing is idempotent
- No batch insert methods exist anywhere in the codebase
- `GetArtist`/`GetAlbum`/`GetTrack` compute full stats (listen count, time listened, rank) on every call — unnecessary during import

### Ecosystem Patterns (from research)

- **ListenBrainz**: Stores raw scrobbles immediately, resolves MBIDs asynchronously via background worker + Typesense index. Uses MessyBrainz as a stable `(artist, track, release) → ID` mapping.
- **Maloja**: Runs every import through the full normalize → dedup → cache-invalidate cycle. Works for live scrobbles, kills bulk import. **This is exactly Koito's current problem.**
- **Last.fm**: Resolves metadata at write time (corrections), batches up to 50 scrobbles per request.
- **General**: DB-level dedup via unique constraint + `ON CONFLICT` is the industry standard.

## Desired End State

1. A `track_lookup` table provides O(1) entity resolution for any `(artist, track, album)` tuple — both live and import scrobbles benefit
2. All 5 importers use a shared `BulkSubmitter` that pre-deduplicates, creates entities in parallel, and batch-inserts listens
3. Image/MBZ enrichment is fully deferred to existing background tasks during import
4. 49k Maloja import completes in **under 30 minutes** (vs 24 hours currently)
5. Live scrobbles are faster too — cache hit skips 18 DB queries, goes straight to 1 SELECT + 1 INSERT

### Verification

- `go build ./...` compiles
- `go test ./...` passes (existing + new tests)
- Manual: import 49k Maloja scrobbles in under 30 minutes on vo-pc
- Manual: verify live scrobbles from multi-scrobbler still work correctly
- Manual: verify album art appears after background image backfill runs

## What We're NOT Doing

- **Replacing the DB engine** (no TimescaleDB, no Redis) — Postgres is fine for self-hosted scale
- **Local MusicBrainz mirror or Typesense index** — overkill for single-user; the live API + background enrichment is sufficient
- **Changing the live `SubmitListen` API path** — the lookup cache makes it faster, but the logic stays the same
- **Parallelizing live scrobbles** — only imports use the worker pool; live scrobbles remain single-threaded through `SubmitListen`
- **Changing the ListenBrainz/Last.fm relay** — multi-scrobbler handles that independently

## Implementation Approach

Adopt ListenBrainz's "MessyBrainz" pattern as a persistent Postgres table: a normalized `(artist, track, album)` tuple maps to resolved `(artist_id, album_id, track_id)`. This is the foundational optimization — everything else builds on it.

```
Before (per scrobble):
  GetArtist (6 queries) → GetAlbum (6 queries) → GetTrack (6 queries) → SaveListen (1 query)
  = 19 queries minimum

After (cache hit):
  SELECT FROM track_lookup (1 query) → SaveListen (1 query)
  = 2 queries

After (bulk import, cache hit):
  In-memory map lookup (0 queries) → batched SaveListen
  = amortized ~0.01 queries per scrobble
```

---

## Phase 1: `track_lookup` Cache Table

### Overview

Add a persistent lookup table that maps normalized `(artist_name, track_title, release_title)` to resolved entity IDs. Integrate into `SubmitListen` so both live and import scrobbles benefit.

### Changes Required

#### 1. New Migration

**File**: `db/migrations/000006_track_lookup.sql`

```sql
-- +goose Up
CREATE TABLE track_lookup (
    lookup_key TEXT NOT NULL PRIMARY KEY,
    artist_id INT NOT NULL REFERENCES artists(id) ON DELETE CASCADE,
    album_id INT NOT NULL REFERENCES releases(id) ON DELETE CASCADE,
    track_id INT NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_track_lookup_track_id ON track_lookup(track_id);
CREATE INDEX idx_track_lookup_artist_id ON track_lookup(artist_id);
CREATE INDEX idx_track_lookup_album_id ON track_lookup(album_id);

-- +goose Down
DROP TABLE IF EXISTS track_lookup;
```

The `lookup_key` is a normalized string: `lower(artist) || '\x00' || lower(track) || '\x00' || lower(album)`. Using a single TEXT key with a null-byte separator is simpler and faster than a multi-column composite key with `citext`.

#### 2. New SQL Queries

**File**: `db/queries/track_lookup.sql`

```sql
-- name: GetTrackLookup :one
SELECT artist_id, album_id, track_id
FROM track_lookup
WHERE lookup_key = $1;

-- name: InsertTrackLookup :exec
INSERT INTO track_lookup (lookup_key, artist_id, album_id, track_id)
VALUES ($1, $2, $3, $4)
ON CONFLICT (lookup_key) DO UPDATE SET
    artist_id = EXCLUDED.artist_id,
    album_id = EXCLUDED.album_id,
    track_id = EXCLUDED.track_id;

-- name: DeleteTrackLookupByArtist :exec
DELETE FROM track_lookup WHERE artist_id = $1;

-- name: DeleteTrackLookupByAlbum :exec
DELETE FROM track_lookup WHERE album_id = $1;

-- name: DeleteTrackLookupByTrack :exec
DELETE FROM track_lookup WHERE track_id = $1;
```

#### 3. Regenerate sqlc

Run `sqlc generate` to create the Go bindings in `internal/repository/`.

#### 4. DB Interface + Psql Implementation

**File**: `internal/db/db.go` — Add to interface:

```go
// Track Lookup Cache
GetTrackLookup(ctx context.Context, key string) (*TrackLookupResult, error)
SaveTrackLookup(ctx context.Context, opts SaveTrackLookupOpts) error
InvalidateTrackLookup(ctx context.Context, opts InvalidateTrackLookupOpts) error
```

**File**: `internal/db/opts.go` — Add types:

```go
type TrackLookupResult struct {
    ArtistID int32
    AlbumID  int32
    TrackID  int32
}

type SaveTrackLookupOpts struct {
    Key      string
    ArtistID int32
    AlbumID  int32
    TrackID  int32
}

type InvalidateTrackLookupOpts struct {
    ArtistID int32
    AlbumID  int32
    TrackID  int32
}
```

**File**: `internal/db/psql/track_lookup.go` — New file implementing the three methods.

#### 5. Lookup Key Helper

**File**: `internal/catalog/lookup_key.go` — New file:

```go
package catalog

import "strings"

// TrackLookupKey builds a normalized cache key for entity resolution.
func TrackLookupKey(artist, track, album string) string {
    return strings.ToLower(artist) + "\x00" + strings.ToLower(track) + "\x00" + strings.ToLower(album)
}
```

#### 6. Integrate into SubmitListen

**File**: `internal/catalog/catalog.go` — Add fast path at the top of `SubmitListen`:

```go
func SubmitListen(ctx context.Context, store db.DB, opts SubmitListenOpts) error {
    l := logger.FromContext(ctx)

    if opts.Artist == "" || opts.TrackTitle == "" {
        return errors.New("track name and artist are required")
    }

    opts.Time = opts.Time.Truncate(time.Second)

    // Fast path: check lookup cache for known entity combo
    if !opts.SkipSaveListen {
        key := TrackLookupKey(opts.Artist, opts.TrackTitle, opts.ReleaseTitle)
        cached, err := store.GetTrackLookup(ctx, key)
        if err == nil && cached != nil {
            l.Debug().Msg("Track lookup cache hit — skipping entity resolution")
            return store.SaveListen(ctx, db.SaveListenOpts{
                TrackID: cached.TrackID,
                Time:    opts.Time,
                UserID:  opts.UserID,
                Client:  opts.Client,
            })
        }
    }

    // ... existing SubmitListen logic (unchanged) ...

    // After successful entity resolution, populate the cache
    store.SaveTrackLookup(ctx, db.SaveTrackLookupOpts{
        Key:      TrackLookupKey(opts.Artist, opts.TrackTitle, opts.ReleaseTitle),
        ArtistID: artists[0].ID,
        AlbumID:  rg.ID,
        TrackID:  track.ID,
    })

    // ... rest of existing logic ...
}
```

Note: The cache only applies when we have a direct artist+track+album text match. Scrobbles with MBZ IDs that resolve to different text representations will still go through the full path on first encounter, then be cached.

#### 7. Invalidation on Merge/Delete

**File**: `internal/db/psql/artist.go` — In `DeleteArtist` and `MergeArtists`, add:
```go
d.q.DeleteTrackLookupByArtist(ctx, id)
```

**File**: `internal/db/psql/album.go` — In `DeleteAlbum` and `MergeAlbums`, add:
```go
d.q.DeleteTrackLookupByAlbum(ctx, id)
```

**File**: `internal/db/psql/track.go` — In `DeleteTrack` and `MergeTracks`, add:
```go
d.q.DeleteTrackLookupByTrack(ctx, id)
```

### Success Criteria

- [ ] `go build ./...` compiles
- [ ] `go test ./...` passes
- [ ] New test: `TestSubmitListen_LookupCacheHit` — second identical scrobble uses cache
- [ ] New test: `TestSubmitListen_LookupCacheInvalidateOnDelete` — deleting entity clears cache
- [ ] Migration applies cleanly on fresh and existing databases
- [ ] Live scrobbles from multi-scrobbler populate the cache on first hit, use it on second

---

## Phase 2: `SaveListensBatch` DB Method

### Overview

Add a batch listen insert method using pgx's `CopyFrom` for high-throughput listen insertion. This is the DB foundation for the BulkSubmitter.

### Changes Required

#### 1. New SQL + DB Interface

**File**: `internal/db/db.go` — Add to interface:

```go
SaveListensBatch(ctx context.Context, opts []SaveListenOpts) (int64, error)
```

Returns the number of rows actually inserted (excluding `ON CONFLICT` duplicates).

#### 2. Psql Implementation

**File**: `internal/db/psql/listen.go` — New method:

```go
func (d *Psql) SaveListensBatch(ctx context.Context, opts []db.SaveListenOpts) (int64, error) {
    if len(opts) == 0 {
        return 0, nil
    }

    // Use a transaction with a temp table + INSERT ... ON CONFLICT pattern
    // since CopyFrom doesn't support ON CONFLICT directly
    tx, err := d.conn.BeginTx(ctx, pgx.TxOptions{})
    if err != nil {
        return 0, fmt.Errorf("SaveListensBatch: BeginTx: %w", err)
    }
    defer tx.Rollback(ctx)

    // Create temp table
    _, err = tx.Exec(ctx, `
        CREATE TEMP TABLE tmp_listens (
            track_id INT,
            listened_at TIMESTAMPTZ,
            user_id INT,
            client TEXT
        ) ON COMMIT DROP
    `)
    if err != nil {
        return 0, fmt.Errorf("SaveListensBatch: create temp table: %w", err)
    }

    // CopyFrom into temp table
    rows := make([][]interface{}, len(opts))
    for i, o := range opts {
        var client interface{}
        if o.Client != "" {
            client = o.Client
        }
        rows[i] = []interface{}{o.TrackID, o.Time, o.UserID, client}
    }

    _, err = tx.CopyFrom(ctx,
        pgx.Identifier{"tmp_listens"},
        []string{"track_id", "listened_at", "user_id", "client"},
        pgx.CopyFromRows(rows),
    )
    if err != nil {
        return 0, fmt.Errorf("SaveListensBatch: CopyFrom: %w", err)
    }

    // Insert from temp table with dedup
    tag, err := tx.Exec(ctx, `
        INSERT INTO listens (track_id, listened_at, user_id, client)
        SELECT track_id, listened_at, user_id, client FROM tmp_listens
        ON CONFLICT DO NOTHING
    `)
    if err != nil {
        return 0, fmt.Errorf("SaveListensBatch: insert: %w", err)
    }

    if err := tx.Commit(ctx); err != nil {
        return 0, fmt.Errorf("SaveListensBatch: Commit: %w", err)
    }

    return tag.RowsAffected(), nil
}
```

This uses the standard `CopyFrom → temp table → INSERT ON CONFLICT` pattern, which is the fastest bulk insert approach with pgx while still supporting deduplication.

### Success Criteria

- [ ] `go build ./...` compiles
- [ ] `go test ./...` passes
- [ ] New test: `TestSaveListensBatch` — insert 1000 listens, verify count
- [ ] New test: `TestSaveListensBatch_Dedup` — insert duplicates, verify no double-counting
- [ ] New test: `TestSaveListensBatch_Empty` — empty input returns 0, no error

---

## Phase 3: `BulkSubmitter` Helper

### Overview

A reusable import accelerator that all importers can use. Pre-deduplicates scrobbles in memory, resolves entities via the `track_lookup` cache (falling back to `SubmitListen` on cache miss), and batch-inserts listens.

### Design

```
BulkSubmitter
├── Accept(SubmitListenOpts)     — buffer a scrobble
├── Flush() (int, error)         — process all buffered scrobbles
│   ├── Phase A: Deduplicate by (artist, track, album) key
│   ├── Phase B: Resolve entities
│   │   ├── Check track_lookup cache (single SELECT)
│   │   ├── On miss: call SubmitListen(SkipSaveListen=true) to create entities
│   │   ├── Worker pool: N goroutines for parallel entity creation
│   │   └── Populate track_lookup cache after creation
│   ├── Phase C: Map all scrobbles to track_ids via resolved cache
│   └── Phase D: SaveListensBatch
└── Progress callback for logging
```

### Changes Required

#### 1. New Package

**File**: `internal/importer/bulk.go`

```go
package importer

import (
    "context"
    "sync"

    "github.com/gabehf/koito/internal/catalog"
    "github.com/gabehf/koito/internal/db"
    "github.com/gabehf/koito/internal/logger"
    "github.com/gabehf/koito/internal/mbz"
)

type BulkSubmitter struct {
    store      db.DB
    mbzc       mbz.MusicBrainzCaller
    ctx        context.Context
    buffer     []catalog.SubmitListenOpts
    workers    int
    onProgress func(imported, total int)
}

type BulkSubmitterOpts struct {
    Store      db.DB
    Mbzc       mbz.MusicBrainzCaller
    Workers    int                        // default 4
    OnProgress func(imported, total int)  // called every 500 items
}

func NewBulkSubmitter(ctx context.Context, opts BulkSubmitterOpts) *BulkSubmitter {
    workers := opts.Workers
    if workers <= 0 {
        workers = 4
    }
    return &BulkSubmitter{
        store:      opts.Store,
        mbzc:       opts.Mbzc,
        ctx:        ctx,
        workers:    workers,
        onProgress: opts.OnProgress,
    }
}

func (bs *BulkSubmitter) Accept(opts catalog.SubmitListenOpts) {
    bs.buffer = append(bs.buffer, opts)
}

func (bs *BulkSubmitter) Flush() (int, error) {
    l := logger.FromContext(bs.ctx)
    if len(bs.buffer) == 0 {
        return 0, nil
    }

    l.Info().Msgf("BulkSubmitter: Processing %d scrobbles", len(bs.buffer))

    // Phase A: Deduplicate — find unique (artist, track, album) tuples
    type entityKey = string
    unique := make(map[entityKey]catalog.SubmitListenOpts)
    for _, opts := range bs.buffer {
        key := catalog.TrackLookupKey(opts.Artist, opts.TrackTitle, opts.ReleaseTitle)
        if _, exists := unique[key]; !exists {
            unique[key] = opts
        }
    }
    l.Info().Msgf("BulkSubmitter: %d unique entity combos from %d scrobbles", len(unique), len(bs.buffer))

    // Phase B: Resolve entities — check cache, create on miss
    resolved := make(map[entityKey]int32) // key → trackID
    var mu sync.Mutex
    var wg sync.WaitGroup
    sem := make(chan struct{}, bs.workers)
    resolveCount := 0

    for key, opts := range unique {
        // Check track_lookup cache first
        cached, err := bs.store.GetTrackLookup(bs.ctx, key)
        if err == nil && cached != nil {
            mu.Lock()
            resolved[key] = cached.TrackID
            resolveCount++
            mu.Unlock()
            continue
        }

        // Cache miss — create entities via SubmitListen (with worker pool)
        wg.Add(1)
        sem <- struct{}{} // acquire worker slot
        go func(k entityKey, o catalog.SubmitListenOpts) {
            defer wg.Done()
            defer func() { <-sem }() // release worker slot

            o.SkipSaveListen = true
            o.SkipCacheImage = true
            err := catalog.SubmitListen(bs.ctx, bs.store, o)
            if err != nil {
                l.Err(err).Msgf("BulkSubmitter: Failed to create entities for '%s' by '%s'", o.TrackTitle, o.Artist)
                return
            }

            // Re-check cache (SubmitListen populates it in Phase 1's integration)
            cached, err := bs.store.GetTrackLookup(bs.ctx, k)
            if err == nil && cached != nil {
                mu.Lock()
                resolved[k] = cached.TrackID
                mu.Unlock()
            }
        }(key, opts)
    }
    wg.Wait()

    l.Info().Msgf("BulkSubmitter: Resolved %d/%d entity combos", len(resolved), len(unique))

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

    // Phase D: Batch insert listens
    inserted, err := bs.store.SaveListensBatch(bs.ctx, batch)
    if err != nil {
        return 0, fmt.Errorf("BulkSubmitter: SaveListensBatch: %w", err)
    }

    l.Info().Msgf("BulkSubmitter: Inserted %d listens (%d duplicates skipped)", inserted, int64(len(batch))-inserted)
    return int(inserted), nil
}
```

#### 2. TOCTOU Safety for Parallel Entity Creation

The worker pool creates entities via `SubmitListen(SkipSaveListen=true)`. Two workers could race on the same artist name. The existing code uses a get-then-save pattern. Mitigations:

- Pre-dedup in Phase A ensures each unique tuple is processed by exactly one goroutine — **no TOCTOU within the worker pool**
- The only remaining race is between the import workers and live scrobbles from multi-scrobbler hitting the same `SubmitListen` path. This is already handled by the DB's unique constraints + `ON CONFLICT` clauses on join tables.

### Success Criteria

- [ ] `go build ./...` compiles
- [ ] `go test ./...` passes
- [ ] New test: `TestBulkSubmitter_BasicImport` — buffer 100 scrobbles, flush, verify all imported
- [ ] New test: `TestBulkSubmitter_Dedup` — buffer 100 scrobbles with 10 unique combos, verify 10 entity creations
- [ ] New test: `TestBulkSubmitter_CacheHit` — pre-populate track_lookup, verify no SubmitListen calls
- [ ] New test: `TestBulkSubmitter_PartialFailure` — one entity creation fails, rest still imported

---

## Phase 4: Migrate All Importers

### Overview

Wire all 5 importers to use BulkSubmitter instead of calling `SubmitListen` directly in a loop.

### Changes Required

#### 1. Maloja Importer

**File**: `internal/importer/maloja.go`

Replace the per-item `catalog.SubmitListen` loop with:

```go
func ImportMalojaFile(ctx context.Context, store db.DB, mbzc mbz.MusicBrainzCaller, filename string) error {
    l := logger.FromContext(ctx)
    // ... file reading and JSON parsing (unchanged) ...

    bs := NewBulkSubmitter(ctx, BulkSubmitterOpts{
        Store: store,
        Mbzc:  mbzc,
        OnProgress: func(imported, total int) {
            l.Info().Msgf("Maloja import progress: %d/%d", imported, total)
        },
    })

    for _, item := range items {
        // ... existing artist parsing, time window check (unchanged) ...

        bs.Accept(catalog.SubmitListenOpts{
            MbzCaller:      mbzc,
            Artist:         item.Track.Artists[0],
            ArtistNames:    artists,
            TrackTitle:     item.Track.Title,
            ReleaseTitle:   releaseTitle,
            Time:           ts.Local(),
            Client:         "maloja",
            UserID:         1,
            SkipCacheImage: true,
        })
    }

    count, err := bs.Flush()
    if err != nil {
        return fmt.Errorf("ImportMalojaFile: %w", err)
    }
    return finishImport(ctx, filename, count)
}
```

#### 2. Spotify Importer

**File**: `internal/importer/spotify.go` — Same pattern: Accept into BulkSubmitter, Flush at end.

#### 3. LastFM Importer

**File**: `internal/importer/lastfm.go` — Same pattern. Note: LastFM scrobbles include MBZ IDs, which will pass through to `SubmitListen` on cache miss for proper entity resolution.

#### 4. ListenBrainz Importer

**File**: `internal/importer/listenbrainz.go` — Same pattern. ListenBrainz data is the richest (full MBZ IDs, MBID mappings) — cache hits will be common after first import.

#### 5. Koito Importer

**File**: `internal/importer/koito.go` — This one currently bypasses `SubmitListen` with direct DB calls. Two options:
- **Option A**: Migrate to BulkSubmitter (consistent, benefits from cache)
- **Option B**: Leave as-is (it's already fast, Koito exports have pre-resolved IDs)

Recommend **Option A** for consistency, with the Koito importer becoming the simplest BulkSubmitter user since its data is pre-resolved.

### Success Criteria

- [ ] `go build ./...` compiles
- [ ] `go test ./...` passes — all existing import tests still pass
- [ ] `TestImportMaloja` — 38 listens imported correctly
- [ ] `TestImportMaloja_NullAlbum` — null album handled
- [ ] `TestImportMaloja_ApiFormat` — list format works
- [ ] `TestImportSpotify` — duration data preserved
- [ ] `TestImportLastFM` — MBZ IDs resolved
- [ ] `TestImportListenBrainz` — MBID mappings applied
- [ ] `TestImportKoito` — aliases preserved
- [ ] Manual: 49k Maloja import on vo-pc completes in under 30 minutes

---

## Phase 5: Skip Image Lookups During Import

### Overview

Short-circuit `GetArtistImage` and `GetAlbumImage` calls when `SkipCacheImage` is true. Currently these functions still make HTTP calls (or call providers that return "no providers enabled") even when the result won't be used. The existing background tasks (`FetchMissingArtistImages`, `FetchMissingAlbumImages`) will backfill images after import.

### Changes Required

#### 1. Early Return in Associate Functions

**File**: `internal/catalog/associate_artists.go`

In `resolveAliasOrCreateArtist` (line ~248) and `matchArtistsByNames` (line ~304):

```go
// Before:
imgUrl, err := images.GetArtistImage(ctx, images.ArtistImageOpts{...})
if err == nil && imgUrl != "" {
    imgid = uuid.New()
    if !opts.SkipCacheImage {
        // download image
    }
}

// After:
var imgUrl string
if !opts.SkipCacheImage {
    imgUrl, err = images.GetArtistImage(ctx, images.ArtistImageOpts{...})
    if err == nil && imgUrl != "" {
        imgid = uuid.New()
        // download image
    }
}
```

**File**: `internal/catalog/associate_album.go`

Same pattern in `createOrUpdateAlbumWithMbzReleaseID` (line ~125) and `matchAlbumByTitle` (line ~220).

### Success Criteria

- [ ] `go build ./...` compiles
- [ ] `go test ./...` passes
- [ ] No `GetArtistImage`/`GetAlbumImage` calls during import (verify via log: no "No image providers" warnings)
- [ ] Background tasks still fetch images after import completes
- [ ] Live scrobbles (SkipCacheImage=false) still fetch images normally

---

## Performance Estimates

| Scenario | Current | After Phase 1 | After All Phases |
|---|---|---|---|
| Repeated live scrobble | ~19 queries | 2 queries (cache hit) | 2 queries |
| New live scrobble | ~19 queries | ~19 queries + 1 cache write | ~19 queries + 1 cache write |
| 49k Maloja import | ~24 hours | ~12 hours (cache helps repeats) | ~15-30 minutes |
| 49k import (second run) | ~24 hours | ~20 minutes (all cache hits) | ~5 minutes |

## Implementation Order

1. **Phase 1** (track_lookup) — standalone, immediate benefit for all scrobbles
2. **Phase 5** (skip image lookups) — standalone, no dependencies, quick win
3. **Phase 2** (SaveListensBatch) — DB layer, needed by Phase 3
4. **Phase 3** (BulkSubmitter) — the core, depends on Phase 1 + 2
5. **Phase 4** (migrate importers) — depends on Phase 3

Phases 1 and 5 can be done first as independent PRs. Phases 2-4 are one PR.

## References

- ListenBrainz architecture: https://listenbrainz.readthedocs.io/en/latest/developers/architecture.html
- ListenBrainz MBID mapping: https://listenbrainz.readthedocs.io/en/latest/developers/mapping.html
- MusicBrainz rate limiting: https://musicbrainz.org/doc/MusicBrainz_API/Rate_Limiting
- PR 1 (importer fixes): https://github.com/gabehf/Koito/pull/228
- PR 2 (MBZ search): https://github.com/gabehf/Koito/pull/229
- Koito native importer (bypass pattern): `internal/importer/koito.go`
- Current SubmitListen: `internal/catalog/catalog.go:70`
- pgx CopyFrom docs: https://pkg.go.dev/github.com/jackc/pgx/v5#Conn.CopyFrom
