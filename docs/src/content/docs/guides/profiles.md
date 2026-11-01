---
title: Listening profiles
description: Share one Koito instance using public profiles and profile-specific API keys.
---

## Browse without an account

The root `/` lists profiles. Open `/<slug>` to view a person's history, charts,
statistics and currently playing track, or use the sidebar profile selector.
The URL determines the profile, so two browser tabs can show different people.
Switching profiles keeps the screen and period where possible and drops filters
that refer to the previous person's media.

Profiles and listening data are public to anyone who can reach the instance.
Selecting a profile does not grant permission to change data. Private profiles
and separate listener passwords are not provided.

## Manage profiles

Open **Settings → Administration**, sign in with the administrator account, and
select a profile in **Manage profile**. Administration stays inside Settings;
`/admin` is also available as a direct entry point.

Create or edit a display name, URL name and optional photo or fallback icon.
Renaming a URL keeps the profile's ID, history, keys and forwarding configuration;
the previous URL returns 404. Administrator login credentials are independent of
the profile name. Images may be up to 5 MiB.

Each profile starts with an API key. Create additional labelled keys for players,
copy them into the appropriate player's settings, or revoke them to stop future
submissions. Keys cannot be transferred between people. A player's key selects
its profile; it does not authorize profile management or other administration.
See [Setting up the Scrobbler](/guides/scrobbler).

Administrators can import/export a selected profile's history and configure its
ListenBrainz forwarding. Deleting a profile requires confirmation of its name
and listen count and removes that profile's listens, keys and relay. Other
profiles and shared catalog entries remain intact. Deleting a profile does not
delete the administrator account.

## Upgrade an existing SQLite installation

1. Stop the existing Koito process and back up its complete configuration/data
   directory, including the SQLite database. Keep the backup for rollback.
2. Install a build containing listening-profile support and reuse that directory.
3. Start Koito. Migration `003_listening_profiles.sql` upgrades schema 2 to 3 in
   one transaction. Do not edit the schema version manually.
4. Sign in to Administration and verify existing history and keys before adding
   more people. A single existing user's history becomes `/default`; multiple
   existing users receive `user-<id>` slugs. Existing credentials and keys are
   preserved.

Existing installations must already use SQLite. The schema migration cannot be
downgraded in place: stop Koito, restore the pre-upgrade directory backup and run
the previous version to roll back.

## API compatibility

Players continue to submit to `/apis/listenbrainz/1/submit-listens` with
`Authorization: Token YOUR_KEY`. The server resolves the destination from the
key rather than from a profile supplied in the payload.

Web reads use `/apis/web/v1/profiles/<slug>/...`. Legacy unscoped read routes
resolve only when exactly one profile exists; otherwise they return `409` asking
for explicit profile selection. Administrative actions require an administrator
session, not a player API key.
