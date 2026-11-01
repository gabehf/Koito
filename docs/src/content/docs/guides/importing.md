---
title: Importing Data
description: How to import your existing listening history from other services into Koito.
---

Koito currently supports the following sources to import data from:

- Koito JSON exports
- Spotify
- Maloja
- LastFM (using https://lastfm.ghan.nl/export/)
- ListenBrainz

:::note
ListenBrainz and LastFM imports can take a long time for large imports due to MusicBrainz requests being throttled at one per second. If you want these imports to go faster, you can [disable MusicBrainz](/reference/configuration/#koito_disable_musicbrainz) in the config while running the importer.
However, this means that artist aliases will not be automatically fetched for imported artists.
You can also use [your own MusicBrainz mirror](https://musicbrainz.org/doc/MusicBrainz_Server/Setup) and [disable MusicBrainz rate limiting](/reference/configuration/#koito_musicbrainz_url) in the config if you want imports to be faster.
:::

## Choose the destination profile

In **Settings → Administration**, select the destination in **Manage profile**,
choose the import format and upload the history file (up to 100 MiB). Confirm
that it belongs to the selected person, then import. ListenBrainz uses a ZIP;
the other formats use JSON. Export history from the same selected profile.

For automatic startup imports, place files under
`<config-directory>/import/<profile-id>/`, not directly under `import/`.
Startup imports run only when `KOITO_SKIP_IMPORT` is not enabled.
The selected directory determines ownership; names inside the file do not change
it. See [Listening profiles](/guides/profiles). Keep a database backup; a failed
import can have saved some records, so check before retrying.

## Spotify

To get your data from Spotify, you first need to request your extended streaming history from [the Spotify privacy page](https://www.spotify.com/us/account/privacy/).
The export could take up to 30 days, according to Spotify. Then, all you have to do is put the `.json` files from your data export into the selected profile's `import/<profile-id>/` folder in your config directory, and restart Koito. The data import will then start automatically.

Koito relies on file names to find files to import. If the files aren't being imported automatically, make sure they contain `Streaming_History_Audio` in the file name.

![The Spotify data export page](../../../assets/spotify_export.png)

## Maloja

You can download your data from Maloja by clicking the `Export` button under Download Data on the `/admin_overview` page of your Maloja instance.
Then, put the resuling `.json` file into the selected profile's `import/<profile-id>/` folder in your config directory, and restart Koito. The data import will then start automatically.

Koito relies on file names to find files to import. If the files aren't being imported automatically, make sure they contain `maloja` in the file name.

:::note
Maloja may have missing or inconsistent track duration information, which means that the 'Hours Listened' statistic may be incorrect after a Maloja import. However, track durations will be filled in as you submit listens using the API.
:::

## LastFM

First, create an export file using [this tool from ghan.nl](https://lastfm.ghan.nl/export/) in JSON format. Then, place the resulting file into the selected profile's `import/<profile-id>/` folder in your config directory.
Once you restart Koito, it will automatically detect the file as a Last FM import, and begin adding your listen activity immediately.

Koito relies on file names to find files to import. If the files aren't being imported automatically, make sure they contain `recenttracks` in the file name.

:::note
LastFM exports do not include track duration information, which means that the 'Hours Listened' statistic may be incorrect after importing.
However, track durations will be filled in as you submit listens using the API.
:::

## ListenBrainz

Create a ListenBrainz export file using [the export tool on the ListenBrainz website](https://listenbrainz.org/settings/export/). Then, place the resulting `.zip` file into the `import`
folder for the destination profile (`import/<profile-id>/`) in your config directory. Once you restart Koito, your ListenBrainz activity will immediately start being imported.

Koito relies on file names to find files to import. If the files aren't being imported automatically, make sure they contain `listenbrainz` in the file name.
