---
title: Jellyfin Quickstart
description: How to set up Koito to work with your Jellyfin instance.
---

## Configure Koito

Point Koito at Jellyfin so album and artist images are fetched from your library.

```yaml title="compose.yaml"
services:
  koito:
    image: gabehf/koito:latest
    container_name: koito
    environment:
      - KOITO_JELLYFIN_URL=https://jellyfin.mydomain.com
      - KOITO_JELLYFIN_API_KEY=<jellyfin_api_key>
      # Optional if Items searches return nothing with the API key alone:
      # - KOITO_JELLYFIN_USER_ID=<jellyfin_user_id>
    ports:
      - "4110:4110"
    volumes:
      - ./koito:/etc/koito
    restart: unless-stopped
```

Create an API key in Jellyfin under **Dashboard → API Keys**.

:::note
Koito uses SQLite for all new installations as of `v0.2.1`. If you are running a version older than that, you will need to set `KOITO_SQLITE_ENABLED=true`.
:::

## Configure Jellyfin scrobbling

Install the [ListenBrainz plugin for Jellyfin](https://github.com/lyarenei/jellyfin-plugin-listenbrainz) and set the ListenBrainz base URL to your Koito instance:

`{your_koito_url}/apis/listenbrainz/1`

Use an API key from the Koito UI as the ListenBrainz token. Listens will be stored in Koito; images for new artists/albums will be resolved from Jellyfin when possible.
