---
title: Web API Reference
description: HTTP API reference for integrating with or building tooling around Koito.
---

This document describes the Koito HTTP API as implemented in the source code. It is intended to help developers integrate with Koito or build tooling around it.

All endpoints are under the base path `/apis/web/v1`.

---

## Authentication

Some endpoints require authentication; others are publicly accessible depending on your instance configuration.

### Session (cookie-based)
Obtained via `POST /apis/web/v1/login`. The server sets a session cookie used by the web UI.

### API Key (token-based)
Pass the key in the `Authorization` header:
```
Authorization: Token <your-api-key>
```
API keys can be managed from the Koito web UI under Settings → API Keys.

### Endpoint access levels

| Level | Applies to |
|---|---|
| **Public** | `GET /config`, `GET /health`, `POST /login`, `POST /logout` |
| **Login gate** — public by default; session or API key required if login is enabled | All read endpoints (`stats`, `top-*`, `listen-activity`, etc.) |
| **Session or API key** (always enforced) | Write/admin endpoints (`export`, `merge`, `delete`, etc.) |

---

## Common Query Parameters

Many endpoints share these parameters.

### Period (`period`)
Shortcuts for common time ranges. Accepted values:

| Value | Range |
|---|---|
| `day` | Last 24 hours |
| `week` | Last 7 days |
| `month` | Last 30 days |
| `year` | Last 365 days |
| `all_time` | All recorded listens |

### Custom timeframe
Fine-grained range selection (takes precedence over `period`):

| Parameter | Type | Description |
|---|---|---|
| `from` | integer (Unix timestamp) | Range start |
| `to` | integer (Unix timestamp) | Range end (defaults to now) |
| `year` | integer | Calendar year (e.g. `2025`) |
| `month` | integer | Calendar month `1–12` (combined with `year`) |
| `week` | integer | ISO week number (combined with `year`) |

### Timezone (`tz`)
IANA timezone name (e.g. `Europe/Paris`). Can also be set via a `tz` cookie. Falls back to UTC if unset. Affects period boundaries and listen-activity bucketing.

### Pagination
Used by list endpoints (`top-*`, `listens`):

| Parameter | Type | Default | Max |
|---|---|---|---|
| `limit` | integer | `100` | `500` |
| `page` | integer | `1` | — |

---

## Endpoints

### `GET /health`
Returns `200 OK` when the server is ready. Returns `503 Service Unavailable` during startup.

No authentication required.

---

### `GET /config`
Returns server configuration visible to clients.

No authentication required.

**Response:**
```json
{
  "default_theme": ""
}
```

> **Note:** `default_theme` is an empty string `""` when not explicitly configured.

---

### `GET /stats`
Returns aggregate listening statistics.

**Authentication:** Public if login is disabled (default); otherwise requires session or API key

**Query parameters:** [Period / timeframe](#common-query-parameters), [`tz`](#timezone-tz)

**Response:**
```json
{
  "listen_count": 47425,
  "track_count": 1894,
  "album_count": 548,
  "artist_count": 355,
  "minutes_listened": 190781
}
```

> **Note:** `minutes_listened` is computed as total seconds listened divided by 60.

---

### `GET /top-artists`
Returns the most-listened artists for a time period, sorted by listen count.

**Authentication:** Public if login is disabled (default); otherwise requires session or API key

**Query parameters:** [Period / timeframe](#common-query-parameters), [Pagination](#pagination), [`tz`](#timezone-tz)

**Response:** Paginated list of ranked artists.
```json
{
  "items": [
    {
      "item": {
        "id": 12,
        "musicbrainz_id": null,
        "name": "Radiohead",
        "aliases": ["radiohead"],
        "image": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
        "listen_count": 58,
        "time_listened": 0,
        "first_listen": 0,
        "all_time_rank": 0
      },
      "rank": 1
    }
  ],
  "total_record_count": 357,
  "items_per_page": 10,
  "has_next_page": true,
  "current_page": 1
}
```

> **Note:** `time_listened`, `first_listen`, and `all_time_rank` are always `0` in this endpoint. Use `GET /summary` or `GET /artist?id=<id>` to get actual time listened.

---

### `GET /top-albums`
Returns the most-listened albums for a time period.

**Authentication:** Public if login is disabled (default); otherwise requires session or API key

**Query parameters:** [Period / timeframe](#common-query-parameters), [Pagination](#pagination), [`tz`](#timezone-tz)

**Response:** Same paginated structure as `/top-artists`, with album objects:
```json
{
  "items": [
    {
      "item": {
        "id": 42,
        "musicbrainz_id": null,
        "title": "OK Computer",
        "image": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
        "artists": [{ "id": 12, "name": "Radiohead" }],
        "is_various_artists": false,
        "listen_count": 15,
        "time_listened": 0,
        "first_listen": 0,
        "all_time_rank": 0
      },
      "rank": 1
    }
  ],
  "total_record_count": 559,
  "items_per_page": 10,
  "has_next_page": true,
  "current_page": 1
}
```

> **Note:** `time_listened`, `first_listen`, and `all_time_rank` are always `0` in this endpoint.

---

### `GET /top-tracks`
Returns the most-listened tracks for a time period.

**Authentication:** Public if login is disabled (default); otherwise requires session or API key

**Query parameters:** [Period / timeframe](#common-query-parameters), [Pagination](#pagination), `artist_id` (integer, filter by artist), `album_id` (integer, filter by album), [`tz`](#timezone-tz)

**Response:** Same paginated structure, with track objects:
```json
{
  "items": [
    {
      "item": {
        "id": 7,
        "title": "Karma Police",
        "artists": [{ "id": 12, "name": "Radiohead" }],
        "musicbrainz_id": null,
        "listen_count": 32,
        "duration": 0,
        "image": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
        "album_id": 42,
        "time_listened": 0,
        "first_listen": 0,
        "all_time_rank": 0
      },
      "rank": 1
    }
  ],
  "total_record_count": 1888,
  "items_per_page": 10,
  "has_next_page": true,
  "current_page": 1
}
```

> **Note:** `duration`, `time_listened`, `first_listen`, and `all_time_rank` are always `0` in this endpoint. Use `GET /track?id=<id>` or `GET /summary` to get actual values.

---

### `GET /listen-activity`
Returns listen counts bucketed by time step, suitable for charts and heatmaps.

**Authentication:** Public if login is disabled (default); otherwise requires session or API key

**Query parameters:**

| Parameter | Type | Default | Description |
|---|---|---|---|
| `step` | string | `day` | Bucket size: `day`, `week`, `month`, `year` |
| `range` | integer | — | Number of steps to return (going back from today). **Required** when using `step`; omitting it returns `null`. |
| `year` | integer | — | Return data for a specific calendar year |
| `month` | integer | — | Combined with `year`, return a specific month |
| `artist_id` | integer | — | Filter by artist |
| `album_id` | integer | — | Filter by album |
| `track_id` | integer | — | Filter by track |
| `tz` | string | — | [Timezone](#timezone-tz) |

> When `year` (and optionally `month`) is specified, `range` and `step` are ignored and the full calendar period is returned.

**Response:** Array of buckets ordered by `start_time`.
```json
[
  { "start_time": "2026-04-01T00:00:00+02:00", "listens": 75 },
  { "start_time": "2026-04-02T00:00:00+02:00", "listens": 42 },
  { "start_time": "2026-04-03T00:00:00+02:00", "listens": 0 }
]
```

Days with zero listens are included as `{ "listens": 0 }` to ensure a continuous series.

> **Note:** `start_time` timestamps reflect the server's local timezone (e.g., `+01:00` or `+02:00` for Europe/Paris with DST), not UTC `Z`.

---

### `GET /listens`
Returns a paginated list of individual listen records (scrobbles), newest first.

**Authentication:** Public if login is disabled (default); otherwise requires session or API key

**Query parameters:** [Period / timeframe](#common-query-parameters), [Pagination](#pagination), `track_id` (integer, filter by track), [`tz`](#timezone-tz)

> **Note:** A time filter (`period`, `from`/`to`, `year`, etc.) is required. Without one, the server internally uses zero timestamps (`0001-01-01`) as the range, which matches no records.

**Response:**
```json
{
  "items": [
    {
      "time": "2026-04-18T21:14:00Z",
      "track": {
        "id": 7,
        "title": "Karma Police",
        "artists": [{ "id": 12, "name": "Radiohead" }],
        "musicbrainz_id": null,
        "listen_count": 0,
        "duration": 0,
        "image": null,
        "album_id": 0,
        "time_listened": 0,
        "first_listen": 0,
        "all_time_rank": 0
      }
    }
  ],
  "total_record_count": 47377,
  "items_per_page": 100,
  "has_next_page": true,
  "current_page": 1
}
```

> **Note:** Track objects within listen records only contain `id`, `title`, and `artists`; all other stat fields (`listen_count`, `duration`, `time_listened`, etc.) are `0` or `null`. Use `GET /track?id=<id>` to fetch full track metadata.

---

### `GET /now-playing`
Returns the track currently being scrobbled, if any.

**Authentication:** Public if login is disabled (default); otherwise requires session or API key

**Response (nothing playing):**
```json
{ "currently_playing": false, "track": {} }
```

**Response (playing):**
```json
{
  "currently_playing": true,
  "track": {
    "id": 1883,
    "title": "Slow Decline",
    "artists": [{ "id": 355, "name": "Lumi Noir" }],
    "musicbrainz_id": null,
    "listen_count": 37,
    "duration": 221,
    "image": "d8da11cc-3768-4301-ab98-d2ada7a43a6d",
    "album_id": 558,
    "time_listened": 8177,
    "first_listen": 1776338132,
    "all_time_rank": 252
  }
}
```

---

### `GET /artist`
Returns a single artist by ID.

**Authentication:** Public if login is disabled (default); otherwise requires session or API key

**Query parameters:**

| Parameter | Type | Required |
|---|---|---|
| `id` | integer | Yes |

**Response:** Artist object (same shape as items in `/top-artists`, but with all stat fields fully populated: `time_listened`, `first_listen`, and `all_time_rank` are real values).

---

### `GET /album`
Returns a single album by ID.

**Authentication:** Public if login is disabled (default); otherwise requires session or API key

**Query parameters:**

| Parameter | Type | Required |
|---|---|---|
| `id` | integer | Yes |

**Response:** Album object (same shape as items in `/top-albums`, but with all stat fields fully populated: `time_listened`, `first_listen`, and `all_time_rank` are real values).

---

### `GET /track`
Returns a single track by ID.

**Authentication:** Public if login is disabled (default); otherwise requires session or API key

**Query parameters:**

| Parameter | Type | Required |
|---|---|---|
| `id` | integer | Yes |

**Response:** Track object (same shape as items in `/top-tracks`, but with all stat fields fully populated: `duration`, `time_listened`, `first_listen`, and `all_time_rank` are real values).

---

### `GET /search`
Searches artists, albums, and tracks by name.

**Authentication:** Public if login is disabled (default); otherwise requires session or API key

**Query parameters:**

| Parameter | Type | Description |
|---|---|---|
| `q` | string | Search query. Prefix with `id:` to look up all entities with a specific numeric ID. |

**Response:**
```json
{
  "artists": [ /* Artist objects */ ],
  "albums":  [ /* Album objects */ ],
  "tracks":  [ /* Track objects */ ]
}
```

> **Note:** When searching by name, stat fields (`listen_count`, `time_listened`, `first_listen`, `all_time_rank`, `duration`) are `0` in results. When searching by ID (`id:<n>`), full stats are returned. "No results" searches may still return partial matches (e.g., `q=zzz` may match tracks titled "Zzz").

---

### `GET /summary`
Returns a rich summary for a time period, including top 5 artists/albums/tracks and aggregate statistics. Unlike `/top-*` endpoints, `time_listened` is populated here. Note that `first_listen` and `all_time_rank` remain `0` even in summary responses.

**Authentication:** Public if login is disabled (default); otherwise requires session or API key

**Query parameters:** [Period / timeframe](#common-query-parameters), [`tz`](#timezone-tz)

**Response:**
```json
{
  "top_artists": [ /* top 5 RankedItem<Artist> — time_listened is populated */ ],
  "top_albums":  [ /* top 5 RankedItem<Album>  — time_listened is populated */ ],
  "top_tracks":  [ /* top 5 RankedItem<Track>  — time_listened is populated */ ],
  "minutes_listened": 1748,
  "avg_minutes_listened_per_day": 291.0,
  "plays": 418,
  "avg_plays_per_day": 69.67,
  "unique_tracks": 139,
  "unique_albums": 67,
  "unique_artists": 59,
  "new_tracks": 22,
  "new_albums": 8,
  "new_artists": 6
}
```

`new_*` fields count items heard for the first time within the requested period.

---

### `GET /interest`
Returns listen counts distributed across N time buckets for a specific artist, album, or track. Used to show how interest in an item has evolved over time.

**Authentication:** Public if login is disabled (default); otherwise requires session or API key

**Query parameters:**

| Parameter | Type | Required | Description |
|---|---|---|---|
| `buckets` | integer | Yes | Number of time buckets to split history into |
| `artist_id` | integer | — | Filter by artist |
| `album_id` | integer | — | Filter by album |
| `track_id` | integer | — | Filter by track |

**Response:** Array of buckets.
```json
[
  {
    "bucket_start": "2025-01-01T00:00:00Z",
    "bucket_end": "2025-04-01T00:00:00Z",
    "listen_count": 120
  }
]
```

---

### `GET /aliases`
Returns all aliases for a given artist, album, or track.

**Authentication:** Public if login is disabled (default); otherwise requires session or API key

**Query parameters:** Exactly one of `artist_id`, `album_id`, or `track_id` (integer).

**Response:** Array of alias objects.

---

### `GET /artists`
Returns all artists associated with an album or track.

**Authentication:** Public if login is disabled (default); otherwise requires session or API key

**Query parameters:** Exactly one of `album_id` or `track_id` (integer).

**Response:** Array of artist objects (same shape as the artist objects within `/top-artists` items — no rank or pagination wrapper).

---

### `GET /export`
Exports the full listen history as a JSON file (Koito's own format). This is the same file available from the web UI.

**Authentication:** Session or API key (always enforced)

**Response:** JSON file download.

---

## Authentication Endpoints

### `POST /login`
Authenticates a user and sets a session cookie.

No authentication required.

**Form parameters:**

| Parameter | Required | Description |
|---|---|---|
| `username` | Yes | — |
| `password` | Yes | — |
| `remember_me` | No | `"true"` extends session to 30 days (default 24 hours) |

**Response:** `204 No Content`. Sets `koito_session` cookie.

---

### `POST /logout`
Clears the current session cookie.

No authentication required.

**Response:** `204 No Content`.

---

## Write / Admin Endpoints

All endpoints in this section require **session or API key** authentication (always enforced, regardless of login configuration).

---

### `POST /listen`
Manually submits a listen by track ID.

**Authentication:** Session or API key (always enforced)

**Form parameters:**

| Parameter | Required | Description |
|---|---|---|
| `track_id` | Yes | Integer ID of the track |
| `unix` | Yes | Unix timestamp (must not be in the future) |
| `client` | No | Client name (defaults to `"Koito Web UI"`) |

**Response:** `201 Created`.

---

### `DELETE /listen`
Deletes a specific listen record.

**Authentication:** Session or API key (always enforced)

**Query parameters:**

| Parameter | Required | Description |
|---|---|---|
| `track_id` | Yes | Integer ID of the track |
| `unix` | Yes | Unix timestamp of the listen to delete |

**Response:** `204 No Content`.

---

### `DELETE /artist`
Deletes an artist by ID.

**Authentication:** Session or API key (always enforced)

**Query parameters:** `id` (integer, required).

**Response:** `204 No Content`.

---

### `DELETE /album`
Deletes an album by ID.

**Authentication:** Session or API key (always enforced)

**Query parameters:** `id` (integer, required).

**Response:** `204 No Content`.

---

### `DELETE /track`
Deletes a track by ID.

**Authentication:** Session or API key (always enforced)

**Query parameters:** `id` (integer, required).

**Response:** `204 No Content`.

---

### `POST /merge/tracks`
Merges one track into another. All listens from `from_id` are reassigned to `to_id`, then `from_id` is deleted.

**Authentication:** Session or API key (always enforced)

**Query parameters:**

| Parameter | Required | Description |
|---|---|---|
| `from_id` | Yes | Track to merge from (will be deleted) |
| `to_id` | Yes | Track to merge into |

**Response:** `204 No Content`. Returns `500` if `from_id` does not exist.

---

### `POST /merge/albums`
Merges one album into another.

**Authentication:** Session or API key (always enforced)

**Query parameters:**

| Parameter | Required | Description |
|---|---|---|
| `from_id` | Yes | Album to merge from (will be deleted) |
| `to_id` | Yes | Album to merge into |
| `replace_image` | No | `"true"` to replace the target album's image with the source's |

**Response:** `204 No Content`. Returns `500` if `from_id` does not exist.

---

### `POST /merge/artists`
Merges one artist into another.

**Authentication:** Session or API key (always enforced)

**Query parameters:**

| Parameter | Required | Description |
|---|---|---|
| `from_id` | Yes | Artist to merge from (will be deleted) |
| `to_id` | Yes | Artist to merge into |
| `replace_image` | No | `"true"` to replace the target artist's image with the source's |

**Response:** `204 No Content`. Returns `500` if `from_id` does not exist.

---

### `PATCH /album`
Updates album metadata.

**Authentication:** Session or API key (always enforced)

**Query parameters:**

| Parameter | Required | Description |
|---|---|---|
| `id` | Yes | Album ID |
| `is_various_artists` | No | `"true"` or `"false"` |

**Response:** `204 No Content`.

---

### `PATCH /mbzid`
Sets the MusicBrainz ID for an artist, album, or track.

**Authentication:** Session or API key (always enforced)

**Form parameters:** `mbz_id` (UUID, required) + exactly one of `artist_id`, `album_id`, or `track_id`.

**Response:** `204 No Content`.

---

### `POST /replace-image`
Replaces the cover image for an artist or album, either from a URL or a file upload.

**Authentication:** Session or API key (always enforced)

**Form parameters:** Exactly one of `artist_id` or `album_id`, plus either `image_url` (string — must be a direct image URL with an image content-type, no redirects) or `image` (multipart file upload).

**Response:**
```json
{ "success": true, "image": "<new-image-uuid>" }
```

---

### `POST /aliases`
Adds an alias to an artist, album, or track.

**Authentication:** Session or API key (always enforced)

**Form parameters:** `alias` (string, required) + exactly one of `artist_id`, `album_id`, or `track_id`.

**Response:** `201 Created`.

---

### `POST /aliases/delete`
Removes an alias from an artist, album, or track.

**Authentication:** Session or API key (always enforced)

**Form parameters:** `alias` (string, required) + exactly one of `artist_id`, `album_id`, or `track_id`.

**Response:** `204 No Content`.

---

### `POST /aliases/primary`
Sets the primary (display) alias for an artist, album, or track.

**Authentication:** Session or API key (always enforced)

**Form parameters:** `alias` (string, required) + exactly one of `artist_id`, `album_id`, or `track_id`.

**Response:** `204 No Content`.

---

### `POST /artists/primary`
Sets whether an artist is the primary artist for an album or track.

**Authentication:** Session or API key (always enforced)

**Form parameters:**

| Parameter | Required | Description |
|---|---|---|
| `artist_id` | Yes | — |
| `is_primary` | Yes | `"true"` or `"false"` |
| `album_id` or `track_id` | Yes | Exactly one |

**Response:** `204 No Content`.

---

### `GET /user/me`
Returns the currently authenticated user.

**Authentication:** Session or API key (always enforced)

**Response:** User object.

---

### `PATCH /user`
Updates the current user's credentials.

**Authentication:** Session or API key (always enforced)

**Form parameters:** At least one of `username` or `password`.

**Response:** `204 No Content`.

---

### `GET /user/apikeys`
Returns all API keys for the current user.

**Authentication:** Session or API key (always enforced)

**Response:** Array of API key objects.

---

### `POST /user/apikeys`
Generates a new API key.

**Authentication:** Session or API key (always enforced)

**Form parameters:** `label` (string, required).

**Response:** `201 Created` with the new API key object (the key value is only returned once).

---

### `PATCH /user/apikeys`
Updates the label of an API key.

**Authentication:** Session or API key (always enforced)

**Form parameters:** `id` (integer, required), `label` (string, required).

**Response:** `200 OK`.

---

### `DELETE /user/apikeys`
Deletes an API key.

**Authentication:** Session or API key (always enforced)

**Query parameters:** `id` (integer, required).

**Response:** `204 No Content`.

---

## ListenBrainz-Compatible Scrobbling API

Koito implements a subset of the ListenBrainz API, allowing any ListenBrainz-compatible scrobbler to submit listens.

Base path: `/apis/listenbrainz/1`

### `POST /submit-listens`
Submits one or more listens.

**Authentication:** API key (`Authorization: Token <key>`)

**Request body:** Standard ListenBrainz [`submit-listens` payload](https://listenbrainz.readthedocs.io/en/latest/users/api/core.html#post--1-submit-listens).

---

### `GET /validate-token`
Validates an API key.

**Authentication:** API key (`Authorization: Token <key>`)

**Response:**
```json
{
  "code": 200,
  "message": "Token valid.",
  "valid": true,
  "user_name": "youruser"
}
```

---

## Images

### `GET /images/{size}/{filename}`
Serves cached cover art. `size` is a dimension hint (e.g. `300`). `filename` is the UUID from a track/album/artist `image` field plus an extension.

No authentication required.

---

## Error responses

All endpoints return errors in this shape:
```json
{ "error": "human-readable error message" }
```

Rate limiting on `POST /login` returns `429 Too Many Requests` (hardcoded at 10 requests/minute) with the same shape.

