# v0.0.13

## Features

## Enhancements
- Track durations will now be updated using MusicBrainz data where possible, if the duration was not provided by the request. (#27)
- You can now search and merge items by their ID! Just preface the id with `id:`. E.g. `id:123` (#26)
- Hovering over any "hours listened" statistic will now also show the minutes listened.
- An experiemental ARM docker image has been added. (#51)

## Fixes
- Navigating from one page directly to another and then changing the image via drag-and-drop now works as expected. (#25)
- Fixed a bug that caused updated usernames with uppercase letters to create login failures.

## Updates
- Migrations are now embedded to allow for a community AUR package. (#37)