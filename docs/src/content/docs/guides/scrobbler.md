---
title: Setting up the Scrobbler
description: Configure per-profile player API keys and optional ListenBrainz forwarding.
---

## Create a profile and get its key

Open Settings using the sidebar button or `\`, then open **Administration** and
sign in as the administrator. Use the credentials configured at installation;
change the initial password if you used the defaults.

Select a person in **Manage profile**, or create a profile with a display name,
URL name and optional photo/icon. Each new profile receives an initial API key.
The **API keys** section lets the administrator copy, label, create and revoke
keys for that profile. A phone and music server can use different keys belonging
to the same person. Revoking a key preserves existing history.

Configure the player with the base URL `{your_koito_address}/apis/listenbrainz/1`
and that person's Koito API key as its ListenBrainz token. The submission endpoint
is `/apis/listenbrainz/1/submit-listens`; requests use
`Authorization: Token YOUR_KEY`. The key determines the destination profile.

Listeners browse public profiles without signing in. The root `/` shows profile
selection; `/<slug>` opens a particular profile. See [Listening profiles](/guides/profiles)
for administration and upgrade details.

## Forward a profile to another server

Under **Settings → Administration → Manage profile**, open **ListenBrainz forwarding**:

1. Enable forwarding for that profile.
2. Enter the destination base URL, such as `https://api.listenbrainz.org/1` or
   `https://another-koito.example/apis/listenbrainz/1`. Do not append `/submit-listens`.
3. Enter that person's token from the destination server. This is a separate
   credential from the Koito key supplied to the player; it is visible to administrators.
4. Use **Test connection** to validate the current fields. The test does not save
   credentials or submit a listen. Save using **Save forwarding**.

The screen shows the saved state and unsaved changes. Each profile forwards only
its own incoming player submissions. Different people can use the same server
with different external tokens; reusing one external token sends their listens
to the same external account. Do not configure a loop between Koito instances.

To stop forwarding, uncheck the option and save. This removes the saved URL/token
and preserves local history. Enabling again requires entering the credentials.
Forwarding is asynchronous and has no persistent delivery queue; historical
imports, edits and deletions are not synchronized to the destination.

The legacy `KOITO_ENABLE_LBZ_RELAY`, `KOITO_LBZ_RELAY_URL` and
`KOITO_LBZ_RELAY_TOKEN` variables only bootstrap a single eligible existing profile
once. With multiple profiles, configure destinations explicitly in Administration.
Changing or disabling these variables does not remove a relay already saved in
the database. See the [configuration reference](/reference/configuration).
