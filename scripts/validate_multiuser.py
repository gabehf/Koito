#!/usr/bin/env python3
"""Exercise the real API with disposable fresh and migrated SQLite databases.

Build first: CGO_ENABLED=1 go build -o /tmp/koito-validation ./cmd/api
Run: python3 scripts/validate_multiuser.py --binary /tmp/koito-validation
An optional --example-db is opened read-only and copied before any mutations.
No player or external music service is contacted. Credentials are never printed.
"""
import argparse
import concurrent.futures
import contextlib
import hashlib
import http.cookiejar
import json
import os
from pathlib import Path
import re
import signal
import socket
import sqlite3
import subprocess
import tempfile
import time
import urllib.error
import urllib.request
import uuid

ROOT = Path(__file__).resolve().parents[1]
WEB = "/apis/web/v1"
LBZ = "/apis/listenbrainz/1"
PASSWORD = "validation-password"
LATEST_SCHEMA_VERSION = max(int(p.name.split("_", 1)[0])
                            for p in (ROOT / "db/migrations_sqlite").glob("[0-9]*_*.sql"))


def snapshot(database):
    """Hash the preserved legacy data without disclosing any credential."""
    with sqlite3.connect(f"{database.resolve().as_uri()}?mode=ro", uri=True) as conn:
        tables = [row[0] for row in conn.execute(
            "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' AND name!='goose_db_version' ORDER BY name")]
        preserved = {}
        for table in tables:
            # These tables gain fields during migration; hash the legacy columns.
            columns = {"listens": "track_id,listened_at,user_id,client",
                       "api_keys": "id,key,user_id,created_at,label"}.get(table, "*")
            values = sorted(repr(row) for row in conn.execute(f'SELECT {columns} FROM "{table}"'))
            preserved[table] = hashlib.sha256("\n".join(values).encode()).hexdigest()
        return preserved


class Server:
    def __init__(self, binary, directory):
        self.binary, self.directory = binary, directory
        self.process = None
        self.cookies = http.cookiejar.CookieJar()
        self.admin = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(self.cookies))
        self.public = urllib.request.build_opener()

    def start(self):
        self.directory.mkdir(parents=True, exist_ok=True)
        with socket.socket() as sock:
            sock.bind(("127.0.0.1", 0))
            port = sock.getsockname()[1]
        self.origin = f"http://127.0.0.1:{port}"
        self.env = {k: v for k, v in os.environ.items() if not k.startswith("KOITO_")}
        self.env.update(KOITO_CONFIG_DIR=str(self.directory), KOITO_BIND_ADDR="127.0.0.1",
                        KOITO_LISTEN_PORT=str(port), KOITO_DEFAULT_USERNAME="validation-admin",
                        KOITO_DEFAULT_PASSWORD=PASSWORD, KOITO_DISABLE_MUSICBRAINZ="true",
                        KOITO_DISABLE_DEEZER="true", KOITO_DISABLE_COVER_ART_ARCHIVE="true",
                        KOITO_SKIP_IMPORT="true", KOITO_DISABLE_RATE_LIMIT="true", KOITO_LOG_LEVEL="warn")
        self.log = open(self.directory / "server.log", "ab")
        os.chmod(self.directory / "server.log", 0o600)
        self.process = subprocess.Popen([str(self.binary)], cwd=ROOT, env=self.env,
                                        stdout=self.log, stderr=self.log)
        for _ in range(300):
            if self.process.poll() is not None:
                raise RuntimeError(f"Validation API exited; see {self.directory / 'server.log'}")
            try:
                self.request("GET", WEB + "/health")
                return
            except (AssertionError, OSError):
                time.sleep(0.1)
        raise RuntimeError("Validation API did not become ready")

    def stop(self):
        if self.process and self.process.poll() is None:
            self.process.send_signal(signal.SIGINT)
            try:
                self.process.wait(timeout=10)
            except subprocess.TimeoutExpired:
                self.process.kill()
                self.process.wait()
        if hasattr(self, "log"):
            self.log.close()

    def request(self, method, path, body=None, *, admin=False, token=None, expected=200, headers=None):
        request_headers = {"Origin": self.origin}
        if body is not None and not isinstance(body, bytes):
            body = json.dumps(body).encode()
            request_headers["Content-Type"] = "application/json"
        if token:
            request_headers["Authorization"] = "Token " + token
        request_headers.update(headers or {})
        req = urllib.request.Request(self.origin + path, body, request_headers, method=method)
        try:
            response = (self.admin if admin else self.public).open(req, timeout=30)
        except urllib.error.HTTPError as error:
            response = error
        with response:
            content = response.read()
            # Do not include bodies in failures: an administrative response may contain tokens.
            assert response.status == expected, f"{method} {path}: expected {expected}, got {response.status}"
            if not content:
                return None
            if "json" in response.headers.get("Content-Type", ""):
                return json.loads(content)
            return content

    def login(self, username="validation-admin", password=PASSWORD, expected=204):
        return self.request("POST", WEB + "/login", {"username": username, "password": password},
                            admin=True, expected=expected)

    def upload(self, path, format_name, data):
        boundary = uuid.uuid4().hex
        body = (f"--{boundary}\r\nContent-Disposition: form-data; name=\"format\"\r\n\r\n{format_name}\r\n"
                f"--{boundary}\r\nContent-Disposition: form-data; name=\"file\"; filename=\"history.json\"\r\n"
                f"Content-Type: application/json\r\n\r\n").encode() + data + f"\r\n--{boundary}--\r\n".encode()
        return self.request("POST", path, body, admin=True, expected=204,
                            headers={"Content-Type": f"multipart/form-data; boundary={boundary}"})


def scrobble(timestamp, kind="single"):
    return {"listen_type": kind, "payload": [{"listened_at": timestamp,
            "track_metadata": {"artist_name": "Validation Artist", "release_name": "Validation Album",
                               "track_name": "Shared validation song", "additional_info": {"duration": 180}}}]}


def validate_journey(server):
    with sqlite3.connect(server.directory / "koito.db") as conn:
        assert conn.execute("SELECT MAX(version_id) FROM goose_db_version WHERE is_applied=1").fetchone()[0] == LATEST_SCHEMA_VERSION
    server.login()
    before = server.request("GET", WEB + "/profiles")
    originals = {p["id"]: server.request("GET", WEB + f'/profiles/{p["slug"]}/stats')["listen_count"] for p in before}
    profiles, keys = [], []
    for slug in ["validation-alpha", "validation-beta"]:
        p = server.request("POST", WEB + "/profiles", {"display_name": slug, "slug": slug, "avatar_icon": "headphones"}, admin=True, expected=201)
        profiles.append(p)
        k = server.request("GET", WEB + f"/profiles/{slug}/user/apikeys", admin=True)
        assert len(k) == 1 and k[0]["profile_id"] == p["id"]
        keys.append(k[0])
    alpha, beta = [WEB + f'/profiles/{p["slug"]}' for p in profiles]
    second_key = server.request("POST", alpha + "/user/apikeys", {"label": "Phone"}, admin=True, expected=201)
    timestamp = int(time.time()) - 3600
    # Identical song and timestamp from different players must produce two rows.
    with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
        list(pool.map(lambda k: server.request("POST", LBZ + "/submit-listens", scrobble(timestamp), token=k["key"]), keys))
    payload = scrobble(timestamp + 1)
    payload["profile_id"] = profiles[1]["id"]
    payload["payload"][0]["profile_id"] = profiles[1]["id"]
    server.request("POST", LBZ + "/submit-listens", payload, token=second_key["key"])
    assert server.request("GET", alpha + "/stats")["listen_count"] == 2
    assert server.request("GET", beta + "/stats")["listen_count"] == 1
    for prefix in [alpha, beta]:
        for route in ["/listens", "/top/artists", "/top/albums", "/top/tracks", "/listen-activity", "/first-activity", "/summary?period=all_time"]:
            server.request("GET", prefix + route)
    for prefix, count in [(alpha, 2), (beta, 1)]:
        assert server.request("GET", prefix + "/listens")["total_record_count"] == count
        for kind, item_type in [("tracks", "track"), ("artists", "artist"), ("albums", "album")]:
            ranking = server.request("GET", prefix + f"/top/{kind}?period=all_time")
            assert len(ranking["items"]) == 1
            item = ranking["items"][0]["item"]
            assert item["listen_count"] == count
            assert server.request("GET", prefix + f'/{item_type}/{item["id"]}')["listen_count"] == count
        activity = server.request("GET", prefix + "/listen-activity?step=day&range=7&tz=UTC")
        assert sum(day["listens"] for day in activity["activity"]) == count
        assert len(server.request("GET", prefix + "/search?q=Shared%20validation")["tracks"]) == 1
    server.request("GET", WEB + "/stats", expected=409)
    # The built client also serves direct links and refreshes, without a Go route per page.
    page = server.request("GET", "/")
    assert b"<!doctype html" in page.lower(), "Build the frontend before this validation"
    for path in ["/validation-alpha", "/validation-beta/chart/top/artists?period=all_time", "/admin"]:
        assert server.request("GET", path) == page
    assets = set(re.findall(rb'"(/assets/[^" ]+\.js)"', page))
    assert assets, "Production JS assets missing from client HTML"
    for asset in assets:
        server.request("GET", asset.decode())
    assert server.request("GET", alpha + "/summary?period=all_time")["plays"] == 2
    assert server.request("GET", beta + "/summary?period=all_time")["plays"] == 1
    server.request("POST", LBZ + "/submit-listens", scrobble(0, "playing_now"), token=keys[1]["key"])
    assert server.request("GET", beta + "/now-playing")["currently_playing"] is True
    assert server.request("GET", alpha + "/now-playing")["currently_playing"] is False
    # Ingestion keys never stand in for an administrator's session.
    for route, method in [("/user/apikeys", "GET"), ("/export", "GET"), ("/import", "POST"), ("", "DELETE")]:
        server.request(method, alpha + route, token=keys[0]["key"], expected=401)
    server.request("POST", WEB + "/profiles", {}, token=keys[0]["key"], expected=401)
    server.request("PATCH", alpha, {"display_name": "Stolen"}, admin=True,
                   headers={"Origin": "http://external.invalid"}, expected=403)
    metadata = json.dumps(server.request("GET", WEB + "/profiles"))
    assert all(k["key"] not in metadata for k in keys + [second_key])
    exported = server.request("GET", alpha + "/export", admin=True)
    assert len(exported["listens"]) == 2
    server.upload(beta + "/import", "koito", json.dumps(exported).encode())
    assert server.request("GET", beta + "/stats")["listen_count"] == 2
    renamed = server.request("PATCH", alpha, {"display_name": "Alpha renamed", "slug": "validation-renamed"}, admin=True)
    assert renamed["id"] == profiles[0]["id"]
    server.request("GET", alpha + "/stats", expected=404)
    alpha = WEB + "/profiles/validation-renamed"
    assert server.request("GET", LBZ + "/validate-token", token=keys[0]["key"])["user_name"] == "validation-renamed"
    server.request("DELETE", alpha + f'/user/apikeys/{keys[0]["id"]}', admin=True, expected=204)
    server.request("POST", LBZ + "/submit-listens", scrobble(timestamp + 2), token=keys[0]["key"], expected=401)
    assert server.request("GET", alpha + "/stats")["listen_count"] == 2
    server.request("POST", LBZ + "/submit-listens", scrobble(timestamp + 2), token=second_key["key"])
    assert server.request("GET", alpha + "/stats")["listen_count"] == 3
    server.request("DELETE", beta, {"display_name": profiles[1]["display_name"], "listen_count": 2}, admin=True, expected=204)
    server.request("GET", beta + "/stats", expected=404)
    assert server.request("GET", alpha + "/stats")["listen_count"] == 3
    for p in before:
        assert server.request("GET", WEB + f'/profiles/{p["slug"]}/stats')["listen_count"] == originals[p["id"]]
    # Account updates are separate from the public profile's display name and slug.
    server.request("PATCH", WEB + "/user", {"username": "validation-renamed-admin"}, admin=True, expected=204)
    server.request("PATCH", WEB + "/user", {"password": "validation-new-password"}, admin=True, expected=204)
    server.login("validation-renamed-admin", "validation-new-password")
    assert server.request("GET", alpha)["display_name"] == "Alpha renamed"
    server.request("POST", LBZ + "/submit-listens", scrobble(0, "playing_now"), token=second_key["key"])
    assert server.request("GET", alpha + "/now-playing")["currently_playing"] is True
    server.stop()
    server.start()
    server.login("validation-renamed-admin", "validation-new-password")
    assert server.request("GET", alpha + "/stats")["listen_count"] == 3
    assert server.request("GET", LBZ + "/validate-token", token=second_key["key"])["valid"] is True
    assert server.request("GET", alpha + "/now-playing")["currently_playing"] is False
    with sqlite3.connect(server.directory / "koito.db") as conn:
        assert conn.execute("PRAGMA integrity_check").fetchall() == [("ok",)]
        assert conn.execute("PRAGMA foreign_key_check").fetchall() == []
        assert conn.execute("SELECT COUNT(*) FROM listens l LEFT JOIN listening_profiles p ON p.id=l.profile_id WHERE p.id IS NULL").fetchone()[0] == 0
        assert conn.execute("SELECT COUNT(*) FROM api_keys k LEFT JOIN listening_profiles p ON p.id=k.profile_id WHERE p.id IS NULL").fetchone()[0] == 0
    print(f"PASS {server.directory.name}: administration → keys → concurrent ingestion → public reads → import/export → revocation/deletion → restart; SQLite integrity OK")


def create_legacy(database, password_hash):
    with sqlite3.connect(database) as conn:
        for migration in ["001_initial_schema.sql", "002_unique_mbid.sql"]:
            text = (ROOT / "db/migrations_sqlite" / migration).read_text().split("-- +goose Down")[0]
            conn.executescript(text)
        conn.executescript("""CREATE TABLE goose_db_version(id INTEGER PRIMARY KEY AUTOINCREMENT, version_id INTEGER NOT NULL, is_applied BOOLEAN NOT NULL, tstamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
          INSERT INTO goose_db_version(version_id,is_applied) VALUES (0,1),(1,1),(2,1);
          INSERT INTO artists(id) VALUES(1);
          INSERT INTO artist_aliases VALUES(1,'Legacy artist','validation',1);
          INSERT INTO releases(id) VALUES(1);
          INSERT INTO release_aliases VALUES(1,'Legacy album','validation',1);
          INSERT INTO artist_releases VALUES(1,1,1);
          INSERT INTO tracks(id,release_id,duration) VALUES(1,1,180);
          INSERT INTO track_aliases VALUES(1,'Legacy track','validation',1);
          INSERT INTO artist_tracks VALUES(1,1,1);""")
        conn.execute("INSERT INTO users(id,username,role,password) VALUES(1,'validation-admin','admin',?)", (password_hash,))
        conn.execute("INSERT INTO api_keys(key,user_id,created_at,label) VALUES('legacy-validation-token',1,1700000000,'Legacy player')")
        conn.execute("INSERT INTO listens(track_id,listened_at,user_id,client) VALUES(1,1700000000,1,'legacy-player')")


def run(binary, root, example=None):
    fresh = Server(binary, root / "fresh")
    with contextlib.ExitStack() as stack:
        stack.callback(fresh.stop)
        fresh.start()
        with sqlite3.connect(fresh.directory / "koito.db") as conn:
            password_hash = conn.execute("SELECT password FROM users WHERE role='admin'").fetchone()[0]
        validate_journey(fresh)
        fresh.stop()
        legacy_dir = root / "migrated-v2"
        legacy_dir.mkdir()
        create_legacy(legacy_dir / "koito.db", password_hash)
        legacy_before = snapshot(legacy_dir / "koito.db")
        migrated = Server(binary, legacy_dir)
        stack.callback(migrated.stop)
        migrated.start()
        after = snapshot(legacy_dir / "koito.db")
        assert all(after[table] == digest for table, digest in legacy_before.items())
        assert list(legacy_dir.glob("koito.db.pre-profiles-*.bak"))
        assert migrated.request("GET", WEB + "/profiles/default/stats")["listen_count"] == 1
        assert migrated.request("GET", LBZ + "/validate-token", token="legacy-validation-token")["valid"] is True
        validate_journey(migrated)
        migrated.stop()
        if example:
            original = snapshot(example)
            sample_dir = root / "example-copy"
            sample_dir.mkdir()
            sample_db = sample_dir / "koito.db"
            with sqlite3.connect(f"{example.resolve().as_uri()}?mode=ro", uri=True) as source, sqlite3.connect(sample_db) as target:
                source.backup(target)
            sample = Server(binary, sample_dir)
            stack.callback(sample.stop)
            sample.start()
            current = snapshot(sample_db)
            assert all(current[table] == digest for table, digest in original.items())
            with sqlite3.connect(sample_db) as conn:
                assert conn.execute("PRAGMA integrity_check").fetchone()[0] == "ok"
                assert not conn.execute("PRAGMA foreign_key_check").fetchall()
            sample.stop()
            # Only the disposable copy gets a temporary known password for the journey.
            reset_env = {**sample.env, "KOITO_LOG_LEVEL": "info"}
            reset = subprocess.run([str(binary), "reset-password"], cwd=ROOT, env=reset_env,
                                   capture_output=True, text=True, check=True)
            password = re.search(r"New password: ([A-Za-z0-9]+)", reset.stdout)
            assert password, "No temporary password returned for example copy"
            with sqlite3.connect(sample_db) as conn:
                username = conn.execute("SELECT username FROM users WHERE role='admin' LIMIT 1").fetchone()[0]
            sample.start()
            sample.login(username, password.group(1))
            # Normalize only the copied admin credentials so the same journey is reusable.
            sample.request("PATCH", WEB + "/user", {"username": "validation-admin", "password": PASSWORD}, admin=True, expected=204)
            validate_journey(sample)
            sample.stop()
            assert snapshot(example) == original, "Example source changed during validation"
            print("PASS example source: profiles, listens, passwords, tokens and catalog preserved (read-only source)")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", required=True, type=Path)
    parser.add_argument("--example-db", type=Path)
    parser.add_argument("--work-dir", type=Path, help="Create a new directory and retain disposable databases/logs for inspection")
    args = parser.parse_args()
    binary = args.binary.resolve()
    assert binary.is_file(), "Build the API binary first"
    if args.work_dir:
        root = args.work_dir.resolve()
        root.mkdir(parents=True, exist_ok=False)
        os.chmod(root, 0o700)
        run(binary, root, args.example_db)
    else:
        with tempfile.TemporaryDirectory(prefix="koito-multiuser-") as directory:
            run(binary, Path(directory), args.example_db)


if __name__ == "__main__":
    main()
