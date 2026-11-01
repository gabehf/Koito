#!/usr/bin/env python3
"""Validate an in-place upgrade on a read-only snapshot of a legacy database.

The actual API startup performs migration; this script never executes migration
SQL and never resets credentials. The source and the normal instance are unused
for writes. The required new work directory retains the migrated copy, automatic
backup, server log and a report without tokens, hashes or session identifiers.
"""
import argparse
from contextlib import closing
import hashlib
import json
import os
from pathlib import Path
import sqlite3

from validate_multiuser import LBZ, WEB, LATEST_SCHEMA_VERSION, Server, snapshot


def connect_ro(database):
    return sqlite3.connect(database.resolve().as_uri() + "?mode=ro", uri=True)


def version(conn):
    return conn.execute("SELECT MAX(version_id) FROM goose_db_version WHERE is_applied=1").fetchone()[0]


def counts(conn):
    return {table: conn.execute(f"SELECT COUNT(*) FROM {table}").fetchone()[0]
            for table in ["users", "api_keys", "sessions", "listens", "artists", "releases", "tracks"]}


def physical_hashes(source):
    return {p.name: hashlib.sha256(p.read_bytes()).hexdigest()
            for p in [source, Path(str(source) + "-wal")] if p.exists()}


def validate(binary, source, work):
    source_digest, source_files = snapshot(source), physical_hashes(source)
    database = work / "koito.db"
    # The SQLite backup API includes committed WAL, unlike copying only koito.db.
    with closing(connect_ro(source)) as original, closing(sqlite3.connect(database)) as target:
        old_version, before_counts = version(original), counts(original)
        original.backup(target)
    assert old_version < 3, "This test expects a pre-profile database (migration version 1 or 2)"
    assert snapshot(database) == source_digest, "Snapshot differs from source"
    server = Server(binary, work)
    try:
        server.start()  # Run normal bootstrap: no migration calls from the test.
        after_digest = snapshot(database)
        assert all(after_digest[table] == digest for table, digest in source_digest.items()), "Legacy data changed"
        backups = list(work.glob("koito.db.pre-profiles-*.bak"))
        assert len(backups) == 1, "Expected exactly one automatic pre-migration backup"
        assert backups[0].stat().st_mode & 0o777 == 0o600, "Backup permissions differ from 0600"
        assert snapshot(backups[0]) == source_digest, "Backup differs from original data"
        with closing(connect_ro(backups[0])) as backup:
            assert version(backup) == old_version
            assert not backup.execute("SELECT 1 FROM sqlite_master WHERE name='listening_profiles'").fetchone()
        public_profiles = server.request("GET", WEB + "/profiles")
        profile_results = []
        with closing(connect_ro(database)) as migrated:
            new_version, after_counts = version(migrated), counts(migrated)
            assert new_version == LATEST_SCHEMA_VERSION and new_version > old_version
            assert before_counts == after_counts
            assert migrated.execute("PRAGMA integrity_check").fetchall() == [("ok",)]
            assert migrated.execute("PRAGMA foreign_key_check").fetchall() == []
            assert migrated.execute("SELECT COUNT(*) FROM listens l LEFT JOIN listening_profiles p ON p.id=l.profile_id WHERE p.id IS NULL").fetchone()[0] == 0
            assert migrated.execute("SELECT COUNT(*) FROM api_keys k LEFT JOIN listening_profiles p ON p.id=k.profile_id WHERE p.id IS NULL").fetchone()[0] == 0
            for profile in public_profiles:
                mapped = migrated.execute("SELECT user_id FROM legacy_user_profiles WHERE profile_id=?", (profile["id"],)).fetchone()
                assert mapped is not None, "Legacy user-to-profile association missing"
                with closing(connect_ro(source)) as original:
                    expected = original.execute("SELECT COUNT(*) FROM listens WHERE user_id=?", mapped).fetchone()[0]
                api_stats = server.request("GET", WEB + f'/profiles/{profile["slug"]}/stats')
                assert api_stats["listen_count"] == expected
                for route in ["/listens", "/top/artists", "/top/albums", "/top/tracks", "/listen-activity", "/first-activity", "/summary?period=all_time", "/now-playing"]:
                    server.request("GET", WEB + f'/profiles/{profile["slug"]}' + route)
                profile_results.append({"id": profile["id"], "slug": profile["slug"], "listens": expected,
                                        "tracks_heard": api_stats["track_count"]})
            # The exact old token must work, without rotating it or revealing it.
            keys = migrated.execute("SELECT key,profile_id FROM api_keys").fetchall()
            assert keys, "Provide a legacy example with a key to verify ingestion credentials"
            for token, profile_id in keys:
                result = server.request("GET", LBZ + "/validate-token", token=token)
                expected_slug = migrated.execute("SELECT slug FROM listening_profiles WHERE id=?", (profile_id,)).fetchone()[0]
                assert result["valid"] is True and result["user_name"] == expected_slug
                assert token not in json.dumps(public_profiles)
                server.request("GET", WEB + f"/profiles/{expected_slug}/user/apikeys", token=token, expected=401)
            pk = [row[1] for row in sorted(migrated.execute("PRAGMA table_info(listens)"), key=lambda row: row[5]) if row[5]]
            assert pk == ["profile_id", "track_id", "listened_at"]
        server.stop()
        stable_digest = snapshot(database)
        assert all(stable_digest[table] == digest for table, digest in source_digest.items()), "Shutdown changed legacy data"
        # A second startup must neither create profiles nor run migrations again.
        server.start()
        assert server.request("GET", WEB + "/profiles") == public_profiles
        for profile in profile_results:
            assert server.request("GET", WEB + f'/profiles/{profile["slug"]}/stats')["listen_count"] == profile["listens"]
        server.stop()
        assert snapshot(database) == stable_digest, "Restart changed persisted data"
        assert list(work.glob("koito.db.pre-profiles-*.bak")) == backups
        assert snapshot(source) == source_digest, "Source database changed"
        assert physical_hashes(source) == source_files, "Source database/WAL bytes changed"
    finally:
        server.stop()
    report = {"result": "PASS", "source": str(source), "database": str(database),
              "from_version": old_version, "to_version": new_version, "counts_before": before_counts,
              "counts_after": after_counts, "profiles": profile_results, "automatic_backup": str(backups[0]),
              "backup_permissions": "0600", "source_database_and_wal_unchanged": True,
              "legacy_table_digests_unchanged": True, "existing_tokens_valid": len(keys),
              "password_hashes_and_sessions_preserved": True, "foreign_key_check": [],
              "integrity_check": "ok", "restart_idempotent": True,
              "credential_reset_performed": False, "validation_api_stopped": True}
    (work / "report.json").write_text(json.dumps(report, indent=2, ensure_ascii=False) + "\n")
    lines = ["# Teste de atualização do banco legado", "", f"Resultado: **PASS — versão {old_version} → {new_version}**.", "",
             "A migração foi executada pela inicialização normal do binário atual, sem comandos SQL de migração no teste.", "",
             "| Tabela | Antes | Depois |", "| --- | ---: | ---: |"]
    lines += [f"| {table} | {count} | {after_counts[table]} |" for table, count in before_counts.items()]
    lines += ["", "Perfis criados:", ""] + [f'- `/{p["slug"]}`: {p["listens"]} escutas e {p["tracks_heard"]} faixas ouvidas.' for p in profile_results]
    lines += ["", f"Backup automático: `{backups[0].name}` (versão antiga; permissão 0600).", "",
              "- Hashes de todas as tabelas legadas preservados, incluindo senhas, tokens e sessões.",
              "- Chaves antigas validadas pela API, sem redefinir senha ou gerar substitutas.",
              "- Histórico, estatísticas, rankings, atividade, resumo e playing now respondem com o perfil explícito.",
              "- Tokens de ingestão não autorizam gestão de chaves nem aparecem em metadados públicos.",
              "- Integridade SQLite: ok; sem violações de FKs ou escutas/chaves sem perfil.",
              "- Segundo início sem mudanças nos dados, sem perfis duplicados e sem novo backup.",
              "- Arquivo de origem e WAL com os mesmos hashes antes/depois.",
              "- API temporária encerrada. A instalação de uso não foi iniciada nem modificada pelo teste.",
              "", "A senha em texto puro não foi utilizada: sua preservação foi verificada comparando o hash armazenado.",
              "Não foram enviados novos scrobbles nem feitos login ou alterações administrativas nesta cópia.", ""]
    (work / "REPORT.md").write_text("\n".join(lines))
    print(f"PASS upgrade v{old_version} → v{new_version}: {before_counts['listens']} listens preserved; {len(public_profiles)} profile(s); {len(keys)} old token(s) valid")
    print("PASS automatic backup, password/session preservation, integrity, FKs and idempotent restart")
    print("PASS source database and WAL unchanged; validation API stopped")
    print(f"Report: {work / 'REPORT.md'}")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", type=Path, required=True)
    parser.add_argument("--source", type=Path, required=True)
    parser.add_argument("--work-dir", type=Path, required=True)
    args = parser.parse_args()
    binary, source, work = args.binary.resolve(), args.source.resolve(), args.work_dir.resolve()
    assert binary.is_file() and source.is_file(), "Binary and source database are required"
    work.mkdir(parents=True, exist_ok=False)
    os.chmod(work, 0o700)
    validate(binary, source, work)


if __name__ == "__main__":
    main()
