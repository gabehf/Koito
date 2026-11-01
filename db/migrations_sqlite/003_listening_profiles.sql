-- +goose Up
-- Goose runs this migration in one transaction. Do not disable foreign keys.
CREATE TABLE listening_profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT NOT NULL UNIQUE CHECK (
        length(slug) BETWEEN 1 AND 64
        AND slug NOT GLOB '*[^a-z0-9-]*'
        AND substr(slug, 1, 1) != '-'
        AND substr(slug, -1, 1) != '-'
        AND slug NOT IN ('admin', 'apis', 'image', 'public', 'artist', 'album',
                         'track', 'chart', 'listens', 'rewind', 'theme-helper')
    ),
    display_name TEXT NOT NULL CHECK (length(trim(display_name)) > 0),
    avatar_image_id TEXT,
    avatar_icon TEXT,
    created_at INTEGER NOT NULL
);

-- Transitional bridge: old callers still submit UserID until step 2.
-- Deleting an account removes this bridge, never its profile/history.
CREATE TABLE legacy_user_profiles (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    profile_id INTEGER NOT NULL UNIQUE REFERENCES listening_profiles(id) ON DELETE CASCADE
);

-- Stable, collision-free slugs even for reserved names or names such as a_b/a.b.
INSERT INTO listening_profiles (id, slug, display_name, created_at)
SELECT id, CASE WHEN (SELECT COUNT(*) FROM users) = 1 THEN 'default'
                ELSE 'user-' || id END,
       username, unixepoch()
FROM users;
INSERT INTO legacy_user_profiles (user_id, profile_id) SELECT id, id FROM users;

CREATE TABLE api_keys_profiles (
    id INTEGER PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    user_id INTEGER, -- legacy attribution only; no cascading account FK
    profile_id INTEGER NOT NULL REFERENCES listening_profiles(id) ON DELETE CASCADE,
    created_at INTEGER NOT NULL,
    label TEXT NOT NULL DEFAULT ''
);
INSERT INTO api_keys_profiles (id, key, user_id, profile_id, created_at, label)
SELECT k.id, k.key, k.user_id,
       (SELECT profile_id FROM legacy_user_profiles WHERE user_id = k.user_id),
       k.created_at, k.label
FROM api_keys k;
DROP TABLE api_keys;
ALTER TABLE api_keys_profiles RENAME TO api_keys;
CREATE INDEX idx_api_keys_profile_id ON api_keys(profile_id);
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);

CREATE TABLE listens_profiles (
    profile_id INTEGER NOT NULL REFERENCES listening_profiles(id) ON DELETE CASCADE,
    track_id INTEGER NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    listened_at INTEGER NOT NULL,
    user_id INTEGER, -- legacy attribution only; no cascading account FK
    client TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (profile_id, track_id, listened_at)
);
INSERT INTO listens_profiles (profile_id, track_id, listened_at, user_id, client)
SELECT (SELECT profile_id FROM legacy_user_profiles WHERE user_id = l.user_id),
       l.track_id, l.listened_at, l.user_id, l.client
FROM listens l;
DROP TABLE listens;
ALTER TABLE listens_profiles RENAME TO listens;
CREATE INDEX idx_listens_profile_listened_at ON listens(profile_id, listened_at);
CREATE INDEX idx_listens_listened_at ON listens(listened_at);
CREATE INDEX idx_listens_track_id ON listens(track_id);
CREATE INDEX idx_listens_track_id_listened_at ON listens(track_id, listened_at);
CREATE INDEX idx_listens_user_id ON listens(user_id);

-- Compatibility for initial admin creation and existing SaveUser callers.
-- This bridge will be retired with explicit profile management, not public signup.
-- +goose StatementBegin
CREATE TRIGGER trg_legacy_user_profile AFTER INSERT ON users
BEGIN
    INSERT INTO listening_profiles (slug, display_name, created_at)
    VALUES (CASE WHEN NOT EXISTS (SELECT 1 FROM listening_profiles) THEN 'default'
                 ELSE 'user-' || NEW.id END, NEW.username, unixepoch());
    INSERT INTO legacy_user_profiles (user_id, profile_id)
    VALUES (NEW.id, last_insert_rowid());
END;
-- +goose StatementEnd

-- Relay configuration and credential ownership are part of the same atomic upgrade.
CREATE TABLE profile_relays (
    profile_id INTEGER PRIMARY KEY REFERENCES listening_profiles(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    token TEXT NOT NULL
);
CREATE TABLE legacy_relay_binding (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    profile_id INTEGER REFERENCES listening_profiles(id) ON DELETE SET NULL
);
-- Prevent reassignment of a credential to another listening identity.
-- +goose StatementBegin
CREATE TRIGGER trg_api_key_profile_immutable
BEFORE UPDATE OF profile_id ON api_keys
WHEN NEW.profile_id != OLD.profile_id
BEGIN
    SELECT RAISE(ABORT, 'API keys cannot be transferred between profiles');
END;
-- +goose StatementEnd

-- +goose Down
-- A downgrade could collapse identical listens across profiles and lose data.
-- Deliberately fail transactionally; restore the pre-migration backup instead.
CREATE TEMP TABLE profiles_downgrade_guard (value INTEGER);
-- +goose StatementBegin
CREATE TEMP TRIGGER profiles_downgrade_block
BEFORE INSERT ON profiles_downgrade_guard
BEGIN
    SELECT RAISE(ABORT, 'Profiles migration cannot be downgraded: restore the pre-profiles backup');
END;
-- +goose StatementEnd
INSERT INTO profiles_downgrade_guard VALUES (1);
