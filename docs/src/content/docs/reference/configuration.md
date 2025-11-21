---
title: Configuration
description: The available configuration options when setting up Koito.
---

Koito is configured using **environment variables**. This is the full list of configuration options supported by Koito.

The suffix `_FILE` is also supported for every environment variable. This allows the use of Docker secrets, for example: `KOITO_DATABASE_URL_FILE=/run/secrets/database-url` will load the content of the file at `/run/secrets/database-url` for the environment variable `KOITO_DATABASE_URL`.

:::caution
If the environment variable is defined without **and** with the suffix at the same time, the content of the environment variable without the `_FILE` suffix will have the higher priority.
:::

##### KOITO_DATABASE_URL
- Required: `true`
- Description: A Postgres connection URI. See https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING-URIS for more information.
##### KOITO_ALLOWED_HOSTS
- Required: `true`
- Description: A list of hosts to allow requests from. E.g. `koito.mydomain.com,192.168.0.100:4110`.
##### KOITO_DEFAULT_USERNAME
- Default: `admin`
- Description: The username for the user that is created on first startup. Only applies when running Koito for the first time.
##### KOITO_DEFAULT_PASSWORD
- Default: `changeme`
- Description: The password for the user that is created on first startup. Only applies when running Koito for the first time.
##### KOITO_DEFAULT_THEME
- Default: `yuu`
- Description: The lowercase name of the default theme to be used by the client. Overridden if a user picks a theme in the theme switcher.
##### KOITO_LOGIN_GATE
- Default: `false`
- Description: When `true`, Koito will not show any statistics unless the user is logged in.
##### KOITO_BIND_ADDR
- Description: The address to bind to. The default blank value is equivalent to `0.0.0.0`.
##### KOITO_LISTEN_PORT
- Default: `4110`
- Description: The port Koito will listen on.
##### KOITO_ENABLE_STRUCTURED_LOGGING
- Default: `false`
- Description: When set to `true`, will log in JSON format.
##### KOITO_ENABLE_FULL_IMAGE_CACHE
- Default: `false`
- Description: When set to `true`, will store the full size downloaded images, which can then be served under `/images/full`.
##### KOITO_LOG_LEVEL
- Default: `info`
- Description: One of `debug | info | warn | error | fatal`
##### KOITO_ARTIST_SEPARATORS_REGEX
- Default: `\s+·\s+`
- Description: The list of regex patterns Koito will use to separate artist strings, separated by two semicolons (`;;`). 
##### KOITO_MUSICBRAINZ_URL
- Default: `https://musicbrainz.org`
- Description: The URL Koito will use to contact MusicBrainz. Replace this value if you have your own MusicBrainz mirror.
##### KOITO_MUSICBRAINZ_RATE_LIMIT
- Default: `1`
- Description: The number of requests to send to the MusicBrainz server per second. Unless you are using your own MusicBrainz mirror, __do not touch this value__.
##### KOITO_ENABLE_LBZ_RELAY
- Default: `false`
- Description: Set to `true` if you want to relay requests from the ListenBrainz endpoints on your Koito server to another ListenBrainz compatible server.
##### KOITO_LBZ_RELAY_URL
- Required: `true` if relays are enabled.
- Description: The URL to which relayed requests will be sent to.
##### KOITO_LBZ_RELAY_TOKEN
- Required: `true` if relays are enabled.
- Description: The user token to send with the relayed ListenBrainz requests.
##### KOITO_CONFIG_DIR
- Default: `/etc/koito`
- Description: The location where import folders and image caches are stored.
##### KOITO_DISABLE_DEEZER
- Default: `false`
- Description: Disables Deezer as a source for finding artist and album images.
##### KOITO_DISABLE_COVER_ART_ARCHIVE
- Default: `false`
- Description: Disables Cover Art Archive as a source for finding album images.
##### KOITO_DISABLE_MUSICBRAINZ
- Default: `false`
##### KOITO_SUBSONIC_URL
- Required: `true` if KOITO_SUBSONIC_PARAMS is set
- Description: The URL of your subsonic compatible music server. For example, `https://navidrome.mydomain.com`.
##### KOITO_SUBSONIC_PARAMS
- Required: `true` if KOITO_SUBSONIC_URL is set
- Description: The `u`, `t`, and `s` authentication parameters to use for authenticated requests to your subsonic server, in the format `u=XXX&t=XXX&s=XXX`. An easy way to find them is to open the network tab in the developer tools of your browser of choice and copy them from a request.
##### KOITO_SKIP_IMPORT
- Default: `false`
- Description: Skips running the importer on startup.
##### KOITO_DISABLE_RATE_LIMIT
- Default: `false`
- Description: When enabled, disables the rate limiter that Koito has on the `/apis/web/v1/login` endpoint.
##### KOITO_THROTTLE_IMPORTS_MS
- Default: `0`
- Description: The amount of time to wait, in milliseconds, between listen imports. Can help when running Koito on low-powered machines.
##### KOITO_IMPORT_BEFORE_UNIX
- Description: A unix timestamp. If an imported listen has a timestamp after this, it will be discarded.
##### KOITO_IMPORT_AFTER_UNIX
- Description: A unix timestamp. If an imported listen has a timestamp before this, it will be discarded.
##### KOITO_FETCH_IMAGES_DURING_IMPORT
- Default: `false`
- Description: When true, images will be downloaded and cached during imports.
##### KOITO_CORS_ALLOWED_ORIGINS
- Default: No CORS policy
- Description: A comma separated list of origins to allow CORS requests from. The special value `*` allows CORS requests from all origins.
